package Simulator

import (
	SimLogger "deniable-im/im-sim/pkg/simulation/simulator/sim_logger"
	SimulatedUser "deniable-im/im-sim/pkg/simulation/simulator/user"
	Types "deniable-im/im-sim/pkg/simulation/types"
	"time"
)

// Should probably return some kind of state, idk
func SimulateTraffic(users []SimulatedUser.SimulatedUser, sim_time int64) {
	var logger SimLogger.SimLogger
	logger.InitLogging()
	end_signal := make(chan int)

	users_to_log := make([]Types.SimUser, len(users))
	for i, user := range users {
		users_to_log[i] = user.User
	}
	logger.LogSimUsers(users_to_log)
	defer logger.EndLogging()

	for _, user := range users {
		go user.StartMessaging(end_signal)
	}

	time.Sleep(time.Duration((sim_time * int64(time.Second))))

	//Kill goroutines
	end_signal <- 1

	time.Sleep(time.Duration(1 * time.Second))
}
