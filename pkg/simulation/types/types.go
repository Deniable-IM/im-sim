package Types

type Msg struct {
	To, From   int32
	MsgContent string
	IsDeniable bool
}

type MsgEvent struct {
	EventType string
	Msg       Msg
}

type SimUser struct {
	OwnID               int32
	RegularContactList  []int32
	DeniableContactList []int32
}
