# Bluefin CLI

A powerful, modern CLI tool for managing Homebrew packages, shell configuration, and development environment customization. Built with beautiful TUIs using [Charm](https://charm.sh/) libraries.

## 🎯 Features

- **✨ Bling**: Toggle modern shell enhancements (eza, bat, ugrep, zoxide, atuin, starship)
- **📰 MOTD**: Beautiful Message of the Day with system info and random tips
- **📦 Brewfile Management**: Create, edit, and apply Brewfile configurations
- **� Bundle Installer**: Install curated bundles (ai, cli, fonts, k8s) from Universal Blue
- **�🐚 Shell Setup**: Interactive shell configuration for Bash, Zsh, and Fish
- **🎨 Starship Themes**: Browse and apply Starship prompt themes
- **📊 Status Command**: View configuration and installed tools at a glance
- **🎪 Beautiful TUIs**: Interactive forms and prompts powered by Charm libraries

## 🚀 Installation

### Prerequisites

- Go 1.21 or later
- Homebrew (for package management features)

### Build from Source

```bash
git clone https://github.com/yourusername/bluefin-cli.git
cd bluefin-cli
go build -o bluefin-cli
sudo mv bluefin-cli /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourusername/bluefin-cli@latest
```

## 📖 Usage

### Check Status

View your current configuration and installed tools:

```bash
bluefin-cli status
```

### Bling - Modern Shell Enhancements

Enable bling for your shell:

```bash
# Interactive mode
bluefin-cli bling

# Enable for specific shell
bluefin-cli bling bash on
bluefin-cli bling zsh on
bluefin-cli bling fish on

# Disable bling
bluefin-cli bling bash off
```

### MOTD - Message of the Day

Show the MOTD:

```bash
bluefin-cli motd show
```

Toggle MOTD for shells:

```bash
# Enable for all shells
bluefin-cli motd toggle all on

# Enable for specific shell
bluefin-cli motd toggle zsh on

# Disable MOTD
bluefin-cli motd toggle all off
```

Configure MOTD theme:

```bash
bluefin-cli motd config
```

### Install Bundles

Install curated Homebrew bundles:

```bash
# List available bundles
bluefin-cli install list

# Install specific bundle
bluefin-cli install ai       # AI tools
bluefin-cli install cli      # CLI essentials
bluefin-cli install fonts    # Development fonts
bluefin-cli install k8s      # Kubernetes tools
bluefin-cli install all      # All bundles

# Install from local Brewfile
bluefin-cli install ./my-custom.Brewfile
```

### Brewfile Management

Initialize a new Brewfile with common development tools:

```bash
bluefin-cli brewfile init
```

Add a package to your Brewfile:

```bash
bluefin-cli brewfile add neovim
```

Install all packages from your Brewfile:

```bash
bluefin-cli brewfile apply
```

### Shell Configuration

Run the interactive shell setup wizard:

```bash
bluefin-cli shell setup
```

This will guide you through:
- Selecting your preferred shell (Bash, Zsh, Fish)
- Installing Oh My Zsh (for Zsh users)
- Setting up Starship prompt
- Configuring modern CLI tools (eza, bat, zoxide, atuin)

### Starship Themes

Browse and apply Starship preset themes:

```bash
bluefin-cli starship theme
```

Install Starship if not already present:

```bash
bluefin-cli starship install
```

## 🔧 What Gets Configured

The shell setup configures these modern CLI tools:

- **eza**: Modern replacement for `ls` with icons and colors
- **bat**: `cat` clone with syntax highlighting
- **zoxide**: Smarter `cd` command that learns your habits
- **atuin**: Magical shell history with sync and search
- **starship**: Fast, customizable prompt for any shell
- **ugrep**: Ultra-fast grep alternative

### Aliases

When bling is sourced in your shell:

```bash
ll      # eza -l --icons=auto --group-directories-first
ls      # eza
cat     # bat --style=plain --pager=never
grep    # ugrep
```

## 🏗️ Project Structure

```
bluefin-cli/
├── main.go              # Application entry point
├── cmd/                 # Cobra commands
│   ├── root.go         # Root command
│   ├── brewfile.go     # Brewfile management
│   ├── shell.go        # Shell configuration
│   └── starship.go     # Starship theme management
├── internal/            # Internal packages
│   ├── brewfile/       # Brewfile logic
│   ├── shell/          # Shell setup logic
│   └── starship/       # Starship integration
└── reference/           # Reference implementations (not distributed)
    ├── packages/       # ublue-os/packages repo
    └── homebrew-experimental-tap/
```

## 📚 Replicates & Extends

This project consolidates and modernizes functionality from:

- **ublue-bling**: Shell aliases and tool initialization scripts
- **bluefin-cli (cask)**: Homebrew package management and MOTD
- **ujust recipes**: Task runner and development environment helpers

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Make (optional)

### Building

```bash
go build -o bluefin-cli
```

### Running Locally

```bash
go run main.go [command]
```

### Dependencies

This project uses:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Huh](https://github.com/charmbracelet/huh) - Forms and prompts

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## 🙏 Acknowledgments

- [Universal Blue](https://universal-blue.org/) - For the original bluefin-cli and ublue-bling
- [Charm](https://charm.sh/) - For the amazing TUI libraries
- The Homebrew community

## 🔗 Related Projects

- [ublue-os/packages](https://github.com/ublue-os/packages) - Original package implementations
- [Starship](https://starship.rs/) - Cross-shell prompt
- [Homebrew](https://brew.sh/) - Package manager for macOS and Linux
