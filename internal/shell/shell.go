package shell

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

//go:embed resources/shell.sh
var shellShScript string

//go:embed resources/shell.fish
var shellFishScript string

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

const shellMaker = "# bluefin-cli shell-config"
const blingMarker = "# bluefin-cli bling"


// Toggle enables or disables bling for the specified shell
func Toggle(shell string, enable bool) error {
	var configFile string
	var rcLine string

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	switch shell {
	case "bash":
		configFile = filepath.Join(home, ".bashrc")
		rcLine = fmt.Sprintf(`eval "$(bluefin-cli init bash)" %s`, shellMaker)
	case "zsh":
		configFile = filepath.Join(home, ".zshrc")
		rcLine = fmt.Sprintf(`eval "$(bluefin-cli init zsh)" %s`, shellMaker)
	case "fish":
		configFile = filepath.Join(home, ".config/fish/config.fish")
		rcLine = fmt.Sprintf(`bluefin-cli init fish | source %s`, shellMaker)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) && enable {
			// Create if doesn't exist and we are enabling
			// For fish, ensure dir exists
			if shell == "fish" {
				if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
					return err
				}
			}
			content = []byte("")
		} else {
			return err
		}
	}

	text := string(content)
	hasLine := strings.Contains(text, shellMaker)

	if enable {
		if hasLine {
			fmt.Println(infoStyle.Render(fmt.Sprintf("%s is already enabled for %s", shell, shell)))
			return nil
		}
		
		f, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		prefix := "\n"
		if len(text) == 0 || strings.HasSuffix(text, "\n") {
			prefix = ""
		}

		if _, err := f.WriteString(prefix + rcLine + "\n"); err != nil {
			return err
		}
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Enabled shell experience for %s", shell)))
	} else {
		if !hasLine {
			fmt.Println(infoStyle.Render(fmt.Sprintf("%s is already disabled for %s", shell, shell)))
			return nil
		}

		// Remove the lines containing the marker
		lines := strings.Split(text, "\n")
		var newLines []string
		for _, line := range lines {
			if !strings.Contains(line, shellMaker) && !strings.Contains(line, blingMarker) {
				newLines = append(newLines, line)
			}
		}
		
		output := strings.Join(newLines, "\n")
		// Trim extra newlines at the end
		output = strings.TrimRight(output, "\n") + "\n"

		if err := os.WriteFile(configFile, []byte(output), 0644); err != nil {
			return err
		}
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Disabled shell experience for %s", shell)))
	}

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
		fmt.Fprintf(&sb, "set -gx BLUEFIN_SHELL_ENABLE_EZA %d\n", boolToInt(config.Eza))
		fmt.Fprintf(&sb, "set -gx BLUEFIN_SHELL_ENABLE_UGREP %d\n", boolToInt(config.Ugrep))
		fmt.Fprintf(&sb, "set -gx BLUEFIN_SHELL_ENABLE_BAT %d\n", boolToInt(config.Bat))
		fmt.Fprintf(&sb, "set -gx BLUEFIN_SHELL_ENABLE_ATUIN %d\n", boolToInt(config.Atuin))
		fmt.Fprintf(&sb, "set -gx BLUEFIN_SHELL_ENABLE_STARSHIP %d\n", boolToInt(config.Starship))
		fmt.Fprintf(&sb, "set -gx BLUEFIN_SHELL_ENABLE_ZOXIDE %d\n", boolToInt(config.Zoxide))
		fmt.Fprintf(&sb, "set -gx BLUEFIN_SHELL_ENABLE_UUTILS %d\n", boolToInt(config.Uutils))
	} else {
		fmt.Fprintf(&sb, "export BLUEFIN_SHELL_ENABLE_EZA=%d\n", boolToInt(config.Eza))
		fmt.Fprintf(&sb, "export BLUEFIN_SHELL_ENABLE_UGREP=%d\n", boolToInt(config.Ugrep))
		fmt.Fprintf(&sb, "export BLUEFIN_SHELL_ENABLE_BAT=%d\n", boolToInt(config.Bat))
		fmt.Fprintf(&sb, "export BLUEFIN_SHELL_ENABLE_ATUIN=%d\n", boolToInt(config.Atuin))
		fmt.Fprintf(&sb, "export BLUEFIN_SHELL_ENABLE_STARSHIP=%d\n", boolToInt(config.Starship))
		fmt.Fprintf(&sb, "export BLUEFIN_SHELL_ENABLE_ZOXIDE=%d\n", boolToInt(config.Zoxide))
		fmt.Fprintf(&sb, "export BLUEFIN_SHELL_ENABLE_UUTILS=%d\n", boolToInt(config.Uutils))
	}
	
	sb.WriteString("\n")

	// 2. Append Shell Script
	if shell == "fish" {
		sb.WriteString(shellFishScript)
	} else {
		sb.WriteString(shellShScript)
	}

	return sb.String(), nil
}

// CheckStatus returns whether bling is enabled for each shell
func CheckStatus() map[string]bool {
	status := make(map[string]bool)
	shells := []string{"bash", "zsh", "fish"}
	home, _ := os.UserHomeDir()

	for _, shell := range shells {
		var configFile string
		switch shell {
		case "bash":
			configFile = filepath.Join(home, ".bashrc")
		case "zsh":
			configFile = filepath.Join(home, ".zshrc")
		case "fish":
			configFile = filepath.Join(home, ".config/fish/config.fish")
		}

		content, err := os.ReadFile(configFile)
		if err != nil {
			status[shell] = false
			continue
		}

		// Check for new marker OR old marker (for compatibility/transitions)
		// We'll consider it enabled if either is present, but Toggle will write the new one.
		status[shell] = strings.Contains(string(content), shellMaker) || strings.Contains(string(content), "# bluefin-cli bling")
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
