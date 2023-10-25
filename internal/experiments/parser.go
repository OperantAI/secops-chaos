package experiments

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ExperimentConfig struct {
	// Name of the experiment
	Name string `json:"name" yaml:"name"`
	// Namespace to apply the experiment to
	Namespace string `json:"namespace" yaml:"namespace"`
	// Type of the experiment
	Type string `json:"type" yaml:"type"`
	// Labels to apply to the experiment in addition to the default labels
	Labels map[string]string `json:"labels" yaml:"labels"`
	// Parameters to apply to the experiment
	Parameters Parameters `json:"parameters" yaml:"parameters"`
}

type Parameters struct {
	Pod Pod `json:"pod"`
}

type Pod struct {
	HostPid     bool `json:"hostPid"`
	HostNetwork bool `json:"hostNetwork"`
	RunAsUser   int  `json:"runAsUser"`
}

// ParseFile parses a JSON or YAML file and returns a slice of ExperimentConfig
func parseFile(file string) ([]ExperimentConfig, error) {
	fileType := filepath.Ext(file)
	switch fileType {
	case "json":
		config, err := parseJSONFile(file)
		if err != nil {
			return nil, err
		}
		return config, nil
	case "yaml":
		config, err := parseYAMLFile(file)
		if err != nil {
			return nil, err
		}
		return config, nil
	default:
		return nil, fmt.Errorf("Unsupported file type: %s", fileType)
	}
}

// parseJSONFile parses a JSON file and returns a slice of ExperimentConfig
func parseJSONFile(file string) ([]ExperimentConfig, error) {
	var config []ExperimentConfig
	err := json.Unmarshal([]byte(file), &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// parseYAMLFile parses a YAML file and returns a slice of ExperimentConfig
func parseYAMLFile(file string) ([]ExperimentConfig, error) {
	var config []ExperimentConfig
	err := yaml.Unmarshal([]byte(file), &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
