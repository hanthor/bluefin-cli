# AGENTS.md

## 🤖 Project Overview
**Bluefin CLI** is a Go-based command-line tool designed for the Bluefin OS (and other Universal Blue derivatives). It serves as a unified interface for system customization, "bling" (shell enhancements), and software installation via Homebrew bundles.

## 🏗 Architecture
The project follows a standard Go CLI structure:
- **`cmd/`**: Contains the main entry point and Cobra commands. Each command file (e.g., `bling.go`, `install.go`) typically handles the CLI arguments and delegates logic to `internal/`.
- **`internal/`**: Contains the core business logic, separated by domain (`bling`, `install`, `motd`, etc.).
- **TUI**: Heavily uses [Bubble Tea](https://charm.land/bubbletea/v2), [Lipgloss](https://charm.land/lipgloss/v2), and [Huh](https://charm.land/huh/v2) for interactive menus.

### Key Components
1.  **Bling (`internal/bling`)**: Manages shell configuration files (`.bashrc`, `.zshrc`, `config.fish`). It embeds shell scripts (`resources/bling.sh`, `resources/bling.fish`) and sources them in the user's shell config.
2.  **Bundles (`internal/install`)**: Defines a hardcoded list of "bundles" (e.g., ai, k8s) mapping to remote Brewfiles hosted on GitHub. It downloads these Brewfiles and runs `brew bundle install`.

## 🛠 Development
The project uses `just` as a task runner.

### Common Commands
- **`just build`**: Builds the binary locally.
- **`just test`**: Runs integration tests inside a container.
- **`just unit-test`**: Runs all Go unit tests inside a container.
- **`just shell-with-bling`**: Spawns a container with the CLI pre-installed and "bling" enabled for manual testing.
- **`just inspect-bling`**: Verifies that shell configuration files are correctly modified.

### Testing Guidelines
- **Unit Tests**: Add unit tests for new logic in the respective `internal/` package. Run them locally with `go test ./internal/...` or via `just unit-test`.
- **Integration Tests**: Integration tests are located in `test/integration_test.go`. Run them via `just test`.
- **Manual Verification**: Use `just shell-with-bling` to verify UI and shell configuration changes in an isolated environment.

### Guidelines for Agents
- **Changing Bundles**: Update `internal/install/install.go` to add/remove bundles or change the source URL.
- **Modifying Bling**:
    - If changing the shell script logic, edit `internal/bling/resources/bling.sh` or `.fish`.
    - If changing how it hooks into shells, edit `internal/bling/bling.go`.
- **UI Changes**: The interactive menu is a persistent Bubble Tea v2 shell in `internal/tui/app` (screen stack + breadcrumb header + contextual footer + ctrl+p palette + the header dino). Menus are `app.MenuScreen`s built in `cmd/menu.go`; every destination is also registered as an `app.Action` for the palette. Legacy flows (huh forms with their own `Run()`, brew/winget installs that stream to stdout) run through `app.RunLegacy`, which releases the terminal and resumes the shell — migrate these to native screens over time rather than adding new `tui.ClearScreen()`-style flows. Colors come from semantic tokens in `internal/tui/theme` (Catppuccin Latte/Mocha, auto light/dark from the terminal background) — never hardcode hex/ANSI colors in views.
