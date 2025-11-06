package motd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

const motdMarker = "# bluefin-cli motd"

var defaultTips = []string{
	"Use `brew search` and `brew install` to install packages. Homebrew will take care of updates automatically",
	"`tldr vim` will give you the basic rundown on commands for a given tool",
	"Performance profiling tools are built-in: try `top`, `htop`, and other debugging tools",
	"Switch shells safely: change your shell in Terminal settings instead of system-wide",
	"Container development is OS-agnostic - your devcontainers work on Linux, macOS, and Windows",
	"Use `docker compose` for multi-container development if devcontainers don't fit your workflow",
	"Bluefin separates the OS from your development environment - embrace the cloud-native workflow",
	"Check out DevPod for open-source, client-only development environments that work with any IDE",
	"Develop with devcontainers! Use `devcontainer.json` files in your projects for isolated, reproducible environments",
	"VS Code comes with devcontainers extension pre-installed - perfect for containerized development",
	"Use `eza -l --icons` for a beautiful file listing with icons and colors",
	"The `bat` command is like `cat` but with syntax highlighting and Git integration",
	"Navigate directories faster with `zoxide` - just use `z <partial-name>` to jump around",
	"Search your shell history with `atuin` using Ctrl+R for a better history search experience",
	"Customize your prompt with `starship config` to modify colors, icons, and modules",
}

var defaultTemplate = `# 󱍢 Welcome to Bluefin CLI
󱋩 %s:%s

|  Command | Description |
| ------- | ----------- |
| ` + "`bluefin-cli bling bash on`" + `  | Enable terminal bling for bash  |
| ` + "`bluefin-cli status`" + ` | Show current configuration |
| ` + "`bluefin-cli help`" + ` | Show all available commands |
| ` + "`brew help`" + ` | Manage command line packages |

%s

- **󰊤** [GitHub Issues](https://github.com/hanthor/bluefin-cli/issues)
- **󰈙** [Documentation](https://github.com/hanthor/bluefin-cli)
`

type ImageInfo struct {
	ImageName     string `json:"image-name"`
	ImageTag      string `json:"image-tag"`
	ImageFlavor   string `json:"image-flavor"`
	ImageVendor   string `json:"image-vendor"`
	FedoraVersion string `json:"fedora-version"`
}

type Config struct {
	TipsDirectory   string `json:"tips-directory"`
	CheckOutdated   string `json:"check-outdated"`
	ImageInfoFile   string `json:"image-info-file"`
	DefaultTheme    string `json:"default-theme"`
	TemplateFile    string `json:"template-file"`
	ThemesDirectory string `json:"themes-directory"`
}

// Toggle enables or disables MOTD for shells
func Toggle(target string, enable bool) error {
	shells := []string{"bash", "zsh", "fish"}
	if target != "all" {
		shells = []string{target}
	}

	for _, shell := range shells {
		if err := toggleForShell(shell, enable); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Error toggling MOTD for %s: %v", shell, err)))
		}
	}

	return nil
}

func toggleForShell(shell string, enable bool) error {
	var configFile string
	var motdLine string
	var linesAfterMarker int

	homeDir := os.Getenv("HOME")
	motdPath := filepath.Join(homeDir, ".local/share/bluefin-cli/motd")

	// Ensure MOTD is set up
	if err := setupMOTD(); err != nil {
		return fmt.Errorf("failed to setup MOTD: %w", err)
	}

	switch shell {
	case "bash":
		configFile = filepath.Join(homeDir, ".bashrc")
		motdLine = fmt.Sprintf("[ -x %s/bluefin-motd.sh ] && %s/bluefin-motd.sh", motdPath, motdPath)
		linesAfterMarker = 1
	case "zsh":
		configFile = filepath.Join(homeDir, ".zshrc")
		motdLine = fmt.Sprintf("[ -x %s/bluefin-motd.sh ] && %s/bluefin-motd.sh", motdPath, motdPath)
		linesAfterMarker = 1
	case "fish":
		configFile = filepath.Join(homeDir, ".config/fish/config.fish")
		motdLine = fmt.Sprintf("if status is-interactive; and test -x %s/bluefin-motd.sh; %s/bluefin-motd.sh; end", motdPath, motdPath)
		linesAfterMarker = 1
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	// Ensure config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		dir := filepath.Dir(configFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
			return err
		}
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	hasMarker := strings.Contains(string(content), motdMarker)

	if enable {
		if hasMarker {
			fmt.Println(infoStyle.Render(fmt.Sprintf("ℹ MOTD already enabled for %s", shell)))
			return nil
		}

		newContent := string(content) + "\n" + motdMarker + "\n" + motdLine + "\n"
		if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
			return err
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("✓ MOTD enabled for %s", shell)))
	} else {
		if !hasMarker {
			fmt.Println(infoStyle.Render(fmt.Sprintf("ℹ MOTD already disabled for %s", shell)))
			return nil
		}

		lines := strings.Split(string(content), "\n")
		var newLines []string
		skipLines := 0

		for _, line := range lines {
			if skipLines > 0 {
				skipLines--
				continue
			}
			if strings.Contains(line, motdMarker) {
				skipLines = linesAfterMarker
				continue
			}
			newLines = append(newLines, line)
		}

		newContent := strings.Join(newLines, "\n")
		if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
			return err
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("✓ MOTD disabled for %s", shell)))
	}

	return nil
}

