package cmd

import (
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/hanthor/bluefin-cli/internal/status"
	"github.com/hanthor/bluefin-cli/internal/tui"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Open the interactive Bluefin main menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		for {
			tui.ClearScreen()
			tui.RenderHeader("Bluefin CLI", "Main Menu")

			// Check status
			shellStatus := shell.CheckStatus()
			hasShell := false
			for _, v := range shellStatus {
				if v {
					hasShell = true
					break
				}
			}

			var shellLabel string
			if hasShell {
				shellLabel = "Bluefin Shell (Enabled)"
			} else {
				shellLabel = "Bluefin Shell (Disabled)"
			}

			// Build options dynamically, include OS scripts if available
			opts := []huh.Option[string]{
				huh.NewOption("📊 Status", "status"),
				huh.NewOption(shellLabel+" ❯", "shell"),
				huh.NewOption("📰 MOTD ❯", "motd"),
				huh.NewOption("📦 Install Tools ❯", "bundles"),
				huh.NewOption("🖼  Wallpapers ❯", "wallpapers"),
				huh.NewOption("🚀 Starship Theme ❯", "starship"),
			}
			opts = append(opts, huh.NewOption("Exit", "exit"))

			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Choose an action").
						Options(opts...).
						Value(&choice),
				),
			).WithTheme(tui.AppTheme)

			if err := form.WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap()).Run(); err != nil {
				// ESC pressed on main menu - exit cleanly
				return nil
			}

			switch choice {
			case "status":
				if err := status.Show(); err != nil {
					return err
				}
				tui.Pause()
			case "shell":
				if err := runShellMenu(); err != nil {
					return err
				}
			case "motd":
				if err := runMotdMenu(); err != nil {
					return err
				}
			case "bundles":
				if err := runBundlesMenu(); err != nil {
					return err
				}
			case "wallpapers":
				if err := runWallpapersMenu(); err != nil {
					return err
				}
			case "starship":
				if err := runStarshipMenu(); err != nil {
					return err
				}
			case "exit":
				return nil
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(menuCmd)
}


