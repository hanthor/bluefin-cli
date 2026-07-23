# ADR 0001: One persistent TUI shell, native screens only

**Status**: accepted (2026-07)

## Context
The menu system grew organically as independent huh `form.Run()` calls that
cleared and took over the terminal; navigation state, headers, and theming
were rebuilt ad hoc per flow, and drill-down/back behavior was inconsistent.

## Decision
One long-lived bubbletea v2 program (`internal/tui/app`) hosts a stack of
`Screen`s (k9s-style push/pop). Every interactive flow renders inside it via
four wrappers: MenuScreen (lists), FormScreen (embedded huh v2 forms),
TextScreen (read-only, scrollable), RunnerScreen (tasks that print — output
is captured by swapping os.Stdout to a pipe; the renderer keeps its own fd).
Interactive CLI commands (`bluefin-cli fonts`…) launch the same shell at the
right screen, so there is exactly one UI codepath. `RunExternal` (tea.Exec)
survives only for launching genuinely external interactive programs.

## Consequences
Consistent chrome, theming, and keybindings everywhere; flows are unit-
testable as models (no pty). Anything that prints or blocks MUST go through
RunnerScreen — synchronous capture deadlocks (glow terminal queries).
