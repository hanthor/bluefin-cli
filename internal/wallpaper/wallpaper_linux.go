package wallpaper

import (
	"fmt"
	"os/exec"
)

// Supported reports whether a native setter is available (GNOME).
func Supported() bool {
	_, err := exec.LookPath("gsettings")
	return err == nil
}

// Set applies the wallpaper for both light and dark GNOME variants.
func Set(path string) error {
	uri := "file://" + path
	for _, key := range []string{"picture-uri", "picture-uri-dark"} {
		if out, err := exec.Command("gsettings", "set", "org.gnome.desktop.background", key, uri).CombinedOutput(); err != nil {
			return fmt.Errorf("gsettings %s: %v: %s", key, err, out)
		}
	}
	return nil
}

// Get returns the current GNOME wallpaper URI.
func Get() (string, error) {
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.background", "picture-uri").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
