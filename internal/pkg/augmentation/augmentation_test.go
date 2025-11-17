package augmentation

import (
	"spn-benchmark-ds/internal/pkg/petrinet"
	"spn-benchmark-ds/internal/pkg/generation"
	"testing"
)

func TestGenerateLambdaVariations(t *testing.T) {
	// This reachability graph corresponds to a simple P-T net:
	// P1 -> T1 -> P2, with initial marking (1, 0)
	pn := petrinet.NewPetriNet(2, 1)
	rg := &generation.ReachabilityGraph{
		Vertices:       []int{1, 0, 0, 1},
		Edges:          []int{0, 1},
		VerticesStride: 2,
		EdgesStride:    2,
		NumVertices:    2,
		NumEdges:       1,
		ArcTransitions: []int{0},
		IsBounded:      true,
	}
	numVariations := 5
	minFiringRate := 1
	maxFiringRate := 10

	variations, lambdaValuesList := GenerateLambdaVariations(pn, rg, numVariations, minFiringRate, maxFiringRate)

	if len(variations) != numVariations {
		t.Errorf("GenerateLambdaVariations returned %d variations, expected %d", len(variations), numVariations)
	}

	if len(lambdaValuesList) != numVariations {
		t.Errorf("GenerateLambdaVariations returned %d lambdaValuesList, expected %d", len(lambdaValuesList), numVariations)
	}
}

func TestGeneratePetriNetVariations(t *testing.T) {
	pn := petrinet.NewPetriNet(5, 3)
	numVariations := 5
	placeUpperBound := 10
	marksLowerLimit := 1
	marksUpperLimit := 100
	minFiringRate := 1
	maxFiringRate := 10

	variations := GeneratePetriNetVariations(pn, placeUpperBound, marksLowerLimit, marksUpperLimit, numVariations, minFiringRate, maxFiringRate)

	if len(variations) != numVariations {
		t.Errorf("GeneratePetriNetVariations returned %d variations, expected %d", len(variations), numVariations)
	}
}
