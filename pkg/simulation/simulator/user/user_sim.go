package User

import (
	Container "deniable-im/im-sim/pkg/container"
	Process "deniable-im/im-sim/pkg/process"
	Behavior "deniable-im/im-sim/pkg/simulation/behavior"
	Types "deniable-im/im-sim/pkg/simulation/types"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const readTimeout = 0.2

type SimulatedUser struct {
	Behavior Behavior.Behavior
	Client   *Container.Container
	User     *Types.SimUser
	stopChan chan bool
	logger   chan Types.MsgEvent
	Process  *Process.Process
}

func (su *SimulatedUser) StartMessaging(stop chan bool, logger chan Types.MsgEvent) {
	var wg sync.WaitGroup

	if su == nil {
		return
	}

	su.stopChan = stop
	su.logger = logger

	args := []string{"./client", su.User.Nickname, fmt.Sprintf("%v", su.User.ID), "false"}

	res, err := su.Client.Exec(args, true)
	if err != nil {
		panic(err)
	}

	su.Process = res
	defer su.Process.Close()

	time.Sleep(2000 * time.Millisecond)

	wg.Add(1)
	go func() {
		defer wg.Done()
		su.MessageListener()
	}()

	for {
		select {
		case <-su.stopChan:
			wg.Wait()
			return
		default:
			time_to_next_message := su.Behavior.GetNextMessageTime()
			dur := time.Duration(time_to_next_message * int(time.Millisecond))
			time.Sleep(dur)
			msgs := su.Behavior.MakeMessages()
			for _, msg := range msgs {
				su.SendMessage(msg)
			}
		}
	}
}

func (su *SimulatedUser) SendMessage(msg Types.Msg) {
	if su == nil {
		return
	}

	err := su.Process.Cmd([]byte(fmt.Sprintf("%v\n", msg.MsgContent)))
	if err != nil {
		panic(err)
	}

	su.logger <- Types.MsgEvent{Msg: msg, EventType: "Send"}
}

func (su *SimulatedUser) OnReceive(msg Types.Msg) {
	if su == nil {
		return
	}

	//Determine if Alice responds to the message
	if !su.Behavior.WillRespond(msg) {
		return
	}

	res := su.Behavior.MakeReply(msg)
	sleep_time := su.Behavior.GetResponseTime()
	time.Sleep(time.Duration(sleep_time * int(time.Millisecond)))

	su.SendMessage(res)
}

func (su *SimulatedUser) MessageListener() {
	for {
		select {
		case <-su.stopChan:
			return
		default:
			time.Sleep(time.Duration(readTimeout * float64(time.Second)))

			err := su.Process.Cmd([]byte("read\n"))
			if err != nil {
				panic(err)
			}

			lines := su.Process.Read(byte('\n'))
			for _, line := range lines {
				msg, err := su.Behavior.ParseIncoming(line)
				if err != nil {
					continue
				}

				msg.To = fmt.Sprintf("%v", su.User.ID)

				su.logger <- Types.MsgEvent{
					Msg:       *msg,
					EventType: "Receive",
				}

				su.OnReceive(*msg)
			}
		}
	}
}

func (su *SimulatedUser) SetDeniableContacts(contacts []string) {
	su.User.DeniableContactList = append(su.User.DeniableContactList, contacts...)
}

// Modifies all passed users' contact lists and ensures each user has a minimum number of contacts
// in the range of min_contact_count and max_contact_count
func CreateCommunicationNetwork(users []*SimulatedUser, min_contact_count, max_contact_count int, r *rand.Rand) []*SimulatedUser {
	max_index := len(users)
	if max_contact_count < min_contact_count {
		panic("Max contact count cannot be less than min contact count")
	}

	if max_index < max_contact_count {
		panic("Max contact count can not be larger than the number of users being assigned contacts")
	}

	for i := range users {
		contact_count := r.Intn(max_contact_count-min_contact_count) + min_contact_count
		(users)[i].User.RegularContactList = make([]string, contact_count)
	}

	for index, alice := range users {
		for i, v := range alice.User.RegularContactList {
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

				for _, id := range alice.User.RegularContactList {
					if id == users[bob_index].User.Nickname {
						retry = true
						break
					}
				}

				if !retry {
					break
				}
			}

			users[index].User.RegularContactList[i] = users[bob_index].User.Nickname

			//Add to Bob's contact list or append if existing contact list is full
			if users[bob_index].User.RegularContactList[len(users[bob_index].User.RegularContactList)-1] == "" {
				for j, val := range users[bob_index].User.RegularContactList {
					if val == "" {
						users[bob_index].User.RegularContactList[j] = alice.User.Nickname
						break
					}
				}
			} else {
				users[bob_index].User.RegularContactList = append(users[bob_index].User.RegularContactList, alice.User.Nickname)
			}
		}
	}

	return users
}

func CreateDeniableNetwork(users []*SimulatedUser, min_contact_count, max_contact_count int, r *rand.Rand) []*SimulatedUser {
	max_index := len(users)
	if max_contact_count < min_contact_count {
		panic("Max contact count cannot be less than min contact count")
	}

	if max_index < max_contact_count {
		panic("Max contact count can not be larger than the number of users being assigned contacts")
	}

	for i := range users {
		contact_count := r.Intn(max_contact_count-min_contact_count) + min_contact_count
		(users)[i].User.DeniableContactList = make([]string, contact_count)
	}

	for index, alice := range users {
		for i, v := range alice.User.DeniableContactList {
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

				for _, id := range alice.User.RegularContactList {
					if id == users[bob_index].User.Nickname {
						retry = true
						break
					}
				}

				for _, id := range alice.User.DeniableContactList {
					if id == users[bob_index].User.Nickname {
						retry = true
						break
					}
				}

				if !retry {
					break
				}
			}

			users[index].User.DeniableContactList[i] = users[bob_index].User.Nickname

			//Add to Bob's contact list or append if existing contact list is full
			if users[bob_index].User.DeniableContactList[len(users[bob_index].User.DeniableContactList)-1] == "" {
				for j, val := range users[bob_index].User.DeniableContactList {
					if val == "" {
						users[bob_index].User.DeniableContactList[j] = alice.User.Nickname
						break
					}
				}
			} else {
				users[bob_index].User.DeniableContactList = append(users[bob_index].User.DeniableContactList, alice.User.Nickname)
			}
		}
	}

	return users
}
