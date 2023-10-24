/*
Copyright © 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/experiments-runtime-tool/internal/experiments"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an experiment",
	Long:  "Run an experiment",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		er := experiments.NewRunner(ctx, []string{""})
		er.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Define the path of the experiment file to run
	runCmd.Flags().StringP("file", "f", "", "Experiment file to run")
	runCmd.MarkFlagRequired("file")

	// Define the namespace(s) to run the experiment in
	runCmd.Flags().StringP("namespace", "n", "", "Namespace to run experiment in")
	runCmd.Flags().StringP("all", "a", "", "Run experiment in all namespaces")
	runCmd.MarkFlagsMutuallyExclusive("namespace", "all")
}
