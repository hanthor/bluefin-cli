package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/bling"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/motd"
	"github.com/hanthor/bluefin-cli/internal/starship"
	"github.com/hanthor/bluefin-cli/internal/status"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Open the interactive Bluefin main menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		for {
			// Build options dynamically, include OS scripts if available
			opts := []huh.Option[string]{
				huh.NewOption("📊 Status", "status"),
				huh.NewOption("✨ Bling", "bling"),
				huh.NewOption("📰 MOTD", "motd"),
				huh.NewOption("📦 Install Tools", "bundles"),
				huh.NewOption("🖼  Wallpapers", "wallpapers"),
				huh.NewOption(" Starship Theme", "starship"),
			}
			if osscriptsAvailable() {
				opts = append(opts, huh.NewOption("⚙️  OS Scripts", "osscripts"))
			}
			opts = append(opts, huh.NewOption("Exit", "exit"))

			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Bluefin CLI – Main Menu").
						Description("Choose an action").
						Options(opts...).
						Value(&choice),
				),
			)
			if err := form.Run(); err != nil {
				// ESC pressed on main menu - exit cleanly
				return nil
			}

			switch choice {
			case "status":
				if err := status.Show(); err != nil {
					return err
				}
			case "bling":
				if err := blingMenu(); err != nil {
					return err
				}
			case "motd":
				if err := motdMenu(); err != nil {
					return err
				}
			case "bundles":
				if err := bundlesMenu(); err != nil {
					return err
				}
			case "wallpapers":
				if err := wallpapersMenu(); err != nil {
					return err
				}
			case "starship":
				if err := starshipMenu(); err != nil {
					return err
				}
			case "osscripts":
				if err := osscriptsMenu(); err != nil {
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

func blingMenu() error {
	// Check current status
	status := bling.CheckStatus()
	
	// Pre-select shells that currently have bling enabled
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
				Title("Toggle Bling for shells").
				Description("Selected = ON, Deselected = OFF (ctrl+c to cancel)").
				Options(
					huh.NewOption("bash", "bash"),
					huh.NewOption("zsh", "zsh"),
					huh.NewOption("fish", "fish"),
				).
				Value(&selected),
		),
	).Run(); err != nil {
		return nil // Interrupted - go back to main menu
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
			if err := bling.Toggle(shell, isEnabled); err != nil {
				return err
			}
		}
	}
	return nil
}

func motdMenu() error {
	for {
		var action string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("MOTD – What do you want to do?").
					Options(
						huh.NewOption("Show MOTD", "show"),
						huh.NewOption("Toggle for shells", "toggle"),
					).
					Value(&action),
			),
		).Run(); err != nil {
			return nil // Ctrl+C pressed - go back to main menu
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
		).Run(); err != nil {
			continue // Ctrl+C pressed - show MOTD menu again
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

func bundlesMenu() error {
	for {
		// Reuse the options from install command
		var selected []string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select tools to install (space to select, enter to confirm)").
					Description("Or press ctrl+c to go back").
					Options(
						huh.NewOption("🤖 AI Tools", "ai"),
						huh.NewOption("💻 CLI Essentials", "cli"),
						huh.NewOption("🔤 Development Fonts", "fonts"),
						huh.NewOption("☸️  Kubernetes Tools", "k8s"),
					).
					Value(&selected),
			),
		)
		if err := form.Run(); err != nil {
			// Ctrl+C or cancel pressed - go back to main menu
			return nil
		}
		
		// If nothing selected, go back to main menu
		if len(selected) == 0 {
			return nil
		}
		
		// Install selected bundles
		for _, b := range selected {
			if err := install.Bundle(b); err != nil {
				return err
			}
		}
		
		// After installation, return to main menu
		return nil
	}
}

func wallpapersMenu() error {
	casks, err := install.GetWallpaperCasks()
	if err != nil {
		return err
	}
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks found")
	}
	opts := make([]huh.Option[string], 0, len(casks))
	for _, c := range casks {
		opts = append(opts, huh.NewOption(c, c))
	}
	var sel []string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select wallpapers to install").
				Options(opts...).
				Value(&sel),
		),
	).Run(); err != nil {
		return err
	}
	if len(sel) == 0 {
		return fmt.Errorf("no wallpapers selected")
	}
	return install.InstallWallpaperCasks(sel)
}

func starshipMenu() error {
	// Ensure Starship is installed
	if err := starship.Install(); err != nil {
		return err
	}
	// Open the existing theme selector UI (same as `bluefin-cli starship theme`)
	if starshipThemeCmd != nil && starshipThemeCmd.RunE != nil {
		return starshipThemeCmd.RunE(starshipThemeCmd, nil)
	}
	return fmt.Errorf("theme selector unavailable")
}