// Show displays the MOTD
func Show() error {
	homeDir := os.Getenv("HOME")
	motdPath := filepath.Join(homeDir, ".local/share/bluefin-cli/motd")

	// Ensure MOTD is set up
	if err := setupMOTD(); err != nil {
		return err
	}

	// Get OS info
	info := getImageInfo()

	// Get random tip
	tip := getRandomTip(filepath.Join(motdPath, "tips"))

	// Format template
	content := fmt.Sprintf(defaultTemplate, info.ImageName, info.ImageTag, tip)

	// Render with glow if available
	if glowPath, err := exec.LookPath("glow"); err == nil {
		cmd := exec.Command(glowPath, "-s", "dark", "-w", "80", "-")
		cmd.Stdin = strings.NewReader(content)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to plain text
	fmt.Println(content)
	return nil
}

// SetTheme sets the MOTD theme
func SetTheme(theme string) error {
	homeDir := os.Getenv("HOME")
	configPath := filepath.Join(homeDir, ".local/share/bluefin-cli/motd/motd.json")

	var config Config
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &config)
	}

	config.DefaultTheme = theme

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✓ MOTD theme set to: %s", theme)))
	return nil
}

// setupMOTD initializes the MOTD system
func setupMOTD() error {
	homeDir := os.Getenv("HOME")
	motdPath := filepath.Join(homeDir, ".local/share/bluefin-cli/motd")
	tipsPath := filepath.Join(motdPath, "tips")

	// Create directories
	if err := os.MkdirAll(tipsPath, 0755); err != nil {
		return err
	}

	// Install tips
	for i, tip := range defaultTips {
		tipFile := filepath.Join(tipsPath, fmt.Sprintf("%02d-tip.md", i+1))
		if _, err := os.Stat(tipFile); os.IsNotExist(err) {
			if err := os.WriteFile(tipFile, []byte(tip), 0644); err != nil {
				return err
			}
		}
	}

	// Create MOTD script
	scriptPath := filepath.Join(motdPath, "bluefin-motd.sh")
	scriptContent := `#!/usr/bin/env bash
bluefin-cli motd show
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return err
	}

	// Create config
	configPath := filepath.Join(motdPath, "motd.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := Config{
			TipsDirectory:   tipsPath,
			CheckOutdated:   "false",
			ImageInfoFile:   filepath.Join(motdPath, "image-info.json"),
			DefaultTheme:    "slate",
			TemplateFile:    filepath.Join(motdPath, "template.md"),
			ThemesDirectory: filepath.Join(motdPath, "themes"),
		}

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func getImageInfo() ImageInfo {
	// Detect OS information
	info := ImageInfo{
		ImageFlavor:   "homebrew",
		ImageVendor:   "bluefin-cli",
		FedoraVersion: "N/A",
	}

	if runtime.GOOS == "darwin" {
		info.ImageName = "macOS"
		if output, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			info.ImageTag = strings.TrimSpace(string(output))
		}
	} else if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "NAME=") {
					info.ImageName = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
				} else if strings.HasPrefix(line, "VERSION_ID=") {
					info.ImageTag = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
				}
			}
		}
	}

	if info.ImageName == "" {
		info.ImageName = runtime.GOOS
	}
	if info.ImageTag == "" {
		info.ImageTag = "unknown"
	}

	return info
}

func getRandomTip(tipsDir string) string {
	files, err := filepath.Glob(filepath.Join(tipsDir, "*.md"))
	if err != nil || len(files) == 0 {
		return ""
	}

	rand.Seed(time.Now().UnixNano())
	tipFile := files[rand.Intn(len(files))]

	content, err := os.ReadFile(tipFile)
	if err != nil {
		return ""
	}

	return "💡 **Tip:** " + string(content)
}

// CheckStatus returns whether MOTD is enabled for each shell
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

		status[shell] = strings.Contains(string(content), motdMarker)
	}

	return status
}
