package bling

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// InstallTools iterates through the config and installs enabled tools
func InstallTools(cfg *Config) {
	tools := []struct {
		enabled bool
		binary  string
		pkg     string
	}{
		{cfg.Eza, "eza", "eza"},
		{cfg.Ugrep, "ug", "ugrep"},
		{cfg.Bat, "bat", "bat"},
		{cfg.Atuin, "atuin", "atuin"},
		{cfg.Starship, "starship", "starship"},
		{cfg.Zoxide, "zoxide", "zoxide"},
		{cfg.Uutils, "uutils", "uutils-coreutils"},
	}

	for _, t := range tools {
		if t.enabled {
			if err := ensureTool(t.binary, t.pkg); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to install %s: %v", t.pkg, err)))
			}
		}
	}
}

// ensureTool checks if a tool is available and installs it via brew if missing
func ensureTool(binary, pkg string) error {
	// Check if tool is already installed
	if _, err := exec.LookPath(binary); err == nil {
		return nil
	}

	// Not found, check if brew is available
	if _, err := exec.LookPath("brew"); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Homebrew not found. Cannot auto-install %s.", pkg)))
		return nil // Don't fail config generation
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("⬇️  Installing %s via Homebrew...", pkg)))
	cmd := exec.Command("brew", "install", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", pkg, err)
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s installed successfully!", pkg)))
	return nil
}

//go:embed resources/bling.sh
var blingShScript string

//go:embed resources/bling.fish
var blingFishScript string

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

const blingMarker = "# bluefin-cli bling"

// Toggle enables or disables bling for the specified shell
// Deprecated: Use 'bluefin-cli init' instead
// Deprecated: Use 'bluefin-cli init' instead
func Toggle(shell string, enable bool) error {
	fmt.Println(infoStyle.Render("ℹ Note: shell integration is now handled via 'bluefin-cli init'"))
	return nil
}

// Init returns the shell initialization script
func Init(shell string) (string, error) {
	config, err := LoadConfig()
	if err != nil {
		// Fallback to default if config can't be loaded (e.g. permission error, though unlikely)
		config = DefaultConfig()
	}

	var sb strings.Builder

	// 1. Generate Environment Variables
	if shell == "fish" {
		fmt.Fprintf(&sb, "set -gx BLING_ENABLE_EZA %d\n", boolToInt(config.Eza))
		fmt.Fprintf(&sb, "set -gx BLING_ENABLE_UGREP %d\n", boolToInt(config.Ugrep))
		fmt.Fprintf(&sb, "set -gx BLING_ENABLE_BAT %d\n", boolToInt(config.Bat))
		fmt.Fprintf(&sb, "set -gx BLING_ENABLE_ATUIN %d\n", boolToInt(config.Atuin))
		fmt.Fprintf(&sb, "set -gx BLING_ENABLE_STARSHIP %d\n", boolToInt(config.Starship))
		fmt.Fprintf(&sb, "set -gx BLING_ENABLE_ZOXIDE %d\n", boolToInt(config.Zoxide))
		fmt.Fprintf(&sb, "set -gx BLING_ENABLE_UUTILS %d\n", boolToInt(config.Uutils))
	} else {
		fmt.Fprintf(&sb, "export BLING_ENABLE_EZA=%d\n", boolToInt(config.Eza))
		fmt.Fprintf(&sb, "export BLING_ENABLE_UGREP=%d\n", boolToInt(config.Ugrep))
		fmt.Fprintf(&sb, "export BLING_ENABLE_BAT=%d\n", boolToInt(config.Bat))
		fmt.Fprintf(&sb, "export BLING_ENABLE_ATUIN=%d\n", boolToInt(config.Atuin))
		fmt.Fprintf(&sb, "export BLING_ENABLE_STARSHIP=%d\n", boolToInt(config.Starship))
		fmt.Fprintf(&sb, "export BLING_ENABLE_ZOXIDE=%d\n", boolToInt(config.Zoxide))
		fmt.Fprintf(&sb, "export BLING_ENABLE_UUTILS=%d\n", boolToInt(config.Uutils))
	}
	
	sb.WriteString("\n")

	// 2. Append Bling Script
	if shell == "fish" {
		sb.WriteString(blingFishScript)
	} else {
		sb.WriteString(blingShScript)
	}

	return sb.String(), nil
}

// CheckStatus returns whether bling is enabled for each shell
func CheckStatus() map[string]bool {
	status := make(map[string]bool)
	shells := []string{"bash", "zsh", "fish"}

	for _, shell := range shells {
		var configFile string
		switch shell {
		case "bash":
			configFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
		case "zsh":
			configFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
		case "fish":
			configFile = filepath.Join(os.Getenv("HOME"), ".config/fish/config.fish")
		}

		content, err := os.ReadFile(configFile)
		if err != nil {
			status[shell] = false
			continue
		}

		status[shell] = strings.Contains(string(content), blingMarker)
	}

	return status
}

// CheckDependencies verifies required tools are installed
func CheckDependencies() map[string]bool {
	tools := []string{"eza", "bat", "zoxide", "atuin", "starship", "ugrep"}
	status := make(map[string]bool)

	for _, tool := range tools {
		_, err := exec.LookPath(tool)
		status[tool] = err == nil
	}

	return status
}
