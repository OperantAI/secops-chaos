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
)

var Experiments = []Experiment{
	&PrivilegedContainer{},
}

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

type JSONOutput struct {
	CLIVersion string     `json:"cli_version"`
	K8sVersion string     `json:"k8s_version"`
	Results    []*Outcome `json:"results"`
}

type Outcome struct {
	ExperimentName string `json:"experiment_name"`
	Success        bool   `json:"success"`
	Message        string `json:"message"`
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
			if _, exists := experimentMap[eConf.Name]; exists {
				experimentConfigMap[eConf.Name] = &eConf
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
func (r *Runner) RunVerifiers(outputJSON bool) {
	headers := []string{"Experiment", "Success", "Message"}
	rows := [][]string{}
	outcomes := []*Outcome{}
	for _, e := range Experiments {
		output.WriteInfo("Verifying experiment %s\n", e.Name())
		outcome, err := e.Verify(r.ctx, r.client, r.experimentsConfig[e.Name()])
		if err != nil {
			output.WriteError("Verifier %s failed: %s", e.Name(), err)
		}
		if outputJSON {
			outcomes = append(outcomes, outcome)
		} else {
			rows = append(rows, []string{outcome.ExperimentName, fmt.Sprintf("%t", outcome.Success), outcome.Message})
		}
	}
	if outputJSON {
		jsonOutput, err := json.Marshal(outcomes)
		if err != nil {
			output.WriteError("Failed to marshal JSON: %s", err)
		}
		fmt.Println(jsonOutput)
		return
	}
	output.WriteTable(headers, rows)
}

// Cleanup cleans up all experiments in the Runner
func (r *Runner) Cleanup() {
	for _, e := range r.experiments {
		output.WriteInfo("Cleaning up experiment %s\n", e.Name())
		fmt.Printf("Cleaning up experiment %s\n", e.Name())
		if err := e.Cleanup(r.ctx, r.client, r.experimentsConfig[e.Name()]); err != nil {
			output.WriteError("Experiment %s cleanup failed: %s", e.Name(), err)
		}

	}
}
