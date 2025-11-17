package analysis

import (
	"fmt"
	"spn-benchmark-ds/internal/pkg/generation"

	"gonum.org/v1/gonum/mat"
)

// SPNAnalysisResult holds the results of the SPN analysis.
type SPNAnalysisResult struct {
	SteadyStateProbs []float64
	AverageMarkings  []float64
	MarkingDensities [][]float64
}

// ComputeStateEquation computes the state equation for the SPN.
func ComputeStateEquation(rg *generation.ReachabilityGraph, lambdaValues []float64) (*mat.Dense, *mat.VecDense) {
	numVertices := len(rg.Vertices)
	data := make([]float64, (numVertices+1)*numVertices)

	for i, edge := range rg.Edges {
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
func ComputeAverageMarkings(vertices [][]int, steadyStateProbs []float64) ([]float64, [][]float64) {
	if len(vertices) == 0 {
		return []float64{}, [][]float64{}
	}
	numPlaces := len(vertices[0])
	avgTokensPerPlace := make([]float64, numPlaces)

	for i, vertex := range vertices {
		for p := 0; p < numPlaces; p++ {
			avgTokensPerPlace[p] += float64(vertex[p]) * steadyStateProbs[i]
		}
	}

	maxTokens := 0
	for _, vertex := range vertices {
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
			for i, vertex := range vertices {
				if vertex[placeIdx] == tokenVal {
					sumProbs += steadyStateProbs[i]
				}
			}
			markingDensityMatrix[placeIdx][tokenVal] = sumProbs
		}
	}
	return avgTokensPerPlace, markingDensityMatrix
}
