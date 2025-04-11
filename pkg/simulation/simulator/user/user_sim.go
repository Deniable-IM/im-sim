package User

import (
	Container "deniable-im/im-sim/pkg/container"
	Process "deniable-im/im-sim/pkg/process"
	Behavior "deniable-im/im-sim/pkg/simulation/behavior"
	Types "deniable-im/im-sim/pkg/simulation/types"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const readTimeout = 1

type SimulatedUser struct {
	Behavior   Behavior.Behavior
	Client     *Container.Container
	User       *Types.SimUser
	stopChan   chan bool
	logger     chan Types.MsgEvent
	Process    *Process.Process
	mu         sync.Mutex
	GlobalLock *sync.Mutex
}

func (su *SimulatedUser) StartMessaging(stop chan bool, logger chan Types.MsgEvent) {
	if su == nil {
		return
	}

	su.stopChan = stop
	su.logger = logger

	args := []string{"./client", su.User.Nickname, fmt.Sprintf("%v", su.User.OwnID), "false"}

	res, err := su.Client.Exec(args, true)
	if err != nil {
		return
	}
	su.Process = res
	defer su.Process.Close()

	time.Sleep(200 * time.Millisecond)

	go su.MessageListener()
	for {
		select {
		case <-su.stopChan:
			su.mu.Lock()
			su.Process.Cmd([]byte("read\n"))
			su.mu.Unlock()
			return
		default:
			time_to_next_message := su.Behavior.GetNextMessageTime()
			time.Sleep(time.Duration(time.Duration(time_to_next_message) * time.Second))
			msgs := su.MakeMessages()
			for _, msg := range msgs {
				go su.SendMessage(msg)
			}
		}

	}
}

func (su *SimulatedUser) makeRegularMessage(target string) Types.Msg {
	if su == nil {
		return Types.Msg{}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "send:%v:Hello %v this is %v sending you a regular message Fuck the alphabet boys reading this", target, target, su.User.Nickname)

	msg := Types.Msg{To: target, From: su.User.Nickname, MsgContent: b.String(), IsDeniable: false}

	return msg
}

func (su *SimulatedUser) makeDeniableMessage(target string) Types.Msg {
	if su == nil {
		return Types.Msg{}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "denim:%v:Hello %v this is %v sending you a deniable message Fuck the alphabet boys reading this", target, target, su.User.Nickname)

	msg := Types.Msg{To: target, From: su.User.Nickname, MsgContent: b.String(), IsDeniable: true}
	return msg
}

func (su *SimulatedUser) MakeMessages() []Types.Msg {
	var msgs []Types.Msg

	//Deniable messages are made first to allow them to piggyback on the regular messages
	if su.Behavior.SendDeniableMsg() {
		den_target := su.User.DeniableContactList[su.Behavior.GetRandomizer().Intn(len(su.User.DeniableContactList))]
		den_msg := su.makeDeniableMessage(den_target)
		msgs = append(msgs, den_msg)
	}

	//Mayhaps make more than one regular message per call? Idk anymore, all of this is horrible to simulate
	if su.Behavior.SendRegularMsg() {
		reg_target := su.User.RegularContactList[su.Behavior.GetRandomizer().Intn(len(su.User.RegularContactList))]
		reg_msg := su.makeRegularMessage(reg_target)
		msgs = append(msgs, reg_msg)
	}

	return msgs
}

func (su *SimulatedUser) SendMessage(msg Types.Msg) {
	if su == nil {
		return
	}

	su.logger <- Types.MsgEvent{Msg: msg, EventType: "Send"}

	// Ensures messages are sent to the docker container at a rate it can handle
	su.GlobalLock.Lock()
	su.mu.Lock()
	su.Process.Cmd([]byte(fmt.Sprintf("%v\n", msg.MsgContent)))
	su.mu.Unlock()
	time.Sleep(50 * time.Millisecond) //Important to avoid fucking processes, if I knew why I would fix the root problem
	su.GlobalLock.Unlock()
}

func (su *SimulatedUser) OnReceive(msg Types.Msg) {
	if su == nil {
		return
	}

	su.logger <- Types.MsgEvent{EventType: "Receive", Msg: msg}

	//Determine if Alice responds to the message
	if !su.Behavior.WillRespond() {
		return
	}

	//TODO: Determine whether the message is regular or the entirety of a deniable message has been received

	res := su.makeRegularMessage(msg.From)
	go su.SendMessage(res)
}

func (su *SimulatedUser) MessageListener() {
	for {
		select {
		case <-su.stopChan:
			break
		default:
			time.Sleep(time.Duration(readTimeout * time.Second))
			su.mu.Lock()
			su.Process.Cmd([]byte("read\n"))
			su.mu.Unlock()
			for su.Process.Buffer.Len() != 0 {
				line, err := su.Process.Buffer.ReadString(byte('\n'))
				if len(line) == 0 {
					fmt.Println("Encountered error: %v \n", err)
					break
				}

				if err != nil {
					break
				}

				if len(line) == 1 {
					continue
				}

				splits := strings.Split(line, ":")
				if splits[0] == "" || splits[0] == "\n" {
					continue
				}

				sender := splits[0]
				if len(sender) < 8 {
					continue
				}

				sender = sender[8:len(sender)]
				if sender == "\n" {
					continue
				}
				for _, name := range su.User.RegularContactList {
					if sender == name {
						msg := Types.Msg{To: su.User.Nickname, From: name, MsgContent: splits[1], IsDeniable: false}
						go su.OnReceive(msg)
						break
					}
				}

				for _, name := range su.User.DeniableContactList {
					if sender == name {
						msg := Types.Msg{To: su.User.Nickname, From: name, MsgContent: splits[1], IsDeniable: true}
						go su.OnReceive(msg)
						break
					}
				}
			}
		}
	}
}

func (su *SimulatedUser) SetDeniableContacts(contacts []string) {
	su.User.DeniableContactList = append(su.User.DeniableContactList, contacts...)
}

// Modifies all passed users' contact lists and ensures each user has a minimum number of contacts
// in the range of min_contact_count and max_contact_count
func CreateCommunicationNetwork(users *[]Types.SimUser, min_contact_count, max_contact_count int, r *rand.Rand) *[]Types.SimUser {
	for i := range *users {
		contact_count := r.Intn(max_contact_count-min_contact_count) + min_contact_count
		(*users)[i].RegularContactList = make([]string, contact_count)
	}

	max_index := len(*users)

	for index, alice := range *users {
		for i, v := range alice.RegularContactList {
			if v != "" {
				continue
			}
			var bob_index int
			//Find new contact not already in Alice's contact list
			for {
				bob_index = r.Intn(max_index)
				if index == bob_index {
					continue
				}

				retry := false

				for _, id := range alice.RegularContactList {
					if id == (*users)[bob_index].Nickname {
						retry = true
						break
					}
				}

				if !retry {
					break
				}
			}

			(*users)[index].RegularContactList[i] = (*users)[bob_index].Nickname

			//Add to Bob's contact list or append if existing contact list is full
			if (*users)[bob_index].RegularContactList[len((*users)[bob_index].RegularContactList)-1] == "" {
				for j, val := range (*users)[bob_index].RegularContactList {
					if val == "" {
						(*users)[bob_index].RegularContactList[j] = alice.Nickname
						break
					}
				}
			} else {
				(*users)[bob_index].RegularContactList = append((*users)[bob_index].RegularContactList, alice.Nickname)
			}
		}
	} //Apparantly the way to get a random number between min_count and max_count...

	return users
}
