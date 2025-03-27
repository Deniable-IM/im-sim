package Behavior

import (
	"fmt"
	"math/rand"
	"strings"

	fuzz "github.com/google/gofuzz"
)

type PureProbabilityDistribution struct {
	Name                string
	Rate                float64
	ProbabilityFunction func(float64, float64) float64
	DeniableModifier    float64
	DeniableProb        float64
	randomizer          *rand.Rand
}

func (q *PureProbabilityDistribution) GetRandomizer() *rand.Rand {
	return q.randomizer
}

func (q *PureProbabilityDistribution) GetBehaviorName() string {
	if q == nil {
		return ""
	}

	var b strings.Builder

	fmt.Fprintf(&b, "%v with lambda %v", q.Name, q.Rate)
	return b.String()
}

func (q *PureProbabilityDistribution) GetNextMessageTime() float64 {
	if q == nil {
		return 0
	}
	return q.ProbabilityFunction(q.Rate, q.DeniableModifier)
}

func (q *PureProbabilityDistribution) SendRegularMsg() bool {
	return false
}

func (q *PureProbabilityDistribution) SendDeniableMsg() bool {
	if q == nil {
		return false
	}

	return q.randomizer.Float64() > (1.0 - q.DeniableProb)
}

func (q *PureProbabilityDistribution) WillRespond() bool {
	return false
}

func NewPureProbabilityDistribution(name string, rate float64, distribution func(float64, float64) float64, deniable_mod float64) *PureProbabilityDistribution {
	return &PureProbabilityDistribution{Name: name, Rate: rate, ProbabilityFunction: distribution, DeniableModifier: deniable_mod}
}

func NewFuzzedPureProbabilityDistribution(fuzzer *fuzz.Fuzzer, distribution func(float64, float64) float64) *PureProbabilityDistribution {
	//TODO: Actually implement fuzzy wuzzy
	var name string
	var rate, deniable_mod float64
	fuzzer.Fuzz(&name)
	fuzzer.Fuzz(&rate)
	fuzzer.Fuzz(&deniable_mod)

	return NewPureProbabilityDistribution(name, rate, distribution, deniable_mod)
}
