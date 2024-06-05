/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"strings"

	"github.com/operantai/secops-chaos/internal/components"
	"github.com/operantai/secops-chaos/internal/output"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage secops-chaos optional components",
	Long:  "Manage secops-chaos optional components",
}

var installComponentCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a component",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := cmd.Flags().GetStringSlice("files")
		if err != nil {
			output.WriteError("Error reading file flag: %v", err)
		}

		ctx := cmd.Context()
		comp := components.New(ctx)
		if err := comp.Add(files); err != nil {
			output.WriteError("Error installing components %s: %v", strings.Join(files, ","), err)
		}
	},
}

var uninstallComponentCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall a component",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := cmd.Flags().GetStringSlice("files")
		if err != nil {
			output.WriteError("Error reading file flag: %v", err)
		}

		ctx := cmd.Context()
		comp := components.New(ctx)
		if err := comp.Remove(files); err != nil {
			output.WriteError("Error uninstalling components %s: %v", strings.Join(files, ","), err)
		}
	},
}

func init() {
	rootCmd.AddCommand(componentCmd)
	componentCmd.AddCommand(installComponentCmd)
	componentCmd.AddCommand(uninstallComponentCmd)

	// Define the path of the experiment file to run
	installComponentCmd.Flags().StringSliceP("files", "f", []string{}, "Component files to install")
	_ = installComponentCmd.MarkFlagRequired("files")

	uninstallComponentCmd.Flags().StringSliceP("files", "f", []string{}, "Component files to install")
	_ = uninstallComponentCmd.MarkFlagRequired("files")
}
