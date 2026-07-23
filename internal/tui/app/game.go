package app

import (
	"fmt"
	"image/color"
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// GameScreen is the hidden chrome-dino-style runner, rendered on a
// half-block PixelCanvas: full-color pixel art at double the vertical
// resolution of plain cells. Jump the kelp; stay grounded under the fish.
// Reached via the ctrl+p palette ("Dino Run") or the hidden
// `bluefin-cli dino` command.

// gameDino* are 19x18-pixel frames downsampled from the real Chrome T-rex
// sprite sheet ('#' = body; the blank eye pixel shows the background).
var gameDinoRun = [][]string{
	{
		"..........########.",
		".........##.######.",
		".........#########.",
		".........#########.",
		".........#########.",
		".........#######...",
		".#.......####......",
		".#.....######......",
		".##...#########....",
		".############......",
		".############......",
		".############......",
		"..##########.......",
		"...########........",
		"....#######........",
		".....##...#........",
		".....##............",
		".....##............",
	},
	{
		"..........########.",
		".........##.######.",
		".........#########.",
		".........#########.",
		".........#########.",
		".........#######...",
		".#.......####......",
		".#.....######......",
		".##...#########....",
		".############......",
		".############......",
		".############......",
		"..##########.......",
		"...########........",
		"....#######........",
		".....##..##........",
		"......#..##........",
		".........##........",
	},
}

var gameDinoStand = []string{
	"..........########.",
	".........##.######.",
	".........#########.",
	".........#########.",
	".........#########.",
	".........#######...",
	".#.......####......",
	".#.....######......",
	".##...#########....",
	".############......",
	".############......",
	".############......",
	"..##########.......",
	"...########........",
	"....#######........",
	".....##..##........",
	".....##..##........",
	".....##..##........",
}

// gameKelp is the seafloor hazard (jump it), 5x12 px.
var gameKelp = []string{
	"..#..",
	".##..",
	"..##.",
	"..#..",
	".##..",
	"..##.",
	"..#..",
	"..##.",
	".###.",
	"..#..",
	".##..",
	".###.",
}

// gameFishPx swims at head height (stay grounded), 10x5 px.
var gameFishPx = []string{
	"..####....",
	".#.####..#",
	"########.#",
	".######..#",
	"..####....",
}

// obstacle is a hazard at pixel column x.
type obstacle struct {
	x   int
	fly bool
}

// bubble is ambient decoration drifting up from the seafloor.
type bubble struct {
	x, y int
}

type GameScreen struct {
	obstacles []obstacle
	bubbles   []bubble
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
	gameJumpLen  = 9  // ticks airborne
	gameDinoX    = 6  // fixed dino column
	gameDinoW    = 19 // sprite width in px
	gameDinoH    = 18 // sprite height in px
	gameKelpW    = 5
	gameKelpH    = 12
	gameFishW    = 10
	gameFishH    = 5
	gameMinGap   = 24
	gameSpawnPct = 20  // % chance per eligible tick
	gameFishFrom = 150 // score at which fish start spawning
)

// jumpArc is the dino's pixel lift per airborne tick — enough at its peak
// to clear the 12px kelp.
var jumpArc = []int{5, 10, 14, 16, 16, 14, 10, 5, 0}

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

func (g *GameScreen) lift() int {
	if g.air > 0 && g.air <= len(jumpArc) {
		return jumpArc[len(jumpArc)-g.air]
	}
	return 0
}

func (g *GameScreen) advanceGame() {
	g.tick++
	g.score++
	if g.air > 0 {
		g.air--
	}

	// March obstacles left; the pace quickens with score, and fish swim
	// a little faster than the current.
	speed := 2
	if g.score > 300 {
		speed = 3
	}
	next := g.obstacles[:0]
	for _, o := range g.obstacles {
		st := speed
		if o.fly {
			st++
		}
		if o.x-st > -gameFishW {
			next = append(next, obstacle{x: o.x - st, fly: o.fly})
		}
	}
	g.obstacles = next

	// Spawn at the right edge with a minimum gap.
	w := max(g.width, 40)
	canSpawn := len(g.obstacles) == 0 || g.obstacles[len(g.obstacles)-1].x < w-gameMinGap
	if canSpawn && rand.Intn(100) < gameSpawnPct {
		fly := g.score > gameFishFrom && rand.Intn(100) < 35
		g.obstacles = append(g.obstacles, obstacle{x: w - 1, fly: fly})
	}

	// Ambient bubbles drift up and pop at the surface.
	nb := g.bubbles[:0]
	for _, b := range g.bubbles {
		b.y -= 1
		b.x -= 1 // carried by the current
		if b.y > 0 {
			nb = append(nb, b)
		}
	}
	g.bubbles = nb
	if g.tick%5 == 0 && len(g.bubbles) < 12 {
		g.bubbles = append(g.bubbles, bubble{x: rand.Intn(max(w, 1)), y: 1 << 20})
	}

	if g.collided() {
		g.over = true
		if g.score > g.best {
			g.best = g.score
			if g.saveBest != nil {
				g.saveBest(g.best)
			}
		}
	}
}

// overlap reports whether [a0,a1] and [b0,b1] intersect.
func overlap(a0, a1, b0, b1 int) bool { return a0 <= b1 && b0 <= a1 }

// collided reports whether any obstacle intersects the dino's body: kelp is
// deadly near the ground (jump it), a fish is deadly while airborne (hold).
// Vertical bands are relative to the ground line; the fish swims just above
// the standing dino's head.
func (g *GameScreen) collided() bool {
	lift := g.lift()
	dinoL, dinoR := gameDinoX+3, gameDinoX+gameDinoW-4
	dinoTop := -gameDinoH - lift // y relative to ground, negative = up
	dinoBot := -lift

	for _, o := range g.obstacles {
		if o.fly {
			// Fish band sits above the standing head.
			fishTop, fishBot := -gameDinoH-gameFishH-2, -gameDinoH-3
			if overlap(o.x, o.x+gameFishW-1, dinoL, dinoR) &&
				overlap(fishTop, fishBot, dinoTop, dinoBot) {
				return true
			}
			continue
		}
		if overlap(o.x, o.x+gameKelpW-1, dinoL, dinoR) &&
			overlap(-gameKelpH+1, 0, dinoTop, dinoBot) {
			return true
		}
	}
	return false
}

// solid maps '#' sprite pixels to one color.
func solid(c color.Color) map[byte]color.Color {
	return map[byte]color.Color{'#': c}
}

func (g *GameScreen) View(width, height int) string {
	g.width = width
	t := theme.DefaultTheme
	dim := lipgloss.NewStyle().Foreground(t.TextFaint)

	rows := max(height-2, 10)
	cv := NewCanvas(width, rows)
	groundY := cv.Height() - 3 // pixel row of the ground line

	// Seafloor: a 2px sandy band with speckles.
	for x := 0; x < width; x++ {
		cv.Set(x, groundY+1, t.Surface)
		if x%7 == 3 {
			cv.Set(x, groundY+2, t.Overlay)
		} else {
			cv.Set(x, groundY+2, t.Surface)
		}
	}

	// Ambient bubbles (spawned with a sentinel y; anchor them to the
	// current seafloor once the canvas size is known).
	for i := range g.bubbles {
		if g.bubbles[i].y >= 1<<20 {
			g.bubbles[i].y = groundY - 1
		}
		cv.Set(g.bubbles[i].x, g.bubbles[i].y, t.Overlay)
	}

	// Obstacles.
	for _, o := range g.obstacles {
		if o.fly {
			cv.Blit(gameFishPx, solid(t.AccentAlt), o.x, groundY-gameDinoH-gameFishH-2)
			continue
		}
		cv.Blit(gameKelp, solid(t.Info), o.x, groundY-gameKelpH+1)
	}

	// Dino.
	sprite := gameDinoRun[g.tick%len(gameDinoRun)]
	if g.over {
		sprite = gameDinoStand
	}
	cv.Blit(sprite, solid(t.Success), gameDinoX, groundY-gameDinoH-g.lift()+1)

	var b strings.Builder
	score := fmt.Sprintf(" score %d", g.score)
	if g.best > 0 {
		score += fmt.Sprintf("   best %d", g.best)
	}
	b.WriteString(dim.Render(score) + "\n")
	b.WriteString(cv.Render())
	if g.over {
		msg := lipgloss.NewStyle().Foreground(t.Error).Bold(true).Render(" 💀 wiped out — ") +
			dim.Render("r to retry, esc to leave")
		b.WriteString("\n" + msg)
	}
	return b.String()
}
