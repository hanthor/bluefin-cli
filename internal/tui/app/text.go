package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// TextScreen renders width-aware text content natively in the shell (with
// j/k scrolling) instead of releasing the terminal to a legacy printer.
type TextScreen struct {
	title     string
	fetch     func(width int) (string, error)
	content   string
	fetchErr  error
	lastWidth int
	scroll    int
}

// NewText creates a text screen; fetch is re-run whenever the screen regains
// focus or the width changes.
func NewText(title string, fetch func(width int) (string, error)) *TextScreen {
	return &TextScreen{title: title, fetch: fetch}
}

func (s *TextScreen) Title() string { return s.title }

func (s *TextScreen) Init() tea.Cmd { return nil }

// Reload invalidates the cached content (Reloader interface).
func (s *TextScreen) Reload() tea.Cmd {
	s.lastWidth = 0
	return nil
}

func (s *TextScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	switch key.String() {
	case "up", "k":
		s.scroll = max(s.scroll-1, 0)
	case "down", "j":
		s.scroll++ // clamped against content height in View
	case "pgup", "ctrl+u":
		s.scroll = max(s.scroll-10, 0)
	case "pgdown", "ctrl+d":
		s.scroll += 10
	case "g", "home":
		s.scroll = 0
	case "G", "end":
		s.scroll = 1 << 30
	}
	return s, nil
}

func (s *TextScreen) View(width, height int) string {
	if width != s.lastWidth {
		// Leave a margin so styled columns never touch the edge.
		s.content, s.fetchErr = s.fetch(max(width-2, 20))
		s.lastWidth = width
	}
	if s.fetchErr != nil {
		return "\n " + lipgloss.NewStyle().Foreground(theme.DefaultTheme.Error).Render(s.fetchErr.Error())
	}
	lines := strings.Split(strings.TrimRight(s.content, "\n"), "\n")
	avail := height
	if len(lines) > height {
		avail = height - 1 // reserve a row for the position indicator
	}
	s.scroll = max(min(s.scroll, len(lines)-avail), 0)
	end := min(s.scroll+avail, len(lines))
	body := make([]string, 0, end-s.scroll+1)
	for _, l := range lines[s.scroll:end] {
		body = append(body, " "+l)
	}
	if len(lines) > height {
		pos := lipgloss.NewStyle().Foreground(theme.DefaultTheme.TextFaint).
			Render(fmt.Sprintf(" — %d-%d of %d —", s.scroll+1, end, len(lines)))
		body = append(body, pos)
	}
	return strings.Join(body, "\n")
}

func (s *TextScreen) KeyHints() []KeyHint {
	return []KeyHint{{"jk", "scroll"}}
}
