package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config holds the configuration for the dataset generation.
type Config struct {
	// NumPlaces is the number of places in the Petri net.
	NumPlaces int `yaml:"num_places"`
	// NumTransitions is the number of transitions in the Petri net.
	NumTransitions int `yaml:"num_transitions"`
	// NumSamples is the number of samples to generate.
	NumSamples int `yaml:"num_samples"`
	// OutputFile is the path to the output file.
	OutputFile string `yaml:"output_file"`
	// Format is the output format (e.g., "jsonl", "protobuf").
	Format string `yaml:"format"`
	// PlaceUpperBound is the maximum number of tokens a place can hold.
	PlaceUpperBound int `yaml:"place_upper_bound"`
	// MarksLowerLimit is the minimum number of markings a Petri net must have.
	MarksLowerLimit int `yaml:"marks_lower_limit"`
	// MarksUpperLimit is the maximum number of markings a Petri net can have.
	MarksUpperLimit int `yaml:"marks_upper_limit"`
	// MinFiringRate is the minimum firing rate of a transition.
	MinFiringRate int `yaml:"min_firing_rate"`
	// MaxFiringRate is the maximum firing rate of a transition.
	MaxFiringRate int `yaml:"max_firing_rate"`
	// EnableTransformations enables or disables transformations.
	EnableTransformations bool `yaml:"enable_transformations"`
	// MaxTransformsPerSample is the maximum number of transformations to apply to a sample.
	MaxTransformsPerSample int `yaml:"max_transforms_per_sample"`
	// EnableStatisticsReport enables or disables the statistics report.
	EnableStatisticsReport bool `yaml:"enable_statistics_report"`
}

// LoadConfig loads the configuration from a YAML file.
// It takes a path to a YAML file and returns a Config struct.
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
