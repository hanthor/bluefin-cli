package shell

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// InstallTools iterates through the config and installs enabled tools
func InstallTools(cfg *Config) {
	// First check if we need to install anything
	needsInstall := false
	for _, tool := range Tools {
		if cfg.IsEnabled(tool.Name) {
			if _, err := exec.LookPath(tool.Binary); err != nil {
				needsInstall = true
				break
			}
		}
	}

	if cfg.IsEnabled("Motd") {
		if _, err := exec.LookPath("glow"); err != nil {
			needsInstall = true
		}
	}

	if !needsInstall {
		return
	}

	// Ensure Homebrew is available
	if err := ensureHomebrew(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Skipping tool installation: %v", err)))
		return
	}

	for _, tool := range Tools {
		if cfg.IsEnabled(tool.Name) {
			if err := ensureTool(tool.Binary, tool.Pkg); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to install %s: %v", tool.Pkg, err)))
			}
		}
	}

	if cfg.IsEnabled("Motd") {
		if err := ensureTool("glow", "glow"); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to install glow: %v", err)))
		}
	}
}

func ensureHomebrew() error {
	if _, err := exec.LookPath("brew"); err == nil {
		return nil
	}

	commonPaths := []string{"/home/linuxbrew/.linuxbrew/bin/brew", "/opt/homebrew/bin/brew", "/usr/local/bin/brew"}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			path := os.Getenv("PATH")
			os.Setenv("PATH", path+string(os.PathListSeparator)+filepath.Dir(p))
			return nil
		}
	}

	fmt.Println(infoStyle.Render("Homebrew is missing. It is required to install enabled components."))
	var install bool
	err := huh.NewConfirm().
		Title("Would you like to install Homebrew?").
		Value(&install).
		Run()
	if err != nil {
		return err
	}

	if !install {
		return fmt.Errorf("homebrew installation declined")
	}

	fmt.Println(infoStyle.Render("⬇️  Installing Homebrew..."))
	
	cmd := exec.Command("/bin/bash", "-c", "curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install homebrew: %w", err)
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			path := os.Getenv("PATH")
			os.Setenv("PATH", path+string(os.PathListSeparator)+filepath.Dir(p))
			fmt.Println(successStyle.Render("✓ Homebrew installed and added to PATH for this session."))
			return nil
		}
	}
	
	return fmt.Errorf("homebrew installed but not found in expected locations")
}

func ensureTool(binary, pkg string) error {
	if _, err := exec.LookPath(binary); err == nil {
		return nil
	}

	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("brew not found")
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

	if cfg, err := LoadConfig(shell); err == nil {
		InstallTools(cfg)
	}

	return nil
}

func Init(shell string, config *Config) (string, error) {
	if config == nil {
		config = DefaultConfig(shell)
	}

	var sb strings.Builder

	for _, tool := range Tools {
		enabled := config.IsEnabled(tool.Name)

		if shell == "fish" {
			fmt.Fprintf(&sb, "set -gx %s %d\n", tool.GetEnvVar(), boolToInt(enabled))
		} else {
			fmt.Fprintf(&sb, "export %s=%d\n", tool.GetEnvVar(), boolToInt(enabled))
		}
	}

	sb.WriteString("\n")

	if shell == "fish" {
		sb.WriteString(shellFishScript)
	} else {
		sb.WriteString(shellShScript)
	}

	return sb.String(), nil
}

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

		status[shell] = strings.Contains(string(content), shellMaker) || strings.Contains(string(content), "# bluefin-cli bling")
	}

	return status
}

func CheckDependencies() map[string]bool {
	status := make(map[string]bool)

	for _, tool := range Tools {
		_, err := exec.LookPath(tool.Binary)
		status[tool.Binary] = err == nil
	}

	return status
}

// GetInstalledShells returns a list of shells that are available in the PATH
func GetInstalledShells() []string {
	var installed []string
	shells := []string{"bash", "zsh", "fish"}

	for _, s := range shells {
		if _, err := exec.LookPath(s); err == nil {
			installed = append(installed, s)
		}
	}

	return installed
}
