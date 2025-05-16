package Behavior

import (
	Types "deniable-im/im-sim/pkg/simulation/types"
	"math/rand"
)

type Behavior interface {
	GetNextMessageTime() int
	GetBehaviorName() string
	GetRandomizer() *rand.Rand
	SendRegularMsg() bool
	SendDeniableMsg() bool
	WillRespond(Types.Msg) bool
	IncrementDeniableCount()
	GetResponseTime() int
	IsBursting() bool
	MakeMessages() []Types.Msg
	MakeReply(Types.Msg) Types.Msg
	ParseIncoming(string) (*Types.Msg, error)
}
