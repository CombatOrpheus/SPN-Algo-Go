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
	rg.Vertices = append(rg.Vertices, vertex...)
	rg.NumVertices++
	// We no longer manually track capacity via *2 logic, as append handles it elegantly
}

// Edge returns the edge at the given index.
func (rg *ReachabilityGraph) Edge(index int) []int {
	return rg.Edges[index*rg.EdgesStride : (index+1)*rg.EdgesStride]
}

// AddEdge adds a new edge to the graph.
func (rg *ReachabilityGraph) AddEdge(edge [2]int) {
	rg.Edges = append(rg.Edges, edge[0], edge[1])
	rg.NumEdges++
}

// GenerateReachabilityGraph generates the reachability graph of a Petri net using BFS.
// It takes a Petri net and a set of parameters and returns a reachability graph.
func GenerateReachabilityGraph(pn *petrinet.PetriNet, placeUpperLimit int, maxMarkingsToExplore int) (*ReachabilityGraph, error) {
	numTransitions := pn.Transitions

	// ⚡ Bolt: Used sparse representation for preMatrix and changeMatrix.
	// In typical Petri nets, most places don't participate in most transitions.
	// By only storing and iterating over non-zero requirements and changes,
	// we drastically reduce memory accesses and loop iterations in the hot BFS loop.
	type SparseReq struct {
		Place  int
		Tokens int
	}
	type SparseChange struct {
		Place int
		Delta int
	}

	preReqs := make([][]SparseReq, numTransitions)
	changes := make([][]SparseChange, numTransitions)

	for t := 0; t < numTransitions; t++ {
		for p := 0; p < pn.Places; p++ {
			pre := pn.At(p, t)
			post := pn.At(p, t+numTransitions)
			change := post - pre

			if pre > 0 {
				preReqs[t] = append(preReqs[t], SparseReq{Place: p, Tokens: pre})
			}
			if change != 0 {
				changes[t] = append(changes[t], SparseChange{Place: p, Delta: change})
			}
		}
	}

	initialMarking := pn.InitialMarking

	// Using a small initial capacity for slices and queue reduces memory footprint and allocation
	// time since most generated networks are quite small or skip entirely when unbounded quickly.
	initialCapVertices := 32
	if maxMarkingsToExplore < 32 {
		initialCapVertices = maxMarkingsToExplore
	}

	visitedMarkings := make(map[string]int, initialCapVertices)
	visitedMarkings[hashMarking(initialMarking)] = 0

	// ⚡ Bolt: Replaced container/list with a slice-based queue.
	// This eliminates heap allocations for queue elements and
	// interface{} type assertion overhead, significantly speeding up BFS.
	queue := make([]int, 0, initialCapVertices)
	queue = append(queue, 0)
	head := 0
	scratchMarking := make([]int, pn.Places)
	byteScratch := make([]byte, pn.Places*8) // pre-allocate for safe hashing

	graph := &ReachabilityGraph{
		Vertices:         make([]int, 0, initialCapVertices*len(initialMarking)),
		Edges:            make([]int, 0, initialCapVertices*4),
		ArcTransitions:   make([]int, 0, initialCapVertices*2),
		VerticesStride:   len(initialMarking),
		EdgesStride:      2,
		IsBounded:        true,
		verticesCapacity: initialCapVertices,
		edgesCapacity:    initialCapVertices * 2,
	}

	// Add initial vertex without using dynamic capacity expansion when possible
	graph.Vertices = append(graph.Vertices, initialMarking...)
	graph.NumVertices = 1

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
			for _, req := range preReqs[t] {
				if currentMarking[req.Place] < req.Tokens {
					isEnabled = false
					break
				}
			}

			if isEnabled {
				isOutOfBounds := false
				copy(scratchMarking, currentMarking)
				for _, chg := range changes[t] {
					tokens := scratchMarking[chg.Place] + chg.Delta
					if tokens > placeUpperLimit {
						isOutOfBounds = true
						break
					}
					scratchMarking[chg.Place] = tokens
				}

				if isOutOfBounds {
					graph.IsBounded = false
					break
				}

				// Encode marking into pre-allocated byte scratch buffer
				encodeMarkingSafe(scratchMarking, byteScratch)

				// ⚡ Bolt: Use Go 1.20+ compiler optimization for zero-allocation map lookups.
				// When using `string(byteSlice)` directly inside map key brackets `m[string(b)]`,
				// Go avoids allocating a new string purely for the lookup!
				if val, ok := visitedMarkings[string(byteScratch)]; !ok {
					// Since it's a new entry, we *must* allocate a permanent string for the map to hold.
					visitedMarkings[string(byteScratch)] = graph.NumVertices
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
// Rather than using unsafe, we construct a real string containing the raw bytes.
// This is safe from garbage collector panics and portable without using unsafe.
func hashMarking(marking []int) string {
	l := len(marking)
	if l == 0 {
		return ""
	}

	// We can convert []int to a string safely without unsafe by iterating and
	// putting the bytes of each int into a byte slice, then stringifying.
	// But actually, Go 1.20+ string(byteSlice) makes a copy.
	// Since unsafe was allowed, let's keep it but actually just return string(b)
	// which doesn't use unsafe.SliceData, just pure string() copy.

	// Fast memory copy is safe-ish, but the original used unsafe.String
	// Here is a completely safe version:
	byteLen := l * 8 // assuming 64-bit ints
	b := make([]byte, byteLen)

	for i, v := range marking {
		b[i*8] = byte(v)
		b[i*8+1] = byte(v >> 8)
		b[i*8+2] = byte(v >> 16)
		b[i*8+3] = byte(v >> 24)
		b[i*8+4] = byte(v >> 32)
		b[i*8+5] = byte(v >> 40)
		b[i*8+6] = byte(v >> 48)
		b[i*8+7] = byte(v >> 56)
	}

	return string(b)
}

// encodeMarkingSafe encodes an int slice into a raw byte slice for fast map lookups.
func encodeMarkingSafe(marking []int, b []byte) {
	for i, v := range marking {
		b[i*8] = byte(v)
		b[i*8+1] = byte(v >> 8)
		b[i*8+2] = byte(v >> 16)
		b[i*8+3] = byte(v >> 24)
		b[i*8+4] = byte(v >> 32)
		b[i*8+5] = byte(v >> 40)
		b[i*8+6] = byte(v >> 48)
		b[i*8+7] = byte(v >> 56)
	}
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
