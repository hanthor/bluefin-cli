// Package terminal sets up a great terminal emulator: install Ghostty, pin
// it to the macOS Dock, and write a themed config. The config uses
// Ghostty's built-in Catppuccin themes with automatic light/dark switching,
// matching the bluefin-cli palette.
package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GhosttyInstalled reports whether Ghostty is present.
func GhosttyInstalled() bool {
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat("/Applications/Ghostty.app"); err == nil {
			return true
		}
	}
	_, err := exec.LookPath("ghostty")
	return err == nil
}

// InstallGhostty installs Ghostty (Homebrew cask on macOS, brew elsewhere
// when available). Output streams to the caller (RunnerScreen captures it).
func InstallGhostty() error {
	if GhosttyInstalled() {
		fmt.Println("Ghostty is already installed.")
		return nil
	}
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew is required to install Ghostty: https://brew.sh")
	}
	args := []string{"install", "--cask", "ghostty"}
	if runtime.GOOS != "darwin" {
		args = []string{"install", "ghostty"}
	}
	cmd := exec.Command("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// dockPlistEntry is the persistent-apps dict for an app.
const dockPlistEntry = `<dict><key>tile-data</key><dict><key>file-data</key><dict><key>_CFURLString</key><string>%s</string><key>_CFURLStringType</key><integer>0</integer></dict></dict></dict>`

// PinToDock pins an app to the macOS Dock (idempotent). Uses dockutil when
// available; otherwise appends to the Dock's persistent-apps and restarts
// the Dock.
func PinToDock(appPath string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("dock pinning is macOS-only")
	}
	if _, err := os.Stat(appPath); err != nil {
		return fmt.Errorf("%s is not installed", appPath)
	}

	name := strings.TrimSuffix(filepath.Base(appPath), ".app")
	current, _ := exec.Command("defaults", "read", "com.apple.dock", "persistent-apps").Output()
	if strings.Contains(string(current), name+".app") {
		fmt.Printf("%s is already in the Dock.\n", name)
		return nil
	}

	if dockutil, err := exec.LookPath("dockutil"); err == nil {
		if out, err := exec.Command(dockutil, "--add", appPath).CombinedOutput(); err != nil {
			return fmt.Errorf("dockutil: %v: %s", err, out)
		}
		fmt.Printf("%s pinned to the Dock.\n", name)
		return nil
	}

	entry := fmt.Sprintf(dockPlistEntry, appPath)
	if out, err := exec.Command("defaults", "write", "com.apple.dock", "persistent-apps", "-array-add", entry).CombinedOutput(); err != nil {
		return fmt.Errorf("defaults write: %v: %s", err, out)
	}
	if err := exec.Command("killall", "Dock").Run(); err != nil {
		return fmt.Errorf("restarting Dock: %w", err)
	}
	fmt.Printf("%s pinned to the Dock.\n", name)
	return nil
}

// managed block markers inside the Ghostty config.
const (
	markerStart = "# --- bluefin-cli managed start ---"
	markerEnd   = "# --- bluefin-cli managed end ---"
)

// ghosttyConfigDir returns Ghostty's config directory (XDG path works on
// macOS too).
func ghosttyConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "ghostty"), nil
}

// detectNerdFont returns a font-family line for a known installed Nerd
// Font, or "" to let Ghostty use its default.
func detectNerdFont() string {
	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, "Library", "Fonts"),
		filepath.Join(home, ".local", "share", "fonts"),
		filepath.Join(home, ".fonts"),
	}
	known := []struct{ file, family string }{
		{"JetBrainsMonoNerdFont", "JetBrainsMono Nerd Font"},
		{"CaskaydiaCoveNerdFont", "CaskaydiaCove Nerd Font"},
		{"CaskaydiaMonoNerdFont", "CaskaydiaMono Nerd Font"},
		{"FiraCodeNerdFont", "FiraCode Nerd Font"},
		{"HackNerdFont", "Hack Nerd Font"},
		{"0xProtoNerdFont", "0xProto Nerd Font"},
	}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			for _, k := range known {
				if strings.Contains(e.Name(), k.file) {
					return k.family
				}
			}
		}
	}
	return ""
}

// WriteGhosttyConfig writes (or refreshes) the managed block in Ghostty's
// config: auto light/dark Catppuccin theme, padding, and a Nerd Font when
// one is installed. Existing user configuration outside the block is
// preserved untouched.
func WriteGhosttyConfig() error {
	dir, err := ghosttyConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	cfgPath := filepath.Join(dir, "config")

	var block strings.Builder
	block.WriteString(markerStart + "\n")
	block.WriteString("theme = light:catppuccin-latte,dark:catppuccin-mocha\n")
	block.WriteString("window-padding-x = 8\nwindow-padding-y = 8\n")
	block.WriteString("cursor-style = block\n")
	if fam := detectNerdFont(); fam != "" {
		fmt.Fprintf(&block, "font-family = %s\n", fam)
	}
	block.WriteString(markerEnd + "\n")

	existing, err := os.ReadFile(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	if start := strings.Index(content, markerStart); start >= 0 {
		if end := strings.Index(content, markerEnd); end >= 0 {
			content = content[:start] + content[end+len(markerEnd)+1:]
		}
	}
	content = strings.TrimRight(content, "\n")
	if content != "" {
		content += "\n\n"
	}
	content += block.String()

	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Ghostty config updated: %s\n", cfgPath)
	fmt.Println("Theme: Catppuccin (auto light/dark). Reload Ghostty with cmd+shift+, or restart it.")
	return nil
}
