package starship

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

// Install downloads and installs Starship
func Install() error {
	// Check if already installed
	if _, err := exec.LookPath("starship"); err == nil {
		fmt.Println(successStyle.Render("✓ Starship is already installed"))
		return nil
	}

	fmt.Println(infoStyle.Render("⬇️  Installing Starship..."))

	// Use Homebrew if available
	if _, err := exec.LookPath("brew"); err == nil {
		cmd := exec.Command("brew", "install", "starship")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("brew install failed: %w", err)
		}

		fmt.Println(successStyle.Render("✓ Starship installed successfully!"))
		return nil
	}

	// Fallback to official installer
	cmd := exec.Command("sh", "-c", "curl -sS https://starship.rs/install.sh | sh -s -- -y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println(successStyle.Render("✓ Starship installed successfully!"))
	return nil
}

// ApplyTheme applies a Starship preset theme
func ApplyTheme(themeName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config")
	starshipConfig := filepath.Join(configDir, "starship.toml")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}


	// Download and apply the preset
	cmd := exec.Command("starship", "preset", themeName, "-o", starshipConfig)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply theme: %w", err)
	}

	return nil
}
