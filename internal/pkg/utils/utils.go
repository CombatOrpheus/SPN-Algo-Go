package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

// LoadJSONLFile loads data from a JSONL file.
func LoadJSONLFile(path string) ([][]byte, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var data [][]byte
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data = append(data, scanner.Bytes())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return data, nil
}

// SaveDataToJSONFile saves data to a JSON file.
func SaveDataToJSONFile(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadJSONFile loads data from a JSON file.
func LoadJSONFile(path string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// SampleJSONFilesFromDirectory samples JSON files from a directory.
func SampleJSONFilesFromDirectory(n int, dir string) ([][]byte, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	var sampledFiles [][]byte
	for i := 0; i < n && i < len(files); i++ {
		file, err := LoadJSONFile(filepath.Join(dir, files[i].Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to load JSON file: %w", err)
		}
		sampledFiles = append(sampledFiles, file)
	}

	return sampledFiles, nil
}
