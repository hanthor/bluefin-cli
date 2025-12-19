package bling

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		Uutils:   false, // Default to false (opt-in)
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

// SaveConfig writes the configuration to file and regenerates env files
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

	return GenerateEnvFiles(config)
}

// GenerateEnvFiles creates the shell environment files with export statements
func GenerateEnvFiles(config *Config) error {
	blingDir := getBlingDir()
	if err := os.MkdirAll(blingDir, 0755); err != nil {
		return fmt.Errorf("failed to create bling directory: %w", err)
	}

	// Generate bash/zsh env file
	shContent := fmt.Sprintf(`export BLING_ENABLE_EZA=%d
export BLING_ENABLE_UGREP=%d
export BLING_ENABLE_BAT=%d
export BLING_ENABLE_ATUIN=%d
export BLING_ENABLE_STARSHIP=%d
export BLING_ENABLE_ZOXIDE=%d
export BLING_ENABLE_UUTILS=%d
`, boolToInt(config.Eza), boolToInt(config.Ugrep), boolToInt(config.Bat), boolToInt(config.Atuin), boolToInt(config.Starship), boolToInt(config.Zoxide), boolToInt(config.Uutils))

	if err := os.WriteFile(filepath.Join(blingDir, "bling-env.sh"), []byte(shContent), 0644); err != nil {
		return fmt.Errorf("failed to write bash env file: %w", err)
	}

	// Generate fish env file
	fishContent := fmt.Sprintf(`set -gx BLING_ENABLE_EZA %d
set -gx BLING_ENABLE_UGREP %d
set -gx BLING_ENABLE_BAT %d
set -gx BLING_ENABLE_ATUIN %d
set -gx BLING_ENABLE_STARSHIP %d
set -gx BLING_ENABLE_ZOXIDE %d
set -gx BLING_ENABLE_UUTILS %d
`, boolToInt(config.Eza), boolToInt(config.Ugrep), boolToInt(config.Bat), boolToInt(config.Atuin), boolToInt(config.Starship), boolToInt(config.Zoxide), boolToInt(config.Uutils))

	if err := os.WriteFile(filepath.Join(blingDir, "bling-env.fish"), []byte(fishContent), 0644); err != nil {
		return fmt.Errorf("failed to write fish env file: %w", err)
	}

	return nil
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bluefin-cli", "bling-config.json"), nil
}

func getBlingDir() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".local/share/bluefin-cli/bling")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
