package generation

import (
	"spn-benchmark-ds/internal/pkg/petrinet"
	"testing"
)

func TestGenerateReachabilityGraph(t *testing.T) {
	// Create a simple Petri net for testing
	pn := &petrinet.PetriNet{
		Places:      2,
		Transitions: 1,
		Matrix: [][]int{
			{1, 0, 1}, // P1 -> T1
			{0, 1, 0}, // T1 -> P2
		},
		InitialMarking: []int{1, 0},
	}

	rg, err := GenerateReachabilityGraph(pn, 10, 100)
	if err != nil {
		t.Fatalf("Error generating reachability graph: %v", err)
	}

	if !rg.IsBounded {
		t.Errorf("Expected the graph to be bounded, but it was not")
	}

	// Expected vertices: [1, 0] and [0, 1]
	if len(rg.Vertices) != 2 {
		t.Errorf("Expected 2 vertices, but got %d", len(rg.Vertices))
	}

	// Expected edge: [1, 0] -> [0, 1]
	if len(rg.Edges) != 1 {
		t.Errorf("Expected 1 edge, but got %d", len(rg.Edges))
	}

	if rg.Edges[0][0] != 0 || rg.Edges[0][1] != 1 {
		t.Errorf("Expected edge from vertex 0 to 1, but got from %d to %d", rg.Edges[0][0], rg.Edges[0][1])
	}

	if len(rg.ArcTransitions) != 1 {
		t.Errorf("Expected 1 arc transition, but got %d", len(rg.ArcTransitions))
	}

	if rg.ArcTransitions[0] != 0 {
		t.Errorf("Expected arc transition to be 0, but got %d", rg.ArcTransitions[0])
	}
}
