package cmd

// Native in-shell flows for everything that used to hand over the terminal:
// selection UIs are FormScreens, read-only views are TextScreens, and
// anything that prints (brew installs, toggles with output) runs in a
// RunnerScreen that captures stdout into a log view.

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/motd"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/starship"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// --- Shell submenu -----------------------------------------------------

func motdMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		cfg, err := shell.LoadConfig(currentShellName())
		if err != nil {
			cfg = shell.DefaultConfig(currentShellName())
		}
		toggle := "Enable MOTD"
		if cfg.IsEnabled("Motd") {
			toggle = "Disable MOTD"
		}
		return []app.MenuItem{
			{Icon: "🔄", Label: toggle, Value: "toggle", Desc: "Message of the day on new terminals"},
			{Icon: "📰", Label: "Show MOTD", Value: "show", Desc: "Preview today's message"},
		}
	}
	return app.NewMenu("MOTD", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "toggle":
			return tea.Sequence(func() tea.Msg {
				sh := currentShellName()
				cfg, err := shell.LoadConfig(sh)
				if err != nil {
					cfg = shell.DefaultConfig(sh)
				}
				enable := !cfg.IsEnabled("Motd")
				cfg.SetEnabled("Motd", enable)
				if err := shell.SaveConfig(cfg); err != nil {
					return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
				}
				if enable {
					return app.ToastMsg{Text: "MOTD enabled."}
				}
				return app.ToastMsg{Text: "MOTD disabled."}
			}, app.ReloadTop())
		case "show":
			return app.Push(app.NewRunner("MOTD", motd.Show))
		}
		return nil
	})
}

