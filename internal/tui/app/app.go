// Package app is the persistent TUI shell: one bubbletea program hosting a
// stack of screens with a shared header (breadcrumb + dino), a contextual
// footer, adaptive theming, toasts, and a help overlay.
//
// Navigation model: screens are pushed onto the stack to drill down and
// popped with esc/left/backspace (k9s-style). Legacy interactive flows run
// through RunLegacy, which releases the terminal and resumes the shell after.
package app

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// KeyHint is a single entry in the contextual footer / help overlay.
type KeyHint struct {
	Keys string
	Desc string
}

// Screen is one level of the navigation stack.
type Screen interface {
	Title() string
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	// View renders the screen body within the given content area.
	View(width, height int) string
	KeyHints() []KeyHint
}

// InputCapturer is implemented by screens that are currently consuming text
// input (e.g. an active filter or an embedded form); while capturing, the
// shell passes keys through instead of treating them as navigation.
type InputCapturer interface {
	CapturingInput() bool
}

// Reloader is implemented by screens whose content should refresh when they
// become the top of the stack again (e.g. enabled/disabled labels).
type Reloader interface {
	Reload() tea.Cmd
}

// Messages for stack navigation and feedback.
type (
	PushMsg  struct{ Screen Screen }
	PopMsg   struct{}
	ToastMsg struct {
		Text  string
		IsErr bool
	}
	toastExpireMsg struct{}
	dinoTickMsg    struct{}
)

// Push returns a command that pushes a screen onto the stack.
func Push(s Screen) tea.Cmd {
	return func() tea.Msg { return PushMsg{Screen: s} }
}

// Pop returns a command that pops the top screen.
func Pop() tea.Cmd {
	return func() tea.Msg { return PopMsg{} }
}

// Toast returns a command that shows a transient status message.
func Toast(text string, isErr bool) tea.Cmd {
	return func() tea.Msg { return ToastMsg{Text: text, IsErr: isErr} }
}

const (
	headerHeight = 3 // title row, dino row, rule row
	footerHeight = 2 // blank + hints
	minBodyLines = 3
)

// Model is the root shell model.
type Model struct {
	stack    []Screen
	width    int
	height   int
	theme    theme.Theme
	showHelp bool
	toast    string
	toastErr bool
	dino     dino
	quitting bool
}

// New creates the shell with the given root screen.
func New(root Screen) Model {
	return Model{
		stack: []Screen{root},
		theme: theme.DefaultTheme,
	}
}

// Run starts the shell program on the terminal.
func Run(root Screen) error {
	_, err := tea.NewProgram(New(root)).Run()
	return err
}

func (m Model) top() Screen { return m.stack[len(m.stack)-1] }

func (m Model) capturing() bool {
	if c, ok := m.top().(InputCapturer); ok {
		return c.CapturingInput()
	}
	return false
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.top().Init(), m.dino.tick(), tea.RequestBackgroundColor)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m.delegate(msg)

	case tea.BackgroundColorMsg:
		m.theme = theme.Resolve(isDarkColor(msg))
		theme.DefaultTheme = m.theme
		return m, nil

	case dinoTickMsg:
		m.dino.advance(m.width)
		return m, m.dino.tick()

	case PushMsg:
		m.stack = append(m.stack, msg.Screen)
		m.showHelp = false
		return m, msg.Screen.Init()

	case PopMsg:
		if len(m.stack) <= 1 {
			m.quitting = true
			return m, tea.Quit
		}
		m.stack = m.stack[:len(m.stack)-1]
		if r, ok := m.top().(Reloader); ok {
			return m, r.Reload()
		}
		return m, nil

	case ToastMsg:
		m.toast, m.toastErr = msg.Text, msg.IsErr
		return m, tea.Tick(4*time.Second, func(time.Time) tea.Msg { return toastExpireMsg{} })

	case toastExpireMsg:
		m.toast = ""
		return m, nil

	case LegacyDoneMsg:
		cmds := []tea.Cmd{}
		if r, ok := m.top().(Reloader); ok {
			cmds = append(cmds, r.Reload())
		}
		if msg.Err != nil {
			cmds = append(cmds, Toast("Error: "+msg.Err.Error(), true))
		}
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		if !m.capturing() {
			switch key {
			case "esc", "left", "backspace", "h":
				return m.Update(PopMsg{})
			case "q":
				m.quitting = true
				return m, tea.Quit
			case "?":
				m.showHelp = true
				return m, nil
			case "ctrl+p":
				if len(registry) > 0 {
					return m.Update(PushMsg{Screen: newPalette()})
				}
			}
		}
		return m.delegate(msg)
	}

	return m.delegate(msg)
}

