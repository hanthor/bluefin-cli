package cmd

import (
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/brewfile"
)

var brewfileCmd = &cobra.Command{
	Use:   "brewfile",
	Short: "Manage Homebrew Brewfiles",
	Long:  `Create, edit, and apply Brewfile configurations for managing Homebrew packages.`,
}

var brewfileInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Brewfile",
	Long:  `Create a new (empty) Brewfile in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return brewfile.Initialize()
	},
}

var brewfileApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply Brewfile configuration",
	Long:  `Install all packages defined in your Brewfile using 'brew bundle'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return brewfile.Apply()
	},
}

var brewfileAddCmd = &cobra.Command{
	Use:   "add [package]",
	Short: "Add a package to your Brewfile",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return brewfile.AddPackage(args[0])
	},
}

func init() {
	rootCmd.AddCommand(brewfileCmd)
	brewfileCmd.AddCommand(brewfileInitCmd)
	brewfileCmd.AddCommand(brewfileApplyCmd)
	brewfileCmd.AddCommand(brewfileAddCmd)
}
