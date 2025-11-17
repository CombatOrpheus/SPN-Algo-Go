package petrinet

import (
	"math/rand"
	"time"
)

// PetriNet represents a Petri Net.
type PetriNet struct {
	Places         int
	Transitions    int
	Matrix         [][]int
	InitialMarking []int
}

// NewPetriNet creates a new PetriNet.
func NewPetriNet(places, transitions int) *PetriNet {
	matrix := make([][]int, places)
	for i := range matrix {
		matrix[i] = make([]int, 2*transitions+1)
	}
	return &PetriNet{
		Places:      places,
		Transitions: transitions,
		Matrix:      matrix,
	}
}

// GenerateRandomPetriNet generates a random Petri net matrix.
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
		pn.Matrix[firstPlace-1][firstTransition-numPlaces-1] = 1
	} else {
		pn.Matrix[firstPlace-1][firstTransition-numPlaces-1+numTransitions] = 1
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
			pn.Matrix[place-1][transition-numPlaces-1] = 1
		} else {
			pn.Matrix[place-1][transition-numPlaces-1+numTransitions] = 1
		}
		subGraph = append(subGraph, node)
	}

	randomPlace := rand.Intn(numPlaces)
	pn.Matrix[randomPlace][2*numTransitions] = 1
	pn.InitialMarking = make([]int, numPlaces)
	for i := 0; i < numPlaces; i++ {
		pn.InitialMarking[i] = pn.Matrix[i][2*numTransitions]
	}

	return pn
}

func removeNode(nodes []int, node int) []int {
	for i, n := range nodes {
		if n == node {
			return append(nodes[:i], nodes[i+1:]...)
		}
	}
	return nodes
}

func filter(nodes []int, condition func(int) bool) []int {
	var result []int
	for _, n := range nodes {
		if condition(n) {
			result = append(result, n)
		}
	}
	return result
}