func (m Model) delegate(msg tea.Msg) (tea.Model, tea.Cmd) {
	s, cmd := m.top().Update(msg)
	m.stack[len(m.stack)-1] = s
	return m, cmd
}

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.width == 0 {
		// Terminal size not reported yet; render for a conventional 80x24.
		m.width, m.height = 80, 24
	}
	t := m.theme

	header := m.renderHeader()
	footer := m.renderFooter()

	bodyHeight := max(m.height-headerHeight-footerHeight, minBodyLines)
	var body string
	if m.showHelp {
		body = m.renderHelp(bodyHeight)
	} else {
		body = m.top().View(m.width, bodyHeight)
	}
	body = lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(body)

	view := tea.NewView(strings.Join([]string{header, body, footer}, "\n"))
	view.BackgroundColor = t.Bg
	return view
}

func (m Model) renderHeader() string {
	t := m.theme
	crumbs := make([]string, 0, len(m.stack))
	for _, s := range m.stack {
		crumbs = append(crumbs, s.Title())
	}

	title := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render(" Bluefin CLI")
	sep := lipgloss.NewStyle().Foreground(t.TextFaint).Render(" › ")
	crumb := lipgloss.NewStyle().Foreground(t.TextMuted).Render(strings.Join(crumbs, " › "))
	titleRow := title + sep + crumb

	dinoRow := m.dino.render(m.width, t)
	rule := lipgloss.NewStyle().Foreground(t.Surface).Render(strings.Repeat("─", max(m.width, 1)))

	return titleRow + "\n" + dinoRow + "\n" + rule
}

func (m Model) renderFooter() string {
	t := m.theme
	if m.toast != "" {
		style := lipgloss.NewStyle().Foreground(t.Success)
		if m.toastErr {
			style = lipgloss.NewStyle().Foreground(t.Error)
		}
		return "\n " + style.Render(m.toast)
	}

	hints := m.top().KeyHints()
	hints = append(hints, globalHints(len(m.stack) > 1)...)

	keyStyle := lipgloss.NewStyle().Foreground(t.Accent)
	descStyle := lipgloss.NewStyle().Foreground(t.TextFaint)
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, keyStyle.Render(h.Keys)+" "+descStyle.Render(h.Desc))
	}
	line := " " + strings.Join(parts, descStyle.Render("  ·  "))
	return "\n" + lipgloss.NewStyle().MaxWidth(m.width).Render(line)
}

func (m Model) renderHelp(height int) string {
	t := m.theme
	hints := append(m.top().KeyHints(), globalHints(len(m.stack) > 1)...)

	keyStyle := lipgloss.NewStyle().Foreground(t.Accent).Width(14)
	descStyle := lipgloss.NewStyle().Foreground(t.TextBase)
	rows := make([]string, 0, len(hints)+2)
	rows = append(rows, lipgloss.NewStyle().Foreground(t.TextMuted).Bold(true).Render("Keys"), "")
	for _, h := range hints {
		rows = append(rows, keyStyle.Render(h.Keys)+descStyle.Render(h.Desc))
	}

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Overlay).
		Padding(1, 3).
		Render(strings.Join(rows, "\n"))
	return lipgloss.Place(m.width, height, lipgloss.Center, lipgloss.Center, panel)
}

func globalHints(canGoBack bool) []KeyHint {
	hints := []KeyHint{}
	if canGoBack {
		hints = append(hints, KeyHint{"esc/←", "back"})
	}
	hints = append(hints,
		KeyHint{"?", "help"},
		KeyHint{"q", "quit"},
	)
	return hints
}

// isDarkColor reports whether a terminal background color is dark, using
// relative luminance.
func isDarkColor(c color.Color) bool {
	if c == nil {
		return true
	}
	r, g, b, _ := c.RGBA()
	lum := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	return lum < 0.5*0xffff
}
