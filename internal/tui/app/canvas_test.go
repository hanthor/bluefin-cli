package app

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestCanvasHalfBlockPairs(t *testing.T) {
	red := lipgloss.Color("#ff0000")
	blue := lipgloss.Color("#0000ff")

	c := NewCanvas(3, 1) // 3x2 px
	c.Set(0, 0, red)     // top only -> ▀
	c.Set(1, 1, blue)    // bottom only -> ▄
	c.Set(2, 0, red)     // both -> ▀ with bg
	c.Set(2, 1, blue)

	out := c.Render()
	plain := stripAnsi(out)
	if plain != "▀▄▀" {
		t.Errorf("rendered glyphs = %q, want ▀▄▀", plain)
	}
	if !strings.Contains(out, "48;") {
		t.Error("both-pixels cell should set a background color")
	}
}

func TestCanvasTransparencyAndBounds(t *testing.T) {
	c := NewCanvas(4, 2)
	// Out-of-bounds writes must be ignored, not panic.
	c.Set(-1, 0, lipgloss.Color("#fff"))
	c.Set(99, 99, lipgloss.Color("#fff"))

	if plain := stripAnsi(c.Render()); strings.TrimRight(strings.ReplaceAll(plain, "\n", ""), " ") != "" {
		t.Errorf("empty canvas should render only blanks, got %q", plain)
	}
}

func TestCanvasBlitClipsAndSkipsTransparent(t *testing.T) {
	green := lipgloss.Color("#00ff00")
	c := NewCanvas(4, 1)
	// Sprite wider than the canvas, blitted partially off the left edge.
	c.Blit([]string{"#.##", "####"}, map[byte]color.Color{}, 0, 0) // empty palette: all transparent
	if plain := strings.TrimSpace(stripAnsi(c.Render())); plain != "" {
		t.Errorf("transparent blit should paint nothing, got %q", plain)
	}

	c2 := NewCanvas(2, 1)
	c2.Blit([]string{"####"}, solid(green), -2, 0)
	if plain := stripAnsi(c2.Render()); plain != "▀▀" {
		t.Errorf("clipped blit = %q, want ▀▀", plain)
	}
}
