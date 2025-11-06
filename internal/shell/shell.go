package shell

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

// Setup configures the shell environment
func Setup(shellType string, installOhMyZsh bool, setupStarship bool) error {
	fmt.Println(infoStyle.Render(fmt.Sprintf("🐚 Setting up %s...", shellType)))

	// Set default shell
	if err := setDefaultShell(shellType); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("⚠️  Warning: %v", err)))
	}

	// Install Oh My Zsh if requested
	if installOhMyZsh && shellType == "zsh" {
		if err := installOhMyZshFramework(); err != nil {
			return fmt.Errorf("failed to install Oh My Zsh: %w", err)
		}
	}

	// Setup Starship if requested
	if setupStarship {
		if err := setupStarshipPrompt(shellType); err != nil {
			return fmt.Errorf("failed to setup Starship: %w", err)
		}
	}

	fmt.Println(successStyle.Render("✓ Shell setup complete!"))
	return nil
}

func setDefaultShell(shellType string) error {
	shellPath, err := exec.LookPath(shellType)
	if err != nil {
		return fmt.Errorf("%s not found: %w", shellType, err)
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("  Shell found at: %s", shellPath)))
	fmt.Println(infoStyle.Render("  To set as default, run: chsh -s " + shellPath))

	return nil
}

func installOhMyZshFramework() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	ohmyzshDir := filepath.Join(homeDir, ".oh-my-zsh")
	if _, err := os.Stat(ohmyzshDir); err == nil {
		fmt.Println(infoStyle.Render("  Oh My Zsh already installed"))
		return nil
	}

	fmt.Println(infoStyle.Render("  Installing Oh My Zsh..."))
	
	// Note: This is a simplified version. In production, you'd want to handle this more carefully
	cmd := exec.Command("sh", "-c", 
		`sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println(successStyle.Render("  ✓ Oh My Zsh installed"))
	return nil
}

func setupStarshipPrompt(shellType string) error {
	fmt.Println(infoStyle.Render("  Configuring Starship for " + shellType + "..."))
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var rcFile string
	var initCommand string

	switch shellType {
	case "bash":
		rcFile = filepath.Join(homeDir, ".bashrc")
		initCommand = `eval "$(starship init bash)"`
	case "zsh":
		rcFile = filepath.Join(homeDir, ".zshrc")
		initCommand = `eval "$(starship init zsh)"`
	case "fish":
		configDir := filepath.Join(homeDir, ".config", "fish")
		rcFile = filepath.Join(configDir, "config.fish")
		initCommand = "starship init fish | source"
		
		// Ensure config directory exists
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported shell: %s", shellType)
	}

	// Check if already configured
	content, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(content) > 0 && contains(string(content), "starship init") {
		fmt.Println(infoStyle.Render("  Starship already configured"))
		return nil
	}

	// Append Starship initialization
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("\n# Starship prompt\n%s\n", initCommand)); err != nil {
		return err
	}

	fmt.Println(successStyle.Render("  ✓ Starship configured"))
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
			len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
