package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/install"
)

var installCmd = &cobra.Command{
	Use:   "install [bundle]",
	Short: "Install Homebrew bundles",
	Long: `Install predefined Homebrew bundles or custom Brewfiles.

Available bundles:
  ai     - AI tools (Goose, Codex, Gemini, Ramalama, etc.)
  cli    - CLI essentials (gh, chezmoi, etc.)
  fonts  - Development fonts (Fira Code, JetBrains Mono, etc.)
  k8s    - Kubernetes tools (kubectl, k9s, kind, etc.)
  
Or provide a path to a local Brewfile.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Interactive mode
			var selectedBundles []string

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Select bundles to install (space to select, enter to confirm)").
						Options(
							huh.NewOption("🤖 AI Tools", "ai"),
							huh.NewOption("💻 CLI Essentials", "cli"),
							huh.NewOption("🔤 Development Fonts", "fonts"),
							huh.NewOption("☸️  Kubernetes Tools", "k8s"),
						).
						Value(&selectedBundles),
				),
			)

			if err := form.Run(); err != nil {
				return fmt.Errorf("form error: %w", err)
			}

			// Install each selected bundle
			for _, bundle := range selectedBundles {
				if err := install.Bundle(bundle); err != nil {
					return err
				}
			}
			return nil
		}

		return install.Bundle(args[0])
	},
}

var installListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available bundles",
	Long:  `Show all available Homebrew bundles with descriptions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		install.ListBundles()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(installListCmd)
}
