package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/status"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
	"github.com/tuna-os/bluefin-cli/internal/update"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Open the interactive Bluefin main menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		registerPaletteActions()
		return app.Run(mainMenuScreen(), checkForUpdate)
	},
}

func init() {
	rootCmd.AddCommand(menuCmd)
}

// checkForUpdate runs in the background when the menu starts and surfaces a
// toast when a newer release exists. Failures are silent — an update nudge
// is never worth an error message.
func checkForUpdate() tea.Msg {
	rel, err := update.Latest()
	if err != nil || !update.IsNewer(version, rel.TagName) {
		return nil
	}
	hint := "bluefin-cli update"
	if h := update.Detect().UpdateHint(); h != "" {
		hint = h
	}
	return app.ToastMsg{Text: fmt.Sprintf("⬆ %s available — run: %s", rel.TagName, hint)}
}

func mainMenuScreen() app.Screen {
	return app.NewMenu("Home", nil, mainMenuItems, mainMenuSelect)
}

func mainMenuItems() []app.MenuItem {
	// Show exactly which shells have the experience enabled, not a vague
	// "Enabled" that may not match the user's current shell.
	var enabledShells []string
	for sh, enabled := range shell.CheckStatus() {
		if enabled {
			enabledShells = append(enabledShells, sh)
		}
	}
	sort.Strings(enabledShells)
	shellHint := "off"
	if len(enabledShells) > 0 {
		shellHint = strings.Join(enabledShells, " ") + " ✓"
	}

	items := []app.MenuItem{
		{Icon: "📊", Label: "Status", Value: "status", Desc: "Environment health at a glance"},
		{Icon: "🐚", Label: "Bluefin Shell", Value: "shell", Desc: "Aliases, prompt, and modern CLI tools", Hint: shellHint, Submenu: true},
		{Icon: "📦", Label: "Install Apps", Value: "bundles", Desc: "Curated tool bundles", Submenu: true},
	}
	items = append(items, extraMenuItems()...)
	items = append(items, app.MenuItem{Icon: "👋", Label: "Exit", Value: "exit", Desc: "Back to your shell"})
	return items
}

func mainMenuSelect(it app.MenuItem) tea.Cmd {
	switch it.Value {
	case "status":
		return app.Push(app.NewText("Status", status.Render))
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
			{Icon: "🔄", Label: toggle, Value: "toggle_current", Desc: "One switch for the whole experience"},
			{Icon: "🔧", Label: "Configure Components", Value: "components", Desc: "Pick which tools load with your shell", Submenu: true},
			{Icon: "📰", Label: "MOTD Settings", Value: "motd", Desc: "Message of the day on new terminals", Submenu: true},
			{Icon: "🐚", Label: "Other Shells", Value: "shells", Desc: "Enable for bash, zsh, or fish", Submenu: true},
			{Icon: "🎨", Label: "Advanced", Value: "advanced", Desc: "Fine-grained shell integration", Submenu: true},
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
			return app.Push(componentsFormScreen())
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

// componentsFormScreen hosts the shell-tools multiselect natively; the save
// is instant, and the (potentially slow, brew-driven) tool installation runs
// through the legacy bridge afterwards.
func componentsFormScreen() app.Screen {
	current := currentShellName()
	var selected []string

	build := func() *huh.Form {
		cfg, err := shell.LoadConfig(current)
		if err != nil {
			cfg = shell.DefaultConfig(current)
		}
		tools := shell.ToolsForShell(current)
		selected = selected[:0]
		options := make([]huh.Option[string], 0, len(tools))
		for _, tool := range tools {
			if cfg.IsEnabled(tool.Name) {
				selected = append(selected, tool.Name)
			}
			options = append(options,
				huh.NewOption(fmt.Sprintf("%s (%s)", tool.Name, tool.Description), tool.Name))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select tools to enable").
				Description("Uncheck to disable specific tools").
				Options(options...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.NewForm("Components", build, func(aborted bool) tea.Cmd {
		if aborted {
			return nil
		}
		newCfg := shell.DefaultConfig(current)
		on := make(map[string]bool, len(selected))
		for _, s := range selected {
			on[s] = true
		}
		for _, tool := range shell.ToolsForShell(current) {
			newCfg.SetEnabled(tool.Name, on[tool.Name])
		}
		if err := shell.SaveConfig(newCfg); err != nil {
			return app.Toast("Error: "+err.Error(), true)
		}
		return tea.Sequence(
			app.RunLegacy(func() error {
				shell.InstallTools(current, newCfg)
				return nil
			}),
			app.Toast("Components saved.", false),
		)
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
			out = append(out, app.MenuItem{Label: cat.Label, Value: cat.ID, Desc: cat.Desc, Submenu: true})
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
