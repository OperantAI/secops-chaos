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

		// Run the verifiers
		ctx := cmd.Context()
		er := experiments.NewRunner(ctx, namespace, allNamespaces, files)
		er.RunVerifiers()
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	// Define the path of the experiment file to run
	verifyCmd.Flags().StringSliceP("file", "f", []string{}, "Experiment file(s) to run")
	verifyCmd.MarkFlagRequired("file")

	// Define the namespace(s) to run the experiment in
	verifyCmd.Flags().StringP("namespace", "n", "", "Namespace to run experiment in")
	verifyCmd.Flags().BoolP("all", "a", false, "Run experiment in all namespaces")
	verifyCmd.MarkFlagsMutuallyExclusive("namespace", "all")
}