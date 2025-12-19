package bling

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigData(t *testing.T) {
	// Setup temp home
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome) // For getBlingDir and getConfigPath (via UserHomeDir usually, but let's mock if needed or rely on HOME env var if Go respects it on linux)
	// On Linux, UserHomeDir reads $HOME.
	
	// Test Default Config
	cfg := DefaultConfig()
	if !cfg.Eza {
		t.Error("Default config should have Eza enabled")
	}

	// Test Save and Load
	cfg.Eza = false
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.Eza {
		t.Error("Expected Eza to be disabled after save")
	}
	if !loadedCfg.Starship {
		t.Error("Expected Starship to be enabled (unchanged)")
	}

	// Test Env File Generation
	envFile := filepath.Join(tmpHome, ".local/share/bluefin-cli/bling/bling-env.sh")
	content, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("Failed to read env file: %v", err)
	}

	if !strings.Contains(string(content), "export BLING_ENABLE_EZA=0") {
		t.Error("Expected env file to contain BLING_ENABLE_EZA=0")
	}
	if !strings.Contains(string(content), "export BLING_ENABLE_STARSHIP=1") {
		t.Error("Expected env file to contain BLING_ENABLE_STARSHIP=1")
	}
	if !strings.Contains(string(content), "export BLING_ENABLE_UUTILS=0") {
		t.Error("Expected env file to contain BLING_ENABLE_UUTILS=0")
	}

	// Test Fish Env File
	fishEnvFile := filepath.Join(tmpHome, ".local/share/bluefin-cli/bling/bling-env.fish")
	fishContent, err := os.ReadFile(fishEnvFile)
	if err != nil {
		t.Fatalf("Failed to read fish env file: %v", err)
	}

	if !strings.Contains(string(fishContent), "set -gx BLING_ENABLE_EZA 0") {
		t.Error("Expected fish env file to contain BLING_ENABLE_EZA 0")
	}
	if !strings.Contains(string(fishContent), "set -gx BLING_ENABLE_UUTILS 0") {
		t.Error("Expected fish env file to contain BLING_ENABLE_UUTILS 0")
	}
}
