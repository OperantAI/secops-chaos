/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/output"
	"k8s.io/client-go/kubernetes"
	"log"
)

var Experiments = []Experiment{
	&PrivilegedContainerExperimentConfig{},
	&HostPathMountExperimentConfig{},
}

type Experiment interface {
	// Type returns the type of the experiment
	Type() string
	// Category returns the MITRE/OWASP category of the experiment
	Category() string
	// Run runs the experiment, returning an error if it fails
	Run(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error
	// Verify verifies the experiment, returning an error if it fails
	Verify(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) (*Outcome, error)
	// Cleanup cleans up the experiment, returning an error if it fails
	Cleanup(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error
}

// Runner runs a set of experiments
type Runner struct {
	ctx               context.Context
	client            *kubernetes.Clientset
	experiments       map[string]Experiment
	experimentsConfig map[string]*ExperimentConfig
}

type JSONOutput struct {
	K8sVersion string     `json:"k8s_version"`
	Results    []*Outcome `json:"results"`
}

type Outcome struct {
	Experiment string `json:"experiment"`
	Category   string `json:"category"`
	Success    bool   `json:"success"`
}

// NewRunner returns a new Runner
func NewRunner(ctx context.Context, namespace string, allNamespaces bool, experimentFiles []string) *Runner {
	// Create a new Kubernetes client
	client, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal("Failed to create Kubernetes client: %s", err)
	}

	experimentMap := make(map[string]Experiment)
	experimentConfigMap := make(map[string]*ExperimentConfig)

	for _, e := range Experiments {
		experimentMap[e.Type()] = e
		log.Println(e.Type())
	}

	for _, e := range experimentFiles {
		experimentConfigs, err := parseExperimentConfigs(e)
		if err != nil {
			output.WriteFatal("Failed to parse experiment configs: %s", err)
		}

		for _, eConf := range experimentConfigs {
			if _, exists := experimentMap[eConf.Metadata.Type]; exists {
				experimentConfigMap[eConf.Metadata.Type] = &eConf
			} else {
				output.WriteError("Experiment %s does not exist", eConf.Metadata.Type)
			}
		}
	}

	return &Runner{
		ctx:               ctx,
		client:            client,
		experiments:       experimentMap,
		experimentsConfig: experimentConfigMap,
	}
}

// Run runs all experiments in the Runner
func (r *Runner) Run() {
	for _, e := range r.experimentsConfig {
		experiment := r.experiments[e.Metadata.Type]
		output.WriteInfo("Running experiment %s\n", e.Metadata.Name)
		if err := experiment.Run(r.ctx, r.client, e); err != nil {
			output.WriteError("Experiment %s failed with error: %s", e.Metadata.Name, err)
		}
	}
}

// RunVerifiers runs all verifiers in the Runner for the provided experiments
func (r *Runner) RunVerifiers(outputJSON bool) {
	headers := []string{"Experiment", "Category", "Result"}
	rows := [][]string{}
	outcomes := []*Outcome{}
	for _, e := range r.experimentsConfig {
		experiment := r.experiments[e.Metadata.Type]
		outcome, err := experiment.Verify(r.ctx, r.client, e)
		if err != nil {
			output.WriteError("Verifier %s failed: %s", e.Metadata.Name, err)
		}
		if outputJSON {
			outcomes = append(outcomes, outcome)
		} else {
			rows = append(rows, []string{outcome.Experiment, outcome.Category, fmt.Sprintf("%t", outcome.Success)})
		}
	}
	if outputJSON {
		k8sVersion, err := k8s.GetK8sVersion(r.client)
		if err != nil {
			output.WriteError("Failed to get Kubernetes version: %s", err)
		}
		out := JSONOutput{
			K8sVersion: k8sVersion.String(),
			Results:    outcomes,
		}
		jsonOutput, err := json.MarshalIndent(out, "", "    ")
		if err != nil {
			output.WriteError("Failed to marshal JSON: %s", err)
		}
		fmt.Println(string(jsonOutput))
		return
	}
	output.WriteTable(headers, rows)
}

// Cleanup cleans up all experiments in the Runner
func (r *Runner) Cleanup() {
	for _, e := range r.experimentsConfig {
		output.WriteInfo("Cleaning up experiment %s\n", e.Metadata.Name)
		experiment := r.experiments[e.Metadata.Type]
		if err := experiment.Cleanup(r.ctx, r.client, e); err != nil {
			output.WriteError("Experiment %s cleanup failed: %s", e.Metadata.Name, err)
		}

	}
}
