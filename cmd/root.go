package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:   "bluefin-cli",
	Short: "A powerful CLI tool for managing Homebrew and shell customization",
	Long: `Bluefin CLI is a modern command-line tool that helps you manage:
  - Homebrew brewfiles
  - Shell configuration and setup
  - Starship theme customization
  - Development environment tools

Built with ❤️ using Charm TUI libraries.`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("bluefin-cli version %s\n", version))
}
