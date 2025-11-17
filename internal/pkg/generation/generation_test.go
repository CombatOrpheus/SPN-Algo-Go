package generation

import (
	"spn-benchmark-ds/internal/pkg/petrinet"
	"testing"
)

func TestGenerateReachabilityGraph(t *testing.T) {
	pn := petrinet.NewPetriNet(2, 2)
	// Manually set up a simple Petri Net for testing
	// P1 -> T1 -> P2
	// T2 -> P1
	pn.Matrix = [][]int{
		{1, 0, 0, 1, 1}, // Place 1
		{0, 1, 1, 0, 0}, // Place 2
	}
	pn.InitialMarking = []int{1, 0}

	graph, err := GenerateReachabilityGraph(pn, 10, 100)
	if err != nil {
		t.Fatalf("Error generating reachability graph: %v", err)
	}

	if !graph.IsBounded {
		t.Error("Expected the graph to be bounded, but it was not")
	}

	// Expected vertices: [1, 0], [0, 1]
	if len(graph.Vertices) != 2 {
		t.Errorf("Expected 2 vertices, but got %d", len(graph.Vertices))
	}

	// Expected edges: [0, 1] (from T1), [1, 0] (from T2)
	if len(graph.Edges) != 2 {
		t.Errorf("Expected 2 edges, but got %d", len(graph.Edges))
	}
}

func TestGenerateReachabilityGraph_Unbounded(t *testing.T) {
	pn := petrinet.NewPetriNet(1, 1)
	// T1 -> P1 (unbounded)
	// No pre-condition for T1, so it can always fire, adding a token to P1.
	pn.Matrix = [][]int{
		{0, 1, 0},
	}
	pn.InitialMarking = []int{0}

	graph, err := GenerateReachabilityGraph(pn, 5, 100)
	if err != nil {
		t.Fatalf("Error generating reachability graph: %v", err)
	}

	if graph.IsBounded {
		t.Error("Expected the graph to be unbounded, but it was not")
	}
}
