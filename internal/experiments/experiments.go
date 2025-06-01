/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"
	"strings"

	"github.com/operantai/woodpecker/internal/output"
	"github.com/operantai/woodpecker/internal/verifier"
)

// Experiment interface
type Experiment interface {
	// Type returns the type of the experiment
	Type() string
	// Description describes the experiment in a brief sentence
	Description() string
	// Framework returns the attack framework e.g., MITRE/OWASP
	Framework() string
	// Tactic returns the attack tactic category
	Tactic() string
	// Technique returns the attack method
	Technique() string
	// Run runs the experiment, returning an error if it fails
	Run(ctx context.Context, experimentConfig *ExperimentConfig) error
	// Verify verifies the experiment, returning an error if it fails
	Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.LegacyOutcome, error)
	// Cleanup cleans up the experiment, returning an error if it fails
	Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error
}

// Runner runs a set of experiments
type Runner struct {
	ctx               context.Context
	experiments       map[string]Experiment
	experimentsConfig map[string]*ExperimentConfig
}

// NewRunner returns a new Runner
func NewRunner(ctx context.Context, experimentFiles []string) *Runner {
	experimentMap := make(map[string]Experiment)
	experimentConfigMap := make(map[string]*ExperimentConfig)

	// Create a map of experiment types to experiments
	for _, e := range ExperimentsRegistry {
		experimentMap[e.Type()] = e
	}

	// Parse the experiment configs
	for _, e := range experimentFiles {
		experimentConfigs, err := parseExperimentConfigs(e)
		if err != nil {
			output.WriteFatal("Failed to parse experiment configs: %s", err)
		}

		for i, eConf := range experimentConfigs {
			if _, exists := experimentMap[eConf.Metadata.Type]; exists {
				experimentConfigMap[eConf.Metadata.Name] = &experimentConfigs[i]
			} else {
				output.WriteError("Experiment %s does not exist", eConf.Metadata.Type)
			}
		}
	}

	return &Runner{
		ctx:               ctx,
		experiments:       experimentMap,
		experimentsConfig: experimentConfigMap,
	}
}

// Run runs all experiments in the Runner
func (r *Runner) Run() {
	for _, e := range r.experimentsConfig {
		experiment := r.experiments[e.Metadata.Type]
		output.WriteInfo("Running experiment %s", e.Metadata.Name)
		if err := experiment.Run(r.ctx, e); err != nil {
			output.WriteError("Experiment %s failed with error: %s", e.Metadata.Name, err)
		}
		output.WriteInfo("Finished running experiment %s. Check results using woodpecker experiment verify command. \n", e.Metadata.Name)
	}
}

// RunVerifiers runs all verifiers in the Runner for the provided experiments
func (r *Runner) RunVerifiers(outputFormat string) {
	if outputFormat != "" {
		// Handle JSON/YAML output
		outcomes := []*verifier.LegacyOutcome{}
		for _, e := range r.experimentsConfig {
			experiment := r.experiments[e.Metadata.Type]
			outcome, err := experiment.Verify(r.ctx, e)
			if err != nil {
				output.WriteFatal("Verifier %s failed: %s", e.Metadata.Name, err)
			}
			outcomes = append(outcomes, outcome)
		}

		structuredOutput := verifier.LegacyStructuredOutput{
			Results: outcomes,
		}
		switch strings.ToLower(outputFormat) {
		case "json":
			output.WriteJSON(structuredOutput)
		case "yaml":
			output.WriteYAML(structuredOutput)
		default:
			output.WriteError("Unknown output format: %s", outputFormat)
		}
		return
	}

	// Handle table output - show each test result as a separate row
	table := output.NewTable([]string{"Experiment", "Description", "Framework", "Tactic", "Technique", "Test", "Result"})

	for _, e := range r.experimentsConfig {
		experiment := r.experiments[e.Metadata.Type]
		outcome, err := experiment.Verify(r.ctx, e)
		if err != nil {
			output.WriteFatal("Verifier %s failed: %s", e.Metadata.Name, err)
		}

		// If there are no specific test results, show overall experiment result
		if len(outcome.Result) == 0 {
			table.AddRow([]string{
				outcome.Experiment,
				outcome.Description,
				outcome.Framework,
				outcome.Tactic,
				outcome.Technique,
				"Overall",
				"No results",
			})
		} else {
			// Show each individual test result as a separate row
			for testName, result := range outcome.Result {
				// Add status emoji for better visual feedback
				status := result
				if result == verifier.Success {
					status = "✓ " + result
				} else if result == verifier.Fail {
					status = "✗ " + result
				}

				table.AddRow([]string{
					outcome.Experiment,
					outcome.Description,
					outcome.Framework,
					outcome.Tactic,
					outcome.Technique,
					testName,
					status,
				})
			}
		}
	}

	table.Render()

	// Show summary
	r.printSummary()
}

// printSummary prints a summary of all experiment results
func (r *Runner) printSummary() {
	totalTests := 0
	passedTests := 0
	failedTests := 0

	for _, e := range r.experimentsConfig {
		experiment := r.experiments[e.Metadata.Type]
		outcome, err := experiment.Verify(r.ctx, e)
		if err != nil {
			continue
		}

		for _, result := range outcome.Result {
			totalTests++
			if result == verifier.Success {
				passedTests++
			} else {
				failedTests++
			}
		}
	}

	fmt.Printf("\nSummary: %d total tests, %d passed, %d failed\n", totalTests, passedTests, failedTests)

	if failedTests == 0 {
		output.WriteSuccess("All tests passed!")
	} else {
		output.WriteWarning("%d test(s) failed", failedTests)
	}
}

// Cleanup cleans up all experiments in the Runner
func (r *Runner) Cleanup() {
	for _, e := range r.experimentsConfig {
		output.WriteInfo("Cleaning up experiment %s", e.Metadata.Name)
		experiment := r.experiments[e.Metadata.Type]
		if err := experiment.Cleanup(r.ctx, e); err != nil {
			output.WriteError("Experiment %s cleanup failed: %s", e.Metadata.Name, err)
		}

	}
}
