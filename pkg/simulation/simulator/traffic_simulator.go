package Simulator

import (
	SimLogger "deniable-im/im-sim/pkg/simulation/simulator/sim_logger"
	SimulatedUser "deniable-im/im-sim/pkg/simulation/simulator/user"
	"deniable-im/im-sim/pkg/tshark"
	"time"
)

// Should probably return some kind of state, idk
func SimulateTraffic(users []*SimulatedUser.SimulatedUser, simTime int64, networkInterface string) {
	end_signal := make(chan bool)
	var logger SimLogger.SimLogger
	msgChan, err := logger.InitLogging(end_signal)
	if err != nil {
		return
	}

	users_to_log := make([]SimLogger.UserInfo, len(users))
	for i, user := range users {
		users_to_log[i].User = (*user.User)
		users_to_log[i].Behavior = user.Behavior
		for _, ip := range (*user).Client.Options.Connections {
			users_to_log[i].UserIP = *ip.IPv4
			break
		}
	}
	logger.LogSimUsers(users_to_log)

	cmd, cerr := tshark.RunTshark(networkInterface, logger.Dir, simTime+3)
	if cerr != nil {
		return
	}
	defer cmd.Wait()
	time.Sleep(2 * time.Second) //Allows tshark to start up properly and start packet capture

	for _, user := range users {
		go user.StartMessaging(end_signal, msgChan)
	}
	time.Sleep(time.Duration((simTime * int64(time.Second))))

	//Kill goroutines
	end_signal <- true

	time.Sleep(time.Duration(5 * time.Second))

	println("Simulation is done")
}
