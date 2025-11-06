package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/yourusername/bluefin-cli/internal/shell"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Manage shell configuration",
	Long:  `Configure and customize your shell environment (bash, zsh, fish).`,
}

var shellSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive shell setup wizard",
	Long:  `Walk through an interactive setup process to configure your shell.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			selectedShell string
			installOhMyZsh bool
			setupStarship bool
		)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose your shell").
					Options(
						huh.NewOption("Bash", "bash"),
						huh.NewOption("Zsh", "zsh"),
						huh.NewOption("Fish", "fish"),
					).
					Value(&selectedShell),
			),
			huh.NewGroup(
				huh.NewConfirm().
					Title("Install Oh My Zsh?").
					Description("A delightful framework for managing your Zsh configuration").
					Value(&installOhMyZsh),
			).WithHideFunc(func() bool {
				return selectedShell != "zsh"
			}),
			huh.NewGroup(
				huh.NewConfirm().
					Title("Setup Starship prompt?").
					Description("The minimal, blazing-fast, and infinitely customizable prompt").
					Value(&setupStarship),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("form error: %w", err)
		}

		return shell.Setup(selectedShell, installOhMyZsh, setupStarship)
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellSetupCmd)
}
