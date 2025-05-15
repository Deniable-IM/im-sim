package Behavior

import (
	"fmt"
	"math/rand"
	"strings"

	fuzz "github.com/google/gofuzz"
)

type SimpleHumanTraits struct {
	Name              string
	SendProp          float64
	ResponseProb      float64
	DeniableProb      float64
	BurstModifier     float64
	DeniableBurstSize int32
	DeniableCount     int32
	nextMsgFunc       func(*SimpleHumanTraits) int
	randomizer        *rand.Rand
}

func (sh *SimpleHumanTraits) GetBehaviorName() string {
	if sh == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Simple Human Traits Behavior with name %v", sh.Name)
	return b.String()
}

func (sh *SimpleHumanTraits) GetNextMessageTime() int {
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

func (sh *SimpleHumanTraits) GetResponseTime(max int64) int {
	if sh == nil {
		return 0
	}

	time := int32(max)
	//Clause to avoid randomizer panicking. 1 ms difference is most likely not a problem
	if time < 1 {
		time = 1
	}

	return int(sh.randomizer.Int31n((time)))
}

func (sh *SimpleHumanTraits) IncrementDeniableCount() {
	sh.DeniableCount += sh.DeniableBurstSize
}

func (sh *SimpleHumanTraits) IsBursting() bool {
	return sh.DeniableCount > 0
}

func NewSimpleHumanTraits(
	name string,
	send_prop, response, deniable_prop, burst_mod float64,
	deniable_burst_size int32,
	next_func func(*SimpleHumanTraits) int,
	r *rand.Rand) *SimpleHumanTraits {
	return &SimpleHumanTraits{
		Name:              name,
		SendProp:          send_prop,
		ResponseProb:      response,
		DeniableProb:      deniable_prop,
		BurstModifier:     burst_mod,
		DeniableBurstSize: deniable_burst_size,
		nextMsgFunc:       next_func,
		randomizer:        r,
	}
}

func FuzzedNewSimpleHumanTraits(
	fuzzer fuzz.Fuzzer,
	next_func func(*SimpleHumanTraits) int,
	r *rand.Rand) *SimpleHumanTraits {
	var sh SimpleHumanTraits
	fuzzer.Fuzz(&sh)
	sh.nextMsgFunc = next_func
	sh.randomizer = r
	sh.DeniableBurstSize = sh.DeniableBurstSize % 10

	return &sh
}
