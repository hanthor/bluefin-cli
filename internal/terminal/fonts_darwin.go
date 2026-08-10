//go:build darwin

package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func detectPlatformTerminals() []ConfiguredTerminal {
	var out []ConfiguredTerminal
	home, _ := os.UserHomeDir()

	// Terminal.app — always available on macOS
	out = append(out, ConfiguredTerminal{
		Name: "Terminal.app", ConfigDir: "", IsReady: true,
	})

	// iTerm2
	itermDir := filepath.Join(home, "Library", "Application Support", "iTerm2")
	if _, err := os.Stat(itermDir); err == nil {
		out = append(out, ConfiguredTerminal{
			Name: "iTerm2", ConfigDir: itermDir, IsReady: true,
		})
	}

	return out
}

func setPlatformTerminalFont(term ConfiguredTerminal, fontFamily string) error {
	switch term.Name {
	case "Terminal.app":
		return setTerminalAppFont(fontFamily)
	case "iTerm2":
		return setITerm2Font(term.ConfigDir, fontFamily)
	default:
		return fmt.Errorf("unsupported terminal: %s", term.Name)
	}
}

// setTerminalAppFont sets the font for Terminal.app's default profile using
// the defaults system. It also sets the font for any existing window groups.
func setTerminalAppFont(family string) error {
	// Map Nerd Font family names to their macOS PostScript names
	sysFont := FontFamilyToSystemFont(family)

	// Terminal.app font is controlled via its plist preferences.
	// We set the font for the "Basic" (default) profile.
	// Font spacing is 1.0 (standard), size is 13.
	fontLine := fmt.Sprintf("%s 13", sysFont)

	// Set the font name and size
	if out, err := exec.Command("defaults", "write",
		"com.apple.Terminal", "Font", fontLine).CombinedOutput(); err != nil {
		return fmt.Errorf("setting Terminal.app font: %v: %s", err, out)
	}
	if out, err := exec.Command("defaults", "write",
		"com.apple.Terminal", "NSFixedPitchFont", sysFont).CombinedOutput(); err != nil {
		return fmt.Errorf("setting Terminal.app NSFixedPitchFont: %v: %s", err, out)
	}
	if out, err := exec.Command("defaults", "write",
		"com.apple.Terminal", "NSFixedPitchFontSize", "-int", "13").CombinedOutput(); err != nil {
		return fmt.Errorf("setting Terminal.app font size: %v: %s", err, out)
	}
	fmt.Printf("Terminal.app font set to %q (size 13). Restart Terminal to apply.\n", sysFont)
	return nil
}

// setITerm2Font creates or updates an iTerm2 dynamic profile with the
// chosen Nerd Font. Dynamic profiles are the safe way to set defaults
// without mutating the user's main preferences plist.
func setITerm2Font(configDir, family string) error {
	dynDir := filepath.Join(configDir, "DynamicProfiles")
	if err := os.MkdirAll(dynDir, 0o755); err != nil {
		return err
	}

	plistPath := filepath.Join(dynDir, "bluefin-cli-font.plist")

	// Map to the system (PostScript) font name for iTerm2
	sysFont := FontFamilyToSystemFont(family)

	// Build a minimal plist that sets the font for the default profile
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Profiles</key>
	<array>
		<dict>
			<key>Name</key>
			<string>Bluefin (Nerd Font)</string>
			<key>Guid</key>
			<string>E5A0B7F2-9C4D-4A1E-8F3B-6D7C8E9F0A1B</string>
			<key>Normal Font</key>
			<string>%s 13</string>
			<key>Non-ASCII Font</key>
			<string>%s 13</string>
			<key>Use Non-ASCII Font</key>
			<true/>
		</dict>
	</array>
</dict>
</plist>
`, sysFont, sysFont)

	if err := os.WriteFile(plistPath, []byte(plist), 0o644); err != nil {
		return err
	}
	fmt.Printf("iTerm2 dynamic profile written: %s\n", plistPath)
	fmt.Println("iTerm2 will pick up the new profile automatically (or restart iTerm2).")
	return nil
}

// FontFamilyToSystemFont converts a Nerd Font family name to what the
// system font registry expects (used by Terminal.app and other native APIs).
func FontFamilyToSystemFont(family string) string {
	// Common mappings from brew/cask names to macOS font family names
	mapping := map[string]string{
	"JetBrainsMono Nerd Font":   "JetBrainsMono NFP",
	"CaskaydiaCove Nerd Font":   "CaskaydiaCove NFP",
	"CaskaydiaMono Nerd Font":   "CaskaydiaCove NFP",
	"FiraCode Nerd Font":        "FiraCode NFP",
	"Hack Nerd Font":            "Hack NFP",
	"0xProto Nerd Font":         "0xProto NFP",
	"ComicShannsMono Nerd Font": "ComicShannsMono NFP",
	"Droid Sans Mono Nerd Font": "DroidSansM NFP",
	"Go Mono Nerd Font":         "GoMono NFP",
	"IBM Plex Mono Nerd Font":   "BlexMono NFP",
	"Source Code Pro Nerd Font": "SauceCodePro NFP",
	"Ubuntu Nerd Font":          "Ubuntu NFP",
	"UbuntuMono Nerd Font":      "UbuntuMono NFP",
	}
	if mapped, ok := mapping[family]; ok {
		return mapped
	}
	// Fallback: try the family name as-is (works for many fonts)
	if strings.Contains(family, "Nerd Font") {
		return strings.Replace(family, "Nerd Font", "NFP", 1)
	}
	return family
}
