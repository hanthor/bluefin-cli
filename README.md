# Bluefin CLI

A powerful, modern CLI tool for managing shell configuration and development environment customization. Built with beautiful TUIs using [Charm](https://charm.sh/) libraries.

## ✨ Features

- **🎨 Interactive Menu**: Default TUI experience for easy navigation
- **✨ Bling**: Toggle modern shell enhancements (eza, bat, ugrep, zoxide, atuin, starship)
- **📰 MOTD**: Beautiful Message of the Day with system info and random tips
- **📦 Bundle Installer**: Install curated tool bundles (ai, cli, fonts, k8s) from Universal Blue
- **�️ Wallpapers**: Install desktop wallpaper collections from ublue-os/tap
- **🎨 Starship Themes**: Browse and apply Starship prompt themes
- **⚙️ OS Scripts**: Run system-provided just recipes and scripts
- **📊 Status Command**: View configuration and installed tools at a glance

## 🚀 Installation

### NOT WORKING YET Via Homebrew

```bash
brew tap ublue-os/homebrew-experimental-tap
brew install bluefin-cli
```

### Build from Source

**Prerequisites:**
- Go 1.21 or later
- Homebrew (for package management features)

```bash
git clone https://github.com/hanthor/bluefin-cli.git
cd bluefin-cli
go build -o bluefin-cli
sudo mv bluefin-cli /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/hanthor/bluefin-cli@latest
```

## 📖 Usage

### Interactive Menu (Default)

Simply run the command to launch the interactive menu:

```bash
bluefin-cli
```

Or explicitly:

```bash
bluefin-cli menu
```

### Command Line Usage

#### Check Status

View your current configuration and installed tools:

```bash
bluefin-cli status
```

#### Bling - Modern Shell Enhancements

Enable/disable bling for your shell:

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

#### MOTD - Message of the Day

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

#### Install Tool Bundles

Install curated Homebrew bundles:

```bash
# List available bundles
bluefin-cli install list

# Install specific bundle
bluefin-cli install ai       # AI tools
bluefin-cli install cli      # CLI essentials
bluefin-cli install fonts    # Development fonts
bluefin-cli install k8s      # Kubernetes tools

# Interactive mode
bluefin-cli install
```

#### Install Wallpapers

Install desktop wallpaper collections:

```bash
# Interactive selection
bluefin-cli install wallpapers

```

#### Starship Themes
you can change your prompy lookks
Browse and apply Starship preset themes:

```bash
bluefin-cli starship theme
```

Install Starship if not already present:

```bash
bluefin-cli starship install
```

#### OS Scripts

Run OS-provided just recipes and shell scripts:

```bash
bluefin-cli osscripts
```

This discovers and lists all available recipes from `/usr/share/*/just/` directories.

## 🔧 What Gets Configured

### Bling Tools

The bling command configures these modern CLI tools:

- **eza**: Modern replacement for `ls` with icons and colors
- **bat**: `cat` clone with syntax highlighting
- **zoxide**: Smarter `cd` command that learns your habits
- **atuin**: Magical shell history with sync and search (optional)
- **starship**: Fast, customizable prompt for any shell
- **ugrep**: Ultra-fast grep alternative (optional)

### Shell Aliases

When bling is enabled in your shell:

```bash
ll      # eza -l --icons=auto --group-directories-first
ls      # eza
cat     # bat --style=plain --pager=never
grep    # ugrep (if installed)
```

## 🏗️ Project Structure

```
bluefin-cli/
├── main.go              # Application entry point
├── cmd/                 # Cobra commands
│   ├── root.go         # Root command & menu default
│   ├── menu.go         # Interactive TUI menu
│   ├── bling.go        # Bling command
│   ├── motd.go         # MOTD command
│   ├── install.go      # Install bundles/wallpapers
│   ├── osscripts.go    # OS scripts discovery
│   ├── starship.go     # Starship theme management
│   └── status.go       # Status display
├── internal/            # Internal packages
│   ├── bling/          # Bling logic & embedded scripts
│   ├── motd/           # MOTD generation
│   ├── install/        # Bundle & wallpaper installation
│   ├── starship/       # Starship integration
│   └── status/         # Status checking
└── test/                # Integration tests
```

## 📚 Inspiration

This project consolidates and modernizes functionality from:

- **ublue-bling**: Shell aliases and tool initialization scripts
- **bluefin-cli (cask)**: Homebrew package management and MOTD
- **ujust recipes**: Task runner and development environment helpers

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Podman (for containerized testing)
- just (for running recipes)

### Building

```bash
just build
```

### Testing

```bash
# Run tests in container
just test

# Run tests locally
go test ./...
```

### Interactive Development

Launch shells with bling pre-configured:

```bash
just bash   # Test in bash
just zsh    # Test in zsh
just fish   # Test in fish
```

### Dependencies

This project uses:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Huh](https://github.com/charmbracelet/huh) - Forms and prompts
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

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
