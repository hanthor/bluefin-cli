//go:build extra

package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

func extraMenuItems() []app.MenuItem {
	items := []app.MenuItem{
		{Icon: "🖼 ", Label: "Wallpapers", Value: "wallpapers", Desc: "Bluefin wallpaper gallery", Submenu: true},
		{Icon: "🔤", Label: "Fonts", Value: "fonts", Desc: "Nerd Fonts for your terminal", Submenu: true},
		{Icon: "🚀", Label: "Starship Theme", Value: "starship", Desc: "Prompt presets & customization", Submenu: true},
	}
	if env.IsWSL() || env.IsWindows() {
		items = append(items, app.MenuItem{Icon: "🌇", Label: "Sunset Switching", Value: "sunset", Desc: "Auto light/dark by time of day", Submenu: true})
	}
	return items
}

func extraMenuDo(value string) tea.Cmd {
	switch value {
	case "wallpapers":
		return app.RunLegacy(runWallpapersMenu)
	case "fonts":
		return app.RunLegacy(runFontsMenu)
	case "starship":
		return app.RunLegacy(runStarshipMenu)
	case "sunset":
		return app.RunLegacy(RunSunsetSetupFlow)
	}
	return nil
}
