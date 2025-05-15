/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"fmt"

	"github.com/operantai/woodpecker/internal/experiments"
	"github.com/operantai/woodpecker/internal/output"
	"github.com/operantai/woodpecker/internal/snippets"
	"github.com/spf13/cobra"
)

// experiment represents the experiment commands
var experimentCmd = &cobra.Command{
	Use:   "experiment",
	Short: "Interact with experiments",
	Long:  "Interact with experiments",
	Run: func(cmd *cobra.Command, args []string) {
		allExperiments := experiments.ListExperiments()
		table := output.NewTable([]string{"Type", "Description"})
		for experimentType, description := range allExperiments {
			table.AddRow([]string{experimentType, description})
		}
		table.Render()
	},
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an experiment",
	Long:  "Run an experiment",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := cmd.Flags().GetStringSlice("file")
		if err != nil {
			output.WriteError("Error reading file flag: %v", err)
		}

		// Run the experiment
		ctx := cmd.Context()
		er := experiments.NewRunner(ctx, files)
		er.Run()
	},
}

// snippetExperimentCmd outputs a template of a given experiment type
var snippetExperimentCmd = &cobra.Command{
	Use:   "snippet",
	Short: "Print a template of an experiment type to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		experiment, err := cmd.Flags().GetString("experiment")
		if err != nil {
			output.WriteError("Error reading experiment flag: %v", err)
		}
		snippet, err := snippets.GetExperimentTemplate(experiment)
		if err != nil {
			output.WriteFatal("Error retrieving experiment template: %v", err)
		}
		fmt.Println(string(snippet))
	},
}

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the outcome of an experiment",
	Long:  "Verify the outcome of an experiment",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := cmd.Flags().GetStringSlice("file")
		if err != nil {
			output.WriteError("Error reading file flag: %v", err)
		}
		outputFormat, err := cmd.Flags().GetString("output")
		if err != nil {
			output.WriteError("Error reading json output flag: %v", err)
		}

		// Run the verifiers
		ctx := cmd.Context()
		er := experiments.NewRunner(ctx, files)
		er.RunVerifiers(outputFormat)
	},
}

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up after an experiment run",
	Long:  "Clean up after an experiment run",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := cmd.Flags().GetStringSlice("file")
		if err != nil {
			output.WriteError("Error reading file flag: %v", err)
		}

		// Create a new experiment runner and clean up
		ctx := cmd.Context()
		er := experiments.NewRunner(ctx, files)
		er.Cleanup()
	},
}

func init() {
	rootCmd.AddCommand(experimentCmd)
	experimentCmd.AddCommand(runCmd)
	experimentCmd.AddCommand(verifyCmd)
	experimentCmd.AddCommand(cleanCmd)
	experimentCmd.AddCommand(snippetExperimentCmd)

	// Define the path of the experiment file to run
	runCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	_ = runCmd.MarkFlagRequired("file")

	verifyCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to verify")
	_ = verifyCmd.MarkFlagRequired("file")

	cleanCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	_ = cleanCmd.MarkFlagRequired("file")

	snippetExperimentCmd.Flags().StringP("experiment", "e", "", "Experiment to generate a template for")
	_ = snippetExperimentCmd.MarkFlagRequired("experiment")

	// Output the results in JSON format
	verifyCmd.Flags().StringP("output", "o", "", "Output results in the provided format (json|yaml)")
}
