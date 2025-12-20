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
	ShellPattern string
	ShellScript  string
	InitShell    func() error
}

var shells = []ShellConfig{
	{
		Name:         "bash",
		ConfigFile:   ".bashrc",
		ShellPattern: "shell.sh",
		ShellScript:  "shell.sh",
		InitShell:    func() error { return touchFile(filepath.Join(os.Getenv("HOME"), ".bashrc")) },
	},
	{
		Name:         "zsh",
		ConfigFile:   ".zshrc",
		ShellPattern: "shell.sh",
		ShellScript:  "shell.sh",
		InitShell:    func() error { return touchFile(filepath.Join(os.Getenv("HOME"), ".zshrc")) },
	},
	{
		Name:         "fish",
		ConfigFile:   ".config/fish/config.fish",
		ShellPattern: "shell.fish",
		ShellScript:  "shell.fish",
		InitShell: func() error {
			dir := filepath.Join(os.Getenv("HOME"), ".config/fish")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return touchFile(filepath.Join(dir, "config.fish"))
		},
	},
}

// Tool configuration for testing shell script content
type ToolConfig struct {
	Name    string
	Pattern string
	Shell   string // "bash", "zsh", or "fish"
}

var tools = []ToolConfig{
	{Name: "eza", Pattern: "alias ll='eza", Shell: "bash"},
	{Name: "bat", Pattern: "alias cat='bat", Shell: "bash"},
	{Name: "starship-bash", Pattern: "starship init ${BLING_SHELL}", Shell: "bash"},
	{Name: "starship-zsh", Pattern: "starship init ${BLING_SHELL}", Shell: "zsh"},
	{Name: "starship-fish", Pattern: "starship init fish", Shell: "fish"},
	{Name: "zoxide", Pattern: "zoxide init", Shell: "bash"},
	{Name: "atuin", Pattern: "atuin init", Shell: "bash"},
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

func TestShellEnableForAllShells(t *testing.T) {
	for _, shell := range shells {
		t.Run(shell.Name, func(t *testing.T) {
			// Enable shell
			_, err := runCommand(t, "shell", shell.Name, "on")
			if err != nil {
				t.Fatalf("Failed to enable shell for %s: %v", shell.Name, err)
			}
			
			// Verify config file contains the eval line
			configPath := filepath.Join(os.Getenv("HOME"), shell.ConfigFile)
			expected := "bluefin-cli init"
			if !fileContains(t, configPath, expected) {
				t.Errorf("Config file %s doesn't contain init command %s", configPath, expected)
			}
		})
	}
}

func TestShellScriptSourcing(t *testing.T) {
	for _, shell := range shells {
		t.Run(shell.Name, func(t *testing.T) {
			// Run init command and check if it outputs the script content
			output, err := runCommand(t, "init", shell.Name)
			if err != nil {
				t.Fatalf("Failed to run init: %v", err)
			}
			
			// We check for some shell specific syntax or standard env vars
			if !strings.Contains(output, "BLUEFIN_SHELL_ENABLE_EZA") {
				t.Errorf("Init output doesn't seem to contain shell script logic")
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

func TestShellToolConfigurations(t *testing.T) {
	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			// Run init to get the script content
			output, err := runCommand(t, "init", tool.Shell)
			if err != nil {
				t.Fatalf("Failed to run init: %v", err)
			}

			if !strings.Contains(output, tool.Pattern) {
				t.Errorf("Shell initialization script doesn't contain configuration for %s (pattern: %s)", tool.Name, tool.Pattern)
			}
		})
	}
}

func TestMOTDSystem(t *testing.T) {
	t.Run("MOTDInBashrc", func(t *testing.T) {
		output, _ := runCommand(t, "init", "bash")
		if !strings.Contains(output, "bluefin-cli motd show") {
			t.Error("MOTD hook missing from init output")
		}
	})
	
	// MOTD resources check removed as it depends on external setup not controlled by CLI logic under test
	
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

func TestShellDisable(t *testing.T) {
	_, err := runCommand(t, "shell", "bash", "off")
	if err != nil {
		t.Fatalf("Failed to disable shell: %v", err)
	}
	
	bashrc := filepath.Join(os.Getenv("HOME"), ".bashrc")
	if fileContains(t, bashrc, "# bluefin-cli shell-config") {
		t.Error("Shell marker still present in bashrc after disable")
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
