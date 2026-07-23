package wallpaper

import "fmt"

// Supported: Windows wallpaper automation is owned by the sunset subsystem.
func Supported() bool { return false }

func Set(path string) error {
	return fmt.Errorf("on Windows, use sunset switching (bluefin-cli sunset) for wallpaper automation")
}

func Get() (string, error) { return "", fmt.Errorf("not supported on Windows") }
