package petri

import "gonum.org/v1/gonum/mat"

// PetriNet represents a Petri net, including its structure and initial marking.
type PetriNet struct {
	// Places is the number of places in the Petri net.
	Places int
	// Transitions is the number of transitions in the Petri net.
	Transitions int
	// Matrix is the incidence matrix of the Petri net, with dimensions
	// (Places, 2*Transitions+1). The last column represents the initial marking.
	Matrix *mat.Dense
}
