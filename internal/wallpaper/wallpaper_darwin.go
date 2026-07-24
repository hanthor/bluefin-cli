package wallpaper

import (
	"fmt"
	"os/exec"
)

// Supported reports whether this platform can set wallpapers natively.
func Supported() bool { return true }

// backendHint tells the user how to get the reliable backend.
const backendHint = "brew install wallpaper"

// Set applies the wallpaper. The `wallpaper` CLI (Homebrew core) drives
// NSWorkspace directly and needs no privacy consent; the osascript fallback
// may trigger a one-time Automation permission prompt.
func Set(path string) error {
	if bin, err := exec.LookPath("wallpaper"); err == nil {
		if out, err := exec.Command(bin, "set", path).CombinedOutput(); err != nil {
			return fmt.Errorf("wallpaper set: %v: %s", err, out)
		}
		return nil
	}
	script := fmt.Sprintf(`tell application "System Events" to tell every desktop to set picture to %q`, path)
	if out, err := exec.Command("osascript", "-e", script).CombinedOutput(); err != nil {
		return fmt.Errorf("osascript (approve the Automation prompt, or %s): %v: %s", backendHint, err, out)
	}
	return nil
}

// Get returns the current wallpaper path when the backend supports reading.
func Get() (string, error) {
	if bin, err := exec.LookPath("wallpaper"); err == nil {
		out, err := exec.Command(bin, "get").Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}
	return "", fmt.Errorf("install the backend to read the current wallpaper: %s", backendHint)
}
