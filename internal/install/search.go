package install

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// SearchPackages queries the package managers available on this platform
// (brew formulae+casks on unix; winget/scoop/choco on Windows) and returns
// a unified result list. Managers run in parallel with a shared timeout;
// an unavailable manager is simply skipped.
func SearchPackages(query string) []Package {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	type source struct {
		bin   string
		args  []string
		parse func(string) []Package
	}
	var sources []source
	if runtime.GOOS == "windows" {
		sources = []source{
			{"winget", []string{"search", "--accept-source-agreements", query}, parseWingetSearch},
			{"cmd", []string{"/c", "scoop", "search", query}, parseScoopSearch},
			{"choco", []string{"search", query, "--limit-output"}, parseChocoSearch},
		}
	} else {
		sources = []source{
			{"brew", []string{"search", "--formula", query}, func(s string) []Package { return parseBrewSearch(s, "brew") }},
			{"brew", []string{"search", "--cask", query}, func(s string) []Package { return parseBrewSearch(s, "cask") }},
		}
	}

	var mu sync.Mutex
	var out []Package
	var wg sync.WaitGroup
	for _, src := range sources {
		probe := src.bin
		if probe == "cmd" && len(src.args) > 1 {
			probe = src.args[1] // the shim behind cmd /c
		}
		if _, err := exec.LookPath(probe); err != nil {
			continue
		}
		wg.Add(1)
		go func(s source) {
			defer wg.Done()
			raw, err := exec.CommandContext(ctx, s.bin, s.args...).Output()
			if err != nil {
				return
			}
			pkgs := s.parse(string(raw))
			mu.Lock()
			out = append(out, pkgs...)
			mu.Unlock()
		}(src)
	}
	wg.Wait()
	return out
}

// parseBrewSearch handles `brew search --formula/--cask` output: one name
// per line, with optional ==> section headers and blank lines.
func parseBrewSearch(raw, kind string) []Package {
	var out []Package
	for _, line := range strings.Split(raw, "\n") {
		name := strings.TrimSpace(line)
		if name == "" || strings.HasPrefix(name, "==>") || strings.Contains(name, "Error:") {
			continue
		}
		out = append(out, Package{Name: name, ID: name, Kind: kind})
	}
	return out
}

// parseWingetSearch handles winget's fixed-width table using the header's
// column offsets (names can contain spaces, so whitespace splitting fails).
func parseWingetSearch(raw string) []Package {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	idStart, verStart := -1, -1
	var out []Package
	for _, line := range lines {
		if idStart < 0 {
			if idx := strings.Index(line, "Id"); idx > 0 && strings.HasPrefix(strings.TrimSpace(line), "Name") {
				idStart = idx
				if v := strings.Index(line, "Version"); v > idStart {
					verStart = v
				}
			}
			continue
		}
		if strings.HasPrefix(line, "---") || len(line) <= idStart {
			continue
		}
		name := strings.TrimSpace(line[:idStart])
		rest := line[idStart:]
		id := strings.TrimSpace(rest)
		if verStart > idStart && len(line) > verStart {
			id = strings.TrimSpace(line[idStart:verStart])
		}
		if id == "" || strings.Contains(id, " ") {
			continue
		}
		out = append(out, Package{Name: name, ID: id, Kind: "winget"})
	}
	return out
}

// parseScoopSearch handles scoop's result table (Name Version Source ...).
func parseScoopSearch(raw string) []Package {
	var out []Package
	headerSeen := false
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		fields := strings.Fields(t)
		if !headerSeen {
			if fields[0] == "Name" {
				headerSeen = true
			}
			continue
		}
		if strings.HasPrefix(fields[0], "---") {
			continue
		}
		out = append(out, Package{Name: fields[0], ID: fields[0], Kind: "scoop"})
	}
	return out
}

// parseChocoSearch handles `choco search --limit-output`: name|version.
func parseChocoSearch(raw string) []Package {
	var out []Package
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 2)
		if len(parts) != 2 || parts[0] == "" {
			continue
		}
		out = append(out, Package{Name: parts[0], ID: parts[0], Kind: "choco"})
	}
	return out
}
