package generation

import (
	"spn-benchmark-ds/internal/pkg/petrinet"
	"strconv"
	"unsafe"
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
	visitedMarkings[hashMarking(initialMarking)] = 0

	// ⚡ Bolt: Replaced container/list with a slice-based queue.
	// This eliminates heap allocations for queue elements and
	// interface{} type assertion overhead, significantly speeding up BFS.
	queue := make([]int, 0, 1024)
	queue = append(queue, 0)
	head := 0
	scratchMarking := make([]int, pn.Places)

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

		// ⚡ Bolt: Inlined getEnabledTransitions to eliminate allocations.
		// Instead of returning new slices, we evaluate transitions directly
		// and use a scratch slice to hash/check bounds before copying.
		for t := 0; t < numTransitions; t++ {
			isEnabled := true
			preT := preMatrix[t]
			for p := 0; p < pn.Places; p++ {
				if currentMarking[p] < preT[p] {
					isEnabled = false
					break
				}
			}

			if isEnabled {
				changeT := changeMatrix[t]
				isOutOfBounds := false
				for p := 0; p < pn.Places; p++ {
					tokens := currentMarking[p] + changeT[p]
					if tokens > placeUpperLimit {
						isOutOfBounds = true
						break
					}
					scratchMarking[p] = tokens
				}

				if isOutOfBounds {
					graph.IsBounded = false
					break
				}

				markingHashView := hashMarkingView(scratchMarking)
				if val, ok := visitedMarkings[markingHashView]; !ok {
					// Allocate permanent string only when inserting a new marking
					markingHash := hashMarking(scratchMarking)
					visitedMarkings[markingHash] = graph.NumVertices
					graph.AddEdge([2]int{currentMarkingIndex, graph.NumVertices})
					graph.AddVertex(scratchMarking)
					queue = append(queue, graph.NumVertices-1)
				} else {
					graph.AddEdge([2]int{currentMarkingIndex, val})
				}
				graph.ArcTransitions = append(graph.ArcTransitions, t)
			}
		}
		if !graph.IsBounded {
			break
		}
	}
	return graph, nil
}

// hashMarkingView creates a zero-allocation string view over a slice of ints.
// ⚡ Bolt: It uses unsafe.String directly over the underlying memory, preventing
// allocations during map lookups. Since it does not copy the data, the resulting
// string must only be used as a map lookup key and never stored or modified.
func hashMarkingView(marking []int) string {
	l := len(marking)
	if l == 0 {
		return ""
	}
	byteLen := l * int(unsafe.Sizeof(int(0)))
	return unsafe.String((*byte)(unsafe.Pointer(&marking[0])), byteLen)
}

// hashMarking creates a fast hash key from a marking.
// ⚡ Bolt: Optimized to bypass base-10 conversion and stringification entirely.
// Using unsafe fast-copying directly from the raw underlying memory into a byte array
// provides a massive performance boost when used for cache keys.
func hashMarking(marking []int) string {
	l := len(marking)
	if l == 0 {
		return ""
	}

	byteLen := l * int(unsafe.Sizeof(int(0)))
	b := make([]byte, byteLen)

	// Fast copy memory
	src := unsafe.Slice((*byte)(unsafe.Pointer(&marking[0])), byteLen)
	copy(b, src)

	return unsafe.String(unsafe.SliceData(b), byteLen)
}

// markingToString converts a marking to a string.
// ⚡ Bolt: Optimized to use direct byte slice preallocation and strconv.AppendInt
// to avoid intermediate string allocations and significantly improve performance
// during state hashing.
func markingToString(marking []int) string {
	l := len(marking)
	if l == 0 {
		return "[]"
	}

	// Preallocate with capacity estimation (approx 3 bytes per number)
	b := make([]byte, 0, l*3+1)
	b = append(b, '[')
	for i, v := range marking {
		if i > 0 {
			b = append(b, ' ')
		}
		b = strconv.AppendInt(b, int64(v), 10)
	}
	b = append(b, ']')
	return string(b)
}
