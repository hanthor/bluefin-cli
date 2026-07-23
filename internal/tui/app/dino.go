package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// dino is the dot-matrix dinosaur that runs along the header's braille
// seafloor. It ambles, pauses to look around, and sprints for a moment
// whenever the user selects something.
var dinoRunFrames = []string{
	"⢀⣠⣾⣿⠟⠁",
	"⢀⣠⣾⣿⠟⠈",
}

// dinoPauseFrame is shown when the dino stops to look around.
const dinoPauseFrame = "⢀⣠⣾⣿⠛⠁"

const (
	dinoFPS        = 8  // ticks per second
	dinoStrideCols = 1  // columns advanced per tick while ambling
	dinoSprintCols = 3  // columns advanced per tick while sprinting
	dinoPauseEvery = 36 // ticks between pauses
	dinoPauseLen   = 10 // ticks a pause lasts
	dinoSprintLen  = 10 // ticks a selection sprint lasts
)

// groundPattern repeats to form the seafloor; mild variation reads as
// terrain rather than a flat rule.
var groundPattern = []rune("⣀⣀⣀⣀⣄⣀⣀⣀⡀⣀⣀⣠⣀⣀⣀⣀")

type dino struct {
	x      int // leading (left) column of the sprite
	ticks  int
	sprint int // remaining sprint ticks
}

func (d *dino) tick() tea.Cmd {
	return tea.Tick(time.Second/dinoFPS, func(time.Time) tea.Msg { return dinoTickMsg{} })
}

// boost makes the dino sprint briefly — fired when the user selects
// something, so navigation has a little kinetic feedback.
func (d *dino) boost() {
	d.sprint = dinoSprintLen
}

func (d *dino) paused() bool {
	return d.sprint == 0 && d.ticks%(dinoPauseEvery+dinoPauseLen) >= dinoPauseEvery
}

func (d *dino) advance(width int) {
	d.ticks++
	if d.paused() {
		return
	}
	stride := dinoStrideCols
	if d.sprint > 0 {
		d.sprint--
		stride = dinoSprintCols
	}
	d.x += stride
	spriteW := lipgloss.Width(dinoRunFrames[0])
	if width > 0 && d.x > width {
		// Ran off the right edge; re-enter from the left.
		d.x = -spriteW
	}
}

// renderGround draws the full-width braille seafloor with the dino spliced
// in at its current position.
func (d *dino) renderGround(width int, t theme.Theme) string {
	if width < 1 {
		return ""
	}
	ground := make([]rune, width)
	for i := range ground {
		ground[i] = groundPattern[i%len(groundPattern)]
	}

	sprite := dinoRunFrames[d.ticks%len(dinoRunFrames)]
	if d.paused() {
		sprite = dinoPauseFrame
	}
	sr := []rune(sprite)

	// Clip the sprite while it enters/leaves the edges.
	start := d.x
	if start < -len(sr) || start >= width {
		return lipgloss.NewStyle().Foreground(t.Surface).Render(string(ground))
	}
	if start < 0 {
		sr = sr[-start:]
		start = 0
	}
	if start+len(sr) > width {
		sr = sr[:width-start]
	}

	groundStyle := lipgloss.NewStyle().Foreground(t.Surface)
	dinoStyle := lipgloss.NewStyle().Foreground(t.Success)
	return groundStyle.Render(string(ground[:start])) +
		dinoStyle.Render(string(sr)) +
		groundStyle.Render(string(ground[start+len(sr):]))
}
