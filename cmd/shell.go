package cmd

import (
	"fmt"
	"os"
	"strings"
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
		tui.RenderHeader("Bluefin CLI", "Main Menu > Shell")

		currentShellPath := os.Getenv("SHELL")
		currentShell := filepath.Base(currentShellPath)
		if currentShell == "" {
			currentShell = "bash" // fallback
		}

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
						huh.NewOption("Configure Components ❯", "components"),
						huh.NewOption("Enable/Disable for other shells ❯", "shells"),
						huh.NewOption("Exit to Main Menu", "exit"),
					).
					Value(&action),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap()).Run(); err != nil {
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
	tui.RenderHeader("Bluefin CLI", "Main Menu > Shell > Shells")

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
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap()).Run(); err != nil {
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
	tui.RenderHeader("Bluefin CLI", "Main Menu > Shell > Components")

	cfg, err := shell.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var selected []string
	for _, tool := range shell.Tools {
		if cfg.IsEnabled(tool.Name) {
			selected = append(selected, tool.Name)
		}
	}
	
	var options []huh.Option[string]
	for _, tool := range shell.Tools {
		label := fmt.Sprintf("%s (%s)", tool.Name, tool.Description)
		options = append(options, huh.NewOption(label, tool.Name))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select tools to enable").
				Description("Uncheck to disable specific tools").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	// Update config
	newCfg := shell.DefaultConfig() 
	// Create a set for selected tools
	selectedSet := make(map[string]bool)
	for _, s := range selected {
		selectedSet[s] = true
	}

	for _, tool := range shell.Tools {
		newCfg.SetEnabled(tool.Name, selectedSet[tool.Name])
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
	// Generate dynamic long description
	var sb strings.Builder
	sb.WriteString("Enable or disable shell experience enhancements (modern aliases and tool initialization).\n\nThe Shell Experience provides:\n")
	for _, tool := range shell.Tools {
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", tool.Name, tool.Description))
	}
	shellCmd.Long = sb.String()

	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellConfigCmd)
}

