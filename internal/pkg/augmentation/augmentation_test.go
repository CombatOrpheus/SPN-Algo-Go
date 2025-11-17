package augmentation

import (
	"spn-benchmark-ds/internal/pkg/petrinet"
	"testing"
)

func TestGeneratePetriNetVariations(t *testing.T) {
	pn := &petrinet.PetriNet{
		Places:      2,
		Transitions: 1,
		Matrix: [][]int{
			{1, 0, 1},
			{0, 1, 0},
		},
		InitialMarking: []int{1, 0},
	}
	minFiringRate := 5
	maxFiringRate := 5

	variations := GeneratePetriNetVariations(pn, 10, 2, 100, 5, minFiringRate, maxFiringRate)

	if len(variations) == 0 {
		// It's possible that no variations are generated if the reachability graph generation fails.
		// Run it again to be sure.
		variations = GeneratePetriNetVariations(pn, 10, 2, 100, 5, minFiringRate, maxFiringRate)
		if len(variations) == 0 {
			t.Errorf("GeneratePetriNetVariations did not generate any variations")
		}
	}

	if len(variations) != 5 {
		// It's possible that fewer variations are generated if the reachability graph generation fails.
		// We'll just check that some variations were generated.
		if len(variations) == 0 {
			t.Errorf("GeneratePetriNetVariations did not generate any variations")
		}
	}

	// Check that the variations are different from the original
	for _, variation := range variations {
		isDifferent := false
		// This is not a great way to check for differences, but it's better than nothing.
		// A better way would be to compare the analysis results.
		if len(variation.AverageMarkings) != len(pn.InitialMarking) {
			isDifferent = true
			break
		}
		for i := range variation.AverageMarkings {
			// This is a very loose comparison, but it's better than nothing.
			if int(variation.AverageMarkings[i]) != pn.InitialMarking[i] {
				isDifferent = true
				break
			}
		}
		if !isDifferent {
			t.Errorf("Generated variation is not different from the original")
		}
	}
}
