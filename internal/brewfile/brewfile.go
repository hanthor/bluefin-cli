package brewfile

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

const defaultBrewfile = `# Taps
tap "homebrew/bundle"
tap "homebrew/cask"
tap "homebrew/core"

# CLI Tools
brew "git"
brew "gh"
brew "wget"
brew "curl"
brew "jq"
brew "ripgrep"
brew "fd"
brew "bat"
brew "eza"
brew "fzf"
brew "zoxide"
brew "starship"

# Development
brew "neovim"
brew "tmux"

# Languages
brew "go"
brew "node"
brew "python"
`

// Initialize creates a new Brewfile
func Initialize(includeDefaults bool) error {
	brewfilePath := filepath.Join(".", "Brewfile")

	// Check if Brewfile already exists
	if _, err := os.Stat(brewfilePath); err == nil {
		return fmt.Errorf("Brewfile already exists in current directory")
	}

	content := ""
	if includeDefaults {
		content = defaultBrewfile
	} else {
		content = "# Brewfile - Add your packages here\n\n"
	}

	if err := os.WriteFile(brewfilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create Brewfile: %w", err)
	}

	fmt.Println(successStyle.Render("✓ Brewfile created successfully!"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("  Location: %s", brewfilePath)))

	return nil
}

// Apply runs brew bundle to install packages
func Apply() error {
	brewfilePath := filepath.Join(".", "Brewfile")

	// Check if Brewfile exists
	if _, err := os.Stat(brewfilePath); os.IsNotExist(err) {
		return fmt.Errorf("Brewfile not found in current directory")
	}

	fmt.Println(infoStyle.Render("📦 Installing packages from Brewfile..."))

	cmd := exec.Command("brew", "bundle", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew bundle failed: %w", err)
	}

	fmt.Println(successStyle.Render("✓ All packages installed successfully!"))
	return nil
}

// AddPackage adds a package to the Brewfile
func AddPackage(packageName string) error {
	brewfilePath := filepath.Join(".", "Brewfile")

	content, err := os.ReadFile(brewfilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Brewfile not found. Run 'bluefin-cli brewfile init' first")
		}
		return fmt.Errorf("failed to read Brewfile: %w", err)
	}

	newEntry := fmt.Sprintf("\nbrew \"%s\"\n", packageName)
	content = append(content, []byte(newEntry)...)

	if err := os.WriteFile(brewfilePath, content, 0644); err != nil {
		return fmt.Errorf("failed to update Brewfile: %w", err)
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Added '%s' to Brewfile", packageName)))
	return nil
}
