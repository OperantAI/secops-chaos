/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/secops-chaos/internal/experiments"
	"github.com/operantai/secops-chaos/internal/output"
	"github.com/spf13/cobra"
)

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
		outputJSON, err := cmd.Flags().GetBool("json")
		if err != nil {
			output.WriteError("Error reading json output flag: %v", err)
		}

		// Run the verifiers
		ctx := cmd.Context()
		er := experiments.NewRunner(ctx, files)
		er.RunVerifiers(outputJSON)
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	// Define the path of the experiment file to run
	verifyCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	_ = verifyCmd.MarkFlagRequired("file")

	// Output the results in JSON format
	verifyCmd.Flags().BoolP("json", "j", false, "Output results in JSON format")
}
