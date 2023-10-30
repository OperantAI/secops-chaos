/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"

	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/output"
	"k8s.io/client-go/kubernetes"
)

// Experiments is a list of all experiments
var Experiments = []Experiment{
	&PrivilegedContainer{},
}

// Experiment is the interface for an experiment
type Experiment interface {
	// Name returns the name of the experiment
	Name() string
	// Category returns the MITRE/OWASP category of the experiment
	Category() string
	// Run runs the experiment, returning an error if it fails
	Run(ctx context.Context, client *kubernetes.Clientset, config *ExperimentConfig) error
	// Verify verifies the experiment, returning an error if it fails
	Verify(ctx context.Context, client *kubernetes.Clientset, config *ExperimentConfig) (*Outcome, error)
	// Cleanup cleans up the experiment, returning an error if it fails
	Cleanup(ctx context.Context, client *kubernetes.Clientset, config *ExperimentConfig) error
}

// Runner runs a set of experiments
type Runner struct {
	ctx               context.Context
	client            *kubernetes.Clientset
	experiments       map[string]Experiment
	experimentsConfig map[string]*ExperimentConfig
}

// Outcome is the result of an experiment
type Outcome struct {
	Experiment string `json:"experiment"`
	Category   string `json:"category"`
	Success    bool   `json:"success"`
}

type JSONOutput struct {
	K8sVersion string     `json:"k8s_version"`
	Results    []*Outcome `json:"results"`
}

// NewRunner returns a new Runner
func NewRunner(ctx context.Context, namespace string, allNamespaces bool, experiments []string) *Runner {
	// Create a new Kubernetes client
	client, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal("Failed to create Kubernetes client: %s", err)
	}

	experimentMap := make(map[string]Experiment)
	experimentConfigMap := make(map[string]*ExperimentConfig)

	for _, e := range Experiments {
		experimentMap[e.Name()] = e
	}

	for _, e := range experiments {
		experimentsConfig, err := parseExperimentConfig(e)
		if err != nil {
			output.WriteFatal("Failed to parse experiment config: %s", err)
		}

		for _, eConf := range experimentsConfig.Experiments {
			if _, exists := experimentMap[eConf.Type]; exists {
				experimentConfigMap[eConf.Type] = &eConf
			} else {
				output.WriteError("Experiment %s does not exist", eConf.Type)
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
	for _, e := range Experiments {
		output.WriteInfo("Running experiment %s\n", e.Name())
		if err := e.Run(r.ctx, r.client, r.experimentsConfig[e.Name()]); err != nil {
			output.WriteError("Experiment %s failed: %s", e.Name(), err)
		}
	}
}

// RunVerifiers runs all verifiers in the Runner for the provided experiments
func (r *Runner) RunVerifiers(writeJSON bool) {
	table := output.NewTable([]string{"Experiment", "Category", "Result"})
	outcomes := []*Outcome{}
	for _, e := range Experiments {
		outcome, err := e.Verify(r.ctx, r.client, r.experimentsConfig[e.Name()])
		if err != nil {
			output.WriteError("Verifier %s failed: %s", e.Name(), err)
		}
		if writeJSON {
			outcomes = append(outcomes, outcome)
		} else {
			table.AddRow([]string{outcome.Experiment, outcome.Category, fmt.Sprintf("%t", outcome.Success)})
		}
	}
	if writeJSON {
		k8sVersion, err := k8s.GetK8sVersion(r.client)
		if err != nil {
			output.WriteError("Failed to get Kubernetes version: %s", err)
		}
		output.WriteJSON(JSONOutput{
			K8sVersion: k8sVersion.String(),
			Results:    outcomes,
		})
		return
	}
	table.Render()
}

// Cleanup cleans up all experiments in the Runner
func (r *Runner) Cleanup() {
	for _, e := range r.experiments {
		output.WriteInfo("Cleaning up experiment %s\n", e.Name())
		if err := e.Cleanup(r.ctx, r.client, r.experimentsConfig[e.Name()]); err != nil {
			output.WriteError("Experiment %s cleanup failed: %s", e.Name(), err)
		}

	}
}
