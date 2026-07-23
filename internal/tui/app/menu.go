package app

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// MenuItem is one selectable row in a MenuScreen.
type MenuItem struct {
	Icon    string
	Label   string
	Value   string
	Hint    string // dimmed annotation, e.g. "Enabled"
	Submenu bool   // renders a › marker
}

// MenuScreen is a filterable list of items; selecting one runs OnSelect.
type MenuScreen struct {
	title     string
	items     []MenuItem
	refresh   func() []MenuItem
	onSelect  func(MenuItem) tea.Cmd
	cursor    int
	filtering bool
	query     string
}

// NewMenu creates a menu screen. refresh recomputes the items whenever the
// screen (re)gains focus, so status labels stay current; it may be nil for
// static menus, in which case items is used as-is.
func NewMenu(title string, items []MenuItem, refresh func() []MenuItem, onSelect func(MenuItem) tea.Cmd) *MenuScreen {
	return &MenuScreen{title: title, items: items, refresh: refresh, onSelect: onSelect}
}

func (s *MenuScreen) Title() string { return s.title }

func (s *MenuScreen) Init() tea.Cmd { return s.Reload() }

// Reload re-runs the refresh function (Reloader interface).
func (s *MenuScreen) Reload() tea.Cmd {
	if s.refresh != nil {
		s.items = s.refresh()
	}
	s.clampCursor()
	return nil
}

// CapturingInput reports whether the filter is consuming keystrokes.
func (s *MenuScreen) CapturingInput() bool { return s.filtering }

func (s *MenuScreen) visible() []int {
	idx := make([]int, 0, len(s.items))
	for i, it := range s.items {
		if s.query == "" || fuzzyMatch(strings.ToLower(it.Label), strings.ToLower(s.query)) {
			idx = append(idx, i)
		}
	}
	return idx
}

func (s *MenuScreen) clampCursor() {
	if n := len(s.visible()); s.cursor >= n {
		s.cursor = max(n-1, 0)
	}
}

func (s *MenuScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}

	if s.filtering {
		switch key.String() {
		case "esc":
			s.filtering, s.query = false, ""
			s.clampCursor()
			return s, nil
		case "enter":
			s.filtering = false
			return s.selectCurrent()
		case "backspace":
			if s.query != "" {
				s.query = s.query[:len(s.query)-1]
			} else {
				s.filtering = false
			}
			s.clampCursor()
			return s, nil
		case "up", "ctrl+p":
			s.moveCursor(-1)
			return s, nil
		case "down", "ctrl+n":
			s.moveCursor(1)
			return s, nil
		default:
			if t := key.Text; t != "" && isPrintable(t) {
				s.query += t
				s.cursor = 0
			}
			return s, nil
		}
	}

	switch key.String() {
	case "up", "k":
		s.moveCursor(-1)
	case "down", "j":
		s.moveCursor(1)
	case "g", "home":
		s.cursor = 0
	case "G", "end":
		s.cursor = max(len(s.visible())-1, 0)
	case "/":
		s.filtering = true
		s.query = ""
	case "enter", "right", "l":
		return s.selectCurrent()
	}
	return s, nil
}

func (s *MenuScreen) moveCursor(delta int) {
	n := len(s.visible())
	if n == 0 {
		return
	}
	s.cursor = (s.cursor + delta + n) % n
}

func (s *MenuScreen) selectCurrent() (Screen, tea.Cmd) {
	vis := s.visible()
	if len(vis) == 0 || s.onSelect == nil {
		return s, nil
	}
	item := s.items[vis[s.cursor]]
	s.query = ""
	return s, s.onSelect(item)
}

func (s *MenuScreen) View(width, height int) string {
	t := theme.DefaultTheme
	var b strings.Builder

	if s.filtering || s.query != "" {
		prompt := lipgloss.NewStyle().Foreground(t.Accent).Render("/") + s.query
		if s.filtering {
			prompt += lipgloss.NewStyle().Foreground(t.Accent).Render("▏")
		}
		b.WriteString(" " + prompt + "\n\n")
	} else {
		b.WriteString("\n")
	}

	// The cursor row gets a full-width background highlight; unselected rows
	// stay quiet with faint hints and submenu markers.
	sel := lipgloss.NewStyle().
		Foreground(t.Accent).
		Background(t.SelectedBackground).
		Bold(true)
	norm := lipgloss.NewStyle().Foreground(t.TextMuted)
	hintStyle := lipgloss.NewStyle().Foreground(t.TextFaint)
	marker := lipgloss.NewStyle().Foreground(t.TextFaint).Render(" ›")

	vis := s.visible()
	if len(vis) == 0 {
		b.WriteString(hintStyle.Render("  no matches"))
	}
	rowWidth := max(width-2, 10)
	for row, i := range vis {
		it := s.items[i]
		line := it.Label
		if it.Icon != "" {
			line = it.Icon + " " + line
		}
		if row == s.cursor {
			text := " ❯ " + line
			if it.Hint != "" {
				text += "  " + it.Hint
			}
			if it.Submenu {
				text += " ›"
			}
			b.WriteString(sel.Width(rowWidth).Render(text))
		} else {
			b.WriteString(norm.Render("   " + line))
			if it.Hint != "" {
				b.WriteString(hintStyle.Render("  " + it.Hint))
			}
			if it.Submenu {
				b.WriteString(marker)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (s *MenuScreen) KeyHints() []KeyHint {
	return []KeyHint{
		{"↑↓/jk", "move"},
		{"enter/→", "select"},
		{"/", "filter"},
	}
}

// fuzzyMatch reports whether needle is a subsequence of haystack.
func fuzzyMatch(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	i := 0
	for _, r := range haystack {
		if rune(needle[i]) == r {
			i++
			if i == len(needle) {
				return true
			}
		}
	}
	return false
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
