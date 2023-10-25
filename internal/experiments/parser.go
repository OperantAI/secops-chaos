package experiments

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ExperimentConfig struct {
	// Name of the experiment
	Name string `yaml:"name"`
	// Namespace to apply the experiment to
	Namespace string `yaml:"namespace"`
	// Type of the experiment
	Type string `yaml:"type"`
	// Labels to apply to the experiment in addition to the default labels
	Labels map[string]string `yaml:"labels"`
	// Parameters to apply to the experiment
	Parameters interface{} `yaml:"parameters"`
}

// ParseFile parses a YAML file and returns a slice of ExperimentConfig
func parseFile(file string) ([]ExperimentConfig, error) {
	var config []ExperimentConfig
	err := yaml.Unmarshal([]byte(file), &config)
	for _, experiment := range config {
		if experiment.Parameters == nil {
			return nil, fmt.Errorf("Parameters in experiment %s cannot be nil", experiment.Name)
		}
		switch experiment.Parameters.(type) {
		case PrivilegedContainer:
			var privilegedContainer PrivilegedContainer
			err = yaml.Unmarshal([]byte(file), &privilegedContainer)
			if err != nil {
				return nil, err
			}
			experiment.Parameters = privilegedContainer
		default:
			return nil, fmt.Errorf("Unsupported experiment type: %s", experiment.Type)
		}
	}
	if err != nil {
		return nil, err
	}
	return config, nil
}
