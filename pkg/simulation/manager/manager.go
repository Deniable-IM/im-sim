package manager

import (
	"deniable-im/im-sim/pkg/container"
	Behavior "deniable-im/im-sim/pkg/simulation/behavior"
	User "deniable-im/im-sim/pkg/simulation/simulator/user"
	Types "deniable-im/im-sim/pkg/simulation/types"
	"fmt"
)

// Creates default user array of the specified size. Panics if there is not enough containers or the nextfunc is nil.
func MakeDefaultSimulation(count int, clientContainers []*container.Container, nextfunc func(*Behavior.SimpleHumanTraits) int) []*User.SimulatedUser {
	if len(clientContainers) < count {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", count, len(clientContainers)))
	}

	if nextfunc == nil {
		panic("No nextfunc was specified as argument")
	}

	users := make([]*User.SimulatedUser, count)
	traits := Behavior.GenerateRealisticSimpleHumanTraits(count, nil, nextfunc)
	for i := 0; i < count; i++ {
		traits[i].ResponseProb += 0.2
		users[i] = &User.SimulatedUser{Behavior: traits[i], User: &Types.SimUser{ID: int32(i), Nickname: fmt.Sprintf("%v", i)}, Client: clientContainers[i]}
	}

	return users
}

// Uses the supplied options struct to generate users. If any critical option is nil, the function will return default users. Panics if there is not enough containers or the nextfunc is nil.
func MakeSimUsersFromOptions(count int, clientContainers []*container.Container, nextfunc func(*Behavior.SimpleHumanTraits) int, options *Types.SimUserOptions) []*User.SimulatedUser {
	if len(clientContainers) < count {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", count, len(clientContainers)))
	}

	if nextfunc == nil {
		panic("No nextfunc was specified as argument")
	}

	if options == nil || options.HasNil() {
		return MakeDefaultSimulation(count, clientContainers, nextfunc)
	}

	var behaviour []Behavior.Behavior

	switch options.Behaviour {
	case Types.BehaviorType(Types.SimpleHuman):
		result := Behavior.GenerateSimpleHumanTraitsFromOptions(count, nextfunc, *options)
		for i := range result {
			behaviour[i] = result[i]
		}

	default:
		panic("Option not set")
	}

	users := make([]*User.SimulatedUser, count)

	for i := 0; i < count; i++ {
		users[i] = &User.SimulatedUser{Behavior: behaviour[i], User: &Types.SimUser{ID: int32(i), Nickname: fmt.Sprintf("%v", i)}, Client: clientContainers[i]}
	}

	return users
}

// Alice, Bob, Charlie and Dorothy example only sending regular messages. Panics if there is not enough containers or the nextfunc is nil.
func MakeAliceBobRegularExampleSimulation(clientContainers []*container.Container, nextfunc func(*Behavior.SimpleHumanTraits) int) []*User.SimulatedUser {
	if len(clientContainers) < 4 {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", 4, len(clientContainers)))
	}

	traits := Behavior.GenerateRealisticSimpleHumanTraits(4, nil, nextfunc)
	for i := range traits {
		traits[i].DeniableBurstSize = 0
	}

	simulatedAlice := User.SimulatedUser{Behavior: traits[0], User: &Types.SimUser{ID: 0, Nickname: "0", RegularContactList: []string{"0"}}, Client: clientContainers[0]}
	simulatedBob := User.SimulatedUser{Behavior: traits[1], User: &Types.SimUser{ID: 2, Nickname: "bob", RegularContactList: []string{"1"}}, Client: clientContainers[1]}
	simulatedCharlie := User.SimulatedUser{Behavior: traits[2], User: &Types.SimUser{ID: 3, Nickname: "charlie", RegularContactList: []string{"4"}}, Client: clientContainers[2]}
	simulatedDorothy := User.SimulatedUser{Behavior: traits[3], User: &Types.SimUser{ID: 4, Nickname: "dorothy", RegularContactList: []string{"3"}}, Client: clientContainers[3]}

	users := []*User.SimulatedUser{&simulatedAlice, &simulatedBob, &simulatedCharlie, &simulatedDorothy}

	return users
}

// Alice, Bob, Charlie and Dorothy example sending both regular messages and deniable messages with bursting. Panics if there is not enough containers or the nextfunc is nil.
func MakeAliceBobDeniableBurstExampleSimulation(clientContainers []*container.Container, nextfunc func(*Behavior.SimpleHumanTraits) int) []*User.SimulatedUser {
	if len(clientContainers) < 4 {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", 4, len(clientContainers)))
	}

	traits := Behavior.GenerateRealisticSimpleHumanTraits(4, nil, nextfunc)

	simulatedAlice := User.SimulatedUser{Behavior: traits[0], User: &Types.SimUser{ID: 1, Nickname: "alice", RegularContactList: []string{"2"}, DeniableContactList: []string{"4"}}, Client: clientContainers[0]}
	simulatedBob := User.SimulatedUser{Behavior: traits[1], User: &Types.SimUser{ID: 2, Nickname: "bob", RegularContactList: []string{"1"}, DeniableContactList: []string{"3"}}, Client: clientContainers[1]}
	simulatedCharlie := User.SimulatedUser{Behavior: traits[2], User: &Types.SimUser{ID: 3, Nickname: "charlie", RegularContactList: []string{"4"}, DeniableContactList: []string{"2"}}, Client: clientContainers[2]}
	simulatedDorothy := User.SimulatedUser{Behavior: traits[3], User: &Types.SimUser{ID: 4, Nickname: "dorothy", RegularContactList: []string{"3"}, DeniableContactList: []string{"1"}}, Client: clientContainers[3]}

	users := []*User.SimulatedUser{&simulatedAlice, &simulatedBob, &simulatedCharlie, &simulatedDorothy}

	return users
}

// Alice, Bob, Charlie and Dorothy example sending both regular messages and deniable messages, but without bursting. Panics if there is not enough containers or the nextfunc is nil.
func MakeAliceBobDeniableNoBurstExampleSimulation(clientContainers []*container.Container, nextfunc func(*Behavior.SimpleHumanTraits) int) []*User.SimulatedUser {
	if len(clientContainers) < 4 {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", 4, len(clientContainers)))
	}

	traits := Behavior.GenerateRealisticSimpleHumanTraits(4, nil, nextfunc)

	for i := range traits {
		traits[i].DeniableBurstSize = 0
	}

	simulatedAlice := User.SimulatedUser{Behavior: traits[0], User: &Types.SimUser{ID: 1, Nickname: "alice", RegularContactList: []string{"2"}, DeniableContactList: []string{"4"}}, Client: clientContainers[0]}
	simulatedBob := User.SimulatedUser{Behavior: traits[1], User: &Types.SimUser{ID: 2, Nickname: "bob", RegularContactList: []string{"1"}, DeniableContactList: []string{"3"}}, Client: clientContainers[1]}
	simulatedCharlie := User.SimulatedUser{Behavior: traits[2], User: &Types.SimUser{ID: 3, Nickname: "charlie", RegularContactList: []string{"4"}, DeniableContactList: []string{"2"}}, Client: clientContainers[2]}
	simulatedDorothy := User.SimulatedUser{Behavior: traits[3], User: &Types.SimUser{ID: 4, Nickname: "dorothy", RegularContactList: []string{"3"}, DeniableContactList: []string{"1"}}, Client: clientContainers[3]}

	users := []*User.SimulatedUser{&simulatedAlice, &simulatedBob, &simulatedCharlie, &simulatedDorothy}

	return users
}
