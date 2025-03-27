package user

import (
	Types "deniable-im/im-sim/pkg/client/types"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestCreateCommunicationNetwork(t *testing.T) {
	//1000 SimUser structs are used as it seems like the maximum we can simulate without problems
	sim_users := make([]SimulatedUser, 1000)
	for i := range sim_users {
		sim_users[i].User.OwnID = int32(i + 1)
	}

	users := make([]Types.SimUser, len(sim_users))
	for i, user := range sim_users {
		users[i] = user.User
	}

	var seed int64 = 696969420
	src := rand.NewSource(seed)
	rng := rand.New(src)

	users = *CreateCommunicationNetwork(&users, 10, 20, rng)

	for _, u := range users {
		values := make(map[int32]int32)
		for _, v := range u.RegularContactList {
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

		if len(u.RegularContactList) == 0 {
			t.Errorf("User %v has no contacts", u.OwnID)
		}
	}
}
