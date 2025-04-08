package Types

import "time"

type Msg struct {
	To, From   string
	MsgContent string
	IsDeniable bool
}

type MsgEvent struct {
	EventType string
	Timestamp time.Time
	Msg       Msg
}

type SimUser struct {
	OwnID               int32
	Nickname            string
	RegularContactList  []string
	DeniableContactList []string
}
