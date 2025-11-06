# Feature Reference from ublue-os/packages

This document outlines the functionality to replicate and extend from the existing ublue-bling scripts and bluefin-cli tools.

## Core Features to Implement

### 1. Bling Shell Configuration (from ublue-bling)
**Source:** `reference/packages/packages/ublue-bling/src/bling.sh` and `bling.fish`

Features:
- **Aliases:**
  - `eza` aliases for modern `ls` replacement (ll, l., ls, l1)
  - `ugrep` aliases for `grep` replacement (grep, egrep, fgrep, xzgrep, etc.)
  - `bat` for `cat` with syntax highlighting
  
- **Shell Initialization:**
  - Atuin (terminal history sync & search) - must init before starship
  - Starship (prompt theming)
  - Zoxide (smart directory jumping)
  - bash-preexec support for bash shells
  
- **Shell Support:**
  - Bash
  - Zsh  
  - Fish

### 2. Brewfile Management
**Source:** `reference/packages/packages/bluefin/schemas/usr/share/ublue-os/homebrew/bluefin-cli.Brewfile`

Default packages to include:
- **Core Tools:**
  - atuin, bat, eza, fd, gh, glab, ripgrep (rg)
  - starship, zoxide, bash-preexec
  - jq, yq (JSON/YAML processors)
  - ugrep, uutils-coreutils

- **Developer Tools:**
  - chezmoi (dotfiles manager)
  - direnv (directory-specific environment)
  - shellcheck (shell script linter)
  - tealdeer (tldr pages)
  
- **System Tools:**
  - dysk (disk info)
  - stress-ng (system stress testing)
  - trash-cli (safe file deletion)
  - television (fuzzy finder)

- **Taps:**
  - valkyrie00/bbrew
  - Other custom taps as needed

### 3. MOTD (Message of the Day) System
**Source:** `reference/homebrew-experimental-tap/Casks/bluefin-cli.rb` (lines 85+)

Features:
- Display system information on shell startup
- Show random tips/tricks
- Support multiple themes (slate, etc.)
- Use glow for markdown rendering
- Integrate with fastfetch for system info
- Custom logo display (sixels, symbols, ASCII)

### 4. Fastfetch Integration
**Source:** CLI logos and fastfetch config in bluefin package

Features:
- Display system information beautifully
- Show Bluefin/Universal Blue branding
- Support for different logo formats:
  - Sixels (for terminals with sixel support)
  - Symbols (Unicode/Nerd Fonts)
  - ASCII/ANSI logos

### 5. Ujust-style Command Runner
**Source:** `reference/packages/packages/ublue-recipes/` and `ublue-os-just/`

Features to consider:
- Task runner for common operations
- Interactive command selection
- Recipe management for development environments
- VM configuration helpers
- Group management for devmode

## Extended Features (New)

### 6. Interactive Setup Wizards
Using Charm TUI libraries (huh, bubbletea):
- Initial shell setup wizard
- Brewfile customization with package selection
- Starship theme browser with live preview
- Tool installation confirmation dialogs

### 7. Shell Customization
- Starship preset theme selector (interactive)
- Shell plugin recommendations
- Dotfile management integration with chezmoi
- Config backup before changes

### 8. Package Management Helpers
- Search for Homebrew packages
- Show package info before installing
- Dependency visualization
- Bulk install/uninstall operations

## Implementation Priority

1. ✅ Basic CLI structure (Cobra)
2. ⏳ Go module setup and dependencies
3. 📝 Brewfile init/apply/add commands
4. 📝 Shell setup command (bling-style initialization)
5. 📝 Starship theme management
6. 📝 MOTD system (optional feature)
7. 📝 Interactive TUI enhancements
8. 📝 Fastfetch integration
9. 📝 Advanced features (chezmoi, task runner, etc.)

## Technical Notes

- Must handle both bash and zsh (fish optional)
- Homebrew prefix detection: `/home/linuxbrew/.linuxbrew`
- Config files location: `~/.config/bluefin-cli/` or `~/.local/share/bluefin-cli/`
- Shell RC files: `.bashrc`, `.zshrc`, `.config/fish/config.fish`
- Atuin must initialize before starship for proper history capture
- Need to check for existing installations before modifying

## Dependencies for Go Project

Required Charm libraries:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/huh` - Forms and prompts
- `github.com/charmbracelet/bubbles` - TUI components (optional)

CLI framework:
- `github.com/spf13/cobra` - Command structure

Utilities:
- Standard library for file operations
- `os/exec` for running shell commands
