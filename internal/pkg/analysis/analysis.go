package analysis

import (
	"fmt"
	"spn-benchmark-ds/internal/pkg/generation"

	"gonum.org/v1/gonum/mat"
)

// SPNAnalysisResult holds the results of the SPN analysis.
type SPNAnalysisResult struct {
	// SteadyStateProbs is a slice of steady-state probabilities for each marking.
	SteadyStateProbs []float64
	// AverageMarkings is a slice of average markings for each place.
	AverageMarkings []float64
	// MarkingDensities is a slice of marking densities for each place.
	MarkingDensities [][]float64
}

// ComputeStateEquation computes the state equation for the SPN.
// It takes a reachability graph and a slice of lambda values and returns a state matrix and a target vector.
func ComputeStateEquation(rg *generation.ReachabilityGraph, lambdaValues []float64) (*mat.Dense, *mat.VecDense) {
	numVertices := rg.NumVertices
	data := make([]float64, (numVertices+1)*numVertices)

	for i := 0; i < rg.NumEdges; i++ {
		edge := rg.Edge(i)
		srcIdx, destIdx := edge[0], edge[1]
		transIdx := rg.ArcTransitions[i]
		rate := lambdaValues[transIdx]
		data[srcIdx*numVertices+srcIdx] -= rate
		data[destIdx*numVertices+srcIdx] += rate
	}

	for i := 0; i < numVertices; i++ {
		data[numVertices*numVertices+i] = 1.0
	}

	stateMatrix := mat.NewDense(numVertices+1, numVertices, data)
	targetVector := mat.NewVecDense(numVertices+1, nil)
	targetVector.SetVec(numVertices, 1.0)

	return stateMatrix, targetVector
}

// SolveForSteadyState solves for steady-state probabilities.
// It takes a state matrix and a target vector and returns a slice of steady-state probabilities.
func SolveForSteadyState(stateMatrix *mat.Dense, targetVector *mat.VecDense) ([]float64, error) {
	_, numVertices := stateMatrix.Dims()

	// We need a square matrix for the solver. Remove one redundant equation.
	data := make([]float64, numVertices*numVertices)
	for r := 1; r < numVertices+1; r++ {
		for c := 0; c < numVertices; c++ {
			data[(r-1)*numVertices+c] = stateMatrix.At(r, c)
		}
	}
	A := mat.NewDense(numVertices, numVertices, data)

	bData := make([]float64, numVertices)
	for i := 0; i < numVertices; i++ {
		bData[i] = targetVector.AtVec(i + 1)
	}
	b := mat.NewVecDense(numVertices, bData)

	var x mat.VecDense
	if err := x.SolveVec(A, b); err != nil {
		return nil, fmt.Errorf("failed to solve linear system: %v", err)
	}

	probs := x.RawVector().Data
	probSum := 0.0
	for _, p := range probs {
		if p < 0 {
			p = 0
		}
		probSum += p
	}
	if probSum > 1e-9 {
		for i := range probs {
			probs[i] /= probSum
		}
	}
	return probs, nil
}

// ComputeAverageMarkings calculates the average number of tokens for each place.
// It takes a reachability graph and a slice of steady-state probabilities and returns a slice of average markings and a slice of marking densities.
func ComputeAverageMarkings(rg *generation.ReachabilityGraph, steadyStateProbs []float64) ([]float64, [][]float64) {
	if rg.NumVertices == 0 {
		return []float64{}, [][]float64{}
	}
	numPlaces := rg.VerticesStride
	avgTokensPerPlace := make([]float64, numPlaces)

	for i := 0; i < rg.NumVertices; i++ {
		vertex := rg.Vertex(i)
		for p := 0; p < numPlaces; p++ {
			avgTokensPerPlace[p] += float64(vertex[p]) * steadyStateProbs[i]
		}
	}

	maxTokens := 0
	for i := 0; i < rg.NumVertices; i++ {
		vertex := rg.Vertex(i)
		for _, tokens := range vertex {
			if tokens > maxTokens {
				maxTokens = tokens
			}
		}
	}

	markingDensityMatrix := make([][]float64, numPlaces)
	for i := range markingDensityMatrix {
		markingDensityMatrix[i] = make([]float64, maxTokens+1)
	}

	for placeIdx := 0; placeIdx < numPlaces; placeIdx++ {
		for tokenVal := 0; tokenVal <= maxTokens; tokenVal++ {
			sumProbs := 0.0
			for i := 0; i < rg.NumVertices; i++ {
				vertex := rg.Vertex(i)
				if vertex[placeIdx] == tokenVal {
					sumProbs += steadyStateProbs[i]
				}
			}
			markingDensityMatrix[placeIdx][tokenVal] = sumProbs
		}
	}
	return avgTokensPerPlace, markingDensityMatrix
}
