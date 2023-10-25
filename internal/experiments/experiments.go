/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"

	"github.com/operantai/experiments-runtime-tool/internal/k8s"
	"github.com/operantai/experiments-runtime-tool/internal/output"
	"k8s.io/client-go/kubernetes"
)

var Experiments = []Experiment{
	&PrivilegedContainer{},
}

type Experiment interface {
	// Name returns the name of the experiment
	Name() string
	// Run runs the experiment, returning an error if it fails
	Run(ctx context.Context, client *kubernetes.Clientset) error
	// Cleanup cleans up the experiment, returning an error if it fails
	Cleanup(ctx context.Context, client *kubernetes.Clientset) error
}

// Runner runs a set of experiments
type Runner struct {
	ctx         context.Context
	client      *kubernetes.Clientset
	experiments map[string]Experiment
}

// NewRunner returns a new Runner
func NewRunner(ctx context.Context, namespace string, allNamespaces bool, experiments []string) *Runner {
	// Create a new Kubernetes client
	client, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal(fmt.Errorf("Failed to create Kubernetes client: %w", err))
	}

	// Check if experiment exists in Experiments slice
	experimentsToRun := make(map[string]Experiment)
	for _, e := range Experiments {
		for _, providedExperiment := range experiments {
			if e.Name() == providedExperiment {
				experimentsToRun[e.Name()] = e
			}
		}
	}

	// Check if all experiments provided are valid
	if len(experimentsToRun) != len(experiments) {
		output.WriteFatal(errors.New("One or more experiments provided are not valid"))
	}

	return &Runner{
		ctx:         ctx,
		client:      client,
		experiments: experimentsToRun,
	}
}

// Run runs all experiments in the Runner
func (r *Runner) Run() {
	for _, e := range r.experiments {
		fmt.Printf("Running experiment %s\n", e.Name())
		output.WriteInfo("Running experiment %s\n", e.Name())
		if err := e.Run(r.ctx, r.client); err != nil {
			output.WriteError(fmt.Errorf("Experiment %s failed: %w", e.Name(), err)
		}
	}
}

// Cleanup cleans up all experiments in the Runner
func (r *Runner) Cleanup() {
	for _, e := range r.experiments {
		output.WriteInfo("Cleaning up experiment %s\n", e.Name())
		fmt.Printf("Cleaning up experiment %s\n", e.Name())
		if err := e.Cleanup(r.ctx, r.client); err != nil {
			output.WriteError(fmt.Errorf("Experiment %s cleanup failed: %w", e.Name(), err)
		}

	}
}
