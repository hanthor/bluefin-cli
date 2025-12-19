package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
)

const (
	commonBaseURL   = "https://raw.githubusercontent.com/projectbluefin/common/main/system_files"
	defaultBrewPath = "shared/usr/share/ublue-os/homebrew"
)

type BundleSpec struct {
	File        string
	Description string
	Path        string // Optional: override defaultBrewPath
}

var bundles = map[string]BundleSpec{
	"ai": {
		File:        "ai-tools.Brewfile",
		Description: "AI tools: Goose, Codex, Gemini, Ramalama, etc.",
	},
	"artwork": {
		File:        "artwork.Brewfile",
		Description: "Artwork and design tools.",
	},
	"cli": {
		File:        "cli.Brewfile",
		Description: "CLI essentials: GitHub CLI, chezmoi, etc.",
	},
	"cncf": {
		File:        "cncf.Brewfile",
		Description: "Cloud Native Computing Foundation tools.",
	},
	"experimental-ide": {
		File:        "experimental-ide.Brewfile",
		Description: "Experimental IDE tools.",
	},
	"fonts": {
		File:        "fonts.Brewfile",
		Description: "Development fonts: Fira Code, JetBrains Mono, etc.",
	},
	"full-desktop": {
		File:        "full-desktop.Brewfile",
		Description: "Full GNOME Desktop apps.",
		Path:        "bluefin/usr/share/ublue-os/homebrew",
	},
	"ide": {
		File:        "ide.Brewfile",
		Description: "IDE tools: VS Code, JetBrains Toolbox, etc.",
	},
	"k8s": {
		File:        "k8s-tools.Brewfile",
		Description: "Kubernetes tools: kubectl, k9s, kubectx, etc.",
	},
}

// Bundle installs a Homebrew bundle
func Bundle(nameOrPath string) error {
	// Check if brew is installed
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew not found. Please install Homebrew first: https://brew.sh")
	}

	var brewfilePath string

	// Special case: install all bundles
	if nameOrPath == "all" {
		fmt.Println(titleStyle.Render("📦 Installing all bundles..."))
		for name := range bundles {
			fmt.Println(infoStyle.Render(fmt.Sprintf("\n Installing bundle: %s", name)))
			if err := Bundle(name); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to install %s: %v", name, err)))
			}
		}
		return nil
	}

	// Check if it's a file path
	if strings.Contains(nameOrPath, "/") || strings.Contains(nameOrPath, "\\") {
		// It's a path
		if _, err := os.Stat(nameOrPath); os.IsNotExist(err) {
			return fmt.Errorf("Brewfile not found: %s", nameOrPath)
		}
		brewfilePath = nameOrPath
	} else {
		// It's a bundle name
		bundle, ok := bundles[nameOrPath]
		if !ok {
			return fmt.Errorf("unknown bundle: %s (available: ai, artwork, cli, cncf, experimental-ide, fonts, full-desktop, ide, k8s, all)", nameOrPath)
		}

		// Ensure Flathub if using full-desktop
		if nameOrPath == "full-desktop" {
			if err := EnsureFlathub(); err != nil {
				return err
			}
		}

		// determine path
		path := defaultBrewPath
		if bundle.Path != "" {
			path = bundle.Path
		}

		// Download the Brewfile
		url := fmt.Sprintf("%s/%s/%s", commonBaseURL, path, bundle.File)
		tmpDir := os.TempDir()
		brewfilePath = filepath.Join(tmpDir, bundle.File)

		fmt.Println(infoStyle.Render(fmt.Sprintf("⬇️  Downloading %s bundle...", nameOrPath)))

		if err := downloadFile(url, brewfilePath); err != nil {
			return fmt.Errorf("failed to download bundle: %w", err)
		}
		defer os.Remove(brewfilePath) // Clean up after installation
	}

	// Install the bundle
	fmt.Println(infoStyle.Render(fmt.Sprintf("📦 Installing packages from: %s", brewfilePath)))

	cmd := exec.Command("brew", "bundle", "install", "--file", brewfilePath)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_ENV_HINTS=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew bundle failed: %w", err)
	}

	fmt.Println(successStyle.Render("✓ Bundle installed successfully!"))
	return nil
}

// ListBundles displays all available bundles
func ListBundles() {
	fmt.Println(titleStyle.Render("📦 Available Homebrew Bundles"))
	fmt.Println()

	for name, bundle := range bundles {
		fmt.Printf("  %s %s\n", 
			lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render(name+":"),
			bundle.Description)
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Usage:"))
	fmt.Println("  bluefin-cli install <bundle-name>")
	fmt.Println("  bluefin-cli install /path/to/Brewfile")
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// IsLinux checks if the OS is Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsGnome checks if the current desktop environment is GNOME
func IsGnome() bool {
	xdgCurrentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
	return strings.Contains(strings.ToUpper(xdgCurrentDesktop), "GNOME")
}

// CheckFlatpak checks if flatpak is installed
func CheckFlatpak() error {
	_, err := exec.LookPath("flatpak")
	return err
}

// EnsureFlathub ensures Flathub remote is added if flatpak is available
func EnsureFlathub() error {
	if err := CheckFlatpak(); err != nil {
		return fmt.Errorf("flatpak not found. Please install flatpak first: https://flatpak.org/setup/")
	}

	// Check if flathub exists
	cmd := exec.Command("flatpak", "remote-list")
	out, err := cmd.Output()
	if err == nil && strings.Contains(string(out), "flathub") {
		return nil
	}

	fmt.Println(infoStyle.Render("Adding Flathub remote..."))
	addCmd := exec.Command("flatpak", "remote-add", "--if-not-exists", "flathub", "https://dl.flathub.org/repo/flathub.flatpakrepo")
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	return addCmd.Run()
}
