package grid

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"spn-benchmark-ds/internal/pkg/petrinet"
	"spn-benchmark-ds/internal/pkg/utils"
	"spn-benchmark-ds/internal/pkg/generation"
	"spn-benchmark-ds/internal/pkg/analysis"
	"spn-benchmark-ds/internal/pkg/augmentation"
)

type TransformedSample struct {
	PetriNet          *petrinet.PetriNet
	ReachabilityGraph *generation.ReachabilityGraph
	Analysis          *analysis.SPNAnalysisResult
	LambdaValues      []float64
}

type GridSample struct {
	PetriNet          petrinet.PetriNet `json:"petri_net"`
	ReachabilityGraph generation.ReachabilityGraph `json:"reachability_graph"`
}

// PartitionDataIntoGrid partitions the raw data into a grid structure.
func PartitionDataIntoGrid(gridDir string, accumulateData bool, rawDataPath string, placesGridBoundaries []int, markingsGridBoundaries []int) error {
	gridDirPath := filepath.Clean(gridDir)
	gridConfig, err := initializeGrid(gridDirPath, accumulateData, placesGridBoundaries, markingsGridBoundaries)
	if err != nil {
		return fmt.Errorf("failed to initialize grid: %w", err)
	}

	allData, err := utils.LoadJSONLFile(rawDataPath)
	if err != nil {
		return fmt.Errorf("failed to load raw data: %w", err)
	}

	for _, data := range allData {
		var sample GridSample
		if err := json.Unmarshal(data, &sample); err != nil {
			return fmt.Errorf("failed to unmarshal grid sample: %w", err)
		}

		pIdx := getGridIndex(sample.PetriNet.Places, placesGridBoundaries)
		mIdx := getGridIndex(sample.ReachabilityGraph.NumVertices, markingsGridBoundaries)
		gridConfig.JSONCount[pIdx-1][mIdx-1]++

		savePath := filepath.Join(gridDirPath, fmt.Sprintf("p%d", pIdx), fmt.Sprintf("m%d", mIdx), fmt.Sprintf("data%d.json", gridConfig.JSONCount[pIdx-1][mIdx-1]))
		if err := utils.SaveDataToJSONFile(savePath, sample); err != nil {
			return fmt.Errorf("failed to save data to JSON file: %w", err)
		}
	}

	if err := utils.SaveDataToJSONFile(filepath.Join(gridDirPath, "config.json"), gridConfig); err != nil {
		return fmt.Errorf("failed to save grid config: %w", err)
	}

	return nil
}

// SampleAndTransformData samples data from the grid and applies transformations.
func SampleAndTransformData(gridDir string, samplesPerGrid int, lambdaVariationsPerSample int, minFiringRate, maxFiringRate int) ([]*TransformedSample, error) {
	gridDataLoc := filepath.Clean(gridDir)
	gridConfigData, err := os.ReadFile(filepath.Join(gridDataLoc, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load grid config: %w", err)
	}

	var gridConfig GridConfig
    if err := json.Unmarshal(gridConfigData, &gridConfig); err != nil {
        return nil, fmt.Errorf("failed to unmarshal grid config: %w", err)
    }

	var allData []*GridSample
	numPlaceBins := len(gridConfig.RowP) + 1
	numMarkingBins := len(gridConfig.ColM) + 1

	for i := 0; i < numPlaceBins; i++ {
		for j := 0; j < numMarkingBins; j++ {
			directoryPath := filepath.Join(gridDataLoc, fmt.Sprintf("p%d", i+1), fmt.Sprintf("m%d", j+1))
			sampledList, err := utils.SampleJSONFilesFromDirectory(samplesPerGrid, directoryPath)
			if err != nil {
				return nil, fmt.Errorf("failed to sample JSON files: %w", err)
			}
			for _, data := range sampledList {
				var sample GridSample
				if err := json.Unmarshal(data, &sample); err != nil {
					return nil, fmt.Errorf("failed to unmarshal grid sample: %w", err)
				}
				allData = append(allData, &sample)
			}
		}
	}

	var transformedData []*TransformedSample
	for _, data := range allData {
		variations, lambdaValuesList := augmentation.GenerateLambdaVariations(&data.PetriNet, &data.ReachabilityGraph, lambdaVariationsPerSample, minFiringRate, maxFiringRate)
		for i, variation := range variations {
			transformedData = append(transformedData, &TransformedSample{
				PetriNet:          &data.PetriNet,
				ReachabilityGraph: &data.ReachabilityGraph,
				Analysis:          variation,
				LambdaValues:      lambdaValuesList[i],
			})
		}
	}

	return transformedData, nil
}

// initializeGrid initializes the grid structure and configuration.
func initializeGrid(gridDir string, accumulateData bool, placesGridBoundaries []int, markingsGridBoundaries []int) (*GridConfig, error) {
	configPath := filepath.Join(gridDir, "config.json")
	if _, err := os.Stat(configPath); err == nil && accumulateData {
		gridConfigData, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read grid config: %w", err)
		}
		var gridConfig GridConfig
		if err := json.Unmarshal(gridConfigData, &gridConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal grid config: %w", err)
		}
		return &gridConfig, nil
	}

	gridConfig := &GridConfig{
		RowP:      placesGridBoundaries,
		ColM:      markingsGridBoundaries,
		JSONCount: make([][]int, len(placesGridBoundaries)+1),
	}
	for i := range gridConfig.JSONCount {
		gridConfig.JSONCount[i] = make([]int, len(markingsGridBoundaries)+1)
	}

	for i := 0; i <= len(placesGridBoundaries); i++ {
		for j := 0; j <= len(markingsGridBoundaries); j++ {
			if err := os.MkdirAll(filepath.Join(gridDir, fmt.Sprintf("p%d", i+1), fmt.Sprintf("m%d", j+1)), os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create grid directory: %w", err)
			}
		}
	}

	return gridConfig, nil
}

// getGridIndex finds the index of the grid cell for a given value.
func getGridIndex(value int, gridBoundaries []int) int {
	for i, boundary := range gridBoundaries {
		if value < boundary {
			return i + 1
		}
	}
	return len(gridBoundaries) + 1
}

// GridConfig holds the configuration for the grid.
type GridConfig struct {
	RowP      []int `json:"row_p"`
	ColM      []int `json:"col_m"`
	JSONCount [][]int `json:"json_count"`
}
