package Simulator

import (
	SimLogger "deniable-im/im-sim/pkg/simulation/simulator/sim_logger"
	SimulatedUser "deniable-im/im-sim/pkg/simulation/simulator/user"
	"fmt"
	"runtime"

	"deniable-im/im-sim/pkg/tshark"
	"sync"
	"time"
)

// Should probably return some kind of state, idk
func SimulateTraffic(users []*SimulatedUser.SimulatedUser, simTime int64, networkInterface string) {
	var wg sync.WaitGroup
	poolSize := 50

	startChan := make(chan struct{})
	stopChan := make(chan bool)

	threads := runtime.NumCPU()
	sendSem := make(chan struct{}, threads)

	var logger SimLogger.SimLogger
	msgChan, err := logger.InitLogging(stopChan)
	if err != nil {
		return
	}

	users_to_log := make([]SimLogger.UserInfo, len(users))
	for i, user := range users {
		users_to_log[i].User = (*user.User)
		users_to_log[i].Behavior = user.Behavior
		users_to_log[i].ContainerName = user.Client.Name
		for _, ip := range (*user).Client.Options.Connections {
			users_to_log[i].UserIP = *ip.IPv4
			break
		}
	}
	logger.LogSimUsers(users_to_log)

	println("Initializing clients")
	time.Sleep(10 * time.Second)

	for i, user := range users {
		wg.Add(1)
		go func(user *SimulatedUser.SimulatedUser) {
			defer wg.Done()
			go user.StartMessaging(startChan, stopChan, sendSem, msgChan)
		}(user)
		if (i+1)%poolSize == 0 {
			wg.Wait()
			time.Sleep(10 * time.Second)
		}
	}
	wg.Wait()
	time.Sleep(10 * time.Second)

	fmt.Printf("Press enter to begin client messaging on %d threads\n", threads)
	fmt.Scanln()

	println("Starting Tshark")
	cmd, cerr := tshark.RunTshark(networkInterface, logger.Dir, simTime+3)
	if cerr != nil {
		return
	}
	defer cmd.Wait()
	time.Sleep(1 * time.Second)

	// Clients now start messaging
	close(startChan)

	// Duration of simulation
	time.Sleep(time.Duration((simTime * int64(time.Second))))

	// Stop all clients
	stopChan <- true

	time.Sleep(time.Duration(5 * time.Second))

	println("Simulation is done")
}
