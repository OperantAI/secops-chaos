/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/experiments-runtime-tool/internal/output"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "experiments-runtime-tool",
	Short: "tbd",
	Long:  "tbd",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		output.WriteError(err.Error())
	}
}
