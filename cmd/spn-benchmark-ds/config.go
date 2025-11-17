package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config holds the configuration for the dataset generation.
type Config struct {
	NumPlaces       int    `yaml:"num_places"`
	NumTransitions  int    `yaml:"num_transitions"`
	NumSamples      int    `yaml:"num_samples"`
	OutputFile      string `yaml:"output_file"`
	Format          string `yaml:"format"`
	PlaceUpperBound int    `yaml:"place_upper_bound"`
	MarksLowerLimit int    `yaml:"marks_lower_limit"`
	MarksUpperLimit int    `yaml:"marks_upper_limit"`
}

// LoadConfig loads the configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}
