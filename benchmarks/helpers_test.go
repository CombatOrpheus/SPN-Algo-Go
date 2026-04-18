package benchmarks

import (
	"math/rand"
	"spn-benchmark-ds/internal/pkg/generation"
	"spn-benchmark-ds/internal/pkg/petrinet"
)

// generateTestPetriNet creates a predictable PetriNet for benchmarking.
// It avoids using random generation inside the benchmark itself.
func generateTestPetriNet(places, transitions int) *petrinet.PetriNet {
	pn := petrinet.NewPetriNet(places, transitions)

	// Create a simple chain/mesh structure to ensure it's connected
	// and bounded.
	for t := 0; t < transitions; t++ {
		prePlace := t % places
		postPlace := (t + 1) % places

		pn.Set(prePlace, t, 1)              // Pre-arc
		pn.Set(postPlace, t+transitions, 1) // Post-arc
	}

	// Set initial marking
	for p := 0; p < places; p++ {
		if p%2 == 0 {
			pn.Set(p, 2*transitions, 1) // Initial marking
			pn.InitialMarking[p] = 1
		}
	}

	return pn
}

// generateTestReachabilityGraph generates a reachability graph for benchmarking filtering.
func generateTestReachabilityGraph(pn *petrinet.PetriNet) *generation.ReachabilityGraph {
	rg, _ := generation.GenerateReachabilityGraph(pn, 10, 1000)
	return rg
}

// generateRandomLambdaValues generates predictable lambda values.
func generateRandomLambdaValues(transitions int) []float64 {
	rand.Seed(42) // Fixed seed for reproducibility
	lambdas := make([]float64, transitions)
	for i := 0; i < transitions; i++ {
		lambdas[i] = float64(1 + rand.Intn(10))
	}
	return lambdas
}
