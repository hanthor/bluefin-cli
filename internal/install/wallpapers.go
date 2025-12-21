package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const wallpapersTap = "ublue-os/tap"

func EnsureBrew() error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew not found. Please install Homebrew first: https://brew.sh")
	}
	return nil
}

func ensureTap(tap string) error {
	if err := EnsureBrew(); err != nil {
		return err
	}
	cmd := exec.Command("brew", "tap", tap)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetWallpaperCasks() ([]string, error) {
	if err := ensureTap(wallpapersTap); err != nil {
		return nil, err
	}

	cmd := exec.Command("brew", "--repository", wallpapersTap)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tap repository path: %w", err)
	}

	tapPath := strings.TrimSpace(string(out))
	casksDir := filepath.Join(tapPath, "Casks")

	entries, err := os.ReadDir(casksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read casks directory at %s: %w", casksDir, err)
	}

	var casks []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".rb") {
			caskName := strings.TrimSuffix(name, ".rb")
			if strings.Contains(strings.ToLower(caskName), "wallpaper") {
				casks = append(casks, caskName)
			}
		}
	}

	return casks, nil
}

func InstallWallpaperCasks(casks []string) error {
	if err := ensureTap(wallpapersTap); err != nil {
		return err
	}
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks selected")
	}
	args := []string{"install", "--cask"}
	for _, c := range casks {
		if strings.Contains(c, "/") {
			args = append(args, c)
		} else {
			args = append(args, wallpapersTap+"/"+c)
		}
	}
	cmd := exec.Command("brew", args...)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_ENV_HINTS=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install wallpaper casks: %w", err)
	}
	fmt.Println(successStyle.Render("✓ Wallpaper casks installed!"))

	// macOS specific instructions
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		fmt.Println("\n" + infoStyle.Render("Wallpapers installed to: "+filepath.Join(home, "Library/Desktop Pictures")))
		fmt.Println(infoStyle.Render("To use: System Settings > Wallpaper > Add Folder"))
	}

	return nil
}
