package user

import (
	"fmt"
	"strings"
	"testing"
)

func TestCreateCommunicationNetwork(t *testing.T) {
	//1000 SimUser structs are used as it seems like the maximum we can simulate without problems
	users := make([]SimUser, 1000)
	for i := range users {
		users[i].OwnID = int32(i + 1)
	}

	users = *CreateCommunicationNetwork(&users, 5, 15)

	for _, u := range users {
		values := make(map[int32]int32)
		for _, v := range u.ContactList {
			values[v] += 1

			if values[v] >= 2 {
				var b strings.Builder
				fmt.Fprintf(&b, "Value %v appears multiple times in User %v's contact list", v, u.OwnID)
				t.Error(b.String())
			}
		}
		if values[u.OwnID] != 0 {
			t.Error("User has itself as contact")
		}

		if len(u.ContactList) == 0 {
			t.Errorf("User %v has no contacts", u.OwnID)
		}
	}
}
