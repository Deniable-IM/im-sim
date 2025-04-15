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
	ProbabilityFunction func(*PureProbabilityDistribution) float64
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
	return q.ProbabilityFunction(q)
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

func NewPureProbabilityDistribution(name string, rate float64, distribution func(*PureProbabilityDistribution) float64, deniable_mod, deniable_prop float64, randomizer *rand.Rand) *PureProbabilityDistribution {
	return &PureProbabilityDistribution{Name: name, Rate: rate, ProbabilityFunction: distribution, DeniableModifier: deniable_mod, DeniableProb: deniable_prop, randomizer: randomizer}
}

func NewFuzzedPureProbabilityDistribution(fuzzer *fuzz.Fuzzer, distribution func(*PureProbabilityDistribution) float64, randomizer *rand.Rand) *PureProbabilityDistribution {
	var q PureProbabilityDistribution
	fuzzer.Fuzz(&q)
	q.ProbabilityFunction = distribution
	q.randomizer = randomizer

	return &q
}
