package shell

import (
	"fmt"
	"strings"
)

// Tool represents a CLI tool that can be managed by bluefin-cli
type Tool struct {
	Name        string // Display name
	Description string // Short description
	Binary      string // Binary name to check for
	Pkg         string // Homebrew package name
	Default     bool   // Whether enabled by default
	ShellDefaults map[string]bool // Per-shell default overrides
}

// GetEnvVar returns the environment variable name for this tool
func (t Tool) GetEnvVar() string {
	return fmt.Sprintf("BLUEFIN_SHELL_ENABLE_%s", strings.ToUpper(t.Name))
}

// Tools is the list of managed tools
var Tools = []Tool{
	{Name: "Eza", Description: "Modern, maintained replacement for ls", Binary: "eza", Pkg: "eza", Default: true},
	{Name: "Ugrep", Description: "Ultra fast grep with interactive mode", Binary: "ug", Pkg: "ugrep", Default: true},
	{Name: "Bat", Description: "A cat clone with wings", Binary: "bat", Pkg: "bat", Default: true},
	{Name: "Atuin", Description: "Magical shell history", Binary: "atuin", Pkg: "atuin", Default: false, ShellDefaults: map[string]bool{"zsh": true, "fish": true}},
	{Name: "Starship", Description: "The minimal, blazing-fast, and infinitely customizable prompt", Binary: "starship", Pkg: "starship", Default: true},
	{Name: "Zoxide", Description: "A smarter cd command", Binary: "zoxide", Pkg: "zoxide", Default: true},
	{Name: "UutilsCoreutils", Description: "Rust rewrite of GNU coreutils", Binary: "hashsum", Pkg: "uutils-coreutils", Default: true},
	{Name: "UutilsFindutils", Description: "Rust rewrite of GNU findutils", Binary: "ufind", Pkg: "uutils-findutils", Default: true},
	{Name: "UutilsDiffutils", Description: "Rust rewrite of GNU diffutils", Binary: "udiffutils", Pkg: "uutils-diffutils", Default: true},
	{Name: "Carapace", Description: "Multi-shell multi-command argument completer", Binary: "carapace", Pkg: "carapace", Default: false},
}
