package grid

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetGridIndex(t *testing.T) {
	boundaries := []int{10, 20, 30}
	testCases := []struct {
		value    int
		expected int
	}{
		{5, 1},
		{15, 2},
		{25, 3},
		{35, 4},
	}

	for _, tc := range testCases {
		actual := getGridIndex(tc.value, boundaries)
		if actual != tc.expected {
			t.Errorf("getGridIndex(%d, %v) = %d, expected %d", tc.value, boundaries, actual, tc.expected)
		}
	}
}

func TestPartitionDataIntoGrid(t *testing.T) {
	// Create a temporary directory for the grid
	gridDir, err := os.MkdirTemp("", "test_grid")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(gridDir)

	// Create a dummy raw data file
	rawData := []map[string]interface{}{
		{
			"petri_net": map[string]interface{}{
				"Places":      5,
				"Transitions": 3,
			},
			"reachability_graph": map[string]interface{}{
				"NumVertices": 15,
				"IsBounded":   true,
			},
		},
		{
			"petri_net": map[string]interface{}{
				"Places":      15,
				"Transitions": 3,
			},
			"reachability_graph": map[string]interface{}{
				"NumVertices": 25,
				"IsBounded":   true,
			},
		},
	}
	rawDataPath := filepath.Join(gridDir, "raw_data.jsonl")
	file, err := os.Create(rawDataPath)
	if err != nil {
		t.Fatalf("failed to create raw data file: %v", err)
	}
	for _, data := range rawData {
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatalf("failed to marshal raw data: %v", err)
		}
		if _, err := file.Write(jsonData); err != nil {
			t.Fatalf("failed to write raw data: %v", err)
		}
		if _, err := file.WriteString("\n"); err != nil {
			t.Fatalf("failed to write newline: %v", err)
		}
	}
	file.Close()

	// Partition the data
	placesBoundaries := []int{10}
	markingsBoundaries := []int{20}
	if err := PartitionDataIntoGrid(gridDir, false, rawDataPath, placesBoundaries, markingsBoundaries); err != nil {
		t.Fatalf("PartitionDataIntoGrid failed: %v", err)
	}

	// Check if the grid was created correctly
	configPath := filepath.Join(gridDir, "config.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("config.json not found: %v", err)
	}

	// Check if the data was partitioned correctly
	data1Path := filepath.Join(gridDir, "p1", "m1", "data1.json")
	if _, err := os.Stat(data1Path); err != nil {
		t.Errorf("data1.json not found: %v", err)
	}
	data2Path := filepath.Join(gridDir, "p2", "m2", "data1.json")
	if _, err := os.Stat(data2Path); err != nil {
		t.Errorf("data2.json not found: %v", err)
	}
}

func TestSampleAndTransformData(t *testing.T) {
	// Create a temporary directory for the grid
	gridDir, err := os.MkdirTemp("", "test_grid")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(gridDir)

	// Create a dummy grid
	placesBoundaries := []int{10}
	markingsBoundaries := []int{20}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if err := os.MkdirAll(filepath.Join(gridDir, fmt.Sprintf("p%d", i+1), fmt.Sprintf("m%d", j+1)), os.ModePerm); err != nil {
				t.Fatalf("failed to create grid directory: %v", err)
			}
		}
	}
	gridConfig := &GridConfig{
		RowP:      placesBoundaries,
		ColM:      markingsBoundaries,
		JSONCount: [][]int{{1, 0}, {0, 0}},
	}
	configPath := filepath.Join(gridDir, "config.json")
	configData, err := json.Marshal(gridConfig)
	if err != nil {
		t.Fatalf("failed to marshal grid config: %v", err)
	}
	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		t.Fatalf("failed to write grid config: %v", err)
	}
	data1 := map[string]interface{}{
		"petri_net": map[string]interface{}{
			"Places":      2,
			"Transitions": 1,
		},
		"reachability_graph": map[string]interface{}{
			"Vertices":       []int{1, 0, 0, 1},
			"Edges":          []int{0, 1},
			"VerticesStride": 2,
			"EdgesStride":    2,
			"NumVertices":    2,
			"NumEdges":       1,
			"ArcTransitions": []int{0},
			"IsBounded":      true,
		},
	}
	data1Path := filepath.Join(gridDir, "p1", "m1", "data1.json")
	data1Data, err := json.Marshal(data1)
	if err != nil {
		t.Fatalf("failed to marshal data1: %v", err)
	}
	if err := os.WriteFile(data1Path, data1Data, 0600); err != nil {
		t.Fatalf("failed to write data1: %v", err)
	}

	// Sample and transform the data
	samples, err := SampleAndTransformData(gridDir, 1, 1, 1, 10)
	if err != nil {
		t.Fatalf("SampleAndTransformData failed: %v", err)
	}

	// Check if the correct number of samples were returned
	if len(samples) != 1 {
		t.Errorf("expected 1 sample, got %d", len(samples))
	}
}
