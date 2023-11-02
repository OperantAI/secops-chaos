/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ExperimentsConfig is a structure which represents the configuration for a set of experiments
type ExperimentsConfig struct {
	ExperimentConfigs []ExperimentConfig `yaml:"experiments"`
}

// ExperimentConfig is a structure which represents the configuration for an experiment
type ExperimentConfig struct {
	// Metadata for the experiment
	Metadata ExperimentMetadata `yaml:"metadata"`
	// Parameters for the experiment
	Parameters interface{} `yaml:"parameters"`
}

// ExperimentMetadata is a structure which represents the metadata required for an experiment
type ExperimentMetadata struct {
	// Name of the experiment
	Name string `yaml:"name"`
	// Namespace to apply the experiment to
	Namespace string `yaml:"namespace"`
	// Type of the experiment
	Type string `yaml:"type"`
}

// parseExperimentConfig parses a YAML file and returns a slice of ExperimentConfig
func parseExperimentConfigs(file string) ([]ExperimentConfig, error) {
	// Read the file and then unmarshal it into a slice of ExperimentConfig
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	configs, err := unmarshalYAML(contents)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func unmarshalYAML(contents []byte) ([]ExperimentConfig, error) {
	var config ExperimentsConfig
	err := yaml.Unmarshal(contents, &config)
	if err != nil {
		return nil, err
	}

	for _, experiment := range config.ExperimentConfigs {
		if experiment.Parameters == nil {
			return nil, fmt.Errorf("Experiment %s is missing parameters", experiment.Metadata.Name)
		}
	}
	return config.ExperimentConfigs, nil
}
