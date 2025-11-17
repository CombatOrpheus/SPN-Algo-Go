package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"spn-benchmark-ds/internal/pkg/analysis"
	"spn-benchmark-ds/internal/pkg/augmentation"
	"spn-benchmark-ds/internal/pkg/generation"
	"spn-benchmark-ds/internal/pkg/grid"
	"spn-benchmark-ds/internal/pkg/petrinet"
	"spn-benchmark-ds/internal/pkg/report"
	"spn-benchmark-ds/internal/pkg/spn"

	"google.golang.org/protobuf/proto"
)

// main is the entry point of the application.
// It parses the command-line arguments, loads the configuration, and runs the generation process.
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

// run is the main function of the application.
// It generates the dataset based on the given configuration.
func run(config *Config) error {
	if config.GenerationMode == "grid" {
		return runGridGeneration(config)
	}
	return runRandomGeneration(config)
}

// runRandomGeneration generates the dataset based on the given configuration.
func runRandomGeneration(config *Config) error {
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

		if !rg.IsBounded || rg.NumVertices < config.MarksLowerLimit {
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

		avgMarkings, markingDensities := analysis.ComputeAverageMarkings(rg, steadyStateProbs)

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

// runGridGeneration generates the dataset based on the given configuration.
func runGridGeneration(config *Config) error {
	// Generate raw data
	rawFilePath := filepath.Join(config.TemporaryGridLocation, "raw_data.jsonl")
	if err := generateRawData(config, rawFilePath); err != nil {
		return fmt.Errorf("error generating raw data: %w", err)
	}

	// Partition data into grid
	if err := grid.PartitionDataIntoGrid(config.TemporaryGridLocation, config.AccumulationData, rawFilePath, config.PlacesGridBoundaries, config.MarkingsGridBoundaries); err != nil {
		return fmt.Errorf("error partitioning data into grid: %w", err)
	}

	// Sample and transform data
	results, err := grid.SampleAndTransformData(config.TemporaryGridLocation, config.SamplesPerGrid, config.LambdaVariationsPerSample, config.MinFiringRate, config.MaxFiringRate)
	if err != nil {
		return fmt.Errorf("error sampling and transforming data: %w", err)
	}

	// Package dataset
	file, err := os.Create(config.OutputGridLocation)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

	for _, result := range results {
		writeSample(file, config.Format, result.PetriNet, result.ReachabilityGraph, result.LambdaValues, result.Analysis.SteadyStateProbs, result.Analysis.AverageMarkings, result.Analysis.MarkingDensities)
	}

	return nil
}

func generateRawData(config *Config, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

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

		if !rg.IsBounded || rg.NumVertices < config.MarksLowerLimit {
			log.Printf("Skipping sample %d: graph is unbounded or has too few markings", i)
			continue
		}
		writeSample(file, config.Format, pn, rg, nil, nil, nil, nil)
	}
	return nil
}

// writeSample writes a sample to the output file in the specified format.
func writeSample(writer io.Writer, format string, pn *petrinet.PetriNet, rg *generation.ReachabilityGraph, lambdaValues, steadyStateProbs, avgMarkings []float64, markingDensities [][]float64) {
	switch format {
	case "jsonl":
		result := map[string]interface{}{
			"petri_net":          pn,
			"reachability_graph": rg,
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
		fmt.Fprintln(writer, string(data))
	case "protobuf":
		spnData := &spn.SPNData{
			PetriNet: &spn.PetriNet{
				Places:      int32(pn.Places),
				Transitions: int32(pn.Transitions),
				Matrix:      toInt32Slice(pn.Matrix),
			},
			ReachabilityGraph: &spn.ReachabilityGraph{
				Vertices:       toProtoVertices(rg),
				Edges:          toProtoEdges(rg),
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
		if _, err := writer.Write(data); err != nil {
			log.Printf("Skipping sample: error writing to file: %v", err)
		}
	default:
		log.Fatalf("Unsupported output format: %s", format)
	}

	fmt.Println("Dataset generation complete.")
}

// toProtoVertices converts the vertices of a reachability graph to the protobuf format.
func toProtoVertices(rg *generation.ReachabilityGraph) []*spn.Vertex {
	var protoVertices []*spn.Vertex
	for i := 0; i < rg.NumVertices; i++ {
		v := rg.Vertex(i)
		protoVertices = append(protoVertices, &spn.Vertex{Marking: toInt32Slice(v)})
	}
	return protoVertices
}

// toProtoEdges converts the edges of a reachability graph to the protobuf format.
func toProtoEdges(rg *generation.ReachabilityGraph) []*spn.Edge {
	var protoEdges []*spn.Edge
	for i := 0; i < rg.NumEdges; i++ {
		e := rg.Edge(i)
		protoEdges = append(protoEdges, &spn.Edge{Src: int32(e[0]), Dest: int32(e[1])})
	}
	return protoEdges
}

// toProtoMarkingDensities converts the marking densities to the protobuf format.
func toProtoMarkingDensities(densities [][]float64) []*spn.MarkingDensity {
	var protoDensities []*spn.MarkingDensity
	for _, d := range densities {
		protoDensities = append(protoDensities, &spn.MarkingDensity{Densities: d})
	}
	return protoDensities
}

// toInt32Slice converts a slice of ints to a slice of int32s.
func toInt32Slice(slice []int) []int32 {
	var result []int32
	for _, v := range slice {
		result = append(result, int32(v))
	}
	return result
}
