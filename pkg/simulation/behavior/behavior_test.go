package Behavior

import (
	Types "deniable-im/im-sim/pkg/simulation/types"
	"math/rand"
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestFuzzPureProbDist(t *testing.T) {
	f := fuzz.NewWithSeed(4206969).NilChance(0)
	r := rand.New(rand.NewSource(42069))

	d := NewFuzzedPureProbabilityDistribution(f, func(test *PureProbabilityDistribution) float64 { return 0 }, r)

	if d == nil {
		t.Error("New function returned nil")
	}

	if d.Rate != d.GetNextMessageTime() {
		t.Error("Error in ProbabilityFunction assigment")
	}

}

func TestRespondLogic(t *testing.T) {
	f := fuzz.NewWithSeed(6942069).NilChance(0)
	r := rand.New(rand.NewSource(42069))

	pp := NewFuzzedPureProbabilityDistribution(f, func(test *PureProbabilityDistribution) float64 { return 0 }, r)

	res1 := pp.WillRespond()
	if res1 {
		t.Error("Pure Probability Distribution Behavior responds to messages")
	}

	sh := SimpleHumanTraits{ResponseProb: 0}
	if sh.WillRespond(Types.Msg{}) {
		t.Error("Simple Human Behavior responds to messages while response prop = 0")
	}

	sh.ResponseProb = 1.0
	if !sh.WillRespond(Types.Msg{}) {
		t.Error("Simple Human Behavior does not respond to messages while response prop = 1")
	}

}
