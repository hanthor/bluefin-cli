package app

import (
	tea "charm.land/bubbletea/v2"

	"charm.land/huh/v2"
)

// FormScreen hosts a huh form natively inside the shell, so submenus keep
// the header, dino, and theming instead of releasing the terminal.
type FormScreen struct {
	title  string
	build  func() *huh.Form
	form   *huh.Form
	onDone func(aborted bool) tea.Cmd
}

// NewForm creates a form screen. build constructs a fresh form each time the
// screen initializes; onDone runs after the form completes or aborts (the
// screen pops itself either way).
func NewForm(title string, build func() *huh.Form, onDone func(aborted bool) tea.Cmd) *FormScreen {
	return &FormScreen{title: title, build: build, onDone: onDone}
}

func (s *FormScreen) Title() string { return s.title }

func (s *FormScreen) Init() tea.Cmd {
	s.form = s.build()
	return s.form.Init()
}

// CapturingInput is always true: the form owns navigation keys; huh maps
// esc/ctrl+c to abort, which pops the screen.
func (s *FormScreen) CapturingInput() bool { return true }

func (s *FormScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if s.form == nil {
		return s, nil
	}
	model, cmd := s.form.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		s.form = f
	}
	switch s.form.State {
	case huh.StateCompleted:
		done := s.onDone
		s.form = nil
		if done != nil {
			return s, tea.Sequence(Pop(), done(false))
		}
		return s, Pop()
	case huh.StateAborted:
		done := s.onDone
		s.form = nil
		if done != nil {
			return s, tea.Sequence(Pop(), done(true))
		}
		return s, Pop()
	}
	return s, cmd
}

func (s *FormScreen) View(width, height int) string {
	if s.form == nil {
		return ""
	}
	return s.form.View()
}

func (s *FormScreen) KeyHints() []KeyHint {
	return []KeyHint{
		{"space", "toggle"},
		{"enter", "confirm"},
		{"esc", "cancel"},
	}
}
