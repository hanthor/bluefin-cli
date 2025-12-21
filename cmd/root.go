package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "0.0.3"
)

var rootCmd = &cobra.Command{
	Use:     "bluefin-cli",
	Short:   "A powerful CLI tool for managing Homebrew and shell customization",
	Long:    `Bluefin CLI brings the bluefin terminal experience to you`,
	Version: version,
	// If no subcommand is provided, open the interactive main menu by default.
	RunE: func(cmd *cobra.Command, args []string) error {
		// Defer to the interactive menu
		if menuCmd != nil && menuCmd.RunE != nil {
			return menuCmd.RunE(menuCmd, nil)
		}
		// Fallback: show help if menu is not available for some reason
		return cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("bluefin-cli version %s\n", version))
}
