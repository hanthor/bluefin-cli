//go:build windows

package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func detectPlatformTerminals() []ConfiguredTerminal {
	var out []ConfiguredTerminal

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return out
	}

	// Windows Terminal (installed via MS Store)
	wtDir := filepath.Join(localAppData, "Packages", "Microsoft.WindowsTerminal_8wekyb3d8bbwe", "LocalState")
	if _, err := os.Stat(filepath.Join(wtDir, "settings.json")); err == nil {
		out = append(out, ConfiguredTerminal{
			Name: "Windows Terminal", ConfigDir: wtDir, IsReady: true,
		})
	}

	// Windows Terminal (unpackaged / winget install)
	wtDir2 := filepath.Join(localAppData, "Microsoft", "Windows Terminal")
	if _, err := os.Stat(filepath.Join(wtDir2, "settings.json")); err == nil {
		// Avoid duplicates
		found := false
		for _, t := range out {
			if t.Name == "Windows Terminal" {
				found = true
				break
			}
		}
		if !found {
			out = append(out, ConfiguredTerminal{
				Name: "Windows Terminal", ConfigDir: wtDir2, IsReady: true,
			})
		}
	}

	return out
}

func setPlatformTerminalFont(term ConfiguredTerminal, fontFamily string) error {
	switch term.Name {
	case "Windows Terminal":
		return setWindowsTerminalFont(term.ConfigDir, fontFamily)
	default:
		return fmt.Errorf("unsupported terminal: %s", term.Name)
	}
}

// setWindowsTerminalFont updates the Windows Terminal settings.json to use
// the given font face in the default profile settings.
func setWindowsTerminalFont(configDir, family string) error {
	cfgPath := filepath.Join(configDir, "settings.json")

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("reading Windows Terminal settings: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parsing Windows Terminal settings.json: %w", err)
	}

	// Set font.face in profiles.defaults
	profiles, ok := settings["profiles"].(map[string]interface{})
	if !ok {
		profiles = map[string]interface{}{}
		settings["profiles"] = profiles
	}

	defaults, ok := profiles["defaults"].(map[string]interface{})
	if !ok {
		defaults = map[string]interface{}{}
		profiles["defaults"] = defaults
	}

	font, ok := defaults["font"].(map[string]interface{})
	if !ok {
		font = map[string]interface{}{}
		defaults["font"] = font
	}

	font["face"] = family
	defaults["font"] = font

	out, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return fmt.Errorf("serializing Windows Terminal settings: %w", err)
	}

	if err := os.WriteFile(cfgPath, out, 0o644); err != nil {
		return fmt.Errorf("writing Windows Terminal settings: %w", err)
	}
	fmt.Printf("Windows Terminal font set to %q in %s\n", family, cfgPath)
	fmt.Println("Windows Terminal will pick up the new font on next launch.")
	return nil
}

// FontBackupPath returns a backup of the settings.json before modification,
// so the user can restore it if needed.
func FontBackupPath(configDir string) string {
	return filepath.Join(configDir, "settings.json.bluefin-backup")
}

// BackupWindowsTerminalSettings creates a backup of settings.json.
func BackupWindowsTerminalSettings(configDir string) error {
	cfgPath := filepath.Join(configDir, "settings.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}
	return os.WriteFile(FontBackupPath(configDir), data, 0o644)
}
