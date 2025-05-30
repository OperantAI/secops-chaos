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

// Generic Verifier that can handle any type
type Verifier[T any] struct {
	outcome *Outcome[T]
}

// Generic Outcome with strongly typed result outputs
type Outcome[T any] struct {
	Experiment    string            `json:"experiment" yaml:"experiment"`
	Description   string            `json:"description" yaml:"description"`
	Framework     string            `json:"framework" yaml:"framework"`
	Tactic        string            `json:"tactic" yaml:"tactic"`
	Technique     string            `json:"technique" yaml:"technique"`
	Result        map[string]string `json:"result" yaml:"result"`
	ResultOutputs map[string][]T    `json:"result_outputs" yaml:"resultOutputs"`
}

// StructuredOutput for multiple outcomes
type StructuredOutput[T any] struct {
	Results []*Outcome[T] `json:"results" yaml:"results"`
}

// AIVerifierResult represents AI verification results
type AIVerifierResult struct {
	Check      string  `json:"check"`
	Detected   bool    `json:"detected"`
	EntityType string  `json:"entityType"`
	Score      float64 `json:"score"`
}

// AIVerifierOutcome represents the outcome of AI verification
type AIVerifierOutcome struct {
	Model                  string             `json:"model" yaml:"model"`
	AIApi                  string             `json:"ai_api" yaml:"ai_api"`
	Prompt                 string             `json:"prompt" yaml:"prompt"`
	APIResponse            string             `json:"api_response" yaml:"api_response"`
	VerifiedPromptChecks   []AIVerifierResult `json:"verified_prompt_checks" yaml:"verified_prompt_checks"`
	VerifiedResponseChecks []AIVerifierResult `json:"verified_response_checks" yaml:"verified_response_checks"`
}

// New returns a new typed Verifier instance
func New[T any](experiment, description, framework, tactic, technique string) *Verifier[T] {
	return &Verifier[T]{
		outcome: &Outcome[T]{
			Experiment:    experiment,
			Description:   description,
			Framework:     framework,
			Tactic:        tactic,
			Technique:     technique,
			Result:        make(map[string]string),
			ResultOutputs: make(map[string][]T),
		},
	}
}

// StoreResultOutputs stores strongly typed results
func (v *Verifier[T]) StoreResultOutputs(experiment string, result T) {
	v.outcome.ResultOutputs[experiment] = append(v.outcome.ResultOutputs[experiment], result)
}

// Success marks an experiment as successful
func (v *Verifier[T]) Success(experiment string) {
	if experiment == "" {
		v.outcome.Result[v.outcome.Experiment] = Success
	} else {
		v.outcome.Result[experiment] = Success
	}
}

// Fail marks an experiment as failed
func (v *Verifier[T]) Fail(experiment string) {
	if experiment == "" {
		v.outcome.Result[v.outcome.Experiment] = Fail
	} else {
		v.outcome.Result[experiment] = Fail
	}
}

// GetOutcome returns the typed Outcome of the Verifier
func (v *Verifier[T]) GetOutcome() *Outcome[T] {
	return v.outcome
}

// GetResultString returns a string representation of the results
func (r *Outcome[T]) GetResultString() string {
	b := new(bytes.Buffer)
	for name, result := range r.Result {
		if result == Success {
			fmt.Fprintf(b, "%s: %s\n", name, Success)
			continue
		}
		fmt.Fprintf(b, "%s: %s\n", name, Fail)
	}
	return b.String()
}

// Type aliases for common verifier types
type AIVerifier = Verifier[AIVerifierOutcome]
type AIOutcome = Outcome[AIVerifierOutcome]

// Factory function for AI verifiers
func NewAIVerifier(experiment, description, framework, tactic, technique string) *AIVerifier {
	return New[AIVerifierOutcome](experiment, description, framework, tactic, technique)
}

// Backward compatibility - keep the old interface for gradual migration
type LegacyVerifier struct {
	outcome *LegacyOutcome
}

type LegacyOutcome struct {
	Experiment    string                   `json:"experiment" yaml:"experiment"`
	Description   string                   `json:"description" yaml:"description"`
	Framework     string                   `json:"framework" yaml:"framework"`
	Tactic        string                   `json:"tactic" yaml:"tactic"`
	Technique     string                   `json:"technique" yaml:"technique"`
	Result        map[string]string        `json:"result" yaml:"result"`
	ResultOutputs map[string][]interface{} `json:"result_outputs" yaml:"resultOutputs"`
}

func NewLegacy(experiment, description, framework, tactic, technique string) *LegacyVerifier {
	return &LegacyVerifier{
		outcome: &LegacyOutcome{
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

func (v *LegacyVerifier) StoreResultOutputs(experiment string, i interface{}) {
	v.outcome.ResultOutputs[experiment] = append(v.outcome.ResultOutputs[experiment], i)
}

func (v *LegacyVerifier) Success(experiment string) {
	if experiment == "" {
		v.outcome.Result[v.outcome.Experiment] = Success
	} else {
		v.outcome.Result[experiment] = Success
	}
}

func (v *LegacyVerifier) Fail(experiment string) {
	if experiment == "" {
		v.outcome.Result[v.outcome.Experiment] = Fail
	} else {
		v.outcome.Result[experiment] = Fail
	}
}

func (v *LegacyVerifier) GetOutcome() *LegacyOutcome {
	return v.outcome
}

func (r *LegacyOutcome) GetResultString() string {
	b := new(bytes.Buffer)
	for name, result := range r.Result {
		if result == Success {
			fmt.Fprintf(b, "%s: %s\n", name, Success)
			continue
		}
		fmt.Fprintf(b, "%s: %s\n", name, Fail)
	}
	return b.String()
}

type LegacyStructuredOutput struct {
	Results []*LegacyOutcome `json:"results" yaml:"results"`
}
