package test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/shell"
)

// These tests verify DESIRED STATE, not exit codes: after a command runs,
// the files it manages must actually contain what the user was promised —
// and undoing the command must actually undo it.

// TestShellToggleRoundTrip: enabling writes the managed init line, `status`
// agrees, disabling removes it again, and `status` agrees again.
func TestShellToggleRoundTrip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix shells only")
	}
	rc := filepath.Join(os.Getenv("HOME"), ".bashrc")
	if err := os.WriteFile(rc, []byte("# base\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := runCommand(t, "shell", "bash", "on"); err != nil {
		t.Fatalf("enable failed: %v", err)
	}
	if !fileContains(t, rc, "bluefin-cli init") {
		t.Fatal("enable did not write the init line to .bashrc")
	}
	out, err := runCommand(t, "status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(out, "bash: enabled") {
		t.Errorf("status disagrees with on-disk state after enable:\n%s", out)
	}

	if _, err := runCommand(t, "shell", "bash", "off"); err != nil {
		t.Fatalf("disable failed: %v", err)
	}
	if fileContains(t, rc, "bluefin-cli init") {
		t.Error("disable left the init line in .bashrc")
	}
	out, err = runCommand(t, "status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(out, "bash: disabled") {
		t.Errorf("status disagrees with on-disk state after disable:\n%s", out)
	}
	// The original content must survive the round trip.
	if !fileContains(t, rc, "# base") {
		t.Error("round trip destroyed pre-existing .bashrc content")
	}
}

// TestThemeFlavorPersists: `theme <flavor>` must land in config.yaml and
// read back through the CLI.
func TestThemeFlavorPersists(t *testing.T) {
	if _, err := runCommand(t, "theme", "mocha"); err != nil {
		t.Fatalf("theme set failed: %v", err)
	}
	cfg := filepath.Join(os.Getenv("HOME"), ".config", "bluefin-cli", "config.yaml")
	if !fileContains(t, cfg, "flavor: mocha") {
		t.Errorf("config.yaml does not contain the chosen flavor")
	}
	out, err := runCommand(t, "theme")
	if err != nil {
		t.Fatalf("theme read failed: %v", err)
	}
	if !strings.Contains(out, "mocha") {
		t.Errorf("theme readback = %q, want mocha", out)
	}
	// Restore auto and verify that persists too.
	if _, err := runCommand(t, "theme", "auto"); err != nil {
		t.Fatalf("theme reset failed: %v", err)
	}
	if !fileContains(t, cfg, "flavor: auto") {
		t.Error("resetting to auto did not persist")
	}
}

// TestMOTDTogglePersists: `motd toggle <shell> off/on` must be reflected in
// the shell config the init script consults.
func TestMOTDTogglePersists(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix shells only")
	}
	if _, err := runCommand(t, "motd", "toggle", "bash", "off"); err != nil {
		t.Fatalf("motd off failed: %v", err)
	}
	cfg, err := shell.LoadConfig("bash")
	if err != nil {
		t.Fatalf("loading shell config: %v", err)
	}
	if cfg.IsEnabled("Motd") {
		t.Error("motd toggle off did not reach the shell config")
	}

	if _, err := runCommand(t, "motd", "toggle", "bash", "on"); err != nil {
		t.Fatalf("motd on failed: %v", err)
	}
	cfg, err = shell.LoadConfig("bash")
	if err != nil {
		t.Fatalf("loading shell config: %v", err)
	}
	if !cfg.IsEnabled("Motd") {
		t.Error("motd toggle on did not reach the shell config")
	}
}

// TestBundleInstallReachesBrew runs `install <bundle>` hermetically with a
// fake `brew` on PATH that records its argv and the merged Brewfile it was
// pointed at. Bundles resolve from embedded Brewfiles (offline-first), so
// this proves the full chain — bundle resolution, merge, brew invocation —
// delivers real package lines to the package manager.
func TestBundleInstallReachesBrew(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("brew path is unix-only")
	}

	fakeBin := t.TempDir()
	log := filepath.Join(fakeBin, "brew.log")
	stub := fmt.Sprintf(`#!/bin/sh
echo "$@" >> %[1]s
for a in "$@"; do
  case "$a" in --file=*) cat "${a#--file=}" >> %[1]s ;; esac
done
exit 0
`, log)
	if err := os.WriteFile(filepath.Join(fakeBin, "brew"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := commandWithEnv(t, []string{
		"PATH=" + fakeBin + string(os.PathListSeparator) + os.Getenv("PATH"),
	}, "install", "cli")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install cli failed: %v\n%s", err, out)
	}

	logged, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("fake brew was never invoked: %v", err)
	}
	got := string(logged)
	if !strings.Contains(got, "bundle install --file=") {
		t.Errorf("brew was not asked to install a bundle:\n%s", got)
	}
	if !strings.Contains(got, `brew "`) {
		t.Errorf("no package lines reached brew:\n%s", got)
	}
}

