package app

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// GameScreen is the hidden chrome-dino-style runner: jump the kelp, rack up
// a score, esc to bail. Reached via the ctrl+p palette ("Dino Run") or the
// hidden `bluefin-cli dino` command.
type GameScreen struct {
	obstacles []int // x column of each obstacle
	tick      int
	score     int
	best      int
	saveBest  func(int) // persists a new high score; may be nil
	air       int       // remaining jump ticks (>0 means airborne)
	over      bool
	width     int
}

type gameTickMsg struct{}

const (
	gameFPS      = 14
	gameJumpLen  = 9 // ticks airborne
	gameDinoX    = 6 // fixed dino column
	gameSpriteW  = 7
	gameMinGap   = 18
	gameSpawnPct = 22 // % chance per eligible tick
)

// jumpArc is the dino's row offset above the ground per airborne tick.
var jumpArc = []int{1, 2, 3, 3, 3, 3, 2, 1, 0}

// NewGame creates the runner screen. best seeds the high score and saveBest
// (optional) is called whenever a run beats it.
func NewGame(best int, saveBest func(int)) *GameScreen {
	return &GameScreen{best: best, saveBest: saveBest}
}

func (g *GameScreen) Title() string { return "Dino Run" }

func (g *GameScreen) Init() tea.Cmd { return g.tickCmd() }

func (g *GameScreen) tickCmd() tea.Cmd {
	return tea.Tick(time.Second/gameFPS, func(time.Time) tea.Msg { return gameTickMsg{} })
}

func (g *GameScreen) KeyHints() []KeyHint {
	return []KeyHint{{"space", "jump"}, {"r", "restart"}}
}

// CapturingInput keeps q from quitting mid-game ("r" restarts, esc exits).
func (g *GameScreen) CapturingInput() bool { return !g.over }

func (g *GameScreen) restart() {
	g.obstacles = nil
	g.tick, g.score, g.air = 0, 0, 0
	g.over = false
}

func (g *GameScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case gameTickMsg:
		if g.over {
			return g, nil
		}
		g.advanceGame()
		return g, g.tickCmd()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "space", "up", "k":
			if !g.over && g.air == 0 {
				g.air = gameJumpLen
			}
		case "r", "enter":
			if g.over {
				g.restart()
				return g, g.tickCmd()
			}
		case "esc":
			return g, Pop()
		}
	}
	return g, nil
}

func (g *GameScreen) advanceGame() {
	g.tick++
	g.score++
	if g.air > 0 {
		g.air--
	}

	// March obstacles left; the pace quickens with score.
	speed := 1
	if g.score > 300 {
		speed = 2
	}
	next := g.obstacles[:0]
	for _, x := range g.obstacles {
		if x-speed > -2 {
			next = append(next, x-speed)
		}
	}
	g.obstacles = next

	// Spawn at the right edge with a minimum gap.
	w := max(g.width, 40)
	canSpawn := len(g.obstacles) == 0 || g.obstacles[len(g.obstacles)-1] < w-gameMinGap
	if canSpawn && rand.Intn(100) < gameSpawnPct {
		g.obstacles = append(g.obstacles, w-1)
	}

	// Collision: an obstacle inside the dino's footprint while grounded.
	height := 0
	if g.air > 0 && g.air <= len(jumpArc) {
		height = jumpArc[len(jumpArc)-g.air]
	}
	if height < 2 {
		for _, x := range g.obstacles {
			if x >= gameDinoX+1 && x <= gameDinoX+gameSpriteW-2 {
				g.over = true
				if g.score > g.best {
					g.best = g.score
					if g.saveBest != nil {
						g.saveBest(g.best)
					}
				}
			}
		}
	}
}

func (g *GameScreen) View(width, height int) string {
	g.width = width
	t := theme.DefaultTheme
	dinoStyle := lipgloss.NewStyle().Foreground(t.Success).Bold(true)
	kelpStyle := lipgloss.NewStyle().Foreground(t.AccentAlt)
	groundStyle := lipgloss.NewStyle().Foreground(t.Surface)
	dim := lipgloss.NewStyle().Foreground(t.TextFaint)

	rows := max(height-2, 8)
	groundRow := rows - 1
	field := make([][]rune, rows)
	owner := make([][]byte, rows) // ' ' empty, 'g' ground, 'k' kelp, 'd' dino
	for i := range field {
		field[i] = []rune(strings.Repeat(" ", width))
		owner[i] = []byte(strings.Repeat(" ", width))
	}
	for i := 0; i < width; i++ {
		field[groundRow][i] = groundPattern[i%len(groundPattern)]
		owner[groundRow][i] = 'g'
	}

	// Obstacles: kelp fronds rising from the seafloor.
	for _, x := range g.obstacles {
		if x >= 0 && x < width {
			field[groundRow][x], owner[groundRow][x] = '⣼', 'k'
			if groundRow-1 >= 0 {
				field[groundRow-1][x], owner[groundRow-1][x] = '⢸', 'k'
			}
		}
	}

	// Dino: three sprite rows, lifted by the jump arc.
	lift := 0
	if g.air > 0 && g.air <= len(jumpArc) {
		lift = jumpArc[len(jumpArc)-g.air]
	}
	sprite := []string{dinoSkyHigh, dinoSkyLow, dinoLegFrames[g.tick%len(dinoLegFrames)]}
	if g.over {
		sprite[2] = dinoLegsPause
	}
	base := groundRow - lift
	for i, row := range sprite {
		y := base - (len(sprite) - 1 - i)
		if y < 0 || y >= rows {
			continue
		}
		for j, r := range []rune(row) {
			x := gameDinoX + j
			if x >= 0 && x < width && r != '⠀' {
				field[y][x], owner[y][x] = r, 'd'
			}
		}
	}

	styleFor := map[byte]lipgloss.Style{'g': groundStyle, 'k': kelpStyle, 'd': dinoStyle, ' ': dim}
	var b strings.Builder
	score := fmt.Sprintf(" score %d", g.score)
	if g.best > 0 {
		score += fmt.Sprintf("   best %d", g.best)
	}
	b.WriteString(dim.Render(score) + "\n")
	for y := range field {
		// Group consecutive same-owner cells into one styled run.
		runStart := 0
		for x := 1; x <= width; x++ {
			if x == width || owner[y][x] != owner[y][runStart] {
				b.WriteString(styleFor[owner[y][runStart]].Render(string(field[y][runStart:x])))
				runStart = x
			}
		}
		if y < len(field)-1 {
			b.WriteString("\n")
		}
	}
	if g.over {
		msg := lipgloss.NewStyle().Foreground(t.Error).Bold(true).Render(" 💀 wiped out — ") +
			dim.Render("r to retry, esc to leave")
		b.WriteString("\n" + msg)
	}
	return b.String()
}
