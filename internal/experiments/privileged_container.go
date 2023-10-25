/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

type PrivilegedContainer struct{}

func (p *PrivilegedContainer) Name() string {
	return "PrivilegedContainer"
}

func (p *PrivilegedContainer) Run(ctx context.Context, client *kubernetes.Clientset) error {
	return nil
}

func (p *PrivilegedContainer) Cleanup(ctx context.Context, client *kubernetes.Clientset) error {
	return nil
}

var _ Experiment = (*PrivilegedContainer)(nil)
