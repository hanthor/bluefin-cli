package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// HomeBrewfile returns the user's Brewfile path, preferring an existing
// ~/Brewfile then ~/.Brewfile; defaults to ~/Brewfile when neither exists.
func HomeBrewfile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "Brewfile"
	}
	for _, name := range []string{"Brewfile", ".Brewfile"} {
		p := filepath.Join(home, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(home, "Brewfile")
}

// PackageKinds are the entry types a Brewfile can hold — Homebrew on
// unix, the three Windows managers on Windows. One file, whole machine.
var PackageKinds = []string{"brew", "cask", "winget", "scoop", "choco"}

// AddToBrewfile appends a package entry (idempotent; creates the file).
func AddToBrewfile(path, name, kind string) error {
	valid := false
	for _, k := range PackageKinds {
		valid = valid || k == kind
	}
	if !valid {
		return fmt.Errorf("kind must be one of %s, got %q", strings.Join(PackageKinds, "/"), kind)
	}
	for _, p := range mustParse(path) {
		if p.ID == name {
			fmt.Printf("%s already in %s\n", name, path)
			return nil
		}
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := fmt.Fprintf(f, "%s \"%s\"\n", kind, name); err != nil {
		return err
	}
	fmt.Printf("Added %s %q to %s\n", kind, name, path)
	return nil
}

// RemoveFromBrewfile deletes the entries for the given package names,
// leaving every other line (taps, comments, unrelated entries) untouched.
func RemoveFromBrewfile(path string, names []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	drop := map[string]bool{}
	for _, n := range names {
		drop[n] = true
	}
	var out []string
	removed := 0
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "brew ") || strings.HasPrefix(t, "cask ") {
			if name := extractQuotedName(t); name != "" && drop[name] {
				removed++
				continue
			}
		}
		out = append(out, line)
	}
	if removed == 0 {
		return fmt.Errorf("no matching entries in %s", path)
	}
	if err := os.WriteFile(path, []byte(strings.Join(out, "\n")), 0o644); err != nil {
		return err
	}
	fmt.Printf("Removed %d entr%s from %s\n", removed, map[bool]string{true: "y", false: "ies"}[removed == 1], path)
	return nil
}

// DumpBrewfile writes the currently installed packages to path — on unix
// via `brew bundle dump`, on Windows by exporting winget (and scoop when
// present) into the unified format.
func DumpBrewfile(path string) error {
	if runtime.GOOS == "windows" {
		return dumpWindows(path)
	}
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew not found: https://brew.sh")
	}
	cmd := exec.Command("brew", "bundle", "dump", "--file="+path, "--force")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew bundle dump: %w", err)
	}
	fmt.Printf("Captured installed packages to %s\n", path)
	return nil
}

// dumpWindows captures winget and scoop inventories into the unified file.
func dumpWindows(path string) error {
	var b strings.Builder
	b.WriteString("# Captured by bluefin-cli brewfile dump\n")

	if _, err := exec.LookPath("winget"); err == nil {
		tmp, err := os.CreateTemp("", "winget-export-*.json")
		if err != nil {
			return err
		}
		tmpName := tmp.Name()
		_ = tmp.Close()
		defer func() { _ = os.Remove(tmpName) }()
		if out, err := exec.Command("winget", "export", "-o", tmpName, "--accept-source-agreements").CombinedOutput(); err != nil {
			fmt.Printf("winget export: %v\n%s", err, out)
		} else if data, err := os.ReadFile(tmpName); err == nil {
			var doc struct {
				Sources []struct {
					Packages []struct {
						PackageIdentifier string `json:"PackageIdentifier"`
					} `json:"Packages"`
				} `json:"Sources"`
			}
			if json.Unmarshal(data, &doc) == nil {
				for _, s := range doc.Sources {
					for _, p := range s.Packages {
						fmt.Fprintf(&b, "winget \"%s\"\n", p.PackageIdentifier)
					}
				}
			}
		}
	}
	if _, err := exec.LookPath("scoop"); err == nil {
		// scoop is a .cmd shim — route through cmd for reliability.
		if out, err := exec.Command("cmd", "/c", "scoop", "list").Output(); err == nil {
			lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
			headerSeen := false
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				if !headerSeen {
					headerSeen = fields[0] == "Name"
					continue
				}
				if strings.HasPrefix(fields[0], "---") {
					continue
				}
				fmt.Fprintf(&b, "scoop \"%s\"\n", fields[0])
			}
		}
	}
	if _, err := exec.LookPath("choco"); err == nil {
		if out, err := exec.Command("choco", "list", "--limit-output").Output(); err == nil {
			for _, p := range parseChocoSearch(string(out)) {
				fmt.Fprintf(&b, "choco \"%s\"\n", p.ID)
			}
		}
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("Captured installed packages to %s\n", path)
	return nil
}

// InstallBrewfileAll installs everything in the Brewfile: brew/cask lines
// go through `brew bundle` in one shot; winget/scoop/choco entries are
// installed per package with their manager (skipped with a note when the
// manager is absent).
func InstallBrewfileAll(path string) error {
	pkgs, err := GetBrewfilePackages(path)
	if err != nil {
		return err
	}
	hasBrewLines := false
	for _, p := range pkgs {
		if p.Kind == "brew" || p.Kind == "cask" {
			hasBrewLines = true
		}
	}
	if hasBrewLines {
		if _, err := exec.LookPath("brew"); err != nil {
			fmt.Println("skipping brew/cask entries: homebrew not found (https://brew.sh)")
		} else {
			cmd := exec.Command("brew", "bundle", "install", "--file="+path)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	managers := map[string][]string{
		"winget": {"install", "--exact", "--accept-package-agreements", "--accept-source-agreements", "--id"},
		"scoop":  {"install"},
		"choco":  {"install", "-y"},
	}
	for _, p := range pkgs {
		args, managed := managers[p.Kind]
		if !managed {
			continue
		}
		if _, err := exec.LookPath(p.Kind); err != nil {
			fmt.Printf("skipping %s %q: %s not installed\n", p.Kind, p.ID, p.Kind)
			continue
		}
		fmt.Printf("%s: installing %s\n", p.Kind, p.ID)
		argv := append(append([]string{}, args...), p.ID)
		var cmd *exec.Cmd
		if p.Kind == "scoop" {
			// scoop is a .cmd shim; exec it through cmd.
			cmd = exec.Command("cmd", append([]string{"/c", "scoop"}, argv...)...)
		} else {
			cmd = exec.Command(p.Kind, argv...)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("  %s %s failed: %v\n", p.Kind, p.ID, err)
		}
	}
	return nil
}

func mustParse(path string) []Package {
	pkgs, err := GetBrewfilePackages(path)
	if err != nil {
		return nil
	}
	return pkgs
}
