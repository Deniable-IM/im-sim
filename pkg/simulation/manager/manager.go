package manager

import (
	"deniable-im/im-sim/pkg/container"
	Behavior "deniable-im/im-sim/pkg/simulation/behavior"
	User "deniable-im/im-sim/pkg/simulation/simulator/user"
	Types "deniable-im/im-sim/pkg/simulation/types"
	"fmt"
)

// Creates default user array of the specified size. Panics if there is not enough containers or the nextfunc is nil.
func MakeDefaultSimulation(
	count int, clientContainers []*container.Container,
	nextfunc func(*Behavior.SimpleHumanTraits) int) []*User.SimulatedUser {
	if len(clientContainers) < count {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", count, len(clientContainers)))
	}

	if nextfunc == nil {
		panic("No nextfunc was specified as argument")
	}

	users := make([]*User.SimulatedUser, count)
	traits := Behavior.GenerateRealisticSimpleHumanTraits(count, nil, nextfunc)
	for i := 0; i < count; i++ {
		user := &Types.SimUser{
			ID:       int32(i),
			Nickname: fmt.Sprintf("%v", i),
		}
		traits[i].ResponseProb += 0.2
		traits[i].User = user
		users[i] = &User.SimulatedUser{
			Behavior: traits[i],
			User:     user,
			Client:   clientContainers[i],
		}
	}

	return users
}

// Uses the supplied options struct to generate users. If any critical option is nil, the function will return default users. Panics if there is not enough containers or the nextfunc is nil.
func MakeSimUsersFromOptions(
	count int,
	clientContainers []*container.Container,
	nextfunc func(*Behavior.SimpleHumanTraits) int,
	options *Types.SimUserOptions) []*User.SimulatedUser {
	if len(clientContainers) < count {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", count, len(clientContainers)))
	}

	if nextfunc == nil {
		panic("No nextfunc was specified as argument")
	}

	if options == nil || options.HasNil() {
		return MakeDefaultSimulation(count, clientContainers, nextfunc)
	}

	behaviour := make([]Behavior.Behavior, count)
	var users []*Types.SimUser

	switch options.Behaviour {
	case Types.BehaviorType(Types.SimpleHuman):
		result := Behavior.GenerateSimpleHumanTraitsFromOptions(count, nextfunc, *options)
		for i := range result {
			user := &Types.SimUser{
				ID:       int32(i),
				Nickname: fmt.Sprintf("%v", i),
			}
			users = append(users, user)
			result[i].User = user
			behaviour[i] = result[i]
		}

	default:
		panic("Option not set")
	}

	sim_users := make([]*User.SimulatedUser, count)

	for i := 0; i < count; i++ {
		sim_users[i] = &User.SimulatedUser{
			Behavior: behaviour[i],
			User:     users[i],
			Client:   clientContainers[i],
		}
	}

	return sim_users
}

// Alice, Bob, Charlie and Dorothy example only sending regular messages. Panics if there is not enough containers or the nextfunc is nil.
func MakeAliceBobRegularExampleSimulation(
	clientContainers []*container.Container,
	nextfunc func(*Behavior.SimpleHumanTraits) int) []*User.SimulatedUser {
	if len(clientContainers) < 4 {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", 4, len(clientContainers)))
	}

	traits := Behavior.GenerateRealisticSimpleHumanTraits(4, nil, nextfunc)
	for i := range traits {
		traits[i].DeniableBurstSize = 0
		traits[i].User = &Types.SimUser{
			ID:       int32(i),
			Nickname: fmt.Sprintf("%v", i),
		}
	}

	simulatedAlice := User.SimulatedUser{
		Behavior: traits[0],
		User:     traits[0].User,
		Client:   clientContainers[0],
	}
	simulatedAlice.User.RegularContactList = []string{"1"}

	simulatedBob := User.SimulatedUser{
		Behavior: traits[1],
		User:     traits[1].User,
		Client:   clientContainers[1],
	}
	simulatedBob.User.RegularContactList = []string{"0"}

	simulatedCharlie := User.SimulatedUser{
		Behavior: traits[2],
		User:     traits[2].User,
		Client:   clientContainers[2],
	}
	simulatedCharlie.User.RegularContactList = []string{"3"}

	simulatedDorothy := User.SimulatedUser{
		Behavior: traits[3],
		User:     traits[3].User,
		Client:   clientContainers[3],
	}
	simulatedDorothy.User.RegularContactList = []string{"2"}

	users := []*User.SimulatedUser{
		&simulatedAlice,
		&simulatedBob,
		&simulatedCharlie,
		&simulatedDorothy,
	}

	return users
}

