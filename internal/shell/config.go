package shell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hanthor/bluefin-cli/internal/env"
)

// Config holds the configuration for shell experience tools
// Config holds the configuration for shell experience tools
type Config map[string]bool

// IsEnabled returns true if the tool with the given name is enabled in the config
func (c Config) IsEnabled(toolName string) bool {
	// We use lowercased tool names as keys
	key := strings.ToLower(toolName)
	if enabled, ok := c[key]; ok {
		return enabled
	}
	// Fallback to default from Tools definition if key missing (though DefaultConfig should populate it)
	// Or maybe false? 
	// Ideally Config should be fully populated. But if we load an old config that misses a new tool, 
	// we should fallback to that tool's default.
	for _, t := range Tools {
		if t.Name == toolName {
			return t.Default
		}
	}
	return false
}

// SetEnabled sets the enabled status of a tool
func (c Config) SetEnabled(toolName string, enabled bool) {
	key := strings.ToLower(toolName)
	c[key] = enabled
}

// DefaultConfig returns a configuration with all tools enabled
func DefaultConfig() *Config {
	cfg := make(Config)
	
	for _, tool := range Tools {
		// Use lowercased name as key for JSON compatibility
		cfg[strings.ToLower(tool.Name)] = tool.Default
	}
	
	return &cfg
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
	
	shellConfig := filepath.Join(dir, "shell.json")
	blingConfig := filepath.Join(dir, "bling.json")

	// Migration: If shell.json doesn't exist but bling.json does, rename it
	if _, err := os.Stat(shellConfig); os.IsNotExist(err) {
		if _, err := os.Stat(blingConfig); err == nil {
			// found old config, rename it
			if err := os.Rename(blingConfig, shellConfig); err != nil {
				// warn but don't fail, we'll just start fresh or read old one if we fell back logic (but we won't complexity here)
				// simplest is just print check? No, we can't print easily here without dep. 
				// Just let it be. If rename fails, we'll just return shell.json path and it will be created as new.
			}
		}
	}

	return shellConfig, nil
}

// getBlingDir removed


func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
