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
	MinFiringRate   int    `yaml:"min_firing_rate"`
	MaxFiringRate   int    `yaml:"max_firing_rate"`
	EnableTransformations bool `yaml:"enable_transformations"`
	MaxTransformsPerSample int `yaml:"max_transforms_per_sample"`
	EnableStatisticsReport bool `yaml:"enable_statistics_report"`
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
