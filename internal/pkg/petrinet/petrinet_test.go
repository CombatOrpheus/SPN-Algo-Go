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

	if len(pn.Matrix) != numPlaces {
		t.Errorf("Expected matrix to have %d rows, but got %d", numPlaces, len(pn.Matrix))
	}

	for i, row := range pn.Matrix {
		if len(row) != 2*numTransitions+1 {
			t.Errorf("Expected row %d to have %d columns, but got %d", i, 2*numTransitions+1, len(row))
		}
	}

	initialMarkingSum := 0
	for _, row := range pn.Matrix {
		initialMarkingSum += row[2*numTransitions]
	}

	if initialMarkingSum == 0 {
		t.Error("Expected at least one token in the initial marking, but got none")
	}
}
