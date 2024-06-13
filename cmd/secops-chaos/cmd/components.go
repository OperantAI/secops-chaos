/*
Copyright 2023 Operant AI
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/operantai/secops-chaos/internal/components"
	"github.com/operantai/secops-chaos/internal/output"
	"github.com/operantai/secops-chaos/internal/snippets"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage secops-chaos optional components",
	Long:  "Manage secops-chaos optional components",
	Run: func(cmd *cobra.Command, args []string) {
		allComponents := components.ListComponents()
		table := output.NewTable([]string{"Type", "Description"})
		for componentType, description := range allComponents {
			table.AddRow([]string{componentType, description})
		}
		table.Render()
	},
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

var snippetComponentCmd = &cobra.Command{
	Use:   "snippet",
	Short: "Print a template of a component out to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		component, err := cmd.Flags().GetString("component")
		if err != nil {
			output.WriteError("Error reading component flag: %v", err)
		}
		snippet, err := snippets.RetrieveComponentTemplate(component)
		if err != nil {
			output.WriteFatal("Error retrieving component template: %s", err)
		}
		fmt.Println(string(snippet))
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
	componentCmd.AddCommand(snippetComponentCmd)

	// Define the path of the experiment file to run
	installComponentCmd.Flags().StringSliceP("files", "f", []string{}, "Component files to install")
	_ = installComponentCmd.MarkFlagRequired("files")

	uninstallComponentCmd.Flags().StringSliceP("files", "f", []string{}, "Component files to install")
	_ = uninstallComponentCmd.MarkFlagRequired("files")

	snippetComponentCmd.Flags().StringP("component", "c", "", "Component to generate a template of")
	_ = snippetComponentCmd.MarkFlagRequired("component")
}
