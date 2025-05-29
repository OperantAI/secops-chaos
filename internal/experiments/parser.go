/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"fmt"
	"os"
	"time"

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

type AIAppRequest struct {
	SystemPrompt string `json:"system_prompt" yaml:"system_prompt"`
	Prompt       string `json:"prompt" yaml:"prompt"`
}

type AIAppResponse struct {
	Message string `json:"message" yaml:"message"`
}

type AIAppPayload struct {
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
	Prompt       string `json:"prompt"`
}

type AIAPIPayload struct {
	Model                string   `json:"model" yaml:"model"`
	AIApi                string   `json:"ai_api" yaml:"ai_api"`
	SystemPrompt         string   `json:"system_prompt" yaml:"system_prompt"`
	Prompt               string   `json:"prompt" yaml:"prompt"`
	Response             string   `json:"response" yaml:"response"`
	VerifyPromptChecks   []string `json:"verify_prompt_checks" yaml:"verify_prompt_checks"`
	VerifyResponseChecks []string `json:"verify_response_checks" yaml:"verify_response_checks"`
}

type AIVerifierResult struct {
	Check      string  `json:"check"`
	Detected   bool    `json:"detected"`
	EntityType string  `json:"entityType"`
	Score      float64 `json:"score"`
}

type AIVerifierAPIResponse struct {
	Model                  string             `json:"model" yaml:"model"`
	AIApi                  string             `json:"ai_api" yaml:"ai_api"`
	Prompt                 string             `json:"prompt" yaml:"prompt"`
	APIResponse            string             `json:"api_response" yaml:"api_response"`
	VerifiedPromptChecks   []AIVerifierResult `json:"verified_prompt_checks" yaml:"verified_prompt_checks"`
	VerifiedResponseChecks []AIVerifierResult `json:"verified_response_checks" yaml:"verified_response_checks"`
}

type ExecuteAIAPI struct {
	Description      string                `yaml:"description"`
	Payload          AIAPIPayload          `yaml:"payload"`
	ExpectedResponse AIVerifierAPIResponse `yaml:"expected_response"`
}

type ExecuteAIAPIResult struct {
	ExperimentName string                `json:"experiment_name"`
	Description    string                `json:"description"`
	Timestamp      time.Time             `json:"timestamp"`
	Status         int                   `json:"status"`
	Response       AIVerifierAPIResponse `json:"response"`
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
