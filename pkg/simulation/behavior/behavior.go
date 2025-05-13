package Behavior

import "math/rand"

type Behavior interface {
	GetNextMessageTime() int
	GetBehaviorName() string
	GetRandomizer() *rand.Rand
	SendRegularMsg() bool
	SendDeniableMsg() bool
	WillRespond() bool
	IncrementDeniableCount()
	GetResponseTime(int64) int
	IsBursting() bool
}
