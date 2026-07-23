package cmd

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
	"github.com/tuna-os/bluefin-cli/internal/update"

	"charm.land/lipgloss/v2"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose common problems with your Bluefin CLI setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctor()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	ok   bool
	warn bool
	name string
	note string // failure detail or fix hint
}

func runDoctor() error {
	t := theme.DefaultTheme
	pass := lipgloss.NewStyle().Foreground(t.Success).Render("✓")
	warn := lipgloss.NewStyle().Foreground(t.Warning).Render("!")
	fail := lipgloss.NewStyle().Foreground(t.Error).Render("✗")
	dim := lipgloss.NewStyle().Foreground(t.TextFaint)

	checks := []checkResult{
		checkBrew(),
		checkShellIntegration(),
		checkTools(),
		checkNetwork(),
		checkVersion(),
	}

	fmt.Println()
	failures := 0
	for _, c := range checks {
		mark := pass
		switch {
		case !c.ok && c.warn:
			mark = warn
		case !c.ok:
			mark = fail
			failures++
		}
		fmt.Printf("  %s %s\n", mark, c.name)
		if c.note != "" {
			fmt.Printf("    %s\n", dim.Render(c.note))
		}
	}
	fmt.Println()
	if failures > 0 {
		return fmt.Errorf("%d check(s) failed", failures)
	}
	fmt.Println(dim.Render("  All checks passed."))
	return nil
}

func checkBrew() checkResult {
	if _, err := exec.LookPath("brew"); err != nil {
		return checkResult{name: "Homebrew on PATH", warn: true,
			note: "brew not found — Install Apps bundles need it: https://brew.sh"}
	}
	return checkResult{ok: true, name: "Homebrew on PATH"}
}

func checkShellIntegration() checkResult {
	current := currentShellName()
	if shell.CheckStatus()[current] {
		return checkResult{ok: true, name: fmt.Sprintf("Shell integration enabled (%s)", current)}
	}
	return checkResult{name: fmt.Sprintf("Shell integration enabled (%s)", current), warn: true,
		note: fmt.Sprintf("run: bluefin-cli shell %s on", current)}
}

func checkTools() checkResult {
	missing := []string{}
	for _, tool := range []string{"eza", "fzf", "starship"} {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}
	if len(missing) == 0 {
		return checkResult{ok: true, name: "Managed tools installed (eza, fzf, starship)"}
	}
	return checkResult{name: "Managed tools installed", warn: true,
		note: fmt.Sprintf("missing: %v — install via the Bluefin Shell menu", missing)}
}

func checkNetwork() checkResult {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head("https://api.github.com")
	if err != nil {
		return checkResult{name: "GitHub reachable", warn: true,
			note: "updates and bundle installs need network access: " + err.Error()}
	}
	_ = resp.Body.Close()
	return checkResult{ok: true, name: "GitHub reachable"}
}

func checkVersion() checkResult {
	if version == "dev" {
		return checkResult{ok: true, name: "Version: dev build"}
	}
	rel, err := update.Latest()
	if err != nil {
		return checkResult{ok: true, name: "Version " + version,
			note: "could not check for updates: " + err.Error()}
	}
	if update.IsNewer(version, rel.TagName) {
		hint := "bluefin-cli update"
		if h := update.Detect().UpdateHint(); h != "" {
			hint = h
		}
		return checkResult{name: "Version " + version, warn: true,
			note: fmt.Sprintf("%s available — run: %s", rel.TagName, hint)}
	}
	return checkResult{ok: true, name: "Version " + version + " (latest)"}
}
