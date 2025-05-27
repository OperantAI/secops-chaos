package verifier

import (
	"bytes"
	"fmt"
)

const (
	// Success is the string used to represent a successful experiment
	Success = "success"
	// Fail is the string used to represent a failed experiment
	Fail = "fail"
)

// Verifier is used to verify the results of an experiment
type Verifier struct {
	outcome *Outcome
}

// Outcome is the result of an experiment
type Outcome struct {
	Experiment    string                   `json:"experiment" yaml:"experiment"`
	Description   string                   `json:"description" yaml:"description"`
	Framework     string                   `json:"framework" yaml:"framework"`
	Tactic        string                   `json:"tactic" yaml:"tactic"`
	Technique     string                   `json:"technique" yaml:"technique"`
	Result        map[string]string        `json:"result" yaml:"result"`
	ResultOutputs map[string][]interface{} `json:"result_outputs" yaml:"resultOutputs"`
}

type AIVerifierResult struct {
	Check      string  `json:"check"`
	Detected   bool    `json:"detected"`
	EntityType string  `json:"entityType"`
	Score      float64 `json:"score"`
}

type AIVerifierOutcome struct {
	Model                  string             `json:"model" yaml:"model"`
	AIApi                  string             `json:"ai_api" yaml:"ai_api"`
	Prompt                 string             `json:"prompt" yaml:"prompt"`
	APIResponse            string             `json:"api_response" yaml:"api_response"`
	VerifiedPromptChecks   []AIVerifierResult `json:"verified_prompt_checks" yaml:"verified_prompt_checks"`
	VerifiedResponseChecks []AIVerifierResult `json:"verified_response_checks" yaml:"verified_response_checks"`
}

// StructuredOutput is a pretty-printed JSON or YAML  output of the verifier
type StructuredOutput struct {
	Results []*Outcome `json:"results" yaml:"results"`
}

// New returns a new Verifier instance
func New(experiment, description, framework, tactic, technique string) *Verifier {
	return &Verifier{
		outcome: &Outcome{
			Experiment:    experiment,
			Description:   description,
			Framework:     framework,
			Tactic:        tactic,
			Technique:     technique,
			Result:        make(map[string]string),
			ResultOutputs: make(map[string][]interface{}),
		},
	}
}

func (v *Verifier) StoreResultOutputs(experiment string, i interface{}) {
	v.outcome.ResultOutputs[experiment] = append(v.outcome.ResultOutputs[experiment], i)
}

// Success increments the successful and total counters
func (v *Verifier) Success(experiment string) {
	if experiment == "" {
		v.outcome.Result[v.outcome.Experiment] = Success
	} else {
		v.outcome.Result[experiment] = Success
	}
}

// Fail increments the total counter
func (v *Verifier) Fail(experiment string) {
	if experiment == "" {
		v.outcome.Result[v.outcome.Experiment] = Fail
	} else {
		v.outcome.Result[experiment] = Fail
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
