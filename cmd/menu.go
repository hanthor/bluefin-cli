package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/status"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Open the interactive Bluefin main menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		registerPaletteActions()
		return app.Run(mainMenuScreen())
	},
}

func init() {
	rootCmd.AddCommand(menuCmd)
}

func mainMenuScreen() app.Screen {
	return app.NewMenu("Home", nil, mainMenuItems, mainMenuSelect)
}

func mainMenuItems() []app.MenuItem {
	shellHint := "Disabled"
	for _, enabled := range shell.CheckStatus() {
		if enabled {
			shellHint = "Enabled"
			break
		}
	}

	items := []app.MenuItem{
		{Icon: "📊", Label: "Status", Value: "status"},
		{Icon: "🐚", Label: "Bluefin Shell", Value: "shell", Hint: shellHint, Submenu: true},
		{Icon: "📦", Label: "Install Apps", Value: "bundles", Submenu: true},
	}
	items = append(items, extraMenuItems()...)
	items = append(items, app.MenuItem{Icon: "👋", Label: "Exit", Value: "exit"})
	return items
}

func mainMenuSelect(it app.MenuItem) tea.Cmd {
	switch it.Value {
	case "status":
		return app.RunLegacy(func() error {
			tui.ClearScreen()
			if err := status.Show(); err != nil {
				return err
			}
			tui.Pause()
			return nil
		})
	case "shell":
		return app.Push(shellMenuScreen())
	case "bundles":
		return app.Push(bundlesMenuScreen())
	case "exit":
		return app.Pop()
	default:
		return extraMenuDo(it.Value)
	}
}

func currentShellName() string {
	name := filepath.Base(os.Getenv("SHELL"))
	if name == "" || name == "." {
		if env.IsWindows() {
			return "powershell"
		}
		return "bash"
	}
	return name
}

func shellMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		current := currentShellName()
		toggle := fmt.Sprintf("Enable for current shell (%s)", current)
		if shell.CheckStatus()[current] {
			toggle = fmt.Sprintf("Disable for current shell (%s)", current)
		}
		return []app.MenuItem{
			{Icon: "🔄", Label: toggle, Value: "toggle_current"},
			{Icon: "🔧", Label: "Configure Components", Value: "components", Submenu: true},
			{Icon: "📰", Label: "MOTD Settings", Value: "motd", Submenu: true},
			{Icon: "🐚", Label: "Other Shells", Value: "shells", Submenu: true},
			{Icon: "🎨", Label: "Advanced", Value: "advanced", Submenu: true},
		}
	}
	return app.NewMenu("Shell", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "toggle_current":
			current := currentShellName()
			enabled := shell.CheckStatus()[current]
			return app.RunLegacy(func() error {
				return shell.Toggle(current, !enabled)
			})
		case "components":
			return app.RunLegacy(configureShellTools)
		case "motd":
			return app.RunLegacy(runMotdMenu)
		case "shells":
			return app.RunLegacy(shellShellsMenu)
		case "advanced":
			return app.RunLegacy(runAdvancedMenu)
		}
		return nil
	})
}

func availableBundleCategories() []bundleCategory {
	out := make([]bundleCategory, 0, len(bundleCategories))
	for _, cat := range bundleCategories {
		if cat.LinuxOnly && (!install.IsLinux() || !install.IsGnome()) {
			continue
		}
		out = append(out, cat)
	}
	return out
}

func bundlesMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		cats := availableBundleCategories()
		out := make([]app.MenuItem, 0, len(cats))
		for _, cat := range cats {
			out = append(out, app.MenuItem{Label: cat.Label, Value: cat.ID, Submenu: true})
		}
		return out
	}
	return app.NewMenu("Install Apps", nil, items, func(it app.MenuItem) tea.Cmd {
		id := it.Value
		return app.RunLegacy(func() error { return runPackageMenu(id) })
	})
}

var paletteOnce sync.Once

// registerPaletteActions exposes every menu destination to the ctrl+p
// command palette.
func registerPaletteActions() {
	paletteOnce.Do(func() {
		app.Register(app.Action{
			ID: "status", Icon: "📊", Label: "Show Status", Section: "Home",
			Do: func() tea.Cmd { return mainMenuSelect(app.MenuItem{Value: "status"}) },
		})
		for _, cat := range availableBundleCategories() {
			id := cat.ID
			app.Register(app.Action{
				ID: "install-" + id, Label: cat.Label, Section: "Install",
				Do: func() tea.Cmd {
					return app.RunLegacy(func() error { return runPackageMenu(id) })
				},
			})
		}
		for _, it := range extraMenuItems() {
			it := it
			app.Register(app.Action{
				ID: it.Value, Icon: it.Icon, Label: it.Label, Section: "Customize",
				Do: func() tea.Cmd { return extraMenuDo(it.Value) },
			})
		}
	})
}
