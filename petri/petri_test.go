package petri

import (
	"testing"
	"gonum.org/v1/gonum/mat"
)

func TestPetriNet(t *testing.T) {
	// This test is just to ensure the code compiles.
	// We will add more tests later.
	_ = &PetriNet{
		Places:      2,
		Transitions: 1,
		Matrix:      mat.NewDense(2, 3, nil),
	}
}

func TestStochasticPetriNet(t *testing.T) {
	// This test is just to ensure the code compiles.
	_ = &StochasticPetriNet{
		PetriNet: &PetriNet{
			Places:      2,
			Transitions: 1,
			Matrix:      mat.NewDense(2, 3, nil),
		},
		ReachabilityGraph: mat.NewDense(2, 2, nil),
		Edges:             [][2]int{{0, 1}},
		ArcTransitions:    []int{0},
		FiringRates:       []float64{1.0},
		SteadyStateProbs:  []float64{0.5, 0.5},
		MarkingDensities:  mat.NewDense(2, 2, nil),
		AverageMarkings:   []float64{0.5, 0.5},
	}
}
