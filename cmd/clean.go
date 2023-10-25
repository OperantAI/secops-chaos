/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/experiments-runtime-tool/internal/experiments"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up after an experiment run",
	Long:  "Clean up after an experiment run",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		er := experiments.NewRunner(ctx, []string{""})
		er.Cleanup()
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	// Define the path of the experiment file to run
	cleanCmd.Flags().StringP("file", "f", "", "Experiment file to run")
	cleanCmd.MarkFlagRequired("file")

	// Define the namespace(s) to run the experiment in
	cleanCmd.Flags().StringP("namespace", "n", "", "Namespace to run experiment in")
	cleanCmd.Flags().StringP("all", "a", "", "Run experiment in all namespaces")
	cleanCmd.MarkFlagsMutuallyExclusive("namespace", "all")
}
