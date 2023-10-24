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
package experiments

import (
	"context"
	"fmt"

	"github.com/operantai/experiments-runtime-tool/internal/k8s"
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
func NewRunner(ctx context.Context, experiments []string) *Runner {
	// Create a new Kubernetes client
	client, err := k8s.NewClient()
	if err != nil {
		panic(err)
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
		panic("One or more experiments provided are not valid")
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
		if err := e.Run(r.ctx, r.client); err != nil {
			fmt.Printf("Experiment %s failed: %s\n", e.Name(), err)
		}
	}
}

// Cleanup cleans up all experiments in the Runner
func (r *Runner) Cleanup() {
	for _, e := range r.experiments {
		fmt.Printf("Cleaning up experiment %s\n", e.Name())
		if err := e.Cleanup(r.ctx, r.client); err != nil {
			fmt.Printf("Experiment %s cleanup failed: %s\n", e.Name(), err)
		}

	}
}
