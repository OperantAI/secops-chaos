package experiments

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ExperimentsConfig struct {
	// Experiments is a slice of ExperimentConfig
	Experiments []ExperimentConfig `yaml:"experiments"`
}

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
func parseFile(file string) (*ExperimentsConfig, error) {
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
			return nil, fmt.Errorf("Parameters in experiment %s cannot be nil", experiment.Name)
		}

		switch experiment.Type {
		case "privileged_container":
			var privilegedContainer PrivilegedContainer
			yamlContents, err := yaml.Marshal(experiment.Parameters)
			if err != nil {
				return nil, err
			}
			err = yaml.Unmarshal(yamlContents, &privilegedContainer)
			if err != nil {
				return nil, err
			}
			experiment.Parameters = privilegedContainer
		default:
			return nil, fmt.Errorf("Unsupported experiment type: %s", experiment.Type)
		}
	}

	return &config, nil
}
