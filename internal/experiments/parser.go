package experiments

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type ExperimentsConfig struct {
	Experiments []ExperimentConfig `yaml:"experiments"`
}

type ExperimentConfig struct {
	// Name of the experiment
	Name string `yaml:"name"`
	// Namespace to apply the experiment to
	Namespace string `yaml:"namespace"`
	// Type of the experiment
	Type string `yaml:"type"`
	// Parameters for the experiment
	Parameters interface{} `yaml:"parameters"`
}

// parseExperimentConfig parses a YAML file and returns a slice of ExperimentConfig
func parseExperimentConfig(file string) (*ExperimentsConfig, error) {
	// Read the file and then unmarshal it into a slice of ExperimentConfig
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	config, err := unmarshalYAML(contents)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func unmarshalYAML(contents []byte) (*ExperimentsConfig, error) {
	var config ExperimentsConfig
	err := yaml.Unmarshal(contents, &config)
	if err != nil {
		return nil, err
	}

	for _, experiment := range config.Experiments {
		if experiment.Parameters == nil {
			return nil, fmt.Errorf("Experiment %s is missing parameters", experiment.Name)
		}
	}

	var specificExperiments []ExperimentConfig

	for _, experiment := range config.Experiments {
		switch experiment.Type {
		case "privileged_container":
			var privilegedContainer PrivilegedContainer
			err := mapstructure.Decode(experiment.Parameters, &privilegedContainer)
			if err != nil {
				return nil, err
			}
			experiment.Parameters = privilegedContainer
			specificExperiments = append(specificExperiments, experiment)
		default:
			return nil, fmt.Errorf("Unsupported experiment type: %s", experiment.Type)
		}
	}

	config.Experiments = specificExperiments

	return &config, nil
}
