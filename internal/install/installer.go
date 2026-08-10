package install

import "github.com/tuna-os/bluefin-cli/internal/env"

// Installer defines the interface for platform-specific installation operations.
type Installer interface {
	InstallBundle(nameOrPath ...string) error
	InstallWallpapers(casks []string) error
	CleanupWallpapers(all bool) error
}

var currentInstaller Installer

// SetInstaller sets the global installer instance.
func SetInstaller(i Installer) {
	currentInstaller = i
}

// GetInstaller returns the current global installer.
func GetInstaller() Installer {
	return currentInstaller
}

// initInstaller selects the installer for the current Unix-like platform.
// Windows has its own installer in windows_installer.go.
func initInstaller(unix, alpine Installer) {
	if env.IsAlpine() {
		SetInstaller(alpine)
		return
	}
	SetInstaller(unix)
}
