/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
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

func (k *KubeExec) Name() string {
	return k.Metadata.Name
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

func (k *KubeExec) DependsOn() []string {
	return k.Metadata.DependsOn
}

func (k *KubeExec) Run(ctx context.Context, client *k8s.Client) error {
	out, errOut, err := client.ExecuteRemoteCommand(
		ctx,
		k.Metadata.Namespace,
		k.Parameters.Target.Pod,
		k.Parameters.Target.Container,
		k.Parameters.Command,
	)
	if err != nil {
		return fmt.Errorf("Error running Kubernetes command in %s/%s in namespace %s: %w", k.Parameters.Target.Pod, k.Parameters.Target.Container, k.Metadata.Namespace, err)
	}

	resultJSON, err := json.Marshal(&KubeExecResult{
		Stdout: out,
		Stderr: errOut,
	})
	if err != nil {
		return fmt.Errorf("Failed to marshal experiment results: %w", err)
	}

	file, err := createTempFile(k.Type(), k.Metadata.Name)
	if err != nil {
		return fmt.Errorf("Unable to create file cache for experiment results %w", err)
	}

	_, err = file.Write(resultJSON)
	if err != nil {
		return fmt.Errorf("Failed to write experiment results: %w", err)
	}
	return nil

}

func (k *KubeExec) Verify(ctx context.Context, client *k8s.Client) (*verifier.Outcome, error) {
	v := verifier.New(
		k.Metadata.Name,
		k.Description(),
		k.Framework(),
		k.Tactic(),
		k.Technique(),
	)
	rawResults, err := getTempFileContentsForExperiment(k.Type(), k.Metadata.Name)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch experiment results: %w", err)
	}

	for _, rawResult := range rawResults {
		var result KubeExecResult
		if err := json.Unmarshal(rawResult, &result); err != nil {
			return nil, fmt.Errorf("Could not parse experiment result: %w", err)
		}
		if len(result.Stderr) > 0 {
			v.Fail(k.Metadata.Name)
			continue
		}
		regex := regexp.MustCompile(k.Parameters.ExpectedOutputRegex)
		if regex.MatchString(result.Stdout) {
			v.Success(k.Metadata.Name)
			continue
		}
		v.Fail(k.Metadata.Name)
	}

	return v.GetOutcome(), nil
}

func (k *KubeExec) Cleanup(ctx context.Context, client *k8s.Client) error {
	if err := removeTempFilesForExperiment(k.Type(), k.Metadata.Name); err != nil {
		return err
	}

	return nil
}
