package Behavior

type Behavior interface {
	GetNextMessageTime(float64) float64
	GetBehaviorName() string
	SendDeniableMsg() bool
	WillRespond() bool
}
