package petrinet

import (
	"math/rand"
	"time"
)

// PetriNet represents a Petri Net.
type PetriNet struct {
	// Places is the number of places in the Petri net.
	Places int
	// Transitions is the number of transitions in the Petri net.
	Transitions int
	// Matrix is the flattened incidence matrix of the Petri net.
	Matrix []int
	// stride is the stride of the matrix.
	stride int
	// InitialMarking is the initial marking of the Petri net.
	InitialMarking []int
}

// At returns the value of the matrix at the given row and column.
func (pn *PetriNet) At(row, col int) int {
	return pn.Matrix[row*pn.stride+col]
}

// Set sets the value of the matrix at the given row and column.
func (pn *PetriNet) Set(row, col, value int) {
	pn.Matrix[row*pn.stride+col] = value
}

// NewPetriNet creates a new PetriNet.
func NewPetriNet(places, transitions int) *PetriNet {
	stride := 2*transitions + 1
	matrix := make([]int, places*stride)
	return &PetriNet{
		Places:         places,
		Transitions:    transitions,
		Matrix:         matrix,
		stride:         stride,
		InitialMarking: make([]int, places),
	}
}

// GenerateRandomPetriNet generates a random Petri net matrix.
// It takes the number of places and transitions and returns a new Petri net.
func GenerateRandomPetriNet(numPlaces, numTransitions int) *PetriNet {
	rand.Seed(time.Now().UnixNano())
	pn := NewPetriNet(numPlaces, numTransitions)

	remainingNodes := make([]int, numPlaces+numTransitions)
	for i := 0; i < numPlaces+numTransitions; i++ {
		remainingNodes[i] = i + 1
	}

	firstPlace := rand.Intn(numPlaces) + 1
	firstTransition := rand.Intn(numTransitions) + numPlaces + 1

	removeNode(remainingNodes, firstPlace)
	removeNode(remainingNodes, firstTransition)

	if rand.Float64() <= 0.5 {
		pn.Set(firstPlace-1, firstTransition-numPlaces-1, 1)
	} else {
		pn.Set(firstPlace-1, firstTransition-numPlaces-1+numTransitions, 1)
	}

	subGraph := []int{firstPlace, firstTransition}
	rand.Shuffle(len(remainingNodes), func(i, j int) {
		remainingNodes[i], remainingNodes[j] = remainingNodes[j], remainingNodes[i]
	})

	for _, node := range remainingNodes {
		subPlaces := filter(subGraph, func(n int) bool { return n <= numPlaces })
		subTransitions := filter(subGraph, func(n int) bool { return n > numPlaces })

		var place, transition int
		if node <= numPlaces {
			place = node
			transition = subTransitions[rand.Intn(len(subTransitions))]
		} else {
			place = subPlaces[rand.Intn(len(subPlaces))]
			transition = node
		}

		if rand.Float64() <= 0.5 {
			pn.Set(place-1, transition-numPlaces-1, 1)
		} else {
			pn.Set(place-1, transition-numPlaces-1+numTransitions, 1)
		}
		subGraph = append(subGraph, node)
	}

	randomPlace := rand.Intn(numPlaces)
	pn.Set(randomPlace, 2*numTransitions, 1)
	pn.InitialMarking = make([]int, numPlaces)
	for i := 0; i < numPlaces; i++ {
		pn.InitialMarking[i] = pn.At(i, 2*numTransitions)
	}

	return pn
}

// removeNode removes a node from a slice of nodes.
func removeNode(nodes []int, node int) []int {
	for i, n := range nodes {
		if n == node {
			return append(nodes[:i], nodes[i+1:]...)
		}
	}
	return nodes
}

// filter filters a slice of nodes based on a condition.
func filter(nodes []int, condition func(int) bool) []int {
	var result []int
	for _, n := range nodes {
		if condition(n) {
			result = append(result, n)
		}
	}
	return result
}

// Prune prunes the Petri net by deleting excess edges and adding missing connections.
func (pn *PetriNet) Prune() {
	pn.deleteExcessEdges()
	pn.addMissingConnections()
}

