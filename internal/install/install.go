package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
)

const baseURL = "https://raw.githubusercontent.com/ublue-os/bluefin/refs/heads/main/brew"

var bundles = map[string]struct {
	File        string
	Description string
}{
	"ai": {
		File:        "bluefin-ai.Brewfile",
		Description: "AI tools: Goose, Codex, Gemini, Ramalama, etc.",
	},
	"cli": {
		File:        "bluefin-cli.Brewfile",
		Description: "CLI essentials: GitHub CLI, chezmoi, etc.",
	},
	"fonts": {
		File:        "bluefin-fonts.Brewfile",
		Description: "Development fonts: Fira Code, JetBrains Mono, etc.",
	},
	"k8s": {
		File:        "bluefin-k8s.Brewfile",
		Description: "Kubernetes tools: kubectl, k9s, kind, etc.",
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
			return fmt.Errorf("unknown bundle: %s (available: ai, cli, fonts, k8s, all)", nameOrPath)
		}

		// Download the Brewfile
		url := fmt.Sprintf("%s/%s", baseURL, bundle.File)
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
