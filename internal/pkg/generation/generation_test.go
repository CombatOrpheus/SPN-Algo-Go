package generation

import (
	"spn-benchmark-ds/internal/pkg/petrinet"
	"testing"
)

func TestGenerateReachabilityGraph(t *testing.T) {
	// Create a simple Petri net for testing
	pn := petrinet.NewPetriNet(2, 1)
	pn.Matrix = []int{
		1, 0, 1, // P1 -> T1
		0, 1, 0, // T1 -> P2
	}
	pn.InitialMarking = []int{1, 0}

	rg, err := GenerateReachabilityGraph(pn, 10, 100)
	if err != nil {
		t.Fatalf("Error generating reachability graph: %v", err)
	}

	if !rg.IsBounded {
		t.Errorf("Expected the graph to be bounded, but it was not")
	}

	// Expected vertices: [1, 0] and [0, 1]
	if rg.NumVertices != 2 {
		t.Errorf("Expected 2 vertices, but got %d", rg.NumVertices)
	}

	// Expected edge: [1, 0] -> [0, 1]
	if rg.NumEdges != 1 {
		t.Errorf("Expected 1 edge, but got %d", rg.NumEdges)
	}

	edge := rg.Edge(0)
	if edge[0] != 0 || edge[1] != 1 {
		t.Errorf("Expected edge from vertex 0 to 1, but got from %d to %d", edge[0], edge[1])
	}

	if len(rg.ArcTransitions) != 1 {
		t.Errorf("Expected 1 arc transition, but got %d", len(rg.ArcTransitions))
	}

	if rg.ArcTransitions[0] != 0 {
		t.Errorf("Expected arc transition to be 0, but got %d", rg.ArcTransitions[0])
	}
}
