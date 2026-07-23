package wallpaper

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

const (
	spiSetDesktopWallpaper = 0x0014
	spifUpdateIniFile      = 0x01
	spifSendChange         = 0x02
)

// Supported: native wallpaper setting works on Windows too.
func Supported() bool { return true }

// Set applies the wallpaper via SystemParametersInfoW — the same API the
// sunset subsystem uses — persisted and broadcast to the shell.
func Set(path string) error {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	ret, _, callErr := systemParametersInfo.Call(
		uintptr(spiSetDesktopWallpaper),
		0,
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(spifUpdateIniFile|spifSendChange),
	)
	if ret == 0 {
		return fmt.Errorf("SystemParametersInfoW failed: %w", callErr)
	}
	return nil
}

// Get reads the current wallpaper from the registry.
func Get() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\Desktop`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer func() { _ = k.Close() }()
	val, _, err := k.GetStringValue("Wallpaper")
	return val, err
}
