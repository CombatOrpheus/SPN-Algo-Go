package main

import (
	"math/rand"
	"testing"
)

func BenchmarkNestedLoop(b *testing.B) {
	numPlaces := 10
	numVertices := 1000
	maxTokens := 50

	vertices := make([][]int, numVertices)
	for i := range vertices {
		vertices[i] = make([]int, numPlaces)
		for j := range vertices[i] {
			vertices[i][j] = rand.Intn(maxTokens + 1)
		}
	}
	steadyStateProbs := make([]float64, numVertices)
	for i := range steadyStateProbs {
		steadyStateProbs[i] = rand.Float64()
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		markingDensityMatrix := make([][]float64, numPlaces)
		for i := range markingDensityMatrix {
			markingDensityMatrix[i] = make([]float64, maxTokens+1)
		}

		for placeIdx := 0; placeIdx < numPlaces; placeIdx++ {
			for tokenVal := 0; tokenVal <= maxTokens; tokenVal++ {
				sumProbs := 0.0
				for i := 0; i < numVertices; i++ {
					vertex := vertices[i]
					if vertex[placeIdx] == tokenVal {
						sumProbs += steadyStateProbs[i]
					}
				}
				markingDensityMatrix[placeIdx][tokenVal] = sumProbs
			}
		}
	}
}

func BenchmarkOptimizedLoop(b *testing.B) {
	numPlaces := 10
	numVertices := 1000
	maxTokens := 50

	vertices := make([][]int, numVertices)
	for i := range vertices {
		vertices[i] = make([]int, numPlaces)
		for j := range vertices[i] {
			vertices[i][j] = rand.Intn(maxTokens + 1)
		}
	}
	steadyStateProbs := make([]float64, numVertices)
	for i := range steadyStateProbs {
		steadyStateProbs[i] = rand.Float64()
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		markingDensityMatrix := make([][]float64, numPlaces)
		for i := range markingDensityMatrix {
			markingDensityMatrix[i] = make([]float64, maxTokens+1)
		}

		for i := 0; i < numVertices; i++ {
			vertex := vertices[i]
			prob := steadyStateProbs[i]
			for placeIdx := 0; placeIdx < numPlaces; placeIdx++ {
				tokenVal := vertex[placeIdx]
				markingDensityMatrix[placeIdx][tokenVal] += prob
			}
		}
	}
}
