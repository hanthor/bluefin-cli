// Package wallpaper sets the desktop wallpaper natively per platform:
// macOS via the `wallpaper` CLI (NSWorkspace — no privacy prompt) with an
// osascript fallback, GNOME via gsettings. Windows automation lives in the
// sunset subsystem.
package wallpaper

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// imageExts are the file types offered by the picker.
var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".heic": true, ".webp": true,
}

// List returns wallpaper candidates from the collections bluefin-cli
// installs (Homebrew caskrooms) plus the usual wallpaper directories.
func List() []string {
	home, _ := os.UserHomeDir()
	var roots []string
	// Only wallpaper casks — every app cask ships icons and DMG art that
	// would drown the picker.
	for _, caskroom := range []string{"/opt/homebrew/Caskroom", "/usr/local/Caskroom"} {
		matches, _ := filepath.Glob(filepath.Join(caskroom, "*wallpaper*"))
		roots = append(roots, matches...)
	}
	roots = append(roots,
		filepath.Join(home, ".local", "share", "backgrounds"),
		filepath.Join(home, "Pictures", "Wallpapers"),
		"/usr/share/backgrounds",
	)

	seen := map[string]bool{}
	var out []string
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if len(out) >= 500 {
				return filepath.SkipAll
			}
			if imageExts[strings.ToLower(filepath.Ext(path))] && !seen[path] {
				seen[path] = true
				out = append(out, path)
			}
			return nil
		})
	}
	sort.Strings(out)
	return out
}
