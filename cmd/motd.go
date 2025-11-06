package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/motd"
)

var motdCmd = &cobra.Command{
	Use:   "motd",
	Short: "Manage Message of the Day",
	Long:  `Configure and display the Message of the Day (MOTD) with system info and tips.`,
}

var motdToggleCmd = &cobra.Command{
	Use:   "toggle [shell|all] [on|off]",
	Short: "Toggle MOTD for shells",
	Long:  `Enable or disable MOTD display on shell startup for bash, zsh, fish, or all shells.`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "all"
		enable := true

		if len(args) > 0 {
			target = args[0]
		}
		if len(args) > 1 {
			enable = args[1] == "on"
		}

		return motd.Toggle(target, enable)
	},
}

var motdShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display the MOTD",
	Long:  `Display the Message of the Day with system information and a random tip.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return motd.Show()
	},
}

var motdConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure MOTD settings",
	Long:  `Interactively configure MOTD theme and settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedTheme string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose MOTD theme").
					Options(
						huh.NewOption("Slate (default)", "slate"),
						huh.NewOption("Dark", "dark"),
						huh.NewOption("Light", "light"),
						huh.NewOption("Dracula", "dracula"),
						huh.NewOption("Pink", "pink"),
					).
					Value(&selectedTheme),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("form error: %w", err)
		}

		return motd.SetTheme(selectedTheme)
	},
}

func init() {
	rootCmd.AddCommand(motdCmd)
	motdCmd.AddCommand(motdToggleCmd)
	motdCmd.AddCommand(motdShowCmd)
	motdCmd.AddCommand(motdConfigCmd)
}
