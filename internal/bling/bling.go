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

// installTools iterates through the config and installs enabled tools
func installTools(cfg *Config) {
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
func Toggle(shell string, enable bool) error {
	shell = strings.ToLower(shell)

	var configFile string
	var sourceLine string

	switch shell {
	case "bash":
		configFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
		blingPath, err := ensureBlingScript("bash")
		if err != nil {
			return err
		}
		// Ensure config/env files exist
		if _, err := LoadConfig(); err != nil {
			return fmt.Errorf("failed to load/init config: %w", err)
		}
		// We re-save to ensure env files are generated matching the config
		cfg, _ := LoadConfig()
		if err := GenerateEnvFiles(cfg); err != nil {
			return fmt.Errorf("failed to generate env files: %w", err)
		}

		if enable {
			installTools(cfg)
		}

		sourceLine = fmt.Sprintf(`if [ -n "${BASH_VERSION:-}" ]; then . %s; fi`, blingPath)
	case "zsh":
		configFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
		blingPath, err := ensureBlingScript("zsh")
		if err != nil {
			return err
		}
		// Ensure config/env files exist
		if _, err := LoadConfig(); err != nil {
			return fmt.Errorf("failed to load/init config: %w", err)
		}
		cfg, _ := LoadConfig()
		if err := GenerateEnvFiles(cfg); err != nil {
			return fmt.Errorf("failed to generate env files: %w", err)
		}

		if enable {
			installTools(cfg)
		}

		sourceLine = fmt.Sprintf(`test -f %s && source %s`, blingPath, blingPath)
	case "fish":
		configFile = filepath.Join(os.Getenv("HOME"), ".config/fish/config.fish")
		blingPath, err := ensureBlingScript("fish")
		if err != nil {
			return err
		}
		// Ensure config/env files exist
		if _, err := LoadConfig(); err != nil {
			return fmt.Errorf("failed to load/init config: %w", err)
		}
		cfg, _ := LoadConfig()
		if err := GenerateEnvFiles(cfg); err != nil {
			return fmt.Errorf("failed to generate env files: %w", err)
		}

		if enable {
			installTools(cfg)
		}

		sourceLine = fmt.Sprintf("source %s", blingPath)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}

	// Ensure config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		dir := filepath.Dir(configFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	hasMarker := strings.Contains(string(content), blingMarker)

	if enable {
		if hasMarker {
			fmt.Println(infoStyle.Render(fmt.Sprintf("ℹ Bling already enabled for %s", shell)))
			return nil
		}

		// Add bling
		newContent := string(content) + "\n" + blingMarker + "\n" + sourceLine + "\n"
		if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to update config file: %w", err)
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Bling enabled for %s", shell)))
		fmt.Println(infoStyle.Render(fmt.Sprintf("  Restart your %s session or run: source %s", shell, configFile)))
	} else {
		if !hasMarker {
			fmt.Println(infoStyle.Render(fmt.Sprintf("ℹ Bling already disabled for %s", shell)))
			return nil
		}

		// Remove bling
		lines := strings.Split(string(content), "\n")
		var newLines []string
		skipNext := false

		for _, line := range lines {
			if skipNext {
				skipNext = false
				continue
			}
			if strings.Contains(line, blingMarker) {
				skipNext = true // Skip the next line (the source command)
				continue
			}
			newLines = append(newLines, line)
		}

		newContent := strings.Join(newLines, "\n")
		if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to update config file: %w", err)
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Bling disabled for %s", shell)))
	}

	return nil
}

// ensureBlingScript ensures the bling script is installed and returns its path
func ensureBlingScript(shell string) (string, error) {
	homeDir := os.Getenv("HOME")
	blingDir := filepath.Join(homeDir, ".local/share/bluefin-cli/bling")

	if err := os.MkdirAll(blingDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bling directory: %w", err)
	}

	var scriptPath string
	var scriptContent string

	if shell == "fish" {
		scriptPath = filepath.Join(blingDir, "bling.fish")
		scriptContent = blingFishScript
	} else {
		scriptPath = filepath.Join(blingDir, "bling.sh")
		scriptContent = blingShScript
	}

	// Write the script
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return "", fmt.Errorf("failed to write bling script: %w", err)
	}

	return scriptPath, nil
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
