package generation

import (
	"container/list"
	"fmt"
	"spn-benchmark-ds/internal/pkg/petrinet"
)

// ReachabilityGraph represents the reachability graph of a Petri Net.
type ReachabilityGraph struct {
	Vertices       [][]int
	Edges          [][2]int
	ArcTransitions []int
	IsBounded      bool
}

// GenerateReachabilityGraph generates the reachability graph of a Petri net using BFS.
func GenerateReachabilityGraph(pn *petrinet.PetriNet, placeUpperLimit int, maxMarkingsToExplore int) (*ReachabilityGraph, error) {
	numTransitions := pn.Transitions
	preMatrix := make([][]int, pn.Places)
	postMatrix := make([][]int, pn.Places)
	changeMatrix := make([][]int, pn.Places)
	for i := 0; i < pn.Places; i++ {
		preMatrix[i] = make([]int, numTransitions)
		postMatrix[i] = make([]int, numTransitions)
		changeMatrix[i] = make([]int, numTransitions)
		for j := 0; j < numTransitions; j++ {
			preMatrix[i][j] = pn.Matrix[i][j]
			postMatrix[i][j] = pn.Matrix[i][j+numTransitions]
			changeMatrix[i][j] = postMatrix[i][j] - preMatrix[i][j]
		}
	}

	initialMarking := pn.InitialMarking

	visitedMarkings := make(map[string]int)
	visitedMarkings[markingToString(initialMarking)] = 0

	queue := list.New()
	queue.PushBack(0)

	graph := &ReachabilityGraph{
		Vertices:  [][]int{initialMarking},
		IsBounded: true,
	}

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		currentMarkingIndex := element.Value.(int)
		currentMarking := graph.Vertices[currentMarkingIndex]

		if len(graph.Vertices) >= maxMarkingsToExplore {
			graph.IsBounded = false
			break
		}

		enabledTransitions, newMarkings := getEnabledTransitions(preMatrix, changeMatrix, currentMarking)

		for i, newMarking := range newMarkings {
			if isMarkingOutOfBounds(newMarking, placeUpperLimit) {
				graph.IsBounded = false
				break
			}

			markingStr := markingToString(newMarking)
			if _, ok := visitedMarkings[markingStr]; !ok {
				visitedMarkings[markingStr] = len(graph.Vertices)
				graph.Vertices = append(graph.Vertices, newMarking)
				queue.PushBack(len(graph.Vertices) - 1)
			}
			graph.Edges = append(graph.Edges, [2]int{currentMarkingIndex, visitedMarkings[markingStr]})
			graph.ArcTransitions = append(graph.ArcTransitions, enabledTransitions[i])
		}
		if !graph.IsBounded {
			break
		}
	}
	return graph, nil
}

func getEnabledTransitions(preMatrix, changeMatrix [][]int, currentMarking []int) ([]int, [][]int) {
	numTransitions := len(preMatrix[0])
	var enabledTransitions []int
	var newMarkings [][]int

	for t := 0; t < numTransitions; t++ {
		isEnabled := true
		for p := 0; p < len(preMatrix); p++ {
			if currentMarking[p] < preMatrix[p][t] {
				isEnabled = false
				break
			}
		}
		if isEnabled {
			newMarking := make([]int, len(currentMarking))
			for p := 0; p < len(currentMarking); p++ {
				newMarking[p] = currentMarking[p] + changeMatrix[p][t]
			}
			enabledTransitions = append(enabledTransitions, t)
			newMarkings = append(newMarkings, newMarking)
		}
	}
	return enabledTransitions, newMarkings
}

func isMarkingOutOfBounds(marking []int, placeUpperLimit int) bool {
	for _, tokens := range marking {
		if tokens > placeUpperLimit {
			return true
		}
	}
	return false
}

func markingToString(marking []int) string {
	return fmt.Sprintf("%v", marking)
}
