package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
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
// is never worth an error message — and the GitHub API is only asked once
// per day.
func checkForUpdate() tea.Msg {
	if last := viper.GetInt64("update.last_check"); time.Since(time.Unix(last, 0)) < 24*time.Hour {
		return nil
	}
	viper.Set("update.last_check", time.Now().Unix())
	_ = config.Save()
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

// launchFlow opens the interactive shell and immediately runs flow — used
// by commands like `bluefin-cli fonts` so every interactive path shares the
// native TUI (esc leads back to the Home menu).
func launchFlow(flow tea.Cmd) error {
	registerPaletteActions()
	return app.Run(mainMenuScreen(), flow, checkForUpdate)
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
		{Icon: "🩺", Label: "Doctor", Value: "doctor", Desc: "Diagnose setup problems with fix hints"},
		{Icon: "🐚", Label: "Bluefin Shell", Value: "shell", Desc: "Aliases, prompt, and modern CLI tools", Hint: shellHint, Submenu: true},
		{Icon: "📦", Label: "Install Apps", Value: "bundles", Desc: "Curated tool bundles", Submenu: true},
		{Icon: "👻", Label: "Terminal Setup", Value: "terminal", Desc: "Ghostty install, Dock pin, themed config", Submenu: true},
	}
	items = append(items, extraMenuItems()...)
	items = append(items, app.MenuItem{Icon: "👋", Label: "Exit", Value: "exit", Desc: "Back to your shell"})
	return items
}

func mainMenuSelect(it app.MenuItem) tea.Cmd {
	switch it.Value {
	case "status":
		return app.Push(app.NewText("Status", status.Render))
	case "doctor":
		return doctorScreenCmd()
	case "shell":
		return app.Push(shellMenuScreen())
	case "bundles":
		return app.Push(bundlesMenuScreen())
	case "terminal":
		return app.Push(terminalMenuScreen())
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
			// Toggling can trigger brew installs (e.g. bash-preexec), which
			// print — so it must run in a RunnerScreen, never inline.
			current := currentShellName()
			enabled := shell.CheckStatus()[current]
			verb := "Enabling"
			if enabled {
				verb = "Disabling"
			}
			return app.Push(app.NewRunner(verb+" "+current, func() error {
				return shell.Toggle(current, !enabled)
			}))
		case "components":
			return app.Push(componentsFormScreen())
		case "motd":
			return app.Push(motdMenuScreen())
		case "shells":
			return app.Push(shellsFormScreen())
		case "advanced":
			return app.Push(advancedMenuScreen())
		}
		return nil
	})
}

// componentsFormScreen hosts the shell-tools multiselect natively; the save
// is instant, and the brew-driven tool installation runs in a RunnerScreen.
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
			app.Push(app.NewRunner("Installing tools", func() error {
				shell.InstallTools(current, newCfg)
				return nil
			})),
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
	withCustom := func() []app.MenuItem {
		out := items()
		for _, name := range install.CustomBundles() {
			path, err := install.CustomBundlePath(name)
			if err != nil {
				continue
			}
			out = append(out, app.MenuItem{Icon: "🧰", Label: name, Value: "file:" + path,
				Desc: "Your bundle (~/.config/bluefin-cli/bundles)", Submenu: true})
		}
		// The `brew bundle` convention: a Brewfile in the home directory
		// becomes a managed package set too.
		home, _ := os.UserHomeDir()
		for _, bf := range []string{"Brewfile", ".Brewfile"} {
			p := filepath.Join(home, bf)
			if _, err := os.Stat(p); err == nil {
				out = append(out, app.MenuItem{Icon: "🏠", Label: "~/" + bf, Value: "file:" + p,
					Desc: "Your home Brewfile — manage its packages", Submenu: true})
			}
		}
		return out
	}
	return app.NewMenu("Install Apps", nil, withCustom, func(it app.MenuItem) tea.Cmd {
		if path, ok := strings.CutPrefix(it.Value, "file:"); ok {
			return brewfileFlow(path, it.Label)
		}
		return packagesFlow(it.Value, it.Label)
	})
}

