package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/bling"
	"github.com/hanthor/bluefin-cli/internal/tui"
)

var blingCmd = &cobra.Command{
	Use:   "bling [shell] [on|off]",
	Short: "Toggle bling shell enhancements",
	Long: `Enable or disable bling shell enhancements (modern aliases and tool initialization).
	
Bling provides:
  - Modern ls replacement with eza (ll, ls aliases)
  - bat for cat with syntax highlighting
  - ugrep for faster grep
  - Initialization for atuin, starship, and zoxide`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runBlingMenu()
		}

		// Args provided
		selectedShell := args[0]
		enable := true // default to on
		if len(args) > 1 {
			enable = args[1] == "on"
		}

		return bling.Toggle(selectedShell, enable)
	},
}

var blingConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure individual bling tools",
	Long:  `Enable or disable specific bling components interactively.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return configureBlingTools()
	},
}

func runBlingMenu() error {
	for {
		tui.ClearScreen()
		tui.RenderHeader("Bluefin CLI", "Bling Configuration")

		// Detect current shell
		currentShellPath := os.Getenv("SHELL")
		currentShell := filepath.Base(currentShellPath)
		if currentShell == "" {
			currentShell = "bash" // fallback
		}

		// Check status for current shell
		status := bling.CheckStatus()
		isEnabled := status[currentShell]
		toggleLabel := fmt.Sprintf("Enable for current shell (%s)", currentShell)
		if isEnabled {
			toggleLabel = fmt.Sprintf("Disable for current shell (%s)", currentShell)
		}

		var action string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose an option").
					Options(
						huh.NewOption(toggleLabel, "toggle_current"),
						huh.NewOption("Configure Components", "components"),
						huh.NewOption("Enable/Disable for other shells", "shells"),
						huh.NewOption("Exit to Main Menu", "exit"),
					).
					Value(&action),
			),
		).WithTheme(tui.AppTheme).Run(); err != nil {
			return nil
		}

		switch action {
		case "toggle_current":
			if err := bling.Toggle(currentShell, !isEnabled); err != nil {
				return err
			}
			tui.Pause()
		case "shells":
			if err := blingShellsMenu(); err != nil {
				return err
			}
		case "components":
			if err := configureBlingTools(); err != nil {
				return err
			}
		case "exit":
			return nil
		}
	}
}

func blingShellsMenu() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Bling > Shells")

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
				Title("Manage other shells").
				Description("Selected = ON, Deselected = OFF").
				Options(
					huh.NewOption("bash", "bash"),
					huh.NewOption("zsh", "zsh"),
					huh.NewOption("fish", "fish"),
				).
				Value(&selected),
		),
	).WithTheme(tui.AppTheme).Run(); err != nil {
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
			tui.Pause()
		}
	}
	return nil
}

func configureBlingTools() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Bling > Components")

	cfg, err := bling.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var selected []string
	if cfg.Eza { selected = append(selected, "eza") }
	if cfg.Ugrep { selected = append(selected, "ugrep") }
	if cfg.Bat { selected = append(selected, "bat") }
	if cfg.Atuin { selected = append(selected, "atuin") }
	if cfg.Starship { selected = append(selected, "starship") }
	if cfg.Zoxide { selected = append(selected, "zoxide") }
	if cfg.Uutils { selected = append(selected, "uutils") }

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select tools to enable").
				Description("Uncheck to disable specific tools").
				Options(
					huh.NewOption("eza (Modern ls)", "eza"),
					huh.NewOption("ugrep (Faster grep)", "ugrep"),
					huh.NewOption("bat (Better cat)", "bat"),
					huh.NewOption("atuin (Shell history)", "atuin"),
					huh.NewOption("starship (Prompt)", "starship"),
					huh.NewOption("zoxide (Smarter cd)", "zoxide"),
					huh.NewOption("uutils (Rust Coreutils)", "uutils"),
				).
				Value(&selected),
		),
	).WithTheme(tui.AppTheme)

	if err := form.Run(); err != nil {
		return fmt.Errorf("form error: %w", err)
	}

	// Update config
	newCfg := &bling.Config{}
	for _, tool := range selected {
		switch tool {
		case "eza": newCfg.Eza = true
		case "ugrep": newCfg.Ugrep = true
		case "bat": newCfg.Bat = true
		case "atuin": newCfg.Atuin = true
		case "starship": newCfg.Starship = true
		case "zoxide": newCfg.Zoxide = true
		case "uutils": newCfg.Uutils = true
		}
	}

	if err := bling.SaveConfig(newCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Install any newly enabled tools
	bling.InstallTools(newCfg)

	fmt.Println(tui.SuccessStyle.Render("Configuration saved! Tools installed/updated."))
	tui.Pause()
	return nil
}

func init() {
	rootCmd.AddCommand(blingCmd)
	blingCmd.AddCommand(blingConfigCmd)
}

