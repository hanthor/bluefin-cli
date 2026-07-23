package app

import (
	"fmt"
	"sort"
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
	Desc    string // one-line description shown under the label
	Hint    string // dimmed annotation, e.g. "bash ✓"
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
	if s.query == "" {
		idx := make([]int, len(s.items))
		for i := range s.items {
			idx[i] = i
		}
		return idx
	}
	type scored struct{ idx, score int }
	matches := make([]scored, 0, len(s.items))
	for i, it := range s.items {
		hay := strings.ToLower(it.Label)
		sc, ok := fuzzyScore(hay, strings.ToLower(s.query))
		if !ok && it.Desc != "" {
			// Fall back to matching the description, ranked below any
			// label match.
			sc, ok = fuzzyScore(strings.ToLower(it.Desc), strings.ToLower(s.query))
			sc -= 1000
		}
		if ok {
			matches = append(matches, scored{i, sc})
		}
	}
	sort.SliceStable(matches, func(a, b int) bool { return matches[a].score > matches[b].score })
	idx := make([]int, len(matches))
	for i, m := range matches {
		idx[i] = m.idx
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
	default:
		// Number keys jump straight to that row and select it.
		if k := key.String(); len(k) == 1 && k[0] >= '1' && k[0] <= '9' {
			if n := int(k[0] - '1'); n < len(s.visible()) {
				s.cursor = n
				return s.selectCurrent()
			}
		}
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
		prompt := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render(" 🔎 ") + s.query
		if s.filtering {
			prompt += lipgloss.NewStyle().Foreground(t.Accent).Render("▏")
		}
		if s.query != "" {
			prompt += lipgloss.NewStyle().Foreground(t.TextFaint).
				Render(fmt.Sprintf("   %d/%d", len(s.visible()), len(s.items)))
		}
		b.WriteString(prompt + "\n\n")
	} else {
		b.WriteString("\n")
	}

	vis := s.visible()
	hintStyle := lipgloss.NewStyle().Foreground(t.TextFaint)
	if len(vis) == 0 {
		b.WriteString(hintStyle.Render("  no matches — esc clears the filter"))
		return b.String()
	}

	// Card layout: each item is a two-line block (label + description).
	// The selected card gets an accent bar and a tinted highlight across
	// both lines. Falls back to compact single-line rows when the terminal
	// is too small for cards.
	cards := width >= 56 && height >= len(vis)*2+2

	// Viewport: when the list is taller than the body, window it around the
	// cursor and mark the clipped ends.
	rowsPer := 1
	if cards {
		rowsPer = 2
	}
	avail := max((height-2)/rowsPer, 1)
	start := 0
	if len(vis) > avail {
		start = min(max(s.cursor-avail/2, 0), len(vis)-avail)
	}
	end := min(start+avail, len(vis))

	rowWidth := max(width-2, 10)
	selLabel := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Background(t.SelectedBackground)
	selDesc := lipgloss.NewStyle().Foreground(t.TextMuted).Background(t.SelectedBackground)
	selHint := lipgloss.NewStyle().Foreground(t.AccentAlt).Background(t.SelectedBackground)
	bar := lipgloss.NewStyle().Foreground(t.Accent).Background(t.SelectedBackground).Render("▎")
	normLabel := lipgloss.NewStyle().Foreground(t.TextBase)
	normDesc := lipgloss.NewStyle().Foreground(t.TextFaint)
	marker := lipgloss.NewStyle().Foreground(t.TextFaint)

	pad := func(styled string, style lipgloss.Style) string {
		if w := lipgloss.Width(styled); w < rowWidth {
			return styled + style.Render(strings.Repeat(" ", rowWidth-w))
		}
		return styled
	}

	if start > 0 {
		b.WriteString(hintStyle.Render("   … more above") + "\n")
	}
	for row := start; row < end; row++ {
		i := vis[row]
		it := s.items[i]
		label := it.Label
		if it.Icon != "" {
			label = it.Icon + " " + label
		}
		suffix := ""
		if it.Submenu {
			suffix = " ›"
		}

		if row == s.cursor {
			line := bar + selLabel.Render(" "+label+suffix)
			if it.Hint != "" {
				line += selHint.Render("  " + it.Hint)
			}
			b.WriteString(pad(line, selLabel) + "\n")
			if cards {
				desc := bar + selDesc.Render("   "+it.Desc)
				b.WriteString(pad(desc, selDesc) + "\n")
			}
		} else {
			line := " " + normLabel.Render(" "+label) + marker.Render(suffix)
			if it.Hint != "" {
				line += hintStyle.Render("  " + it.Hint)
			}
			b.WriteString(line + "\n")
			if cards {
				b.WriteString(normDesc.Render("    " + it.Desc))
				b.WriteString("\n")
			}
		}
	}
	if end < len(vis) {
		b.WriteString(hintStyle.Render("   … more below") + "\n")
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
	_, ok := fuzzyScore(haystack, needle)
	return ok
}

// fuzzyScore reports whether needle is a subsequence of haystack and how
// good the match is: tighter spans, earlier starts, and prefix matches
// score higher, so "sta" ranks "Starship" above "Install Apps".
func fuzzyScore(haystack, needle string) (int, bool) {
	if needle == "" {
		return 0, true
	}
	i, first, last := 0, -1, -1
	for pos, r := range haystack {
		if rune(needle[i]) == r {
			if first < 0 {
				first = pos
			}
			last = pos
			i++
			if i == len(needle) {
				score := -(last - first) - first
				if first == 0 {
					score += 100
				}
				return score, true
			}
		}
	}
	return 0, false
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
