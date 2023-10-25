/*
Copyright 2023 Operant AI
*/
package verifiers

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

type PrivilegedContainer struct{}

func (p *PrivilegedContainer) Name() string {
	return "PrivilegedContainer"
}

func (p *PrivilegedContainer) Verify(ctx context.Context, client *kubernetes.Clientset) (*VerifierOutput, error) {
	return nil, nil
}

var _ Verifier = (*PrivilegedContainer)(nil)
