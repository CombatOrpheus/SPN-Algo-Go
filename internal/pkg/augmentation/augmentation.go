package augmentation

import (
	"math/rand"
	"spn-benchmark-ds/internal/pkg/petrinet"
	"spn-benchmark-ds/internal/pkg/analysis"
	"spn-benchmark-ds/internal/pkg/generation"
)

// GeneratePetriNetVariations generates variations of a Petri net by adding or removing tokens.
// It takes a Petri net and a set of parameters and returns a slice of SPN analysis results.
func GeneratePetriNetVariations(pn *petrinet.PetriNet, placeUpperBound, marksLowerLimit, marksUpperLimit, numVariations, minFiringRate, maxFiringRate int) []*analysis.SPNAnalysisResult {
	var variations []*analysis.SPNAnalysisResult

	for i := 0; i < numVariations; i++ {
		variationPN := deepCopyPetriNet(pn)
		if rand.Float64() < 0.5 {
			// Add tokens
			place := rand.Intn(variationPN.Places)
			if variationPN.InitialMarking[place] < placeUpperBound {
				variationPN.InitialMarking[place]++
			}
		} else {
			// Remove tokens
			place := rand.Intn(variationPN.Places)
			if variationPN.InitialMarking[place] > 0 {
				variationPN.InitialMarking[place]--
			}
		}

		rg, err := generation.GenerateReachabilityGraph(variationPN, placeUpperBound, marksUpperLimit)
		if err != nil {
			continue
		}

		if !rg.IsBounded || rg.NumVertices < marksLowerLimit {
			continue
		}

		lambdaValues := make([]float64, variationPN.Transitions)
		for i := range lambdaValues {
			lambdaValues[i] = float64(minFiringRate + rand.Intn(maxFiringRate-minFiringRate+1))
		}

		stateMatrix, targetVector := analysis.ComputeStateEquation(rg, lambdaValues)
		steadyStateProbs, err := analysis.SolveForSteadyState(stateMatrix, targetVector)
		if err != nil {
			continue
		}

		avgMarkings, markingDensities := analysis.ComputeAverageMarkings(rg, steadyStateProbs)
		variations = append(variations, &analysis.SPNAnalysisResult{
			SteadyStateProbs: steadyStateProbs,
			AverageMarkings:  avgMarkings,
			MarkingDensities: markingDensities,
		})
	}

	return variations
}

// deepCopyPetriNet creates a deep copy of a Petri net.
func deepCopyPetriNet(pn *petrinet.PetriNet) *petrinet.PetriNet {
	newPN := petrinet.NewPetriNet(pn.Places, pn.Transitions)
	copy(newPN.Matrix, pn.Matrix)
	copy(newPN.InitialMarking, pn.InitialMarking)
	return newPN
}

// GenerateLambdaVariations generates variations of a Petri net by changing the lambda values.
func GenerateLambdaVariations(pn *petrinet.PetriNet, rg *generation.ReachabilityGraph, numVariations, minFiringRate, maxFiringRate int) ([]*analysis.SPNAnalysisResult, [][]float64) {
	var variations []*analysis.SPNAnalysisResult
	var lambdaValuesList [][]float64

	for i := 0; i < numVariations; i++ {
		lambdaValues := make([]float64, pn.Transitions)
		for i := range lambdaValues {
			lambdaValues[i] = float64(minFiringRate + rand.Intn(maxFiringRate-minFiringRate+1))
		}

		stateMatrix, targetVector := analysis.ComputeStateEquation(rg, lambdaValues)
		steadyStateProbs, err := analysis.SolveForSteadyState(stateMatrix, targetVector)
		if err != nil {
			continue
		}

		avgMarkings, markingDensities := analysis.ComputeAverageMarkings(rg, steadyStateProbs)
		variations = append(variations, &analysis.SPNAnalysisResult{
			SteadyStateProbs: steadyStateProbs,
			AverageMarkings:  avgMarkings,
			MarkingDensities: markingDensities,
		})
		lambdaValuesList = append(lambdaValuesList, lambdaValues)
	}

	return variations, lambdaValuesList
}
