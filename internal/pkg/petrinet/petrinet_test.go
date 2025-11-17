package petrinet

import (
	"testing"
)

func TestGenerateRandomPetriNet(t *testing.T) {
	numPlaces := 5
	numTransitions := 3
	pn := GenerateRandomPetriNet(numPlaces, numTransitions)

	if pn.Places != numPlaces {
		t.Errorf("Expected %d places, but got %d", numPlaces, pn.Places)
	}
	if pn.Transitions != numTransitions {
		t.Errorf("Expected %d transitions, but got %d", numTransitions, pn.Transitions)
	}
	if len(pn.Matrix) != numPlaces*(2*numTransitions+1) {
		t.Errorf("Expected matrix to have %d rows, but got %d", numPlaces, len(pn.Matrix))
	}

	// Check if there is an initial marking
	initialMarkingSum := 0
	for i := 0; i < numPlaces; i++ {
		initialMarkingSum += pn.At(i, 2*numTransitions)
	}
	if initialMarkingSum == 0 {
		t.Errorf("Expected at least one initial marking, but got none")
	}

	// Check if the initial marking in the struct matches the one in the matrix
	for i := 0; i < numPlaces; i++ {
		if pn.InitialMarking[i] != pn.At(i, 2*numTransitions) {
			t.Errorf("Initial marking in struct does not match matrix at place %d", i)
		}
	}
}

func TestPrune(t *testing.T) {
	// Create a Petri net with excess edges
	pn := NewPetriNet(2, 2)
	pn.Matrix = []int{
		1, 1, 1, 1, 0,
		1, 1, 1, 1, 0,
	}

	initialEdgeCount := 0
	for i := 0; i < pn.Places; i++ {
		for j := 0; j < 2*pn.Transitions; j++ {
			initialEdgeCount += pn.At(i, j)
		}
	}

	pn.Prune()

	// Check that the number of edges has been reduced
	finalEdgeCount := 0
	for i := 0; i < pn.Places; i++ {
		for j := 0; j < 2*pn.Transitions; j++ {
			finalEdgeCount += pn.At(i, j)
		}
	}

	if finalEdgeCount >= initialEdgeCount {
		t.Errorf("Pruning did not reduce the number of edges")
	}

	// Create a Petri net with missing connections
	pn = NewPetriNet(2, 2)
	pn.Matrix = []int{
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
	}

	pn.Prune()

	// Check that connections have been added
	for j := 0; j < 2*pn.Transitions; j++ {
		colSum := 0
		for i := 0; i < pn.Places; i++ {
			colSum += pn.At(i, j)
		}
		if colSum == 0 {
			t.Errorf("Pruning failed to add connection to transition %d", j)
		}
	}
}

func TestAddTokensRandomly(t *testing.T) {
	pn := NewPetriNet(10, 5)

	pn.AddTokensRandomly()

	// Check that some tokens have been added
	tokenSum := 0
	for _, marking := range pn.InitialMarking {
		tokenSum += marking
	}
	if tokenSum == 0 {
		// It's possible, but unlikely, that no tokens are added.
		// Run it again to be sure.
		pn.AddTokensRandomly()
		for _, marking := range pn.InitialMarking {
			tokenSum += marking
		}
		if tokenSum == 0 {
			t.Errorf("AddTokensRandomly did not add any tokens")
		}
	}
}
