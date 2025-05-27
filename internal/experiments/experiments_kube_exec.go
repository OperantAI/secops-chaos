/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/operantai/woodpecker/internal/categories"
	"github.com/operantai/woodpecker/internal/k8s"
	"github.com/operantai/woodpecker/internal/verifier"
)

type KubeExec struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters KubeExecParameters `yaml:"parameters"`
}

// KubeExec is an experiment that attempts to run a command in a target container
type KubeExecParameters struct {
	Target struct {
		Pod       string `yaml:"pod"`
		Container string `yaml:"container"`
	} `yaml:"target"`
	Command             []string `yaml:"command"`
	ExpectedOutputRegex string   `yaml:"expectedOutputRegex"`
}

type KubeExecResult struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func (k *KubeExec) Type() string {
	return "kube-exec"
}

func (k *KubeExec) Description() string {
	return "Run a command in a container"
}

func (k *KubeExec) Technique() string {
	return categories.MITRE.PrivilegeEscalation.PrivilegedContainer.Technique
}

func (k *KubeExec) Tactic() string {
	return categories.MITRE.PrivilegeEscalation.PrivilegedContainer.Tactic
}

func (k *KubeExec) Framework() string {
	return string(categories.Mitre)
}

func (k *KubeExec) Run(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	var config KubeExec
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}
	out, errOut, err := client.ExecuteRemoteCommand(
		ctx,
		config.Metadata.Namespace,
		config.Parameters.Target.Pod,
		config.Parameters.Target.Container,
		config.Parameters.Command,
	)
	if err != nil {
		return fmt.Errorf("Error running Kubernetes command in %s/%s in namespace %s: %w", config.Parameters.Target.Pod, config.Parameters.Target.Container, config.Metadata.Namespace, err)
	}

	resultJSON, err := json.Marshal(&KubeExecResult{
		Stdout: out,
		Stderr: errOut,
	})
	if err != nil {
		return fmt.Errorf("Failed to marshal experiment results: %w", err)
	}

	file, err := createTempFile(k.Type(), config.Metadata.Name)
	if err != nil {
		return fmt.Errorf("Unable to create file cache for experiment results %w", err)
	}

	_, err = file.Write(resultJSON)
	if err != nil {
		return fmt.Errorf("Failed to write experiment results: %w", err)
	}
	return nil

}

func (k *KubeExec) Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config KubeExec
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return nil, err
	}
	v := verifier.New(
		config.Metadata.Name,
		config.Description(),
		config.Framework(),
		config.Tactic(),
		config.Technique(),
	)
	rawResults, err := getTempFileContentsForExperiment(k.Type(), config.Metadata.Name)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch experiment results: %w", err)
	}

	for _, rawResult := range rawResults {
		var result KubeExecResult
		if err := json.Unmarshal(rawResult, &result); err != nil {
			return nil, fmt.Errorf("Could not parse experiment result: %w", err)
		}
		if len(result.Stderr) > 0 {
			v.Fail(config.Metadata.Name)
			continue
		}
		regex := regexp.MustCompile(config.Parameters.ExpectedOutputRegex)
		if regex.MatchString(result.Stdout) {
			v.Success(config.Metadata.Name)
			continue
		}
		v.Fail(config.Metadata.Name)
	}

	return v.GetOutcome(), nil
}

func (k *KubeExec) Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error {
	var config KubeExec
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	if err := removeTempFilesForExperiment(k.Type(), config.Metadata.Name); err != nil {
		return err
	}

	return nil
}
