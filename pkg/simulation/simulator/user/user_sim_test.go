package User

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestCreateCommunicationNetwork(t *testing.T) {
	//1000 SimUser structs are used as it seems like the maximum we can simulate without problems
	sim_users := make([]*SimulatedUser, 1000)
	for i := range sim_users {
		sim_users[i].User.ID = int32(i + 1)
		sim_users[i].User.Nickname = fmt.Sprintf("%v", i+1)
	}

	var seed int64 = 696969420
	src := rand.NewSource(seed)
	rng := rand.New(src)

	users := CreateCommunicationNetwork(sim_users, 10, 20, rng)

	for _, u := range users {
		values := make(map[string]int32)
		for _, v := range u.User.RegularContactList {
			values[v] += 1

			if values[v] >= 2 {
				var b strings.Builder
				fmt.Fprintf(&b, "Value %v appears multiple times in User %v's contact list", v, u.User.ID)
				t.Error(b.String())
			}
		}
		if values[u.User.Nickname] != 0 {
			t.Error("User has itself as contact")
		}

		if len(u.User.RegularContactList) == 0 {
			t.Errorf("User %v has no contacts", u.User.ID)
		}
	}
}
