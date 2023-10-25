/*
Copyright 2023 Operant AI
*/
package verifiers

import (
	"context"
	"fmt"
	"time"

	"github.com/operantai/experiments-runtime-tool/internal/k8s"
	"github.com/operantai/experiments-runtime-tool/internal/output"
	"k8s.io/client-go/kubernetes"
)

var Verifiers = []Verifier{
	&PrivilegedContainer{},
}

type Verifier interface {
	// Name returns the name of the verifier
	Name() string
	// Verify verifies the experiment
	Verify(ctx context.Context, client *kubernetes.Clientset) (*VerifierOutput, error)
}

type VerifierOutput struct {
	Timestamp      time.Time         `json:"timestamp"`
	ExperimentName string            `json:"verifier_name"`
	Category       string            `json:"category"`
	Outcome        ExperimentOutcome `json:"success"`
}

type ExperimentOutcome struct {
	ExperimentsRun    int `json:"experiments_run"`
	ExperimentsPassed int `json:"experiments_passed"`
}

type Runner struct {
	ctx       context.Context
	client    *kubernetes.Clientset
	verifiers []Verifier
}

func NewRunner(ctx context.Context, namespace string, allNamespaces bool, verifiers []string) *Runner {
	client, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal("Failed to create Kubernetes client: %w", err)
	}

	// Check if verifiers exists in Verifier slice
	verifiersToRun := make(map[string]Verifier)
	for _, v := range Verifiers {
		for _, providedVerifier := range verifiers {
			if v.Name() == providedVerifier {
				verifiersToRun[v.Name()] = v
			}
		}
	}

	// Check if all verifiers provided exist
	if len(verifiersToRun) != len(verifiers) {
		output.WriteFatal("One or more verifiers provided do not exist")
	}

	return &Runner{
		ctx:       ctx,
		client:    client,
		verifiers: Verifiers,
	}
}

func (r *Runner) Run() {
	headers := []string{"Timestamp", "Name", "Category", "Success"}
	var rows [][]string
	for _, v := range r.verifiers {
		result, err := v.Verify(r.ctx, r.client)
		if err != nil {
			output.WriteError("Failed to verify experiment %s: %w", v.Name(), err)
			continue
		}
		rows = append(rows, []string{result.Timestamp.String(), result.ExperimentName, result.Category, fmt.Sprintf("%s/%s", result.Outcome.ExperimentsPassed, result.Outcome.ExperimentsRun)})
	}
	output.WriteTable(headers, rows)
}
