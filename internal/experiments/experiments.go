/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/output"
	"github.com/operantai/secops-chaos/internal/verifier"
	"k8s.io/client-go/kubernetes"
)

// Experiment is the interface for an experiment
type Experiment interface {
	// Type returns the type of the experiment
	Type() string
	// Description describes the experiment in a brief sentence
	Description() string
	// Framework returns the attack framework e.g., MITRE/OWASP
	Framework() string
	// Tactic returns the attack tactic category
	Tactic() string
	// Technique returns the attack method
	Technique() string
	// Run runs the experiment, returning an error if it fails
	Run(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error
	// Verify verifies the experiment, returning an error if it fails
	Verify(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) (*verifier.Outcome, error)
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

// NewRunner returns a new Runner
func NewRunner(ctx context.Context, experimentFiles []string) *Runner {
	// Create a new Kubernetes client
	client, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal("Failed to create Kubernetes client: %s", err)
	}

	experimentMap := make(map[string]Experiment)
	experimentConfigMap := make(map[string]*ExperimentConfig)

	// Create a map of experiment types to experiments
	for _, e := range ExperimentsRegistry {
		experimentMap[e.Type()] = e
	}

	// Parse the experiment configs
	for _, e := range experimentFiles {
		experimentConfigs, err := parseExperimentConfigs(e)
		if err != nil {
			output.WriteFatal("Failed to parse experiment configs: %s", err)
		}

		for i, eConf := range experimentConfigs {
			if _, exists := experimentMap[eConf.Metadata.Type]; exists {
				experimentConfigMap[eConf.Metadata.Name] = &experimentConfigs[i]
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
func (r *Runner) RunVerifiers(writeJSON bool) {
	table := output.NewTable([]string{"Experiment", "Description", "Framework", "Tactic", "Technique", "Result"})
	outcomes := []*verifier.Outcome{}
	for _, e := range r.experimentsConfig {
		experiment := r.experiments[e.Metadata.Type]
		outcome, err := experiment.Verify(r.ctx, r.client, e)
		if err != nil {
			output.WriteFatal("Verifier %s failed: %s", e.Metadata.Name, err)
		}
		// if JSON flag is set, append to JSON output
		if writeJSON {
			outcomes = append(outcomes, outcome)
		} else {
			table.AddRow([]string{
				outcome.Experiment,
				outcome.Description,
				outcome.Framework,
				outcome.Tactic,
				outcome.Technique,
				outcome.GetResultString(),
			})
		}
	}

	// if JSON flag is set, print JSON output
	if writeJSON {
		k8sVersion, err := k8s.GetK8sVersion(r.client)
		if err != nil {
			output.WriteError("Failed to get Kubernetes version: %s", err)
		}
		output.WriteJSON(verifier.JSONOutput{
			K8sVersion: k8sVersion.String(),
			Results:    outcomes,
		})
		return
	}

	table.Render()
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
