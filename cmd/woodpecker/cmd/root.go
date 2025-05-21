/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"github.com/operantai/woodpecker/internal/output"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "woodpecker",
	Short: "",
	Long:  "",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		output.WriteError("%s", err.Error())
	}
}
