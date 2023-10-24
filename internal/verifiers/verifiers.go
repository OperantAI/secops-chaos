/*
Copyright © 2023 Operant AI

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package verifiers

import (
	"context"
	"fmt"

	"github.com/operantai/experiments-runtime-tool/internal/k8s"
	"k8s.io/client-go/kubernetes"
)

var Verifiers = []Verifier{
	&PrivilegedContainer{},
}

type Verifier interface {
	// Name returns the name of the verifier
	Name() string
	// Verify verifies the experiment
	Verify(ctx context.Context, client *kubernetes.Clientset) error
}

type Runner struct {
	ctx       context.Context
	client    *kubernetes.Clientset
	verifiers []Verifier
}

func NewRunner(ctx context.Context, verifiers []string) *Runner {
	client, err := k8s.NewClient()
	if err != nil {
		panic(err)
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
		panic("One or more verifiers provided do not exist")
	}

	return &Runner{
		ctx:       ctx,
		client:    client,
		verifiers: Verifiers,
	}
}

func (r *Runner) Run() {
	for _, v := range r.verifiers {
		fmt.Printf("Running verifier: %s\n", v.Name())
		if err := v.Verify(r.ctx, r.client); err != nil {
			fmt.Printf("Failed to verify experiment %s: %s\n", v.Name(), err)
		}
	}
}
