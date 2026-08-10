package app

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
	"github.com/tuna-os/bluefin-cli/internal/wallpaper"
)

// WallpaperPickerScreen is a custom TUI screen that displays a list of wallpaper
// candidates on the left and a half-block/graphics PixelCanvas preview on the right.
type WallpaperPickerScreen struct {
	title       string
	items       []MenuItem
	onSelect    func(MenuItem) tea.Cmd
	cursor      int
	filtering   bool
	query       string
	scaledCache map[string]image.Image
}

func NewWallpaperPicker(title string, items []MenuItem, onSelect func(MenuItem) tea.Cmd) *WallpaperPickerScreen {
	return &WallpaperPickerScreen{
		title:       title,
		items:       items,
		onSelect:    onSelect,
		scaledCache: make(map[string]image.Image),
	}
}

func (s *WallpaperPickerScreen) Title() string { return s.title }

func (s *WallpaperPickerScreen) Init() tea.Cmd {
	s.clampCursor()
	return nil
}

func (s *WallpaperPickerScreen) CapturingInput() bool { return s.filtering }

func (s *WallpaperPickerScreen) visible() []int {
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
			sc, ok = fuzzyScore(strings.ToLower(it.Desc), strings.ToLower(s.query))
			sc -= 1000
		}
		if ok {
			matches = append(matches, scored{i, sc})
		}
	}
	idx := make([]int, len(matches))
	for i, m := range matches {
		idx[i] = m.idx
	}
	return idx
}

func (s *WallpaperPickerScreen) clampCursor() {
	if n := len(s.visible()); s.cursor >= n {
		s.cursor = max(n-1, 0)
	}
}

func (s *WallpaperPickerScreen) moveCursor(delta int) {
	n := len(s.visible())
	if n == 0 {
		return
	}
	s.cursor = (s.cursor + delta + n) % n
}

func (s *WallpaperPickerScreen) selectCurrent() (Screen, tea.Cmd) {
	vis := s.visible()
	if len(vis) == 0 || s.onSelect == nil {
		return s, nil
	}
	item := s.items[vis[s.cursor]]
	s.query = ""
	return s, s.onSelect(item)
}

func (s *WallpaperPickerScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
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
		if k := key.String(); len(k) == 1 && k[0] >= '1' && k[0] <= '9' {
			if n := int(k[0] - '1'); n < len(s.visible()) {
				s.cursor = n
				return s.selectCurrent()
			}
		}
	}
	return s, nil
}

func (s *WallpaperPickerScreen) View(width, height int) string {
	t := theme.DefaultTheme

	// Determine layout: if wide enough (>= 70), split into list (left) + preview (right)
	split := width >= 70
	listWidth := width
	previewWidth := 0
	if split {
		listWidth = width / 2
		previewWidth = width - listWidth - 2
	}

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

	cards := listWidth >= 35 && height >= len(vis)*2+2
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

	rowWidth := max(listWidth-2, 10)
	selLabel := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Background(t.SelectedBackground)
	selDesc := lipgloss.NewStyle().Foreground(t.TextMuted).Background(t.SelectedBackground)
	bar := lipgloss.NewStyle().Foreground(t.Accent).Background(t.SelectedBackground).Render("▎")
	normLabel := lipgloss.NewStyle().Foreground(t.TextBase)
	normDesc := lipgloss.NewStyle().Foreground(t.TextFaint)

	pad := func(styled string, style lipgloss.Style) string {
		if w := lipgloss.Width(styled); w < rowWidth {
			return styled + style.Render(strings.Repeat(" ", rowWidth-w))
		}
		return styled
	}

	var listLines []string
	if start > 0 {
		listLines = append(listLines, hintStyle.Render("   … more above"))
	}
	for row := start; row < end; row++ {
		i := vis[row]
		it := s.items[i]
		label := it.Label
		if it.Icon != "" {
			label = it.Icon + " " + label
		}

		if row == s.cursor {
			line := bar + selLabel.Render(" "+label)
			listLines = append(listLines, pad(line, selLabel))
			if cards {
				desc := bar + selDesc.Render("   "+it.Desc)
				listLines = append(listLines, pad(desc, selDesc))
			}
		} else {
			line := " " + normLabel.Render(" "+label)
			listLines = append(listLines, line)
			if cards {
				listLines = append(listLines, normDesc.Render("    "+it.Desc))
			}
		}
	}
	if end < len(vis) {
		listLines = append(listLines, hintStyle.Render("   … more below"))
	}

	if !split {
		return strings.Join(listLines, "\n")
	}

	// Render preview panel on right side
	previewLines := s.renderPreview(vis[s.cursor], previewWidth, height-2)

	listBlock := lipgloss.NewStyle().Width(listWidth).Render(strings.Join(listLines, "\n"))
	previewBlock := lipgloss.NewStyle().Width(previewWidth).Render(previewLines)

	return lipgloss.JoinHorizontal(lipgloss.Top, listBlock, "  ", previewBlock)
}

func (s *WallpaperPickerScreen) renderPreview(itemIndex int, w, rows int) string {
	if itemIndex < 0 || itemIndex >= len(s.items) {
		return ""
	}
	path := s.items[itemIndex].Value
	if path == "" {
		return ""
	}

	t := theme.DefaultTheme
	dim := lipgloss.NewStyle().Foreground(t.TextFaint)

	// Check scale cache
	cacheKey := fmt.Sprintf("%s_%dx%d", path, w, rows*2)
	scaled, ok := s.scaledCache[cacheKey]
	if !ok {
		var err error
		scaled, err = wallpaper.DecodeAndScale(path, w, rows*2)
		if err != nil {
			return dim.Render("Preview unavailable:\n" + filepath.Base(path))
		}
		s.scaledCache[cacheKey] = scaled
	}

	cv := NewRasterCanvas(w, rows, DefaultRasterBackend)
	bounds := scaled.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			cv.Set(x, y, scaled.At(x, y))
		}
	}

	return cv.RenderImage()
}

func (s *WallpaperPickerScreen) KeyHints() []KeyHint {
	return []KeyHint{
		{"↑↓/jk", "move"},
		{"enter/→", "select"},
		{"/", "filter"},
	}
}
