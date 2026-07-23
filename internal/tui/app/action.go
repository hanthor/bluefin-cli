package app

import (
	tea "charm.land/bubbletea/v2"
)

// Action is a single invokable capability. Menus and the ctrl+p command
// palette both render from the same registry, so anything reachable through
// the menu tree is also one fuzzy search away.
type Action struct {
	ID        string
	Icon      string
	Label     string
	Section   string
	Available func() bool // nil means always available
	Do        func() tea.Cmd
}

var registry []Action

// Register adds actions to the global registry. Called from init-time wiring
// in cmd; the extra build variant registers additional actions.
func Register(actions ...Action) {
	registry = append(registry, actions...)
}

// Actions returns the currently-available registered actions.
func Actions() []Action {
	out := make([]Action, 0, len(registry))
	for _, a := range registry {
		if a.Available == nil || a.Available() {
			out = append(out, a)
		}
	}
	return out
}

// newPalette builds the ctrl+p command palette from the registry.
func newPalette() Screen {
	byID := map[string]Action{}
	items := func() []MenuItem {
		actions := Actions()
		out := make([]MenuItem, 0, len(actions))
		for _, a := range actions {
			byID[a.ID] = a
			out = append(out, MenuItem{Icon: a.Icon, Label: a.Label, Value: a.ID, Desc: a.Section})
		}
		return out
	}
	menu := NewMenu("Palette", nil, items, func(it MenuItem) tea.Cmd {
		a, ok := byID[it.Value]
		if !ok {
			return nil
		}
		return tea.Sequence(Pop(), a.Do())
	})
	menu.filtering = true
	return menu
}
