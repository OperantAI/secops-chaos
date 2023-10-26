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
		// Read the flags
		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			output.WriteError("Error reading namespace flag: %v", err)
		}
		allNamespaces, err := cmd.Flags().GetBool("all")
		if err != nil {
			output.WriteError("Error reading all flag: %v", err)
		}
		files, err := cmd.Flags().GetStringSlice("file")
		if err != nil {
			output.WriteError("Error reading file flag: %v", err)
		}

		// Run the experiment
		ctx := cmd.Context()
		er := experiments.NewRunner(ctx, namespace, allNamespaces, files)
		er.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Define the path of the experiment file to run
	runCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	runCmd.MarkFlagRequired("file")

	// Define the namespace(s) to run the experiment in
	runCmd.Flags().StringP("namespace", "n", "", "Namespace to run experiment in")
	runCmd.Flags().BoolP("all", "a", false, "Run experiment in all namespaces")
	runCmd.MarkFlagsMutuallyExclusive("namespace", "all")
}
