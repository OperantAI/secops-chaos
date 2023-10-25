/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/experiments-runtime-tool/internal/verifiers"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the outcome of an experiment",
	Long:  "Verify the outcome of an experiment",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		vr := verifiers.NewRunner(ctx, []string{""})
		vr.Run()
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	// Define the path of the experiment file to run
	verifyCmd.Flags().StringP("file", "f", "", "Experiment file to run")
	verifyCmd.MarkFlagRequired("file")

	// Define the namespace(s) to run the experiment in
	verifyCmd.Flags().StringP("namespace", "n", "", "Namespace to run experiment in")
	verifyCmd.Flags().StringP("all", "a", "", "Run experiment in all namespaces")
	verifyCmd.MarkFlagsMutuallyExclusive("namespace", "all")
}
