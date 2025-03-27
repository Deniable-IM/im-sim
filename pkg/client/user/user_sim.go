package user

import (
	Client "deniable-im/im-sim/pkg/client"
	Behavior "deniable-im/im-sim/pkg/client/behavior"
	Logger "deniable-im/im-sim/pkg/client/simulator/sim_logger"
	Types "deniable-im/im-sim/pkg/client/types"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type SimulatedUser struct {
	Behavior Behavior.Behavior
	Client   Client.Client
	User     Types.SimUser
	stopChan chan int32
	logger   Logger.SimLogger
}

func (su *SimulatedUser) StartMessaging() {
	if su == nil {
		return
	}

	for len(su.stopChan) == 0 {
		time_to_next_message := su.Behavior.GetNextMessageTime()
		time.Sleep(time.Duration(time_to_next_message * float64(time.Second)))
		msgs := su.MakeMessages()
		for _, msg := range msgs {
			go su.SendMessage(msg)
			go su.logger.LogMsgEvent(Types.MsgEvent{Msg: msg, EventType: "send"})
		}

	}

}

func (su *SimulatedUser) makeRegularMessage(target int32) Types.Msg {
	if su == nil {
		return Types.Msg{}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "send:%v: Hello %v, this is %v sending you regular a message. Fuck the alphabet boys reading this.", target, target, su.User.OwnID)

	msg := Types.Msg{To: target, From: su.User.OwnID, MsgContent: b.String(), IsDeniable: false}

	return msg
}

func (su *SimulatedUser) makeDeniableMessage(target int32) Types.Msg {
	if su == nil {
		return Types.Msg{}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "den:%v: Hello %v, this is %v sending you a deniable message. Fuck the alphabet boys reading this.", target, target, su.User.OwnID)

	msg := Types.Msg{To: target, From: su.User.OwnID, MsgContent: b.String(), IsDeniable: true}
	return msg
}

func (su *SimulatedUser) MakeMessages() []Types.Msg {
	var msgs []Types.Msg

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

	//IDK how to actually send stuff

}

func (su *SimulatedUser) OnReceive(msg Types.Msg) {
	if su == nil {
		return
	}

	//Determine if Alice responds to the message
	if !su.Behavior.WillRespond() {
		return
	}

	//TODO: Determine whether the message is regular or the entirety of a deniable message has been received
	// reg_msg := su.makeRegularMessage(sender.OwnID)

	// su.SendMessage(reg_msg)

}

func (su *SimulatedUser) SetDeniableContacts(contacts []int32) {
	su.User.DeniableContactList = append(su.User.DeniableContactList, contacts...)
}

// Modifies all passed users' contact lists and ensures each user has a minimum number of contacts
// in the range of min_contact_count and max_contact_count
func CreateCommunicationNetwork(users *[]Types.SimUser, min_contact_count, max_contact_count int, r *rand.Rand) *[]Types.SimUser {
	for i := range *users {
		//Apparantly the way to get a random number between min_count and max_count...
		contact_count := r.Intn(max_contact_count-min_contact_count) + min_contact_count
		(*users)[i].RegularContactList = make([]int32, contact_count)
	}

	max_index := len(*users)

	for index, alice := range *users {
		for i, v := range alice.RegularContactList {
			if v != 0 {
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
					if id == (*users)[bob_index].OwnID {
						retry = true
						break
					}
				}

				if !retry {
					break
				}
			}

			(*users)[index].RegularContactList[i] = (*users)[bob_index].OwnID

			//Add to Bob's contact list or append if existing contact list is full
			if (*users)[bob_index].RegularContactList[len((*users)[bob_index].RegularContactList)-1] == 0 {
				for j, val := range (*users)[bob_index].RegularContactList {
					if val == 0 {
						(*users)[bob_index].RegularContactList[j] = alice.OwnID
						break
					}
				}
			} else {
				(*users)[bob_index].RegularContactList = append((*users)[bob_index].RegularContactList, alice.OwnID)
			}
		}
	}
	return users
}
