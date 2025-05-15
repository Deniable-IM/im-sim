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
	ID                  int32
	Nickname            string
	RegularContactList  []string
	DeniableContactList []string
}

type BehaviorType int

const (
	SimpleHuman int = iota
	//Insert other behaviour types as they come
	Other
)

type FloatTuple struct {
	First  float64
	Second float64
}

type SimUserOptions struct {
	Behaviour                 BehaviorType
	MinMaxRegularProbabiity   *FloatTuple
	MinMaxDeniableProbability *FloatTuple
	MinMaxReplyProbability    *FloatTuple
	BurstModifier             *float64
	BurstSize                 *int
	Seed                      *int64
}

func (options *SimUserOptions) HasNil() bool {
	return (options.MinMaxRegularProbabiity != nil &&
		options.MinMaxDeniableProbability != nil &&
		options.MinMaxReplyProbability != nil &&
		options.BurstModifier != nil &&
		options.BurstSize != nil)
}
