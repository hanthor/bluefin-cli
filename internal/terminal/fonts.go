package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfiguredTerminal holds a detected terminal and its font-ready status.
type ConfiguredTerminal struct {
	Name      string // e.g. "Alacritty", "Kitty", "Windows Terminal"
	ConfigDir string // directory containing config (empty if not found)
	IsReady   bool   // true if the terminal has a config we can safely edit
}

// DetectTerminals returns terminals found on this system whose configs
// we can safely update with a Nerd Font.
func DetectTerminals() []ConfiguredTerminal {
	var out []ConfiguredTerminal
	out = append(out, detectPlatformTerminals()...)

	// Cross-platform terminals
	home, _ := os.UserHomeDir()
	if home == "" {
		return out
	}

	// Ghostty — already handled by WriteGhosttyConfig, but listed here
	// so the fonts flow can update its font-family.
	if dir, err := ghosttyConfigDir(); err == nil {
		if _, err := os.Stat(dir); err == nil {
			out = append(out, ConfiguredTerminal{
				Name: "Ghostty", ConfigDir: dir, IsReady: true,
			})
		}
	}

	// Alacritty
	ad := filepath.Join(home, ".config", "alacritty")
	if fi, err := os.Stat(filepath.Join(ad, "alacritty.toml")); err == nil && !fi.IsDir() {
		out = append(out, ConfiguredTerminal{
			Name: "Alacritty", ConfigDir: ad, IsReady: true,
		})
	}

	// Kitty
	kd := filepath.Join(home, ".config", "kitty")
	if fi, err := os.Stat(filepath.Join(kd, "kitty.conf")); err == nil && !fi.IsDir() {
		out = append(out, ConfiguredTerminal{
			Name: "Kitty", ConfigDir: kd, IsReady: true,
		})
	}

	// WezTerm — config lives at ~/.wezterm.lua
	if _, err := os.Stat(filepath.Join(home, ".wezterm.lua")); err == nil {
		out = append(out, ConfiguredTerminal{
			Name: "WezTerm", ConfigDir: home, IsReady: true,
		})
	}

	return out
}

// SetTerminalFont updates the given terminal's config to use the provided
// font family. It returns a human-readable summary of what was done.
func SetTerminalFont(term ConfiguredTerminal, fontFamily string) error {
	switch term.Name {
	case "Ghostty":
		return updateGhosttyFont(fontFamily)
	case "Alacritty":
		return updateAlacrittyFont(term.ConfigDir, fontFamily)
	case "Kitty":
		return updateKittyFont(term.ConfigDir, fontFamily)
	case "WezTerm":
		return updateWezTermFont(term.ConfigDir, fontFamily)
	default:
		return setPlatformTerminalFont(term, fontFamily)
	}
}

// --- Ghostty -----------------------------------------------------------

func updateGhosttyFont(family string) error {
	dir, err := ghosttyConfigDir()
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(dir, "config")

	existing, err := os.ReadFile(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(existing)
	// Remove existing managed block if any
	if start := strings.Index(content, markerStart); start >= 0 {
		if end := strings.Index(content, markerEnd); end >= 0 {
			content = content[:start] + content[end+len(markerEnd)+1:]
		}
	}
	content = strings.TrimRight(content, "\n")

	var block strings.Builder
	block.WriteString(markerStart + "\n")
	block.WriteString("theme = light:catppuccin-latte,dark:catppuccin-mocha\n")
	block.WriteString("window-padding-x = 8\nwindow-padding-y = 8\n")
	block.WriteString("cursor-style = block\n")
	fmt.Fprintf(&block, "font-family = %s\n", family)
	block.WriteString(markerEnd + "\n")

	if content != "" {
		content += "\n\n"
	}
	content += block.String()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Ghostty font set to %q in %s\n", family, cfgPath)
	return nil
}

// --- Alacritty ---------------------------------------------------------

func updateAlacrittyFont(configDir, family string) error {
	cfgPath := filepath.Join(configDir, "alacritty.toml")
	existing, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	content := string(existing)
	newFontLine := fmt.Sprintf(`family = "%s"`, family)

	// Check if [font] section exists and has a "normal" sub-table
	if strings.Contains(content, "[font]") {
		content = replaceOrInsertTOMLKey(content, "[font]", "family", newFontLine)
	} else {
		// Add a [font] section at the end
		content = strings.TrimRight(content, "\n") + "\n\n[font]\n" + newFontLine + "\n"
	}

	if strings.Contains(content, "[font.normal]") {
		content = replaceOrInsertTOMLKey(content, "[font.normal]", "family", newFontLine)
	}

	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Alacritty font set to %q in %s\n", family, cfgPath)
	return nil
}

// --- Kitty -------------------------------------------------------------

func updateKittyFont(configDir, family string) error {
	cfgPath := filepath.Join(configDir, "kitty.conf")
	existing, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	content := string(existing)
	newLine := fmt.Sprintf("font_family %s", family)

	if strings.Contains(content, "font_family") {
		content = replaceKeyLine(content, "font_family", newLine)
	} else {
		content = strings.TrimRight(content, "\n") + "\n" + newLine + "\n"
	}

	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Kitty font set to %q in %s\n", family, cfgPath)
	return nil
}

// --- WezTerm -----------------------------------------------------------

func updateWezTermFont(configDir, family string) error {
	cfgPath := filepath.Join(configDir, ".wezterm.lua")
	existing, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	content := string(existing)
	newLine := fmt.Sprintf("config.font = wezterm.font '%s'", family)

	// Match config.font = ... but not config.font_size etc.
	fontKey := "config.font "
	if strings.Contains(content, fontKey) {
		content = replaceKeyLine(content, fontKey, newLine)
	} else if strings.Contains(content, "config.font=") {
		content = replaceKeyLine(content, "config.font=", newLine)
	} else {
		// Insert before the return config line
		insertBefore := "\nreturn config"
		if idx := strings.Index(content, insertBefore); idx >= 0 {
			content = content[:idx] + newLine + "\n" + content[idx:]
		} else {
			content = strings.TrimRight(content, "\n") + "\n" + newLine + "\n"
		}
	}

	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("WezTerm font set to %q in %s\n", family, cfgPath)
	return nil
}

// --- helpers -----------------------------------------------------------

// replaceKeyLine replaces lines starting with 'key' in content with newLine.
func replaceKeyLine(content, key, newLine string) string {
	lines := strings.Split(content, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key) {
			// Preserve leading whitespace
			lead := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			out = append(out, lead+newLine)
		} else {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// replaceOrInsertTOMLKey replaces a key under the given TOML section header.
// If the key doesn't exist after the header, it appends it before the next
// section or EOF.
func replaceOrInsertTOMLKey(content, section, key, line string) string {
	secIdx := strings.Index(content, section)
	if secIdx < 0 {
		return content
	}
	// Find the start of the next section (a line starting with [)
	rest := content[secIdx+len(section):]
	nextSec := strings.Index(rest, "\n[")
	var sectionBody string
	if nextSec >= 0 {
		sectionBody = rest[:nextSec]
	} else {
		sectionBody = rest
	}

	// Replace existing key within the section body
	if strings.Contains(sectionBody, key) {
		newBody := replaceKeyLine(sectionBody, key, line)
		return content[:secIdx+len(section)] + newBody + content[secIdx+len(section)+len(sectionBody):]
	}

	// Insert after the section header
	insertPos := secIdx + len(section)
	return content[:insertPos] + "\n" + line + content[insertPos:]
}
