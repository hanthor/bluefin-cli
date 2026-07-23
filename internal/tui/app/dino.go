package app

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// dino is the little dot-matrix dinosaur that ambles across the header.
// Frames are braille-cell sprites (hand-tuned): tail on the left, head up on
// the right, with the trailing cell alternating to suggest leg movement.
var dinoRunFrames = []string{
	"⢀⣠⣾⣿⠟⠁",
	"⢀⣠⣾⣿⠟⠈",
}

// dinoPauseFrame is shown when the dino stops to look around.
const dinoPauseFrame = "⢀⣠⣾⣿⠛⠁"

const (
	dinoFPS        = 8  // ticks per second
	dinoStrideCols = 1  // columns advanced per tick
	dinoPauseEvery = 36 // ticks between pauses
	dinoPauseLen   = 10 // ticks a pause lasts
)

type dino struct {
	x     int // leading (left) column of the sprite
	ticks int
}

func (d *dino) tick() tea.Cmd {
	return tea.Tick(time.Second/dinoFPS, func(time.Time) tea.Msg { return dinoTickMsg{} })
}

func (d *dino) paused() bool {
	return d.ticks%(dinoPauseEvery+dinoPauseLen) >= dinoPauseEvery
}

func (d *dino) advance(width int) {
	d.ticks++
	if d.paused() {
		return
	}
	d.x += dinoStrideCols
	spriteW := lipgloss.Width(dinoRunFrames[0])
	if width > 0 && d.x > width {
		// Ran off the right edge; re-enter from the left.
		d.x = -spriteW
	}
}

func (d *dino) render(width int, t theme.Theme) string {
	sprite := dinoRunFrames[d.ticks%len(dinoRunFrames)]
	if d.paused() {
		sprite = dinoPauseFrame
	}
	spriteW := lipgloss.Width(sprite)
	if width < spriteW+2 {
		return ""
	}

	// Clip while entering/leaving the edges.
	runes := []rune(sprite)
	start := d.x
	if start < 0 {
		clip := min(-start, len(runes))
		runes = runes[clip:]
		start = 0
	}
	if start+len(runes) > width {
		over := start + len(runes) - width
		if over >= len(runes) {
			return ""
		}
		runes = runes[:len(runes)-over]
	}

	return strings.Repeat(" ", start) +
		lipgloss.NewStyle().Foreground(t.Success).Render(string(runes))
}
