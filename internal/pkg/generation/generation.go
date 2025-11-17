package generation

import (
	"container/list"
	"fmt"
	"spn-benchmark-ds/internal/pkg/petrinet"
)

// ReachabilityGraph represents the reachability graph of a Petri Net.
type ReachabilityGraph struct {
	// Vertices is a flattened slice of vertices.
	Vertices []int
	// Edges is a flattened slice of edges.
	Edges []int
	// VerticesStride is the stride of the vertices slice.
	VerticesStride int
	// EdgesStride is the stride of the edges slice.
	EdgesStride int
	// NumVertices is the number of vertices in the graph.
	NumVertices int
	// NumEdges is the number of edges in the graph.
	NumEdges int
	// ArcTransitions is a slice of arc transitions.
	ArcTransitions []int
	// IsBounded is true if the graph is bounded.
	IsBounded bool
	// verticesCapacity is the capacity of the vertices slice.
	verticesCapacity int
	// edgesCapacity is the capacity of the edges slice.
	edgesCapacity int
}

// Vertex returns the vertex at the given index.
func (rg *ReachabilityGraph) Vertex(index int) []int {
	return rg.Vertices[index*rg.VerticesStride : (index+1)*rg.VerticesStride]
}

// AddVertex adds a new vertex to the graph.
func (rg *ReachabilityGraph) AddVertex(vertex []int) {
	if rg.NumVertices >= rg.verticesCapacity {
		rg.verticesCapacity *= 2
		newVertices := make([]int, rg.verticesCapacity*rg.VerticesStride)
		copy(newVertices, rg.Vertices)
		rg.Vertices = newVertices
	}
	copy(rg.Vertices[rg.NumVertices*rg.VerticesStride:], vertex)
	rg.NumVertices++
}

// Edge returns the edge at the given index.
func (rg *ReachabilityGraph) Edge(index int) []int {
	return rg.Edges[index*rg.EdgesStride : (index+1)*rg.EdgesStride]
}

// AddEdge adds a new edge to the graph.
func (rg *ReachabilityGraph) AddEdge(edge [2]int) {
	if rg.NumEdges >= rg.edgesCapacity {
		rg.edgesCapacity *= 2
		newEdges := make([]int, rg.edgesCapacity*rg.EdgesStride)
		copy(newEdges, rg.Edges)
		rg.Edges = newEdges
	}
	rg.Edges[rg.NumEdges*rg.EdgesStride] = edge[0]
	rg.Edges[rg.NumEdges*rg.EdgesStride+1] = edge[1]
	rg.NumEdges++
}

// GenerateReachabilityGraph generates the reachability graph of a Petri net using BFS.
// It takes a Petri net and a set of parameters and returns a reachability graph.
func GenerateReachabilityGraph(pn *petrinet.PetriNet, placeUpperLimit int, maxMarkingsToExplore int) (*ReachabilityGraph, error) {
	numTransitions := pn.Transitions
	preMatrix := make([][]int, pn.Places)
	postMatrix := make([][]int, pn.Places)
	changeMatrix := make([][]int, pn.Places)
	for i := 0; i < pn.Places; i++ {
		preMatrix[i] = make([]int, numTransitions)
		postMatrix[i] = make([]int, numTransitions)
		changeMatrix[i] = make([]int, numTransitions)
		for j := 0; j < numTransitions; j++ {
			preMatrix[i][j] = pn.At(i, j)
			postMatrix[i][j] = pn.At(i, j+numTransitions)
			changeMatrix[i][j] = postMatrix[i][j] - preMatrix[i][j]
		}
	}

	initialMarking := pn.InitialMarking

	visitedMarkings := make(map[string]int)
	visitedMarkings[markingToString(initialMarking)] = 0

	queue := list.New()
	queue.PushBack(0)

	graph := &ReachabilityGraph{
		Vertices:         make([]int, 1*len(initialMarking)),
		Edges:            make([]int, 10),
		VerticesStride:   len(initialMarking),
		EdgesStride:      2,
		IsBounded:        true,
		verticesCapacity: 1,
		edgesCapacity:    10,
	}
	graph.AddVertex(initialMarking)

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		currentMarkingIndex := element.Value.(int)
		currentMarking := graph.Vertex(currentMarkingIndex)

		if graph.NumVertices >= maxMarkingsToExplore {
			graph.IsBounded = false
			break
		}

		enabledTransitions, newMarkings := getEnabledTransitions(preMatrix, changeMatrix, currentMarking)

		for i, newMarking := range newMarkings {
			if isMarkingOutOfBounds(newMarking, placeUpperLimit) {
				graph.IsBounded = false
				break
			}

			markingStr := markingToString(newMarking)
			if _, ok := visitedMarkings[markingStr]; !ok {
				visitedMarkings[markingStr] = graph.NumVertices
				graph.AddVertex(newMarking)
				queue.PushBack(graph.NumVertices - 1)
			}
			graph.AddEdge([2]int{currentMarkingIndex, visitedMarkings[markingStr]})
			graph.ArcTransitions = append(graph.ArcTransitions, enabledTransitions[i])
		}
		if !graph.IsBounded {
			break
		}
	}
	return graph, nil
}

// getEnabledTransitions returns the enabled transitions and the new markings.
func getEnabledTransitions(preMatrix, changeMatrix [][]int, currentMarking []int) ([]int, [][]int) {
	numTransitions := len(preMatrix[0])
	var enabledTransitions []int
	var newMarkings [][]int

	for t := 0; t < numTransitions; t++ {
		isEnabled := true
		for p := 0; p < len(preMatrix); p++ {
			if currentMarking[p] < preMatrix[p][t] {
				isEnabled = false
				break
			}
		}
		if isEnabled {
			newMarking := make([]int, len(currentMarking))
			for p := 0; p < len(currentMarking); p++ {
				newMarking[p] = currentMarking[p] + changeMatrix[p][t]
			}
			enabledTransitions = append(enabledTransitions, t)
			newMarkings = append(newMarkings, newMarking)
		}
	}
	return enabledTransitions, newMarkings
}

// isMarkingOutOfBounds returns true if the marking is out of bounds.
func isMarkingOutOfBounds(marking []int, placeUpperLimit int) bool {
	for _, tokens := range marking {
		if tokens > placeUpperLimit {
			return true
		}
	}
	return false
}

// markingToString converts a marking to a string.
func markingToString(marking []int) string {
	return fmt.Sprintf("%v", marking)
}
