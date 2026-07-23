package app

import (
	"errors"
	"io"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// ExternalDoneMsg reports the result of an external program run via RunExternal.
type ExternalDoneMsg struct{ Err error }

// execFn adapts a plain function to tea.ExecCommand so legacy flows (huh
// forms with their own Run loop, brew/winget commands writing to stdout) can
// run with the terminal released, then return control to the shell.
type execFn struct {
	fn func() error
}

func (e *execFn) Run() error          { return e.fn() }
func (e *execFn) SetStdin(io.Reader)  {}
func (e *execFn) SetStdout(io.Writer) {}
func (e *execFn) SetStderr(io.Writer) {}

// RunExternal releases the terminal for a genuinely external interactive
// program (e.g. the WSL->Windows CLI delegation), then resumes the shell. A
// huh.ErrUserAborted from fn is treated as a normal "back" and swallowed.
func RunExternal(fn func() error) tea.Cmd {
	return tea.Exec(&execFn{fn: fn}, func(err error) tea.Msg {
		if errors.Is(err, huh.ErrUserAborted) {
			err = nil
		}
		return ExternalDoneMsg{Err: err}
	})
}
