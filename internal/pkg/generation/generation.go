package generation

import (
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
	preMatrix := make([][]int, numTransitions)
	postMatrix := make([][]int, numTransitions)
	changeMatrix := make([][]int, numTransitions)
	for t := 0; t < numTransitions; t++ {
		preMatrix[t] = make([]int, pn.Places)
		postMatrix[t] = make([]int, pn.Places)
		changeMatrix[t] = make([]int, pn.Places)
		for p := 0; p < pn.Places; p++ {
			preMatrix[t][p] = pn.At(p, t)
			postMatrix[t][p] = pn.At(p, t+numTransitions)
			changeMatrix[t][p] = postMatrix[t][p] - preMatrix[t][p]
		}
	}

	initialMarking := pn.InitialMarking

	visitedMarkings := make(map[string]int)
	visitedMarkings[markingToString(initialMarking)] = 0

	// ⚡ Bolt: Replaced container/list with a slice-based queue.
	// This eliminates heap allocations for queue elements and
	// interface{} type assertion overhead, significantly speeding up BFS.
	queue := make([]int, 0, 1024)
	queue = append(queue, 0)
	head := 0

	graph := &ReachabilityGraph{
		Vertices:         make([]int, 1*len(initialMarking)),
		Edges:            make([]int, 20),
		VerticesStride:   len(initialMarking),
		EdgesStride:      2,
		IsBounded:        true,
		verticesCapacity: 1,
		edgesCapacity:    10,
	}
	graph.AddVertex(initialMarking)

	for head < len(queue) {

		currentMarkingIndex := queue[head]
		head++
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
				queue = append(queue, graph.NumVertices-1)
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
// ⚡ Bolt: Optimized by pre-allocating return slices to eliminate allocations in loop.
// Also changed preMatrix/changeMatrix orientation to [transitions][places]
// to improve cache locality when evaluating each transition's requirements.
func getEnabledTransitions(preMatrix, changeMatrix [][]int, currentMarking []int) ([]int, [][]int) {
	numTransitions := len(preMatrix)
	enabledTransitions := make([]int, 0, numTransitions)
	newMarkings := make([][]int, 0, numTransitions)
	numPlaces := len(currentMarking)

	for t := 0; t < numTransitions; t++ {
		isEnabled := true
		preT := preMatrix[t]
		for p := 0; p < numPlaces; p++ {
			if currentMarking[p] < preT[p] {
				isEnabled = false
				break
			}
		}
		if isEnabled {
			newMarking := make([]int, numPlaces)
			changeT := changeMatrix[t]
			for p := 0; p < numPlaces; p++ {
				newMarking[p] = currentMarking[p] + changeT[p]
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
	l := len(marking)
	if l == 0 {
		return "[]"
	}

	// Preallocate with exact 4 bytes per 32-bit integer.
	// We use direct byte mapping to serialize the marking into a compact string.
	// This generates unique byte signatures for markings faster than formatting methods,
	// drastically improving hashing performance during BFS state space exploration.
	b := make([]byte, l*4)
	for i, v := range marking {
		idx := i * 4
		b[idx] = byte(v >> 24)
		b[idx+1] = byte(v >> 16)
		b[idx+2] = byte(v >> 8)
		b[idx+3] = byte(v)
	}
	return string(b)
}
