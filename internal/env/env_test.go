package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	// Mock HOME for consistent test
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	// Mock HOMEBREW_PREFIX
	prefix := filepath.Join(tmpHome, "homebrew")
	os.Setenv("HOMEBREW_PREFIX", prefix)
	defer os.Unsetenv("HOMEBREW_PREFIX")

	homeConfig := filepath.Join(tmpHome, ".config", "bluefin-cli")
	brewConfig := filepath.Join(prefix, "etc", "bluefin-cli")

	// 1. Test Default: No local config, HOMEBREW_PREFIX set
	// Should return Brew config
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}
	if dir != brewConfig {
		t.Errorf("Expected Brew config %s, got %s", brewConfig, dir)
	}

	// 2. Test Override: Local config exists
	// Create local config dir
	if err := os.MkdirAll(homeConfig, 0755); err != nil {
		t.Fatalf("Failed to create mock local config: %v", err)
	}

	dir, err = GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}
	// Should now prefer local config
	if dir != homeConfig {
		t.Errorf("Expected Local config %s (override), got %s", homeConfig, dir)
	}

	// 3. Test Fallback: No HOMEBREW_PREFIX
	os.Unsetenv("HOMEBREW_PREFIX")
	dir, err = GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}
	if dir != homeConfig {
		t.Errorf("Expected Local config %s (fallback), got %s", homeConfig, dir)
	}
}