// brewfileFlow opens any on-disk Brewfile as a managed multiselect —
// installed packages pre-checked, uncheck to uninstall, same diff+confirm
// flow as the curated bundles.
func brewfileFlow(path, label string) tea.Cmd {
	return tea.Batch(
		app.Toast("Reading "+label+"…", false),
		func() tea.Msg {
			pkgs, err := install.GetBrewfilePackages(path)
			if err != nil {
				return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
			}
			if len(pkgs) == 0 {
				return app.ToastMsg{Text: label + " has no brew/cask entries.", IsErr: true}
			}
			pkgs = install.MarkInstalled(pkgs)
			return app.PushMsg{Screen: packagesFormScreen(label, pkgs)}
		},
	)
}

// packagesFlow loads a bundle in the background (the menu stays live), then
// pushes a native multiselect; confirmed changes run in a native RunnerScreen
// that captures brew output.
func packagesFlow(id, label string) tea.Cmd {
	return tea.Batch(
		app.Toast("Loading "+label+"…", false),
		func() tea.Msg {
			pkgs, err := install.GetBundlePackages(id)
			if err != nil {
				return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
			}
			pkgs = install.MarkInstalled(pkgs)
			return app.PushMsg{Screen: packagesFormScreen(label, pkgs)}
		},
	)
}

func packagesFormScreen(label string, pkgs []install.Package) app.Screen {
	var selected []string

	build := func() *huh.Form {
		selected = selected[:0]
		opts := make([]huh.Option[string], 0, len(pkgs))
		for _, p := range pkgs {
			if p.Installed {
				selected = append(selected, p.ID)
			}
			opts = append(opts, huh.NewOption(p.Name, p.ID))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select packages").
				Description("Pre-checked = already installed. Space toggles, enter confirms.").
				Options(opts...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.NewForm(label, build, func(aborted bool) tea.Cmd {
		if aborted {
			return nil
		}
		on := make(map[string]bool, len(selected))
		for _, id := range selected {
			on[id] = true
		}
		var toInstall, toRemove []install.Package
		for _, p := range pkgs {
			switch {
			case on[p.ID] && !p.Installed:
				toInstall = append(toInstall, p)
			case !on[p.ID] && p.Installed:
				toRemove = append(toRemove, p)
			}
		}
		if len(toInstall) == 0 && len(toRemove) == 0 {
			return app.Toast("No changes selected.", false)
		}
		return app.Push(confirmChangesScreen(toInstall, toRemove))
	})
}

func confirmChangesScreen(toInstall, toRemove []install.Package) app.Screen {
	var lines []string
	for _, p := range toInstall {
		lines = append(lines, "+ "+p.Name)
	}
	for _, p := range toRemove {
		lines = append(lines, "- "+p.Name)
	}
	confirmed := false

	build := func() *huh.Form {
		confirmed = false
		return huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Apply these changes?").
				Description(strings.Join(lines, "\n")).
				Value(&confirmed),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.ConfirmKeyMap())
	}

	return app.NewForm("Confirm", build, func(aborted bool) tea.Cmd {
		if aborted || !confirmed {
			return nil
		}
		return app.Push(app.NewRunner("Applying changes", func() error {
			applyPackageChanges(toInstall, toRemove)
			return nil
		}))
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
		app.Register(app.Action{
			ID: "dino", Icon: "🦕", Label: "Dino Run", Section: "Fun",
			Do: func() tea.Cmd { return app.Push(gameScreen()) },
		})
		app.Register(app.Action{
			ID: "doctor", Icon: "🩺", Label: "Doctor", Section: "Home",
			Do: func() tea.Cmd { return doctorScreenCmd() },
		})
		app.Register(app.Action{
			ID: "terminal", Icon: "👻", Label: "Terminal Setup", Section: "Home",
			Do: func() tea.Cmd { return app.Push(terminalMenuScreen()) },
		})
		app.Register(app.Action{
			ID: "update", Icon: "⬆", Label: "Check for Updates", Section: "Home",
			Do: func() tea.Cmd {
				return app.Push(app.NewRunner("Update", func() error { return runUpdate(false) }))
			},
		})
		for _, cat := range availableBundleCategories() {
			id, label := cat.ID, cat.Label
			app.Register(app.Action{
				ID: "install-" + id, Label: label, Section: "Install",
				Do: func() tea.Cmd { return packagesFlow(id, label) },
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
