package Behavior

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	fuzz "github.com/google/gofuzz"
)

type SimpleHumanTraits struct {
	Name              string
	SendRate          float64
	ForgetRate        float64
	Rhythm            float64
	LastSentMessageAt float64
	SendProp          float64
	ResponseProb      float64
	DeniableProb      float64
	nextMsgFunc       func(*SimpleHumanTraits) float64
	randomizer        *rand.Rand
}

func (sh *SimpleHumanTraits) GetBehaviorName() string {
	if sh == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Simple Human Traits Behavior with name %v, Forget_rate = %v, Rhythm = %v", sh.Name, sh.ForgetRate, sh.Rhythm)
	return b.String()
}

func (sh *SimpleHumanTraits) GetNextMessageTime() float64 {
	if sh == nil {
		return 0
	}

	next := sh.nextMsgFunc(sh)

	return next
}

func (sh *SimpleHumanTraits) GetRandomizer() *rand.Rand {
	return sh.randomizer
}

func (sh *SimpleHumanTraits) WillRespond() bool {
	if sh == nil {
		return false
	}

	return sh.randomizer.Float64() > (1.0 - sh.ResponseProb)
}

func (sh *SimpleHumanTraits) SendRegularMsg() bool {
	if sh == nil {
		return false
	}

	return sh.randomizer.Float64() > (1.0 - sh.SendProp)
}

func (sh *SimpleHumanTraits) SendDeniableMsg() bool {
	if sh == nil {
		return false
	}

	return sh.randomizer.Float64() > (1.0 - sh.DeniableProb)
}

func (sh *SimpleHumanTraits) GetResponseTime(max float64) float64 {
	if sh == nil {
		return 0
	}

	return math.Min(max, sh.nextMsgFunc(sh))
}

func NewSimpleHumanTraits(name string, send_rate, forget_rate, rhythm, send_prop, response, deniable_prop float64, next_func func(*SimpleHumanTraits) float64, r *rand.Rand) *SimpleHumanTraits {
	return &SimpleHumanTraits{Name: name, SendRate: send_rate, ForgetRate: forget_rate, Rhythm: rhythm, SendProp: send_prop, ResponseProb: response, DeniableProb: deniable_prop, nextMsgFunc: next_func, randomizer: r}
}

func FuzzedNewSimpleHumanTraits(fuzzer fuzz.Fuzzer, next_func func(*SimpleHumanTraits) float64, r *rand.Rand) *SimpleHumanTraits {
	var sh SimpleHumanTraits
	fuzzer.Fuzz(&sh)
	sh.nextMsgFunc = next_func
	sh.randomizer = r

	return &sh
}
