package app

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// PixelCanvas is a half-block (▀) pixel framebuffer: each terminal cell
// holds two vertically stacked pixels with independent 24-bit colors — the
// portable "high definition" mode used by chafa, notcurses, and timg when
// real graphics protocols aren't available. One pixel per column
// horizontally, two per row vertically; a nil pixel is transparent and
// shows the terminal background.
type PixelCanvas struct {
	w, h int // h is in pixels (2x the cell rows)
	px   []color.Color
}

// NewCanvas creates a canvas covering w cells by rows cell-rows.
func NewCanvas(w, rows int) *PixelCanvas {
	return &PixelCanvas{w: w, h: rows * 2, px: make([]color.Color, w*rows*2)}
}

// Width returns the canvas width in pixels (== cells).
func (c *PixelCanvas) Width() int { return c.w }

// Height returns the canvas height in pixels (2x the cell rows).
func (c *PixelCanvas) Height() int { return c.h }

// Set paints one pixel; out-of-bounds is ignored.
func (c *PixelCanvas) Set(x, y int, col color.Color) {
	if x < 0 || x >= c.w || y < 0 || y >= c.h {
		return
	}
	c.px[y*c.w+x] = col
}

// Blit paints a sprite grid at (x, y): each row string is one pixel row,
// and each byte indexes palette; bytes not in the palette are transparent.
func (c *PixelCanvas) Blit(sprite []string, palette map[byte]color.Color, x, y int) {
	for dy, row := range sprite {
		for dx := 0; dx < len(row); dx++ {
			if col, ok := palette[row[dx]]; ok {
				c.Set(x+dx, y+dy, col)
			}
		}
	}
}

// Render flattens the canvas to styled half-block rows.
func (c *PixelCanvas) Render() string {
	var b strings.Builder
	for row := 0; row < c.h/2; row++ {
		top, bot := row*2, row*2+1
		x := 0
		for x < c.w {
			tc, bc := c.px[top*c.w+x], c.px[bot*c.w+x]
			// Group a run of columns with identical top/bottom colors into
			// one styled segment to keep the output compact.
			run := x + 1
			for run < c.w && sameColor(c.px[top*c.w+run], tc) && sameColor(c.px[bot*c.w+run], bc) {
				run++
			}
			n := run - x
			switch {
			case tc == nil && bc == nil:
				b.WriteString(strings.Repeat(" ", n))
			case tc != nil && bc != nil:
				b.WriteString(lipgloss.NewStyle().Foreground(tc).Background(bc).Render(strings.Repeat("▀", n)))
			case tc != nil:
				b.WriteString(lipgloss.NewStyle().Foreground(tc).Render(strings.Repeat("▀", n)))
			default:
				b.WriteString(lipgloss.NewStyle().Foreground(bc).Render(strings.Repeat("▄", n)))
			}
			x = run
		}
		if row < c.h/2-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	ar, ag, ab2, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab2 == bb && aa == ba
}
