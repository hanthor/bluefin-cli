package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/motd"
	"github.com/hanthor/bluefin-cli/internal/tui"
)

var motdCmd = &cobra.Command{
	Use:   "motd",
	Short: "Manage Message of the Day",
	Long:  `Configure and display the Message of the Day (MOTD) with system info and tips.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMotdMenu()
	},
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
		if len(args) > 0 {
			return motd.SetTheme(args[0])
		}

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
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
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

func runMotdMenu() error {
	for {
		var action string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("MOTD – What do you want to do?").
					Options(
						huh.NewOption("Show MOTD", "show"),
						huh.NewOption("Toggle for shells ❯", "toggle"),
					).
					Value(&action),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap()).Run(); err != nil {
			// Abort or ctrl+c -> go back/exit
			return nil
		}

		if action == "show" {
			if err := motd.Show(); err != nil {
				return err
			}
			continue // Show menu again after displaying MOTD
		}

		// Toggle mode - check current status
		status := motd.CheckStatus()
		
		// Pre-select shells that currently have MOTD enabled
		var selected []string
		for _, shell := range []string{"bash", "zsh", "fish"} {
			if status[shell] {
				selected = append(selected, shell)
			}
		}
		
		// Store initial state
		initialSelected := make(map[string]bool)
		for _, sh := range selected {
			initialSelected[sh] = true
		}
		
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Toggle MOTD for shells").
					Description("Selected = ON, Deselected = OFF (ctrl+c to cancel)").
					Options(
						huh.NewOption("bash", "bash"),
						huh.NewOption("zsh", "zsh"),
						huh.NewOption("fish", "fish"),
					).
					Value(&selected),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap()).Run(); err != nil {
			// Abort -> go back to motd menu
			continue
		}

		// Build map of final selections
		finalSelected := make(map[string]bool)
		for _, sh := range selected {
			finalSelected[sh] = true
		}

		// Apply changes for shells that changed state
		for _, shell := range []string{"bash", "zsh", "fish"} {
			wasEnabled := initialSelected[shell]
			isEnabled := finalSelected[shell]
			
			// Only toggle if state changed
			if wasEnabled != isEnabled {
				if err := motd.Toggle(shell, isEnabled); err != nil {
					return err
				}
			}
		}
		
		// After toggling, return to main menu
		return nil
	}
}
