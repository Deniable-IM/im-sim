package Simulator

import (
	SimLogger "deniable-im/im-sim/pkg/simulation/simulator/sim_logger"
	SimulatedUser "deniable-im/im-sim/pkg/simulation/simulator/user"
	Types "deniable-im/im-sim/pkg/simulation/types"
	"time"
)

// Should probably return some kind of state, idk
func SimulateTraffic(users *[]SimulatedUser.SimulatedUser, sim_time int64) {
	end_signal := make(chan bool)
	var logger SimLogger.SimLogger
	msgChan, err := logger.InitLogging(end_signal)
	if err != nil {
		return
	}

	users_to_log := make([]Types.SimUser, len(*users))
	for i, user := range *users {
		users_to_log[i] = user.User

	}
	logger.LogSimUsers(users_to_log)
	for _, user := range *users {
		go user.StartMessaging(end_signal, msgChan)
	}
	time.Sleep(time.Duration((sim_time * int64(time.Second))))

	//Kill goroutines
	end_signal <- true

	time.Sleep(time.Duration(5 * time.Second))

	println("Simulation is done")
}