func shellsFormScreen() app.Screen {
	var installed []string
	var selected []string
	initial := map[string]bool{}

	build := func() *huh.Form {
		installed = shell.GetInstalledShells()
		status := shell.CheckStatus()
		selected = selected[:0]
		options := make([]huh.Option[string], 0, len(installed))
		for _, sh := range installed {
			initial[sh] = status[sh]
			if status[sh] {
				selected = append(selected, sh)
			}
			options = append(options, huh.NewOption(sh, sh))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Manage other shells").
				Description("Selected = enabled. Space toggles, enter applies.").
				Options(options...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.NewForm("Other Shells", build, func(aborted bool) tea.Cmd {
		if aborted {
			return nil
		}
		want := map[string]bool{}
		for _, sh := range selected {
			want[sh] = true
		}
		changed := 0
		for _, sh := range installed {
			if want[sh] != initial[sh] {
				changed++
			}
		}
		if changed == 0 {
			return app.Toast("No changes.", false)
		}
		shells := append([]string(nil), installed...)
		return app.Push(app.NewRunner("Applying", func() error {
			for _, sh := range shells {
				if want[sh] != initial[sh] {
					fmt.Printf("%s -> %v\n", sh, want[sh])
					if err := shell.Toggle(sh, want[sh]); err != nil {
						return err
					}
				}
			}
			return nil
		}))
	})
}

// --- Extras ------------------------------------------------------------

func wallpapersFlow() tea.Cmd {
	return tea.Batch(
		app.Toast("Loading wallpapers…", false),
		func() tea.Msg {
			casks, err := install.GetWallpaperCasks()
			if err != nil {
				return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
			}
			if len(casks) == 0 {
				return app.ToastMsg{Text: "No wallpaper casks found in ublue-os/tap.", IsErr: true}
			}
			return app.PushMsg{Screen: wallpapersFormScreen(casks, install.InstalledCasks())}
		},
	)
}

func wallpapersFormScreen(casks []string, installed map[string]bool) app.Screen {
	var selected []string
	build := func() *huh.Form {
		selected = selected[:0]
		opts := make([]huh.Option[string], 0, len(casks))
		for _, c := range casks {
			label := c
			if installed[c] {
				label += " (installed)"
			}
			opts = append(opts, huh.NewOption(label, c))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select wallpapers to install").
				Description("Space toggles, enter confirms.").
				Options(opts...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}
	return app.NewForm("Wallpapers", build, func(aborted bool) tea.Cmd {
		if aborted || len(selected) == 0 {
			return nil
		}
		picks := append([]string(nil), selected...)
		return app.Push(app.NewRunner("Installing wallpapers", func() error {
			return install.InstallWallpaperCasks(picks)
		}))
	})
}

func fontsFlow() tea.Cmd {
	var selected []string
	build := func() *huh.Form {
		selected = selected[:0]
		opts := make([]huh.Option[string], 0, len(availableFonts))
		for _, f := range availableFonts {
			opts = append(opts, huh.NewOption(f.Name, f.Cask))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select fonts to install").
				Description("Space toggles, enter confirms.").
				Options(opts...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}
	return app.Push(app.NewForm("Fonts", build, func(aborted bool) tea.Cmd {
		if aborted || len(selected) == 0 {
			return nil
		}
		picks := append([]string(nil), selected...)
		return app.Push(app.NewRunner("Installing fonts", func() error {
			installFontCasks(picks)
			return maybeHandlePostFontInstall()
		}))
	}))
}

func starshipFlow() tea.Cmd {
	var selectedTheme string
	build := func() *huh.Form {
		return huh.NewForm(huh.NewGroup(
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
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}
	return app.Push(app.NewForm("Starship Theme", build, func(aborted bool) tea.Cmd {
		if aborted || selectedTheme == "" {
			return nil
		}
		pick := selectedTheme
		return app.Push(app.NewRunner("Applying "+pick, func() error {
			if err := starship.Install(); err != nil {
				return err
			}
			return starship.ApplyTheme(pick)
		}))
	}))
}

// --- Advanced ----------------------------------------------------------

func advancedMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		dark := "🌙 Dark Mode: On"
		if !viper.GetBool("ui.dark_mode") {
			dark = "☀️  Dark Mode: Off"
		}
		return []app.MenuItem{
			{Label: dark, Value: "toggle_dark", Desc: "Hint for legacy CLI output styling"},
			{Icon: "🎨", Label: "Flavor: " + viper.GetString("ui.flavor"), Value: "flavor",
				Desc: "Cycle Catppuccin flavor (auto follows the terminal)"},
		}
	}
	return app.NewMenu("Advanced", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "toggle_dark":
			return tea.Sequence(func() tea.Msg {
				viper.Set("ui.dark_mode", !viper.GetBool("ui.dark_mode"))
				if err := config.Save(); err != nil {
					return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
				}
				return app.ToastMsg{Text: "Saved."}
			}, app.ReloadTop())
		case "flavor":
			return tea.Sequence(func() tea.Msg {
				flavors := theme.Flavors()
				cur := viper.GetString("ui.flavor")
				next := flavors[0]
				for i, f := range flavors {
					if f == cur {
						next = flavors[(i+1)%len(flavors)]
						break
					}
				}
				viper.Set("ui.flavor", next)
				if err := config.Save(); err != nil {
					return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
				}
				theme.Flavor = ""
				if next != "auto" {
					theme.Flavor = next
				}
				theme.DefaultTheme = theme.Resolve(theme.DefaultTheme.IsDark)
				return app.ToastMsg{Text: "Flavor: " + next}
			}, app.ReloadTop())
		}
		return nil
	})
}

// --- Doctor & game wiring ----------------------------------------------

// doctorScreenCmd runs the doctor checks in a RunnerScreen (they hit the
// network, so they must not block the render loop).
func doctorScreenCmd() tea.Cmd {
	return app.Push(app.NewRunner("Doctor", func() error {
		report, _ := doctorReport()
		fmt.Println(report)
		return nil
	}))
}

// gameScreen builds the dino runner with a config-persisted high score.
func gameScreen() app.Screen {
	return app.NewGame(viper.GetInt("game.high_score"), func(n int) {
		viper.Set("game.high_score", n)
		_ = config.Save()
	})
}