// TestProfileRoundTripRestoresState: export must capture real state, and
// import must drive a drifted machine back to it — rc lines, config, theme.
func TestProfileRoundTripRestoresState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix shells only")
	}
	home := os.Getenv("HOME")
	rc := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(rc, []byte("# base\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Establish a distinctive state.
	mustRun := func(args ...string) {
		t.Helper()
		if out, err := runCommand(t, args...); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}
	mustRun("shell", "bash", "on")
	mustRun("theme", "mocha")

	profilePath := filepath.Join(t.TempDir(), "profile.json")
	mustRun("profile", "export", profilePath)
	if !fileContains(t, profilePath, `"bash"`) || !fileContains(t, profilePath, `"flavor": "mocha"`) {
		data, _ := os.ReadFile(profilePath)
		t.Fatalf("export did not capture state:\n%s", data)
	}

	// Drift away from it.
	mustRun("shell", "bash", "off")
	mustRun("theme", "auto")
	if fileContains(t, rc, "bluefin-cli init") {
		t.Fatal("precondition: drift did not disable bash")
	}

	// Import must restore the captured state on disk.
	mustRun("profile", "import", profilePath)
	if !fileContains(t, rc, "bluefin-cli init") {
		t.Error("import did not restore the bash init line")
	}
	cfg := filepath.Join(home, ".config", "bluefin-cli", "config.yaml")
	if !fileContains(t, cfg, "flavor: mocha") {
		t.Error("import did not restore the theme flavor")
	}
	out, err := runCommand(t, "status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(out, "bash: enabled") {
		t.Errorf("status disagrees after import:\n%s", out)
	}

	// Leave the sandbox tidy for other tests.
	mustRun("shell", "bash", "off")
	mustRun("theme", "auto")
}

// TestDoctorFixReachesState: --fix must actually enable shell integration,
// and the subsequent report must agree.
func TestDoctorFixReachesState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix shells only")
	}
	rc := filepath.Join(os.Getenv("HOME"), ".bashrc")
	if err := os.WriteFile(rc, []byte("# base\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// doctor exit code is nonzero when hard failures remain (e.g. no
	// network in a sandbox), so assert on state, not on the exit code.
	cmd := commandWithEnv(t, []string{"SHELL=/bin/bash"}, "doctor", "--fix")
	out, _ := cmd.CombinedOutput()

	if !fileContains(t, rc, "bluefin-cli init") {
		t.Errorf("doctor --fix did not enable shell integration:\n%s", out)
	}
	if !strings.Contains(string(out), "Shell integration enabled (bash)") {
		t.Errorf("doctor report disagrees with the fix it just applied:\n%s", out)
	}

	// Tidy the sandbox.
	if _, err := runCommand(t, "shell", "bash", "off"); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}

// TestCustomBundleInstall: a user Brewfile in the config dir becomes an
// installable bundle whose packages actually reach brew.
func TestCustomBundleInstall(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("brew path is unix-only")
	}
	home := os.Getenv("HOME")
	dir := filepath.Join(home, ".config", "bluefin-cli", "bundles")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mystuff.Brewfile"), []byte("brew \"cowsay\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	fakeBin := t.TempDir()
	log := filepath.Join(fakeBin, "brew.log")
	stub := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %[1]s\nfor a in \"$@\"; do case \"$a\" in --file=*) cat \"${a#--file=}\" >> %[1]s ;; esac; done\nexit 0\n", log)
	if err := os.WriteFile(filepath.Join(fakeBin, "brew"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := commandWithEnv(t, []string{
		"PATH=" + fakeBin + string(os.PathListSeparator) + os.Getenv("PATH"),
	}, "install", "mystuff")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install mystuff failed: %v\n%s", err, out)
	}
	logged, _ := os.ReadFile(log)
	if !strings.Contains(string(logged), `brew "cowsay"`) {
		t.Errorf("custom bundle content never reached brew:\n%s", logged)
	}
}
