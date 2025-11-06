package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/starship"
)

var starshipCmd = &cobra.Command{
	Use:   "starship",
	Short: "Manage Starship prompt themes",
	Long:  `Install, configure, and customize Starship prompt themes.`,
}

var starshipThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Select and apply a Starship theme",
	Long:  `Choose from popular Starship preset themes interactively.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedTheme string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose a Starship theme").
					Description("Select a preset theme for your terminal prompt").
					Options(
						huh.NewOption("Nerd Font Symbols", "nerd-font-symbols"),
						huh.NewOption("No Runtime Versions", "no-runtime-versions"),
						huh.NewOption("Plain Text Symbols", "plain-text-symbols"),
						huh.NewOption("Pure Preset", "pure-preset"),
						huh.NewOption("Tokyo Night", "tokyo-night"),
						huh.NewOption("Gruvbox Rainbow", "gruvbox-rainbow"),
						huh.NewOption("Catppuccin Powerline", "catppuccin-powerline"),
						huh.NewOption("Jetpack", "jetpack"),
						huh.NewOption("No Empty Icons", "no-empty-icons"),
						huh.NewOption("No Nerd Font", "no-nerd-font"),
						huh.NewOption("Pastel Powerline", "pastel-powerline"),
					).
					Value(&selectedTheme),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("form error: %w", err)
		}

		return starship.ApplyTheme(selectedTheme)
	},
}

var starshipInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Starship prompt",
	Long:  `Download and install the Starship prompt if not already installed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return starship.Install()
	},
}

func init() {
	rootCmd.AddCommand(starshipCmd)
	starshipCmd.AddCommand(starshipThemeCmd)
	starshipCmd.AddCommand(starshipInstallCmd)
}
