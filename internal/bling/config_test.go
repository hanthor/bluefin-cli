package bling

import (
	"os"
	"testing"
)

func TestConfigData(t *testing.T) {
	// Setup temp home
	tmpHome := t.TempDir()
	os.Setenv("HOMEBREW_PREFIX", tmpHome) // Mock Homebrew Prefix
	defer os.Unsetenv("HOMEBREW_PREFIX")
	
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

}

