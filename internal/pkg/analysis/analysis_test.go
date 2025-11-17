package analysis

import (
	"spn-benchmark-ds/internal/pkg/generation"
	"testing"
)

func TestSPNAnalysis(t *testing.T) {
	rg := &generation.ReachabilityGraph{
		Vertices: [][]int{
			{1, 0},
			{0, 1},
		},
		Edges: [][2]int{
			{0, 1},
			{1, 0},
		},
		ArcTransitions: []int{0, 1},
	}
	lambdaValues := []float64{1.0, 1.0}

	stateMatrix, targetVector := ComputeStateEquation(rg, lambdaValues)

	if stateMatrix.At(0, 0) != -1.0 {
		t.Errorf("Expected stateMatrix.At(0, 0) to be -1.0, but got %f", stateMatrix.At(0, 0))
	}
	if stateMatrix.At(1, 0) != 1.0 {
		t.Errorf("Expected stateMatrix.At(1, 0) to be 1.0, but got %f", stateMatrix.At(1, 0))
	}
	if targetVector.AtVec(2) != 1.0 {
		t.Errorf("Expected targetVector.AtVec(2) to be 1.0, but got %f", targetVector.AtVec(2))
	}

	steadyStateProbs, err := SolveForSteadyState(stateMatrix, targetVector)
	if err != nil {
		t.Fatalf("Error solving for steady state: %v", err)
	}

	if len(steadyStateProbs) != 2 {
		t.Fatalf("Expected 2 steady state probabilities, but got %d", len(steadyStateProbs))
	}

	// For a symmetric system, we expect equal probabilities
	if steadyStateProbs[0] < 0.49 || steadyStateProbs[0] > 0.51 {
		t.Errorf("Expected steadyStateProbs[0] to be ~0.5, but got %f", steadyStateProbs[0])
	}
	if steadyStateProbs[1] < 0.49 || steadyStateProbs[1] > 0.51 {
		t.Errorf("Expected steadyStateProbs[1] to be ~0.5, but got %f", steadyStateProbs[1])
	}

	avgMarkings, _ := ComputeAverageMarkings(rg.Vertices, steadyStateProbs)

	if len(avgMarkings) != 2 {
		t.Fatalf("Expected 2 average markings, but got %d", len(avgMarkings))
	}

	// Expected average marking for each place is 0.5
	if avgMarkings[0] < 0.49 || avgMarkings[0] > 0.51 {
		t.Errorf("Expected avgMarkings[0] to be ~0.5, but got %f", avgMarkings[0])
	}
	if avgMarkings[1] < 0.49 || avgMarkings[1] > 0.51 {
		t.Errorf("Expected avgMarkings[1] to be ~0.5, but got %f", avgMarkings[1])
	}
}
