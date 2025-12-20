package install

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "strings"
)

const wallpapersTap = "ublue-os/tap"

// EnsureBrew ensures Homebrew exists in PATH
func EnsureBrew() error {
    if _, err := exec.LookPath("brew"); err != nil {
        return fmt.Errorf("Homebrew not found. Please install Homebrew first: https://brew.sh")
    }
    return nil
}

// ensureTap ensures a tap is added (idempotent)
func ensureTap(tap string) error {
    if err := EnsureBrew(); err != nil {
        return err
    }
    cmd := exec.Command("brew", "tap", tap)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

// GetWallpaperCasks returns the list of casks available in the ublue-os/tap tap
func GetWallpaperCasks() ([]string, error) {
    if err := ensureTap(wallpapersTap); err != nil {
        return nil, err
    }
    // Query casks in the tap. We use `brew search --casks tap/` and parse lines.
    // Example output lines may look like: tapname/caskname
    cmd := exec.Command("brew", "search", "--casks", wallpapersTap+"/")
    out, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to search casks: %w", err)
    }
    var casks []string
    scanner := bufio.NewScanner(strings.NewReader(string(out)))
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        if strings.Contains(line, "/") {
            parts := strings.Split(line, "/")
            line = parts[len(parts)-1]
        }
        casks = append(casks, line)
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    var filtered []string
    for _, c := range casks {
        if strings.Contains(strings.ToLower(c), "wallpaper") {
            filtered = append(filtered, c)
        }
    }
    if len(filtered) > 0 {
        return filtered, nil
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
    // Install all selected casks in a single brew command for better UX
    args := []string{"install", "--cask"}
    for _, c := range casks {
        // Accept both plain and tap-qualified; ensure tap-qualified
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
    return nil
}