// deleteExcessEdges deletes excess edges from the Petri net.
func (pn *PetriNet) deleteExcessEdges() {
	// Delete excess edges from places
	for i := 0; i < pn.Places; i++ {
		rowSum := 0
		for j := 0; j < 2*pn.Transitions; j++ {
			rowSum += pn.At(i, j)
		}
		if rowSum >= 3 {
			var edgeIndices []int
			for j := 0; j < 2*pn.Transitions; j++ {
				if pn.At(i, j) == 1 {
					edgeIndices = append(edgeIndices, j)
				}
			}
			rand.Shuffle(len(edgeIndices), func(k, l int) {
				edgeIndices[k], edgeIndices[l] = edgeIndices[l], edgeIndices[k]
			})
			for k := 0; k < len(edgeIndices)-2; k++ {
				// Only remove the edge if it doesn't disconnect the graph
				pn.Set(i, edgeIndices[k], 0)
				if !pn.isConnected() {
					pn.Set(i, edgeIndices[k], 1)
				}
			}
		}
	}

	// Delete excess edges from transitions
	for j := 0; j < 2*pn.Transitions; j++ {
		colSum := 0
		for i := 0; i < pn.Places; i++ {
			colSum += pn.At(i, j)
		}
		if colSum >= 3 {
			var edgeIndices []int
			for i := 0; i < pn.Places; i++ {
				if pn.At(i, j) == 1 {
					edgeIndices = append(edgeIndices, i)
				}
			}
			rand.Shuffle(len(edgeIndices), func(k, l int) {
				edgeIndices[k], edgeIndices[l] = edgeIndices[l], edgeIndices[k]
			})
			for k := 0; k < len(edgeIndices)-2; k++ {
				// Only remove the edge if it doesn't disconnect the graph
				pn.Set(edgeIndices[k], j, 0)
				if !pn.isConnected() {
					pn.Set(edgeIndices[k], j, 1)
				}
			}
		}
	}
}

// isConnected returns true if the Petri net is connected.
func (pn *PetriNet) isConnected() bool {
	// Check for isolated places
	for i := 0; i < pn.Places; i++ {
		rowSum := 0
		for j := 0; j < 2*pn.Transitions; j++ {
			rowSum += pn.At(i, j)
		}
		if rowSum == 0 {
			return false
		}
	}

	// Check for isolated transitions
	for j := 0; j < 2*pn.Transitions; j++ {
		colSum := 0
		for i := 0; i < pn.Places; i++ {
			colSum += pn.At(i, j)
		}
		if colSum == 0 {
			return false
		}
	}

	return true
}

// addMissingConnections adds missing connections to the Petri net.
func (pn *PetriNet) addMissingConnections() {
	// Ensure each transition has at least one connection
	for j := 0; j < 2*pn.Transitions; j++ {
		colSum := 0
		for i := 0; i < pn.Places; i++ {
			colSum += pn.At(i, j)
		}
		if colSum == 0 {
			randomRow := rand.Intn(pn.Places)
			pn.Set(randomRow, j, 1)
		}
	}

	// Ensure each place has at least one incoming and one outgoing edge
	for i := 0; i < pn.Places; i++ {
		preSum := 0
		postSum := 0
		for j := 0; j < pn.Transitions; j++ {
			preSum += pn.At(i, j)
			postSum += pn.At(i, j+pn.Transitions)
		}
		if preSum == 0 {
			randomCol := rand.Intn(pn.Transitions)
			pn.Set(i, randomCol, 1)
		}
		if postSum == 0 {
			randomCol := rand.Intn(pn.Transitions) + pn.Transitions
			pn.Set(i, randomCol, 1)
		}
	}
}

// AddTokensRandomly adds tokens to random places in the Petri net.
func (pn *PetriNet) AddTokensRandomly() {
	for i := 0; i < pn.Places; i++ {
		if rand.Intn(10) <= 2 {
			pn.Set(i, 2*pn.Transitions, pn.At(i, 2*pn.Transitions)+1)
		}
	}
	pn.updateInitialMarking()
}

// updateInitialMarking updates the initial marking of the Petri net.
func (pn *PetriNet) updateInitialMarking() {
	for i := 0; i < pn.Places; i++ {
		pn.InitialMarking[i] = pn.At(i, 2*pn.Transitions)
	}
}
