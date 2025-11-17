package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	// Create a temporary config file
	configFile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(configFile.Name())

	configContent := `
num_places: 1
num_transitions: 1
num_samples: 1
output_file: "test_output.jsonl"
format: "jsonl"
place_upper_bound: 10
marks_lower_limit: 1
marks_upper_limit: 100
min_firing_rate: 5
max_firing_rate: 5
enable_transformations: false
enable_statistics_report: true
`
	if _, err := configFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	configFile.Close()

	config, err := LoadConfig(configFile.Name())
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if err := run(config); err != nil {
		t.Fatalf("Error running generation: %v", err)
	}

	// Check that the output file was created and has content
	content, err := os.ReadFile("test_output.jsonl")
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if len(content) == 0 {
		t.Errorf("Output file is empty")
	}

	// Check that the report file was created and has content
	reportContent, err := os.ReadFile("test_output.jsonl.html")
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}
	if len(reportContent) == 0 {
		t.Errorf("Report file is empty")
	}
	if !strings.Contains(string(reportContent), "<td>1</td>") {
		t.Errorf("Report does not contain the correct number of samples")
	}

	// Clean up
	os.Remove("test_output.jsonl")
	os.Remove("test_output.jsonl.html")
}

func TestRandomFiringRates(t *testing.T) {
	// Create a temporary config file
	configFile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(configFile.Name())

	configContent := `
num_places: 2
num_transitions: 1
num_samples: 1
output_file: "test_output.jsonl"
format: "jsonl"
place_upper_bound: 10
marks_lower_limit: 1
marks_upper_limit: 100
min_firing_rate: 5
max_firing_rate: 5
enable_transformations: false
enable_statistics_report: false
`
	if _, err := configFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	configFile.Close()

	config, err := LoadConfig(configFile.Name())
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if err := run(config); err != nil {
		t.Fatalf("Error running generation: %v", err)
	}

	// Check that the output file was created and has content
	content, err := os.ReadFile("test_output.jsonl")
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check the firing rates in the output
	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}
	lambdaValues := result["lambda_values"].([]interface{})
	for _, val := range lambdaValues {
		if val.(float64) != 5.0 {
			t.Errorf("Expected firing rate to be 5.0, but got %f", val.(float64))
		}
	}

	// Clean up
	os.Remove("test_output.jsonl")
}

func TestRunGridGeneration(t *testing.T) {
	// Create a temporary config file
	configFile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(configFile.Name())

	configContent := `
generation_mode: "grid"
num_places: 1
num_transitions: 1
num_samples: 1
output_file: "test_output.jsonl"
format: "jsonl"
place_upper_bound: 10
marks_lower_limit: 1
marks_upper_limit: 100
min_firing_rate: 5
max_firing_rate: 5
enable_transformations: false
enable_statistics_report: false
places_grid_boundaries: [5]
markings_grid_boundaries: [50]
samples_per_grid: 1
lambda_variations_per_sample: 1
temporary_grid_location: "test_grid"
output_grid_location: "test_grid_output.jsonl"
`
	if _, err := configFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	configFile.Close()

	config, err := LoadConfig(configFile.Name())
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if err := os.MkdirAll(config.TemporaryGridLocation, os.ModePerm); err != nil {
		t.Fatalf("Failed to create temp grid location: %v", err)
	}

	if err := run(config); err != nil {
		t.Fatalf("Error running generation: %v", err)
	}

	// Check that the output file was created and has content
	content, err := os.ReadFile("test_grid_output.jsonl")
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if len(content) == 0 {
		t.Errorf("Output file is empty")
	}

	// Clean up
	os.Remove("test_output.jsonl")
	os.RemoveAll("test_grid")
	os.Remove("test_grid_output.jsonl")
}
