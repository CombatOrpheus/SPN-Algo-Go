package petri

import "gonum.org/v1/gonum/mat"

// StochasticPetriNet represents the results of a stochastic analysis of a Petri net.
type StochasticPetriNet struct {
	// PetriNet is the underlying Petri net structure.
	PetriNet *PetriNet
	// ReachabilityGraph is the set of all reachable states from the initial marking.
	// Each row in the matrix represents a state.
	ReachabilityGraph *mat.Dense
	// Edges defines the transitions between states in the reachability graph.
	// It is a slice of pairs, where each pair [from, to] represents an edge.
	Edges [][2]int
	// ArcTransitions maps each edge in the reachability graph to a transition
	// in the Petri net.
	ArcTransitions []int
	// FiringRates are the rates of the transitions.
	FiringRates []float64
	// SteadyStateProbs are the steady-state probabilities for each state in the
	// reachability graph.
	SteadyStateProbs []float64
	// MarkingDensities is a matrix where each row corresponds to a place and each
	// column corresponds to a possible token count. The value at (i, j) is the
	// probability of place i having j tokens.
	MarkingDensities *mat.Dense
	// AverageMarkings is a vector containing the average number of tokens for
	// each place.
	AverageMarkings []float64
	// TotalAverageMarking is the sum of the average markings over all places.
	TotalAverageMarking float64
}
