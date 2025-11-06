package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/yourusername/bluefin-cli/internal/bling"
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
		var selectedShell string
		var enable bool

		// If no args, show interactive prompt
		if len(args) == 0 {
			form := huh.NewForm(
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
		} else {
			selectedShell = args[0]
			enable = true // default to on
			if len(args) > 1 {
				enable = args[1] == "on"
			}
		}

		return bling.Toggle(selectedShell, enable)
	},
}

func init() {
	rootCmd.AddCommand(blingCmd)
}
