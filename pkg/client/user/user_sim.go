package user

import (
	Client "deniable-im/im-sim/pkg/client"
	Behavior "deniable-im/im-sim/pkg/client/behavior"
	"fmt"
	"math/rand/v2"
	"strings"
)

type SimUser struct {
	Behavior    Behavior.Behavior
	OwnID       int32
	ContactList []int32
	Client      Client.Client
}

func (su *SimUser) StartMessaging() {
	if su == nil {
		return
	}

	target := su.ContactList[rand.IntN(len(su.ContactList))]
	msg := su.MakeRegularMessage(target)

	//TODO: Send the message to the client and log it
	su.SendMessage(msg)

}

func (su *SimUser) MakeRegularMessage(target int32) string {
	if su == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "send:%v: Hello %v, this is %v sending you a message. Fuck the alphabet boys reading this.", target, target, su.OwnID)

	return b.String()
}

func (su *SimUser) MakeDeniableMessage(target int32) string {
	if su == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "den:%v: Hello %v, this is %v sending you a deniable message. Fuck the alphabet boys reading this.", target, target, su.OwnID)
	return b.String()
}

func (su *SimUser) SendMessage(msg string) {

}

func (su *SimUser) OnReceive(sender SimUser, msg_type int) {
	if su == nil {
		return
	}

	//Determine if Alice responds to the message
	if !su.Behavior.WillRespond() {
		return
	}

	//TODO: Determine whether the message is regular or a complete deniable message has been received
	reg_msg := su.MakeRegularMessage(sender.OwnID)

	go su.SendMessage(reg_msg)

}

// Modifies all passed users' contact lists and ensures each user has a minimum number of contacts
// in the range of min_contact_count and max_contact_count
func CreateCommunicationNetwork(users *[]SimUser, min_contact_count, max_contact_count int) *[]SimUser {
	for i := range *users {
		//Apparantly the way to get a random number between min_count and max_count
		contact_count := rand.IntN(max_contact_count-min_contact_count) + min_contact_count
		(*users)[i].ContactList = make([]int32, contact_count)
	}

	max_index := len(*users)

	for index, alice := range *users {
		for i, v := range alice.ContactList {
			if v != 0 {
				continue
			}
			var bob_index int
			//Find new contact not already in Alice's contact list
			for {
				bob_index = rand.IntN(max_index)
				if index == bob_index {
					continue
				}

				retry := false

				for _, id := range alice.ContactList {
					if id == (*users)[bob_index].OwnID {
						retry = true
						break
					}
				}

				if !retry {
					break
				}

			}

			(*users)[index].ContactList[i] = (*users)[bob_index].OwnID

			//Add to Bob's contact list or append if existing contact list is full
			if (*users)[bob_index].ContactList[len((*users)[bob_index].ContactList)-1] == 0 {
				for j, val := range (*users)[bob_index].ContactList {
					if val == 0 {
						(*users)[bob_index].ContactList[j] = alice.OwnID
						break
					}
				}
			} else {
				(*users)[bob_index].ContactList = append((*users)[bob_index].ContactList, alice.OwnID)
			}

		}
	}
	return users
}
