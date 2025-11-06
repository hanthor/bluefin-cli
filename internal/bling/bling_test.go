package bling

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToggle(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	tests := []struct {
		name    string
		shell   string
		enable  bool
		wantErr bool
	}{
		{"Enable bash bling", "bash", true, false},
		{"Enable zsh bling", "zsh", true, false},
		{"Enable fish bling", "fish", true, false},
		{"Disable bash bling", "bash", false, false},
		{"Invalid shell", "invalid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Toggle(tt.shell, tt.enable)
			if (err != nil) != tt.wantErr {
				t.Errorf("Toggle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.shell != "invalid" {
				// Verify config file was created/modified
				var configFile string
				switch tt.shell {
				case "bash":
					configFile = filepath.Join(tmpHome, ".bashrc")
				case "zsh":
					configFile = filepath.Join(tmpHome, ".zshrc")
				case "fish":
					configFile = filepath.Join(tmpHome, ".config/fish/config.fish")
				}

				content, err := os.ReadFile(configFile)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}

				hasMarker := strings.Contains(string(content), blingMarker)
				if tt.enable && !hasMarker {
					t.Error("Expected bling marker in config file when enabling")
				}
				if !tt.enable && hasMarker {
					t.Error("Expected no bling marker in config file when disabling")
				}
			}
		})
	}
}

func TestCheckStatus(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	// Enable bling for bash
	if err := Toggle("bash", true); err != nil {
		t.Fatalf("Failed to enable bling: %v", err)
	}

	status := CheckStatus()

	if !status["bash"] {
		t.Error("Expected bash bling to be enabled")
	}
	if status["zsh"] {
		t.Error("Expected zsh bling to be disabled")
	}
	if status["fish"] {
		t.Error("Expected fish bling to be disabled")
	}
}

func TestCheckDependencies(t *testing.T) {
	deps := CheckDependencies()

	// We can't guarantee what's installed, but the function should return a map
	if deps == nil {
		t.Error("Expected non-nil dependencies map")
	}

	// Check expected tools are in the map
	expectedTools := []string{"eza", "bat", "zoxide", "atuin", "starship", "ugrep"}
	for _, tool := range expectedTools {
		if _, exists := deps[tool]; !exists {
			t.Errorf("Expected tool %s to be in dependencies map", tool)
		}
	}
}

func TestEnsureBlingScript(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	tests := []struct {
		name      string
		shell     string
		wantFile  string
		wantErr   bool
	}{
		{"Bash script", "bash", "bling.sh", false},
		{"Zsh script", "zsh", "bling.sh", false},
		{"Fish script", "fish", "bling.fish", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ensureBlingScript(tt.shell)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureBlingScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file exists
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("Expected bling script to exist at %s", path)
				}

				// Verify it's in the right directory
				if !strings.Contains(path, ".local/share/bluefin-cli/bling") {
					t.Errorf("Expected path to contain .local/share/bluefin-cli/bling, got %s", path)
				}

				// Verify filename
				if !strings.HasSuffix(path, tt.wantFile) {
					t.Errorf("Expected path to end with %s, got %s", tt.wantFile, path)
				}

				// Verify file is executable
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("Failed to stat file: %v", err)
				}
				if info.Mode().Perm()&0111 == 0 {
					t.Error("Expected bling script to be executable")
				}
			}
		})
	}
}

func TestToggleIdempotency(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	// Enable twice, should be idempotent
	if err := Toggle("bash", true); err != nil {
		t.Fatalf("First enable failed: %v", err)
	}

	if err := Toggle("bash", true); err != nil {
		t.Fatalf("Second enable failed: %v", err)
	}

	// Check config file doesn't have duplicate entries
	configFile := filepath.Join(tmpHome, ".bashrc")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	markerCount := strings.Count(string(content), blingMarker)
	if markerCount != 1 {
		t.Errorf("Expected 1 bling marker, found %d", markerCount)
	}

	// Disable twice, should be idempotent
	if err := Toggle("bash", false); err != nil {
		t.Fatalf("First disable failed: %v", err)
	}

	if err := Toggle("bash", false); err != nil {
		t.Fatalf("Second disable failed: %v", err)
	}

	// Verify marker is gone
	content, err = os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if strings.Contains(string(content), blingMarker) {
		t.Error("Expected bling marker to be removed after disable")
	}
}
