package Behavior

import (
	"fmt"
	"math/rand/v2"
	"strings"

	fuzz "github.com/google/gofuzz"
)

type SimpleHumanTraits struct {
	Name              string
	SendRate          float64
	ForgetRate        float64
	Rhythm            float64
	LastSentMessageAt float64
	ResponseProb      float64
	DeniableProb      float64
	NextMsgFunc       func(float64, float64, float64, float64) float64
}

func (sh *SimpleHumanTraits) GetBehaviorName() string {
	if sh == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Simple Human Traits Behavior with name %v, Forget_rate = %v, Rhythm = %v", sh.Name, sh.ForgetRate, sh.Rhythm)
	return b.String()
}

func (sh *SimpleHumanTraits) GetNextMessageTime(current_time float64) float64 {
	if sh == nil {
		return 0
	}

	next := sh.NextMsgFunc(sh.SendRate, sh.ForgetRate, sh.Rhythm, current_time)
	sh.LastSentMessageAt = current_time + next

	return next
}

func (sh *SimpleHumanTraits) WillRespond() bool {
	if sh == nil {
		return false
	}

	return rand.Float64() > (1.0 - sh.ResponseProb)
}

func (sh *SimpleHumanTraits) SendDeniableMsg() bool {
	if sh == nil {
		return false
	}

	return rand.Float64() > (1.0 - sh.DeniableProb)
}

func NewSimpleHumanTraits(name string, send_rate, forget_rate, rhythm float64, next_func func(float64, float64, float64, float64) float64) *SimpleHumanTraits {
	return &SimpleHumanTraits{Name: name, SendRate: send_rate, ForgetRate: forget_rate, Rhythm: rhythm, NextMsgFunc: next_func}
}

func FuzzedNewSimpleHumanTraits(fuzzer fuzz.Fuzzer, next_func func(float64, float64, float64, float64) float64) *SimpleHumanTraits {
	var sh SimpleHumanTraits
	fuzzer.Fuzz(&sh)
	sh.NextMsgFunc = next_func
	return &sh
}
