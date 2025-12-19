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
  ai               - AI tools (Goose, Codex, Gemini, Ramalama, etc.)
  artwork          - Artwork and design tools.
  cli              - CLI essentials (gh, chezmoi, etc.)
  cncf             - Cloud Native Computing Foundation tools.
  experimental-ide - Experimental IDE tools.
  fonts            - Development fonts (Fira Code, JetBrains Mono, etc.)
  ide              - IDE tools: VS Code, JetBrains Toolbox, etc.
  k8s              - Kubernetes tools: kubectl, k9s, kubectx, etc.
  
Or provide a path to a local Brewfile.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runBundlesMenu()
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
	installCmd.AddCommand(installWallpapersCmd)
}

// Wallpapers: install casks from ublue-os/tap
var installWallpapersCmd = &cobra.Command{
	Use:   "wallpapers [cask...]",
	Short: "Install wallpaper casks from ublue-os/tap",
	Long:  "Install wallpapers published as Homebrew casks from the ublue-os/tap tap.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return install.InstallWallpaperCasks(args)
		}

		return runWallpapersMenu()
	},
}

func runBundlesMenu() error {
	var selectedBundles []string

	opts := []huh.Option[string]{
		huh.NewOption("🤖 AI Tools", "ai"),
		huh.NewOption("🎨 Artwork", "artwork"),
		huh.NewOption("💻 CLI Essentials", "cli"),
		huh.NewOption("☁️  CNCF Tools", "cncf"),
		huh.NewOption("🧪 Experimental IDE", "experimental-ide"),
		huh.NewOption("🔤 Development Fonts", "fonts"),
		huh.NewOption("📝 IDE Tools", "ide"),
		huh.NewOption("☸️  Kubernetes Tools", "k8s"),
	}

	// Conditionally add Full GNOME Desktop
	if install.IsLinux() && install.IsGnome() {
		opts = append(opts, huh.NewOption("🖥️  Full GNOME Desktop", "full-desktop"))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select bundles to install (space to select, enter to confirm)").
				Options(opts...).
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

func runWallpapersMenu() error {
	// Interactive mode: discover available casks and let user multi-select
	casks, err := install.GetWallpaperCasks()
	if err != nil {
		return fmt.Errorf("failed to discover wallpaper casks: %w", err)
	}
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks found in ublue-os/tap")
	}

	// Build options list
	opts := make([]huh.Option[string], 0, len(casks))
	for _, c := range casks {
		// Show pretty labels, value is the plain cask name
		opts = append(opts, huh.NewOption(c, c))
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select wallpapers to install (space to select, enter to confirm)").
				Options(opts...).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("form error: %w", err)
	}
	if len(selected) == 0 {
		return fmt.Errorf("no wallpapers selected")
	}
	return install.InstallWallpaperCasks(selected)
}