// Alice, Bob, Charlie and Dorothy example sending both regular messages and deniable messages with bursting. Panics if there is not enough containers or the nextfunc is nil.
func MakeAliceBobDeniableBurstExampleSimulation(
	clientContainers []*container.Container,
	nextfunc func(*Behavior.SimpleHumanTraits) int) []*User.SimulatedUser {
	if len(clientContainers) < 4 {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", 4, len(clientContainers)))
	}

	traits := Behavior.GenerateRealisticSimpleHumanTraits(4, nil, nextfunc)
	for i := range traits {
		traits[i].User = &Types.SimUser{
			ID:       int32(i),
			Nickname: fmt.Sprintf("%v", i),
		}
	}

	simulatedAlice := User.SimulatedUser{
		Behavior: traits[0],
		User:     traits[0].User,
		Client:   clientContainers[0],
	}
	simulatedAlice.User.RegularContactList = []string{"1"}
	simulatedAlice.User.DeniableContactList = []string{"3"}

	simulatedBob := User.SimulatedUser{
		Behavior: traits[1],
		User:     traits[1].User,
		Client:   clientContainers[1],
	}
	simulatedBob.User.RegularContactList = []string{"0"}

	simulatedCharlie := User.SimulatedUser{
		Behavior: traits[2],
		User:     traits[2].User,
		Client:   clientContainers[2],
	}
	simulatedCharlie.User.RegularContactList = []string{"3"}

	simulatedDorothy := User.SimulatedUser{
		Behavior: traits[3],
		User:     traits[3].User,
		Client:   clientContainers[3],
	}
	simulatedDorothy.User.RegularContactList = []string{"2"}
	simulatedDorothy.User.DeniableContactList = []string{"0"}

	users := []*User.SimulatedUser{
		&simulatedAlice,
		&simulatedBob,
		&simulatedCharlie,
		&simulatedDorothy,
	}

	return users
}

// Alice, Bob, Charlie and Dorothy example sending both regular messages and deniable messages, but without bursting. Panics if there is not enough containers or the nextfunc is nil.
func MakeAliceBobDeniableNoBurstExampleSimulation(
	clientContainers []*container.Container,
	nextfunc func(*Behavior.SimpleHumanTraits) int) []*User.SimulatedUser {
	if len(clientContainers) < 4 {
		panic(fmt.Sprintf("Insufficient number of clientContainers provided as argument. Expected %v, got %v", 4, len(clientContainers)))
	}

	traits := Behavior.GenerateRealisticSimpleHumanTraits(4, nil, nextfunc)

	for i := range traits {
		traits[i].User = &Types.SimUser{
			ID:       int32(i),
			Nickname: fmt.Sprintf("%v", i),
		}
	}

	simulatedAlice := User.SimulatedUser{
		Behavior: traits[0],
		User:     traits[0].User,
		Client:   clientContainers[0],
	}
	simulatedAlice.User.RegularContactList = []string{"1"}
	simulatedAlice.User.DeniableContactList = []string{"3"}

	simulatedBob := User.SimulatedUser{
		Behavior: traits[1],
		User:     traits[1].User,
		Client:   clientContainers[1],
	}
	simulatedBob.User.RegularContactList = []string{"0"}

	simulatedCharlie := User.SimulatedUser{
		Behavior: traits[2],
		User:     traits[2].User,
		Client:   clientContainers[2],
	}
	simulatedCharlie.User.RegularContactList = []string{"3"}

	simulatedDorothy := User.SimulatedUser{
		Behavior: traits[3],
		User:     traits[3].User,
		Client:   clientContainers[3],
	}
	simulatedDorothy.User.RegularContactList = []string{"2"}
	simulatedDorothy.User.DeniableContactList = []string{"0"}

	users := []*User.SimulatedUser{
		&simulatedAlice,
		&simulatedBob,
		&simulatedCharlie,
		&simulatedDorothy,
	}

	return users
}
