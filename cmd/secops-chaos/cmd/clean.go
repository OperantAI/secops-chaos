/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/secops-chaos/internal/experiments"
	"github.com/operantai/secops-chaos/internal/output"
	"github.com/spf13/cobra"
)

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
	rootCmd.AddCommand(cleanCmd)

	// Define the path of the experiment file to run
	cleanCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	_ = cleanCmd.MarkFlagRequired("file")
}
