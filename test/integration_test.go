package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	binaryPath = "../bluefin-cli"
)

// Shell configuration for parameterized tests
type ShellConfig struct {
	Name         string
	ConfigFile   string
	BlingPattern string
	BlingScript  string
	InitShell    func() error
}

var shells = []ShellConfig{
	{
		Name:         "bash",
		ConfigFile:   ".bashrc",
		BlingPattern: "bling.sh",
		BlingScript:  "bling.sh",
		InitShell:    func() error { return touchFile(filepath.Join(os.Getenv("HOME"), ".bashrc")) },
	},
	{
		Name:         "zsh",
		ConfigFile:   ".zshrc",
		BlingPattern: "bling.sh",
		BlingScript:  "bling.sh",
		InitShell:    func() error { return touchFile(filepath.Join(os.Getenv("HOME"), ".zshrc")) },
	},
	{
		Name:         "fish",
		ConfigFile:   ".config/fish/config.fish",
		BlingPattern: "bling.fish",
		BlingScript:  "bling.fish",
		InitShell: func() error {
			dir := filepath.Join(os.Getenv("HOME"), ".config/fish")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return touchFile(filepath.Join(dir, "config.fish"))
		},
	},
}

// Tool configuration for testing bling script content
type ToolConfig struct {
	Name    string
	Pattern string
	Script  string // "bling.sh" or "bling.fish"
}

var tools = []ToolConfig{
	{Name: "eza", Pattern: "alias ls='eza'", Script: "bling.sh"},
	{Name: "bat", Pattern: "alias cat='bat", Script: "bling.sh"},
	{Name: "starship-bash", Pattern: "starship init bash", Script: "bling.sh"},
	{Name: "starship-zsh", Pattern: "starship init zsh", Script: "bling.sh"},
	{Name: "starship-fish", Pattern: "starship init fish", Script: "bling.fish"},
	{Name: "zoxide", Pattern: "zoxide init", Script: "bling.sh"},
	{Name: "atuin", Pattern: "atuin init", Script: "bling.sh"},
	{Name: "command-check", Pattern: "command -v eza", Script: "bling.sh"},
}

func TestMain(m *testing.M) {
	// Setup
	if os.Getenv("HOME") == "" {
		os.Setenv("HOME", "/root")
	}
	
	// Initialize shell configs
	for _, shell := range shells {
		if err := shell.InitShell(); err != nil {
			panic(err)
		}
	}
	
	// Run tests
	code := m.Run()
	os.Exit(code)
}

func touchFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

func runCommand(t *testing.T, args ...string) (string, error) {
	cmd := exec.Command(binaryPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func fileContains(t *testing.T, filepath, pattern string) bool {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), pattern)
}

func TestBinaryExecutes(t *testing.T) {
	_, err := runCommand(t, "--version")
	if err != nil {
		t.Fatalf("Binary failed to execute: %v", err)
	}
}

func TestStatusCommand(t *testing.T) {
	_, err := runCommand(t, "status")
	if err != nil {
		t.Fatalf("Status command failed: %v", err)
	}
}

func TestBlingEnableForAllShells(t *testing.T) {
	for _, shell := range shells {
		t.Run(shell.Name, func(t *testing.T) {
			// Enable bling
			_, err := runCommand(t, "bling", shell.Name, "on")
			if err != nil {
				t.Fatalf("Failed to enable bling for %s: %v", shell.Name, err)
			}
			
			// Verify config file contains bling pattern
			configPath := filepath.Join(os.Getenv("HOME"), shell.ConfigFile)
			if !fileContains(t, configPath, shell.BlingPattern) {
				t.Errorf("Config file %s doesn't contain bling pattern %s", configPath, shell.BlingPattern)
			}
		})
	}
}

func TestBlingScriptSourcing(t *testing.T) {
	for _, shell := range shells {
		t.Run(shell.Name, func(t *testing.T) {
			configPath := filepath.Join(os.Getenv("HOME"), shell.ConfigFile)
			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read config: %v", err)
			}
			
			if !strings.Contains(string(content), shell.BlingScript) {
				t.Errorf("Shell config doesn't source bling script")
			}
		})
	}
}

func TestShellSyntax(t *testing.T) {
	tests := []struct {
		shell      string
		configFile string
		validator  string
	}{
		{"bash", ".bashrc", "bash"},
		{"zsh", ".zshrc", "zsh"},
		{"fish", ".config/fish/config.fish", "fish"},
	}
	
	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			configPath := filepath.Join(os.Getenv("HOME"), tt.configFile)
			cmd := exec.Command(tt.validator, "-n", configPath)
			if err := cmd.Run(); err != nil {
				t.Errorf("%s config has syntax errors: %v", tt.shell, err)
			}
		})
	}
}

func TestBlingToolConfigurations(t *testing.T) {
	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			scriptPath := filepath.Join(os.Getenv("HOME"), ".local/share/bluefin-cli/bling", tool.Script)
			if !fileContains(t, scriptPath, tool.Pattern) {
				t.Errorf("Bling script doesn't contain configuration for %s", tool.Name)
			}
		})
	}
}

func TestMOTDSystem(t *testing.T) {
	t.Run("EnableMOTD", func(t *testing.T) {
		_, err := runCommand(t, "motd", "toggle", "bash", "on")
		if err != nil {
			t.Fatalf("Failed to enable MOTD: %v", err)
		}
	})
	
	t.Run("MOTDInBashrc", func(t *testing.T) {
		bashrc := filepath.Join(os.Getenv("HOME"), ".bashrc")
		if !fileContains(t, bashrc, "bluefin-motd.sh") {
			t.Error("MOTD not configured in bashrc")
		}
	})
	
	t.Run("MOTDResourcesInstalled", func(t *testing.T) {
		motdDir := filepath.Join(os.Getenv("HOME"), ".local/share/bluefin-cli/motd")
		tipsDir := filepath.Join(motdDir, "tips")
		
		files, err := filepath.Glob(filepath.Join(tipsDir, "*.md"))
		if err != nil || len(files) < 10 {
			t.Errorf("Not enough MOTD tips found: got %d, want >= 10", len(files))
		}
	})
	
	t.Run("MOTDShowCommand", func(t *testing.T) {
		output, _ := runCommand(t, "motd", "show")
		if !strings.Contains(output, "Bluefin") {
			t.Error("MOTD show command didn't display expected content")
		}
	})
}

func TestStatusReflectsChanges(t *testing.T) {
	output, err := runCommand(t, "status")
	if err != nil {
		t.Fatalf("Status command failed: %v", err)
	}
	
	if !strings.Contains(output, "bash: enabled") {
		t.Error("Status doesn't show bash bling as enabled")
	}
	
	if !strings.Contains(output, "Message of the Day") {
		t.Error("Status doesn't show MOTD section")
	}
}

func TestBlingDisable(t *testing.T) {
	_, err := runCommand(t, "bling", "bash", "off")
	if err != nil {
		t.Fatalf("Failed to disable bling: %v", err)
	}
	
	bashrc := filepath.Join(os.Getenv("HOME"), ".bashrc")
	if fileContains(t, bashrc, "bling.sh") {
		t.Error("Bling still present in bashrc after disable")
	}
}

func TestInstallList(t *testing.T) {
	output, err := runCommand(t, "install", "list")
	if err != nil {
		t.Fatalf("Install list command failed: %v", err)
	}
	
	if !strings.Contains(output, "Available Homebrew Bundles") {
		t.Error("Install list doesn't show expected content")
	}
}
