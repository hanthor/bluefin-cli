//go:build !windows

package install

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AlpineInstaller installs the formula portion of Brewfile bundles using the
// Alpine package universe. Casks and flatpaks have no general Alpine
// equivalent and are reported as unavailable rather than sent to apk.
type AlpineInstaller struct{}

func (i *AlpineInstaller) InstallBundle(names ...string) error {
	var packages []Package
	for _, name := range names {
		path, cleanup, err := GetBrewfile(name)
		if err != nil {
			return err
		}
		defer cleanup()
		bundlePackages, err := GetBrewfilePackages(path)
		if err != nil {
			return err
		}
		packages = append(packages, bundlePackages...)
	}
	return InstallAlpinePackages(packages)
}

func (i *AlpineInstaller) InstallWallpapers([]string) error {
	return fmt.Errorf("wallpaper casks are not available on Alpine/postmarketOS")
}

func (i *AlpineInstaller) CleanupWallpapers(bool) error { return nil }

// InstallAlpinePackages installs each supported package independently so one
// package absent from Alpine does not prevent the rest of a bundle installing.
func InstallAlpinePackages(packages []Package) error {
	manager, err := alpinePackageManager()
	if err != nil {
		return err
	}

	installed, skipped, failed := 0, 0, 0
	for _, pkg := range packages {
		if !isAlpinePackage(pkg) {
			skipped++
			continue
		}
		name := alpinePackageName(pkg.ID)
		fmt.Println(infoStyle.Render(fmt.Sprintf("⬇️  Installing %s via %s...", name, manager)))
		if err := runAlpineInstall(manager, name); err != nil {
			failed++
			fmt.Println(errorStyle.Render(fmt.Sprintf("  unavailable or failed: %s (%v)", name, err)))
			continue
		}
		installed++
	}

	if skipped > 0 || failed > 0 {
		fmt.Println(infoStyle.Render(fmt.Sprintf("Alpine bundle result: %d installed, %d unavailable, %d failed.", installed, skipped, failed)))
	}
	if installed == 0 && (skipped > 0 || failed > 0) {
		return fmt.Errorf("no supported packages from bundle were installed")
	}
	return nil
}

func UninstallAlpinePackages(packages []Package) error {
	manager, err := alpinePackageManager()
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		if !isAlpinePackage(pkg) {
			continue
		}
		name := alpinePackageName(pkg.ID)
		var cmd *exec.Cmd
		if manager == "coldbrew" {
			cmd = exec.Command(manager, "uninstall", name)
		} else {
			cmd = exec.Command("sudo", "apk", "del", name)
		}
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Failed to uninstall %s: %v", name, err)))
		}
	}
	return nil
}

func alpinePackageManager() (string, error) {
	if _, err := exec.LookPath("coldbrew"); err == nil {
		return "coldbrew", nil
	}
	if _, err := exec.LookPath("apk"); err != nil {
		return "", fmt.Errorf("neither coldbrew nor apk was found; install one before installing Alpine bundles")
	}
	return "apk", nil
}

func runAlpineInstall(manager, name string) error {
	var cmd *exec.Cmd
	if manager == "coldbrew" {
		cmd = exec.Command(manager, "install", name)
	} else {
		cmd = exec.Command("sudo", "apk", "add", name)
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if manager == "coldbrew" {
		wrap := exec.Command(manager, "wrap", name)
		wrap.Stdout, wrap.Stderr = os.Stdout, os.Stderr
		return wrap.Run()
	}
	return nil
}

var alpinePackageAliases = map[string]string{
	"gh":                  "github-cli",
	"kubernetes-cli":      "kubectl",
	"dapr/tap/dapr-cli":   "dapr",
	"buildpacks/tap/pack": "pack-cli",
}

func alpinePackageName(id string) string {
	if name, ok := alpinePackageAliases[id]; ok {
		return name
	}
	if idx := strings.LastIndex(id, "/"); idx >= 0 {
		id = id[idx+1:]
	}
	return id
}

func isAlpinePackage(pkg Package) bool {
	return pkg.Kind == "brew"
}
