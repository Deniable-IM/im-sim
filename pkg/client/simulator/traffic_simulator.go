package Simulator

import (
	SimLogger "deniable-im/im-sim/pkg/client/simulator/sim_logger"
	Types "deniable-im/im-sim/pkg/client/types"
	SimulatedUser "deniable-im/im-sim/pkg/client/user"
	"time"
)

// Should probably return some kind of state, idk
// Maybe should generate the users itself, also idk
// Users should absolutely be initialised and given contacts beforehand
// TODO: Figure out how the random seed works in go
func SimulateTraffic(users []SimulatedUser.SimulatedUser, sim_time int64) {

	//Init
	var logger SimLogger.SimLogger
	logger.InitLogging()
	end_signal := make(chan int32)

	users_to_log := make([]Types.SimUser, len(users))
	for i, user := range users {
		users_to_log[i] = user.User
	}
	logger.LogSimUsers(users_to_log)
	defer logger.EndLogging()

	//All users start messaging
	for _, user := range users {
		go user.StartMessaging()
	}

	time.Sleep(time.Duration((sim_time * int64(time.Second))))

	end_signal <- 1

	time.Sleep(time.Duration(1 * time.Second))
}
