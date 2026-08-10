package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestWallpaperPickerScreen(t *testing.T) {
	items := []MenuItem{
		{Label: "wallpaper1.jpg", Value: "/path/to/wallpaper1.jpg", Desc: "/path/to"},
		{Label: "wallpaper2.png", Value: "/path/to/wallpaper2.png", Desc: "/path/to"},
	}

	selected := ""
	picker := NewWallpaperPicker("Set Wallpaper", items, func(it MenuItem) tea.Cmd {
		selected = it.Value
		return nil
	})

	if picker.Title() != "Set Wallpaper" {
		t.Errorf("Title() = %q, want %q", picker.Title(), "Set Wallpaper")
	}

	m := drive(picker, special(tea.KeyEnter))
	_ = m
	if selected != "/path/to/wallpaper1.jpg" {
		t.Errorf("selected = %q, want %q", selected, "/path/to/wallpaper1.jpg")
	}
}

func TestWallpaperPickerRenderNoCrash(t *testing.T) {
	items := []MenuItem{
		{Label: "test.jpg", Value: "/nonexistent/test.jpg"},
	}
	picker := NewWallpaperPicker("Set Wallpaper", items, nil)
	view := picker.View(80, 24)
	if !strings.Contains(view, "test.jpg") {
		t.Errorf("View missing label, got %q", view)
	}
}
