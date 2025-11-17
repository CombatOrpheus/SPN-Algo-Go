package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"spn-benchmark-ds/internal/pkg/analysis"
	"spn-benchmark-ds/internal/pkg/augmentation"
	"spn-benchmark-ds/internal/pkg/generation"
	"spn-benchmark-ds/internal/pkg/petrinet"
	"spn-benchmark-ds/internal/pkg/report"
	"spn-benchmark-ds/internal/pkg/spn"

	"google.golang.org/protobuf/proto"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := run(config); err != nil {
		log.Fatalf("Error running generation: %v", err)
	}
}

func run(config *Config) error {
	file, err := os.Create(config.OutputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

	var results []*report.SampleResult
	for i := 0; i < config.NumSamples; i++ {
		pn := petrinet.GenerateRandomPetriNet(config.NumPlaces, config.NumTransitions)
		log.Printf("Generated Petri net with %d places and %d transitions", pn.Places, pn.Transitions)
		pn.Prune()
		log.Printf("Pruned Petri net")
		pn.AddTokensRandomly()
		log.Printf("Added tokens randomly")
		rg, err := generation.GenerateReachabilityGraph(pn, config.PlaceUpperBound, config.MarksUpperLimit)
		if err != nil {
			log.Printf("Skipping sample %d: error generating reachability graph: %v", i, err)
			continue
		}

		if !rg.IsBounded || len(rg.Vertices) < config.MarksLowerLimit {
			log.Printf("Skipping sample %d: graph is unbounded or has too few markings", i)
			continue
		}

		lambdaValues := make([]float64, pn.Transitions)
		for i := range lambdaValues {
			lambdaValues[i] = float64(config.MinFiringRate + rand.Intn(config.MaxFiringRate-config.MinFiringRate+1))
		}

		stateMatrix, targetVector := analysis.ComputeStateEquation(rg, lambdaValues)
		steadyStateProbs, err := analysis.SolveForSteadyState(stateMatrix, targetVector)
		if err != nil {
			log.Printf("Skipping sample %d: error solving for steady state: %v", i, err)
			continue
		}

		avgMarkings, markingDensities := analysis.ComputeAverageMarkings(rg.Vertices, steadyStateProbs)

		analysisResult := &analysis.SPNAnalysisResult{
			SteadyStateProbs: steadyStateProbs,
			AverageMarkings:  avgMarkings,
			MarkingDensities: markingDensities,
		}

		if config.EnableTransformations {
			variations := augmentation.GeneratePetriNetVariations(pn, config.PlaceUpperBound, config.MarksLowerLimit, config.MarksUpperLimit, config.MaxTransformsPerSample, config.MinFiringRate, config.MaxFiringRate)
			for _, variation := range variations {
				writeSample(file, config.Format, pn, rg, lambdaValues, variation.SteadyStateProbs, variation.AverageMarkings, variation.MarkingDensities)
				results = append(results, &report.SampleResult{
					NumPlaces:      pn.Places,
					NumTransitions: pn.Transitions,
					Analysis:       variation,
				})
			}
		} else {
			writeSample(file, config.Format, pn, rg, lambdaValues, steadyStateProbs, avgMarkings, markingDensities)
			results = append(results, &report.SampleResult{
				NumPlaces:      pn.Places,
				NumTransitions: pn.Transitions,
				Analysis:       analysisResult,
			})
		}
	}

	if config.EnableStatisticsReport {
		reportFile, err := os.Create(config.OutputFile + ".html")
		if err != nil {
			return fmt.Errorf("error creating report file: %w", err)
		}
		defer reportFile.Close()

		stats := report.CalculateStats(results)
		if err := report.GenerateReport(reportFile, stats); err != nil {
			return fmt.Errorf("error generating report: %w", err)
		}
	}
	return nil
}

func writeSample(file *os.File, format string, pn *petrinet.PetriNet, rg *generation.ReachabilityGraph, lambdaValues, steadyStateProbs, avgMarkings []float64, markingDensities [][]float64) {
	switch format {
	case "jsonl":
		result := map[string]interface{}{
			"petri_net":          pn.Matrix,
			"vertices":           rg.Vertices,
			"edges":              rg.Edges,
			"arc_transitions":    rg.ArcTransitions,
			"lambda_values":      lambdaValues,
			"steady_state_probs": steadyStateProbs,
			"average_markings":   avgMarkings,
			"marking_densities":  markingDensities,
		}
		data, err := json.Marshal(result)
		if err != nil {
			log.Printf("Skipping sample: error marshalling to JSON: %v", err)
			return
		}
		fmt.Fprintln(file, string(data))
	case "protobuf":
		spnData := &spn.SPNData{
			PetriNet: &spn.PetriNet{
				Places:      int32(pn.Places),
				Transitions: int32(pn.Transitions),
				Matrix:      flattenMatrix(pn.Matrix),
			},
			ReachabilityGraph: &spn.ReachabilityGraph{
				Vertices:       toProtoVertices(rg.Vertices),
				Edges:          toProtoEdges(rg.Edges),
				ArcTransitions: toInt32Slice(rg.ArcTransitions),
			},
			LambdaValues:     lambdaValues,
			SteadyStateProbs: steadyStateProbs,
			AverageMarkings:  avgMarkings,
			MarkingDensities: toProtoMarkingDensities(markingDensities),
		}
		data, err := proto.Marshal(spnData)
		if err != nil {
			log.Printf("Skipping sample: error marshalling to protobuf: %v", err)
			return
		}
		file.Write(data)
	default:
		log.Fatalf("Unsupported output format: %s", format)
	}

	fmt.Println("Dataset generation complete.")
}

func flattenMatrix(matrix [][]int) []int32 {
	var flat []int32
	for _, row := range matrix {
		for _, val := range row {
			flat = append(flat, int32(val))
		}
	}
	return flat
}

func toProtoVertices(vertices [][]int) []*spn.Vertex {
	var protoVertices []*spn.Vertex
	for _, v := range vertices {
		protoVertices = append(protoVertices, &spn.Vertex{Marking: toInt32Slice(v)})
	}
	return protoVertices
}

func toProtoEdges(edges [][2]int) []*spn.Edge {
	var protoEdges []*spn.Edge
	for _, e := range edges {
		protoEdges = append(protoEdges, &spn.Edge{Src: int32(e[0]), Dest: int32(e[1])})
	}
	return protoEdges
}

func toProtoMarkingDensities(densities [][]float64) []*spn.MarkingDensity {
	var protoDensities []*spn.MarkingDensity
	for _, d := range densities {
		protoDensities = append(protoDensities, &spn.MarkingDensity{Densities: d})
	}
	return protoDensities
}

func toInt32Slice(slice []int) []int32 {
	var result []int32
	for _, v := range slice {
		result = append(result, int32(v))
	}
	return result
}
