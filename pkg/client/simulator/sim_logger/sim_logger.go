package simlogger

import (
	Types "deniable-im/im-sim/pkg/client/types"
	"encoding/json"
	"time"

	"fmt"
	"os"
)

type SimLogger struct {
	dir string
	fp  *os.File
}

func (sl *SimLogger) InitLogging() {
	ts := time.Now().GoString()
	dirname := fmt.Sprintf("logs/%v", ts)
	err := os.MkdirAll(dirname, 0750)
	if err != nil {
		return
	}

	sl.dir = dirname
	filename := fmt.Sprintf("%v/messages.json", dirname)
	file, ferr := os.Create(filename)
	if ferr != nil {
		fmt.Println("Error creating file:", ferr)
		return
	}

	sl.fp = file
}

func (sl *SimLogger) LogMsgEvent(logEvent Types.MsgEvent) {

	jsonData, err := json.MarshalIndent(logEvent, "", " ")
	if err != nil {
		fmt.Println("Error marshalling JSON", err)
		return
	}

	_, werr := sl.fp.Write(jsonData)
	if werr != nil {
		fmt.Println("Error writing msg event to file", werr)
		return
	}
}

func (sl *SimLogger) LogSimUsers(users []Types.SimUser) {

	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	filename := fmt.Sprintf("%v/users.json", sl.dir)

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

	fmt.Println("Successfully logged users")
}

func (sl *SimLogger) EndLogging() {
	sl.fp.Close()
}
