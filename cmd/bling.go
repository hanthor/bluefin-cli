package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/bling"
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
		// If no args, show interactive prompt
		if len(args) == 0 {
			var action string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Bling Configuration").
						Options(
							huh.NewOption("Enable/Disable Bling for Shells", "toggle"),
							huh.NewOption("Configure Tools", "config"),
						).
						Value(&action),
				),
			)

			if err := form.Run(); err != nil {
				return fmt.Errorf("form error: %w", err)
			}

			if action == "config" {
				return configureBlingTools()
			}

			// Continue to toggle flow
			var selectedShell string
			var enable bool

			form = huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Select shell").
						Options(
							huh.NewOption("Bash", "bash"),
							huh.NewOption("Zsh", "zsh"),
							huh.NewOption("Fish", "fish"),
						).
						Value(&selectedShell),
				),
				huh.NewGroup(
					huh.NewConfirm().
						Title("Enable bling?").
						Description("Enable modern shell aliases and tool initialization").
						Value(&enable),
				),
			)

			if err := form.Run(); err != nil {
				return fmt.Errorf("form error: %w", err)
			}
			
			return bling.Toggle(selectedShell, enable)
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

func configureBlingTools() error {
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
	)

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

	fmt.Println("Configuration saved! Changes will take effect on next shell start.")
	return nil
}

func init() {
	rootCmd.AddCommand(blingCmd)
}
