package verifier

import "fmt"

// Verifier is used to verify the results of an experiment
type Verifier struct {
	outcome       *Outcome
	numAssertions int
}

// Outcome is the result of an experiment
type Outcome struct {
	Experiment  string `json:"experiment"`
	Description string `json:"description"`
	Framework   string `json:"framework"`
	Tactic      string `json:"tactic"`
	Technique   string `json:"technique"`
	Result      Result `json:"result"`
}

// Result is the result of an experiment
type Result struct {
	Successful int `json:"successful"`
	Total      int `json:"total"`
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
		},
	}
}

// AssertEqual compares the actual and expected values and sets the Result accordingly
func (v *Verifier) AssertEqual(actual, expected interface{}) bool {
	v.numAssertions++

	// Update the Result based on the assertion
	if actual == expected {
		v.outcome.Result.Successful++
	}
	v.outcome.Result.Total++

	return actual == expected
}

// Success increments the successful and total counters
func (v *Verifier) Success() {
	v.outcome.Result.Successful++
	v.outcome.Result.Total++
}

// Fail increments the total counter
func (v *Verifier) Fail() {
	v.outcome.Result.Total++
}

// GetOutcome returns the Outcome of the Verifier
func (v *Verifier) GetOutcome() *Outcome {
	return v.outcome
}

// String returns a string representation of the Verifier result
func (r *Result) String() string {
	return fmt.Sprintf("%d/%d", r.Successful, r.Total)
}
