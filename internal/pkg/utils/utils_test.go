package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndSaveJSONFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_json")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy data structure
	data := map[string]interface{}{
		"foo": "bar",
		"baz": 123,
	}

	// Save the data to a JSON file
	filePath := filepath.Join(tmpDir, "data.json")
	if err := SaveDataToJSONFile(filePath, data); err != nil {
		t.Fatalf("SaveDataToJSONFile failed: %v", err)
	}

	// Load the data from the JSON file
	loadedData, err := LoadJSONFile(filePath)
	if err != nil {
		t.Fatalf("LoadJSONFile failed: %v", err)
	}

	// Check if the loaded data is the same as the original data
	var loadedDataMap map[string]interface{}
	if err := json.Unmarshal(loadedData, &loadedDataMap); err != nil {
		t.Fatalf("failed to unmarshal loaded data: %v", err)
	}
	if loadedDataMap["foo"] != "bar" || loadedDataMap["baz"].(float64) != 123 {
		t.Errorf("loaded data is not the same as the original data")
	}
}

func TestLoadAndSaveJSONLFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_jsonl")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy data structure
	data := []map[string]interface{}{
		{
			"foo": "bar",
			"baz": 123,
		},
		{
			"foo": "qux",
			"baz": 456,
		},
	}

	// Save the data to a JSONL file
	filePath := filepath.Join(tmpDir, "data.jsonl")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	for _, item := range data {
		jsonData, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("failed to marshal data: %v", err)
		}
		if _, err := file.Write(jsonData); err != nil {
			t.Fatalf("failed to write data: %v", err)
		}
		if _, err := file.WriteString("\n"); err != nil {
			t.Fatalf("failed to write newline: %v", err)
		}
	}
	file.Close()

	// Load the data from the JSONL file
	loadedData, err := LoadJSONLFile(filePath)
	if err != nil {
		t.Fatalf("LoadJSONLFile failed: %v", err)
	}

	// Check if the loaded data is the same as the original data
	if len(loadedData) != 2 {
		t.Fatalf("expected 2 items, got %d", len(loadedData))
	}
	var loadedDataMap map[string]interface{}
	if err := json.Unmarshal(loadedData[0], &loadedDataMap); err != nil {
		t.Fatalf("failed to unmarshal loaded data: %v", err)
	}
	if loadedDataMap["foo"] != "bar" || loadedDataMap["baz"].(float64) != 123 {
		t.Errorf("loaded data is not the same as the original data")
	}
	if err := json.Unmarshal(loadedData[1], &loadedDataMap); err != nil {
		t.Fatalf("failed to unmarshal loaded data: %v", err)
	}
	if loadedDataMap["foo"] != "qux" || loadedDataMap["baz"].(float64) != 456 {
		t.Errorf("loaded data is not the same as the original data")
	}
}

func TestSampleJSONFilesFromDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_sample")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some dummy JSON files
	for i := 0; i < 5; i++ {
		data := map[string]interface{}{
			"index": i,
		}
		filePath := filepath.Join(tmpDir, fmt.Sprintf("data%d.json", i))
		if err := SaveDataToJSONFile(filePath, data); err != nil {
			t.Fatalf("SaveDataToJSONFile failed: %v", err)
		}
	}

	// Sample 3 files from the directory
	sampledFiles, err := SampleJSONFilesFromDirectory(3, tmpDir)
	if err != nil {
		t.Fatalf("SampleJSONFilesFromDirectory failed: %v", err)
	}

	// Check if the correct number of files were sampled
	if len(sampledFiles) != 3 {
		t.Errorf("expected 3 files, got %d", len(sampledFiles))
	}
}
