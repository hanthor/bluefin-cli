package bling

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hanthor/bluefin-cli/internal/env"
)

// Config holds the configuration for bling tools
type Config struct {
	Eza      bool `json:"eza"`
	Ugrep    bool `json:"ugrep"`
	Bat      bool `json:"bat"`
	Atuin    bool `json:"atuin"`
	Starship bool `json:"starship"`
	Zoxide   bool `json:"zoxide"`
	Uutils   bool `json:"uutils"`
}

// DefaultConfig returns a configuration with all tools enabled
func DefaultConfig() *Config {
	return &Config{
		Eza:      true,
		Ugrep:    true,
		Bat:      true,
		Atuin:    true,
		Starship: true,
		Zoxide:   true,
		Uutils:   true, // Default to true (opt-out)
	}
}

// LoadConfig reads the configuration from file or returns default if not found
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig writes the configuration to file
func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GenerateEnvFiles was removed as we now use 'bluefin-cli init'


func getConfigPath() (string, error) {
	dir, err := env.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "bling.json"), nil
}

// getBlingDir removed


func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
