package verifier

import (
	"bytes"
	"fmt"
)

// Verifier is used to verify the results of an experiment
type Verifier struct {
	outcome *Outcome
}

// Outcome is the result of an experiment
type Outcome struct {
	Experiment  string          `json:"experiment"`
	Description string          `json:"description"`
	Framework   string          `json:"framework"`
	Tactic      string          `json:"tactic"`
	Technique   string          `json:"technique"`
	Result      map[string]bool `json:"result"`
	// Result      Result `json:"result"`
}

// Result is the result of an experiment
//type Result struct {
//	Successful int `json:"successful"`
//	Total      int `json:"total"`
//}

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
		},
	}
}

// AssertEqual compares the actual and expected values and sets the Result accordingly
func (v *Verifier) AssertEqual(test string, actual, expected interface{}) bool {
	// Update the Result based on the assertion
	if actual == expected {
		v.Success(test)
	} else {
		v.Fail(test)
	}

	return actual == expected
}

// Success increments the successful and total counters
func (v *Verifier) Success(test string) {
	if test == "" {
		v.outcome.Result[v.outcome.Experiment] = true
	} else {
		v.outcome.Result[test] = true
	}
}

// Fail increments the total counter
func (v *Verifier) Fail(test string) {
	if test == "" {
		v.outcome.Result[v.outcome.Experiment] = false
	} else {
		v.outcome.Result[test] = false
	}
}

// GetOutcome returns the Outcome of the Verifier
func (v *Verifier) GetOutcome() *Outcome {
	return v.outcome
}

// String returns a string representation of the Verifier result
func (r *Outcome) GetResultString() string {
	b := new(bytes.Buffer)
	for name, result := range r.Result {
		if result {
			fmt.Fprintf(b, "%s: %s\n", name, "success")
			continue
		}
		fmt.Fprintf(b, "%s: %s\n", name, "fail")
	}
	return b.String()
	// return fmt.Sprintf("%d/%d", r.Successful, r.Total)
}
