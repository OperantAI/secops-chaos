/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"strings"
	"sync"

	"github.com/heimdalr/dag"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/output"
	"github.com/operantai/secops-chaos/internal/verifier"
)

// Experiment is the interface for an experiment
type Experiment interface {
	// Name returns the name of the experiment
	Name() string
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
	// DepeondsOn returns the dependencies of the experiment
	DependsOn() []string
	// Run runs the experiment, returning an error if it fails
	Run(ctx context.Context, client *k8s.Client) error
	// Verify verifies the experiment, returning an error if it fails
	Verify(ctx context.Context, client *k8s.Client) (*verifier.Outcome, error)
	// Cleanup cleans up the experiment, returning an error if it fails
	Cleanup(ctx context.Context, client *k8s.Client) error
}

// Runner runs a set of experiments
type Runner struct {
	ctx         context.Context
	Client      *k8s.Client
	DAG         *dag.DAG
	Mutex       sync.Mutex
	Experiments map[string]Experiment
	// ExperimentStatus map[string]DAGStatus
	ExperimentStatus sync.Map
	WaitGroup        *sync.WaitGroup
	Cond             *sync.Cond
}

// NewRunner returns a new Runner
func NewRunner(ctx context.Context) *Runner {
	client, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal("Failed to create Kubernetes client: %s", err)
	}

	return &Runner{
		ctx: ctx,
		Client: &k8s.Client{
			Clientset:  client.Clientset,
			RestConfig: client.RestConfig,
		},
		DAG:         dag.NewDAG(),
		Experiments: make(map[string]Experiment),
		WaitGroup:   new(sync.WaitGroup),
		Cond:        sync.NewCond(&sync.Mutex{}),
		// ExperimentStatus: make(map[string]DAGStatus),
	}
}

// ParseExperiments parses the files to it, and adds them to the Runner
func (r *Runner) ParseExperiments(files []string) error {
	for _, f := range files {
		experimentConfigs, err := parseExperimentConfigs(f)
		if err != nil {
			output.WriteFatal("Failed to parse experiment configs: %s", err)
		}

		for _, config := range experimentConfigs {
			experiment, err := ExperimentFactory(&config)
			if err != nil {
				return err
			}
			r.AddExperiment(experiment)
		}
	}
	return nil
}

// AddExperiment adds a experiment to the Runner
func (r *Runner) AddExperiment(experiment Experiment) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	// Add the experiment
	r.Experiments[experiment.Name()] = experiment
	r.DAG.AddVertex(experiment.Name())
	r.ExperimentStatus.Store(experiment.Name(), PendingStatus)

	// Add any dependencies the experiment has
	for _, dependency := range experiment.DependsOn() {
		if err := r.DAG.AddEdge(dependency, experiment.Name()); err != nil {
			return err
		}
	}
	return nil
}

// Run runs all experiments in the Runner
func (r *Runner) Run() {
	visitor := &experimentVisitor{r}
	r.DAG.OrderedWalk(visitor)
	r.WaitGroup.Wait()
	output.WriteInfo("All experiments completed")
}

// RunVerifiers runs all verifiers in the Runner for the provided experiments
func (r *Runner) RunVerifiers(outputFormat string) {
	table := output.NewTable([]string{"Experiment", "Description", "Framework", "Tactic", "Technique", "Result"})
	outcomes := []*verifier.Outcome{}
	for _, experiment := range r.Experiments {
		outcome, err := experiment.Verify(r.ctx, r.Client)
		if err != nil {
			output.WriteFatal("Verifier %s failed: %s", experiment.Name(), err)
		}
		// if JSON flag is set, append to JSON output
		if outputFormat != "" {
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

	// if output flag is set, print JSON or YAML output
	if outputFormat != "" {
		k8sVersion, err := r.Client.GetK8sVersion()
		if err != nil {
			output.WriteError("Failed to get Kubernetes version: %s", err)
		}
		structuredOutput := verifier.StructuredOutput{
			K8sVersion: k8sVersion.String(),
			Results:    outcomes,
		}
		switch strings.ToLower(outputFormat) {
		case "json":
			output.WriteJSON(structuredOutput)
		case "yaml":
			output.WriteYAML(structuredOutput)
		default:
			output.WriteError("Unknown output format: %s", outputFormat)
		}
		return
	}

	table.Render()
}

// Cleanup cleans up all experiments in the Runner
func (r *Runner) Cleanup() {
	for _, experiment := range r.Experiments {
		output.WriteInfo("Cleaning up experiment %s", experiment.Name())
		if err := experiment.Cleanup(r.ctx, r.Client); err != nil {
			output.WriteError("Experiment %s cleanup failed: %s", experiment.Name(), err)
		}
	}
}
