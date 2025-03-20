package Behavior

import (
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestFuzzPureProbDist(t *testing.T) {
	f := fuzz.NewWithSeed(4206969).NilChance(0)
	d := NewFuzzedPureProbabilityDistribution(f, func(f1, f2 float64) float64 { return f1 })

	if d == nil {
		t.Error("New function returned nil")
	}

	if d.Rate != d.GetNextMessageTime(0) {
		t.Error("Error in ProbabilityFunction assigment")
	}

}

func TestRespondLogic(t *testing.T) {
	f := fuzz.NewWithSeed(6942069).NilChance(0)
	pp := NewFuzzedPureProbabilityDistribution(f, func(x, y float64) float64 { return y })

	res1 := pp.WillRespond()
	if res1 {
		t.Error("Pure Probability Distribution Behavior responds to messages")
	}

	sh := SimpleHumanTraits{ResponseProb: 0}
	if sh.WillRespond() {
		t.Error("Simple Human Behavior responds to messages while response prop = 0")
	}

	sh.ResponseProb = 1.0
	if !sh.WillRespond() {
		t.Error("Simple Human Behavior does not respond to messages while response prop = 1")
	}

}
