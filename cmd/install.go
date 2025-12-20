package cmd

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/tui"
)

var installCmd = &cobra.Command{
	Use:   "install [bundle]",
	Short: "Install Homebrew bundles",
	Long: `Install predefined Homebrew bundles or custom Brewfiles.

Available bundles:
  ai               - AI tools (Goose, Codex, Gemini, Ramalama, etc.)
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

	for {
		tui.ClearScreen()
		tui.RenderHeader("Bluefin CLI", "Main Menu > Install Apps")
		// Reset selection
		selectedBundles = []string{}

		opts := []huh.Option[string]{
			huh.NewOption("🤖 AI Tools", "ai"),
			huh.NewOption("💻 CLI Essentials", "cli"),
			huh.NewOption("☁️  CNCF Tools", "cncf"),
			huh.NewOption("🧪 Experimental IDE", "experimental-ide"),
			huh.NewOption("🔤 Development Fonts", "fonts"),
			huh.NewOption("📝 IDE Tools", "ide"),
			huh.NewOption("☸️  Kubernetes Tools", "k8s"),
		}

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
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				return nil
			}
			return fmt.Errorf("form error: %w", err)
		}

		if len(selectedBundles) > 0 {
			break
		}

		// Popup error message
		msg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Render("No Selection") + "\n\n" +
			"You must select at least one bundle to install.\nUse Space to select items."

		fmt.Println()
		fmt.Println(tui.PopupStyle.Render(msg))
		fmt.Println()
		
		time.Sleep(3 * time.Second)
	}

	var brewfiles []string
	var cleanups []func()

	defer func() {
		for _, c := range cleanups {
			c()
		}
	}()

	for _, bundle := range selectedBundles {
		path, cleanup, err := install.GetBrewfile(bundle)
		if err != nil {
			return err
		}
		brewfiles = append(brewfiles, path)
		cleanups = append(cleanups, cleanup)
	}

	if len(brewfiles) > 0 {
		if err := install.EnsureBbrew(); err != nil {
			return err
		}

		var finalPath string
		if len(brewfiles) > 1 {
			mergedPath, cleanup, err := install.MergeBrewfiles(brewfiles)
			if err != nil {
				return err
			}
			cleanups = append(cleanups, cleanup)
			finalPath = mergedPath
			fmt.Println(tui.InfoStyle.Render("🍺 Merged Brewfiles into single view..."))
		} else {
			finalPath = brewfiles[0]
		}

		fmt.Println(tui.InfoStyle.Render(fmt.Sprintf("🍺 Opening apps in bbrew...")))
		if err := install.RunBbrew(finalPath); err != nil {
			return err
		}
	}

	return nil
}

func runWallpapersMenu() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Main Menu > Wallpapers")
	casks, err := install.GetWallpaperCasks()
	if err != nil {
		return fmt.Errorf("failed to discover wallpaper casks: %w", err)
	}
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks found in ublue-os/tap")
	}

	opts := make([]huh.Option[string], 0, len(casks))
	for _, c := range casks {
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
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}
	if len(selected) == 0 {
		return fmt.Errorf("no wallpapers selected")
	}
	return install.InstallWallpaperCasks(selected)
}
