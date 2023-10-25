/*
Copyright 2023 Operant AI
*/
package verifiers

import (
	"context"
	"errors"
	"fmt"

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
	VerifierName string `json:"verifier_name"`
	Success      bool   `json:"success"`
	Message      string `json:"message"`
}

type Runner struct {
	ctx       context.Context
	client    *kubernetes.Clientset
	verifiers []Verifier
}

func NewRunner(ctx context.Context, namespace string, allNamespaces bool, verifiers []string) *Runner {
	client, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal(fmt.Errorf("Failed to create Kubernetes client: %w", err))
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
		output.WriteFatal(errors.New("One or more verifiers provided do not exist"))
	}

	return &Runner{
		ctx:       ctx,
		client:    client,
		verifiers: Verifiers,
	}
}

func (r *Runner) Run() {
	headers := []string{"Name", "Success", "Message"}
	var rows [][]string
	for _, v := range r.verifiers {
		verifierOutcome, err := v.Verify(r.ctx, r.client)
		if err != nil {
			output.WriteError(fmt.Errorf("Failed to verify experiment %s: %w", v.Name(), err))
			continue
		}
		rows = append(rows, []string{verifierOutcome.VerifierName, fmt.Sprintf("%t", verifierOutcome.Success), verifierOutcome.Message})
	}
	output.WriteTable(headers, rows)
}
