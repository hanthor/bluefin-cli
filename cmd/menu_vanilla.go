//go:build !extra

package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

func extraMenuItems() []app.MenuItem {
	return []app.MenuItem{
		{Icon: "🖼 ", Label: "Wallpapers", Value: "wallpapers", Submenu: true},
		{Icon: "🔤", Label: "Fonts", Value: "fonts", Submenu: true},
		{Icon: "🚀", Label: "Starship Theme", Value: "starship", Submenu: true},
	}
}

func extraMenuDo(value string) tea.Cmd {
	switch value {
	case "wallpapers":
		return app.RunLegacy(runWallpapersMenu)
	case "fonts":
		return app.RunLegacy(runFontsMenu)
	case "starship":
		return app.RunLegacy(runStarshipMenu)
	}
	return nil
}
