package verifier

import (
	"bytes"
	"fmt"
)

const (
	// Success is the string used to represent a successful test
	Success = "success"
	// Fail is the string used to represent a failed test
	Fail = "fail"
)

// Verifier is used to verify the results of an experiment
type Verifier struct {
	outcome *Outcome
}

// Outcome is the result of an experiment
type Outcome struct {
	Experiment  string            `json:"experiment"`
	Description string            `json:"description"`
	Framework   string            `json:"framework"`
	Tactic      string            `json:"tactic"`
	Technique   string            `json:"technique"`
	Result      map[string]string `json:"result"`
}

// JSONOutput is a pretty-printed JSON output of the verifier
type JSONOutput struct {
	K8sVersion string     `json:"k8s_version"`
	Results    []*Outcome `json:"results"`
}

// New returns a new Verifier instance
func New(experiment, description, framework, tactic, technique string) *Verifier {
	return &Verifier{
		outcome: &Outcome{
			Experiment:  experiment,
			Description: description,
			Framework:   framework,
			Tactic:      tactic,
			Technique:   technique,
			Result:      make(map[string]string),
		},
	}
}

// Success increments the successful and total counters
func (v *Verifier) Success(test string) {
	if test == "" {
		v.outcome.Result[v.outcome.Experiment] = Success
	} else {
		v.outcome.Result[test] = Success
	}
}

// Fail increments the total counter
func (v *Verifier) Fail(test string) {
	if test == "" {
		v.outcome.Result[v.outcome.Experiment] = Fail
	} else {
		v.outcome.Result[test] = Fail
	}
}

// GetOutcome returns the Outcome of the Verifier
func (v *Verifier) GetOutcome() *Outcome {
	return v.outcome
}

// String returns a string representation of the Verifier result
func (r *Outcome) GetResultString() string {
	b := new(bytes.Buffer)

	// For each result in the Outcome, add a line to the output
	for name, result := range r.Result {
		if result == Success {
			fmt.Fprintf(b, "%s: %s\n", name, Success)
			continue
		}

		fmt.Fprintf(b, "%s: %s\n", name, Fail)
	}
	return b.String()
}
