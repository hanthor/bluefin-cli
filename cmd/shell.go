package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/hanthor/bluefin-cli/internal/tui"
)

var shellCmd = &cobra.Command{
	Use:   "shell [shell] [on|off]",
	Short: "Toggle shell experience enhancements",
	Long: `Enable or disable shell experience enhancements (modern aliases and tool initialization).
	
The Shell Experience provides:
  - Modern ls replacement with eza (ll, ls aliases)
  - bat for cat with syntax highlighting
  - ugrep for faster grep
  - Initialization for atuin, starship, and zoxide`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runShellMenu()
		}

		// Args provided
		selectedShell := args[0]
		enable := true // default to on
		if len(args) > 1 {
			enable = args[1] == "on"
		}

		return shell.Toggle(selectedShell, enable)
	},
}

var shellConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure individual shell experience tools",
	Long:  `Enable or disable specific shell experience components interactively.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return configureShellTools()
	},
}

func runShellMenu() error {
	for {
		tui.ClearScreen()
		tui.RenderHeader("Bluefin CLI", "Shell Configuration")

		// Detect current shell
		currentShellPath := os.Getenv("SHELL")
		currentShell := filepath.Base(currentShellPath)
		if currentShell == "" {
			currentShell = "bash" // fallback
		}

		// Check status for current shell
		status := shell.CheckStatus()
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
			if err := shell.Toggle(currentShell, !isEnabled); err != nil {
				return err
			}
			tui.Pause()
		case "shells":
			if err := shellShellsMenu(); err != nil {
				return err
			}
		case "components":
			if err := configureShellTools(); err != nil {
				return err
			}
		case "exit":
			return nil
		}
	}
}

func shellShellsMenu() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Shell > Shells")

	// Check current status
	status := shell.CheckStatus()
	
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
	for _, shName := range []string{"bash", "zsh", "fish"} {
		wasEnabled := initialSelected[shName]
		isEnabled := finalSelected[shName]
		
		// Only toggle if state changed
		if wasEnabled != isEnabled {
			if err := shell.Toggle(shName, isEnabled); err != nil {
				return err
			}
			tui.Pause()
		}
	}
	return nil
}

func configureShellTools() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Shell > Components")

	cfg, err := shell.LoadConfig()
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
	newCfg := &shell.Config{}
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

	if err := shell.SaveConfig(newCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Install any newly enabled tools
	shell.InstallTools(newCfg)

	fmt.Println(tui.SuccessStyle.Render("Configuration saved! Tools installed/updated."))
	tui.Pause()
	return nil
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellConfigCmd)
}

