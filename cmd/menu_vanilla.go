//go:build !extra

package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

func extraMenuItems() []app.MenuItem {
	return []app.MenuItem{
		{Icon: "🖼 ", Label: "Wallpapers", Value: "wallpapers", Desc: "Bluefin wallpaper gallery", Submenu: true},
		{Icon: "🔤", Label: "Fonts", Value: "fonts", Desc: "Nerd Fonts for your terminal", Submenu: true},
		{Icon: "🚀", Label: "Starship Theme", Value: "starship", Desc: "Prompt presets & customization", Submenu: true},
	}
}

func extraMenuDo(value string) tea.Cmd {
	switch value {
	case "wallpapers":
		return wallpapersFlow()
	case "fonts":
		return fontsFlow()
	case "starship":
		return starshipFlow()
	}
	return nil
}
