package verifier

import "fmt"

// Verifier is used to verify the results of an experiment
type Verifier struct {
	Outcome       *Outcome
	Experiment    string
	Category      string
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
func New(experiment, description string) *Verifier {
	return &Verifier{
		Outcome:    &Outcome{},
		Experiment: experiment,
		Category:   category,
	}
}

// AssertEqual compares the actual and expected values and sets the Result accordingly
func (v *Verifier) AssertEqual(actual, expected interface{}) bool {
	v.numAssertions++

	// Create a new Outcome if one doesn't exist
	if v.Outcome == nil {
		v.Outcome = &Outcome{
			Experiment: v.Experiment,
			Category:   v.Category,
			Result:     Result{},
		}
	}

	// Update the Result based on the assertion
	if actual == expected {
		v.Outcome.Result.Successful++
	}
	v.Outcome.Result.Total++

	return actual == expected
}

// GetOutcome returns the Outcome of the Verifier
func (v *Verifier) GetOutcome() *Outcome {
	return v.Outcome
}

// String returns a string representation of the Verifier result
func (r *Result) String() string {
	return fmt.Sprintf("%d/%d", r.Successful, r.Total)
}
