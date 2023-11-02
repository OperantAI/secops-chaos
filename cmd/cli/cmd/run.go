/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/secops-chaos/internal/experiments"
	"github.com/operantai/secops-chaos/internal/output"
	"github.com/spf13/cobra"
)

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

func init() {
	rootCmd.AddCommand(runCmd)

	// Define the path of the experiment file to run
	runCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	_ = runCmd.MarkFlagRequired("file")
}
