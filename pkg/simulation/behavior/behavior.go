package Behavior

import "math/rand"

type Behavior interface {
	GetNextMessageTime() float64
	GetBehaviorName() string
	GetRandomizer() *rand.Rand
	SendRegularMsg() bool
	SendDeniableMsg() bool
	WillRespond() bool
}
