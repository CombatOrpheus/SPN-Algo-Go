package benchmarks

import (
	"math/rand"
	"spn-benchmark-ds/internal/pkg/analysis"
	"spn-benchmark-ds/internal/pkg/augmentation"
	"spn-benchmark-ds/internal/pkg/generation"
	"spn-benchmark-ds/internal/pkg/petrinet"
	"testing"
)

func BenchmarkGeneration_Small(b *testing.B) {
	pn := generateTestPetriNet(5, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generation.GenerateReachabilityGraph(pn, 10, 1000)
	}
}

func BenchmarkGeneration_Medium(b *testing.B) {
	pn := generateTestPetriNet(20, 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generation.GenerateReachabilityGraph(pn, 10, 1000)
	}
}

func BenchmarkGeneration_Large(b *testing.B) {
	pn := generateTestPetriNet(50, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generation.GenerateReachabilityGraph(pn, 10, 10000)
	}
}

func BenchmarkAnalysis_Small(b *testing.B) {
	pn := generateTestPetriNet(5, 5)
	rg := generateTestReachabilityGraph(pn)
	lambdas := generateRandomLambdaValues(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stateMatrix, targetVector := analysis.ComputeStateEquation(rg, lambdas)
		steadyStateProbs, err := analysis.SolveForSteadyState(stateMatrix, targetVector)
		if err != nil {
			continue // Skip singular matrices which can happen with synthetic nets
		}
		_, _ = analysis.ComputeAverageMarkings(rg, steadyStateProbs)
	}
}

func BenchmarkAnalysis_Medium(b *testing.B) {
	pn := generateTestPetriNet(20, 20)
	rg := generateTestReachabilityGraph(pn)
	lambdas := generateRandomLambdaValues(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stateMatrix, targetVector := analysis.ComputeStateEquation(rg, lambdas)
		steadyStateProbs, err := analysis.SolveForSteadyState(stateMatrix, targetVector)
		if err != nil {
			continue // Skip singular matrices which can happen with synthetic nets
		}
		_, _ = analysis.ComputeAverageMarkings(rg, steadyStateProbs)
	}
}

func BenchmarkAugmentation_PetriNet_Small(b *testing.B) {
	pn := generateTestPetriNet(5, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = augmentation.GeneratePetriNetVariations(pn, 10, 1, 1000, 5, 1, 10)
	}
}

func BenchmarkAugmentation_PetriNet_Medium(b *testing.B) {
	pn := generateTestPetriNet(15, 15) // Slightly smaller than medium to keep tests fast

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = augmentation.GeneratePetriNetVariations(pn, 10, 1, 1000, 5, 1, 10)
	}
}

func BenchmarkAugmentation_Lambda_Medium(b *testing.B) {
	pn := generateTestPetriNet(20, 20)
	rg := generateTestReachabilityGraph(pn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = augmentation.GenerateLambdaVariations(pn, rg, 5, 1, 10)
	}
}

func BenchmarkWholeProgram_Pipeline(b *testing.B) {
	// Re-implements the core inner loop of runRandomGeneration
	// to avoid I/O bottlenecks and benchmark pure execution time
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Initial setup per sample
		pn := petrinet.GenerateRandomPetriNet(10, 10)
		pn.Prune()
		pn.AddTokensRandomly()
		b.StartTimer()

		rg, err := generation.GenerateReachabilityGraph(pn, 10, 1000)
		if err != nil || !rg.IsBounded || rg.NumVertices < 1 {
			continue
		}

		lambdaValues := make([]float64, pn.Transitions)
		for j := range lambdaValues {
			lambdaValues[j] = float64(1 + rand.Intn(10))
		}

		stateMatrix, targetVector := analysis.ComputeStateEquation(rg, lambdaValues)
		steadyStateProbs, err := analysis.SolveForSteadyState(stateMatrix, targetVector)
		if err != nil {
			continue
		}

		_, _ = analysis.ComputeAverageMarkings(rg, steadyStateProbs)

		// Include simple transformation step like runRandomGeneration when EnableTransformations=true
		_ = augmentation.GeneratePetriNetVariations(pn, 10, 1, 1000, 3, 1, 10)
	}
}
