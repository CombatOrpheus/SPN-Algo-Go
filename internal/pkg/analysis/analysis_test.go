package analysis

import (
	"math"
	"spn-benchmark-ds/internal/pkg/generation"
	"testing"
)

const float64EqualityThreshold = 1e-9

func float64Equals(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func TestAnalysisFunctions(t *testing.T) {
	// This reachability graph corresponds to a simple P-T net:
	// P1 -> T1 -> P2, with initial marking (1, 0)
	rg := &generation.ReachabilityGraph{
		Vertices:       [][]int{{1, 0}, {0, 1}},
		Edges:          [][2]int{{0, 1}},
		ArcTransitions: []int{0},
		IsBounded:      true,
	}
	lambdaValues := []float64{1.0}

	// 1. Test ComputeStateEquation
	stateMatrix, targetVector := ComputeStateEquation(rg, lambdaValues)

	// Expected state matrix (transposed generator matrix + row of 1s)
	// -1.0  0.0
	//  1.0  0.0  -> This system is wrong, T1 is not connected back to P1
	//  1.0  1.0
	// The system is M0 -> M1, where M1 is an absorbing state.
	// The generator matrix Q is:
	// -1  1
	//  0  0
	// The python code computes Q^T:
	// -1  0
	//  1  0
	// The stateMatrix in Go should be this Q^T plus a row of 1s for the probability sum constraint.
	expectedStateMatrixData := []float64{
		-1.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
	}
	for i, val := range expectedStateMatrixData {
		if r, c := i/2, i%2; !float64Equals(stateMatrix.At(r, c), val) {
			t.Errorf("Expected stateMatrix.At(%d, %d) to be %f, but got %f", r, c, val, stateMatrix.At(r, c))
		}
	}

	expectedTargetVectorData := []float64{0.0, 0.0, 1.0}
	for i, val := range expectedTargetVectorData {
		if !float64Equals(targetVector.AtVec(i), val) {
			t.Errorf("Expected targetVector.AtVec(%d) to be %f, but got %f", i, val, targetVector.AtVec(i))
		}
	}

	// 2. Test SolveForSteadyState
	steadyStateProbs, err := SolveForSteadyState(stateMatrix, targetVector)
	if err != nil {
		t.Fatalf("Error solving for steady state: %v", err)
	}

	// Expected steady state probabilities: [0.0, 1.0] (state {0,1} is absorbing)
	expectedProbs := []float64{0.0, 1.0}
	if len(steadyStateProbs) != len(expectedProbs) {
		t.Fatalf("Expected %d steady state probabilities, but got %d", len(expectedProbs), len(steadyStateProbs))
	}
	for i, prob := range expectedProbs {
		if !float64Equals(steadyStateProbs[i], prob) {
			t.Errorf("Expected probability %d to be %f, but got %f", i, prob, steadyStateProbs[i])
		}
	}

	// 3. Test ComputeAverageMarkings
	avgMarkings, markingDensities := ComputeAverageMarkings(rg.Vertices, steadyStateProbs)

	// Expected average markings: [0.0, 1.0]
	expectedAvgMarkings := []float64{0.0, 1.0}
	if len(avgMarkings) != len(expectedAvgMarkings) {
		t.Fatalf("Expected %d average markings, but got %d", len(expectedAvgMarkings), len(avgMarkings))
	}
	for i, marking := range expectedAvgMarkings {
		if !float64Equals(avgMarkings[i], marking) {
			t.Errorf("Expected average marking %d to be %f, but got %f", i, marking, avgMarkings[i])
		}
	}

	// Expected marking densities:
	// Place 0: 100% prob of 0 tokens, 0% prob of 1 token
	// Place 1: 0% prob of 0 tokens, 100% prob of 1 token
	expectedMarkingDensities := [][]float64{
		{1.0, 0.0},
		{0.0, 1.0},
	}
	if len(markingDensities) != len(expectedMarkingDensities) {
		t.Fatalf("Expected %d marking densities, but got %d", len(expectedMarkingDensities), len(markingDensities))
	}
	for i, densities := range expectedMarkingDensities {
		if len(markingDensities[i]) != len(densities) {
			t.Fatalf("Expected %d densities for place %d, but got %d", len(densities), i, len(markingDensities[i]))
		}
		for j, density := range densities {
			if !float64Equals(markingDensities[i][j], density) {
				t.Errorf("Expected marking density for place %d, token %d to be %f, but got %f", i, j, density, markingDensities[i][j])
			}
		}
	}
}
