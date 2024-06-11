/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/secops-chaos/internal/experiments"
	"github.com/operantai/secops-chaos/internal/output"
	"github.com/spf13/cobra"
)

// experiment represents the experiment commands
var experimentCmd = &cobra.Command{
	Use:   "experiment",
	Short: "Interact with experiments",
	Long:  "Interact with experiments",
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

	// Define the path of the experiment file to run
	runCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	_ = runCmd.MarkFlagRequired("file")

	verifyCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to verify")
	_ = verifyCmd.MarkFlagRequired("file")

	cleanCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	_ = cleanCmd.MarkFlagRequired("file")

	// Output the results in JSON format
	verifyCmd.Flags().StringP("output", "o", "", "Output results in provided format (json|yaml)")
}
