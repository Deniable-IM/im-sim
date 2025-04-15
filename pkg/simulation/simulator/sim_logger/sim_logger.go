package simlogger

import (
	Types "deniable-im/im-sim/pkg/simulation/types"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type SimLogger struct {
	Dir      string
	killChan chan bool
}

func (sl *SimLogger) InitLogging(kill chan bool) (chan Types.MsgEvent, error) {
	sl.killChan = kill

	ts := time.Now().String()
	ts = strings.ReplaceAll(ts, " ", "")
	ts = strings.ReplaceAll(ts, ":", "")
	ts = ts[:16]
	dirname := fmt.Sprintf("logs/%v", ts)
	err := os.MkdirAll(dirname, 0750)
	if err != nil {
		return nil, fmt.Errorf("Failed to create logs/%v", ts)
	}

	sl.Dir = dirname

	msgLogChan := make(chan Types.MsgEvent)
	go sl.LogMsgEvent(msgLogChan)
	return msgLogChan, nil
}

func (sl *SimLogger) LogMsgEvent(eventChan chan Types.MsgEvent) {
	path := fmt.Sprintf("%v/messages.json", sl.Dir)
	f, ferr := os.Create(path)
	if ferr != nil {
		fmt.Println("Error creating file:", ferr)
		return
	}
	defer f.Close()

	for {
		select {
		case <-sl.killChan:
			return
		default:
			logEvent := <-eventChan
			logEvent.Timestamp = time.Now()
			jsonData, err := json.MarshalIndent(logEvent, "", " ")
			if err != nil {
				fmt.Println("Error marshalling JSON", err)
				return
			}

			_, werr := f.Write(jsonData)
			if werr != nil {
				fmt.Println("Error writing msg event to file", werr)
				return
			}
		}
	}

}

func (sl *SimLogger) LogSimUsers(users []Types.SimUser) {

	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	filename := fmt.Sprintf("%v/users.json", sl.Dir)

	file, ferr := os.Create(filename)
	if ferr != nil {
		fmt.Println("Error creating file:", ferr)
		return
	}
	defer file.Close()

	_, werr := file.Write(jsonData)
	if werr != nil {
		fmt.Println("Error writing to file", werr)
		return
	}
}
