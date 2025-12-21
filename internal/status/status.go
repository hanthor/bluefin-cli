package status

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hanthor/bluefin-cli/internal/motd"
	"github.com/hanthor/bluefin-cli/internal/shell"
)

var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Underline(true)
	enabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

// Show displays the current configuration status
func Show() error {
	fmt.Println(titleStyle.Render("Bluefin CLI Status"))
	fmt.Println()

	// --- Left Column ---
	var leftCol string

	// Shell status
	leftCol += labelStyle.Render("Shell Experience:") + "\n"
	shellStatus := shell.CheckStatus()
	installedShells := shell.GetInstalledShells()

	// Get default shell
	defaultShellPath := os.Getenv("SHELL")
	defaultShell := filepath.Base(defaultShellPath)

	// Get current shell (heuristic using parent process)
	// We use `ps -p $PPID -o comm=` to get the command name of the parent process
	var currentShell string
	if ppid := os.Getppid(); ppid > 0 {
		cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=")
		if out, err := cmd.Output(); err == nil {
			comm := strings.TrimSpace(string(out))
			// Handle e.g. /bin/zsh or -zsh
			comm = strings.TrimPrefix(comm, "-")
			currentShell = filepath.Base(comm)
		}
	}

	if len(installedShells) == 0 {
		leftCol += "  (no compatible shells found)\n"
	}

	for _, s := range installedShells {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"

		if shellStatus[s] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		markers := ""
		isDefault := s == defaultShell
		isCurrent := s == currentShell

		if isDefault && isCurrent {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" ★ (default, current)")
		} else if isDefault {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" ★ (default)")
		} else if isCurrent {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(" ● (current)")
		}

		leftCol += fmt.Sprintf("  %s %s: %s%s\n",
			style.Render(symbol),
			s,
			style.Render(status),
			markers)
	}
	leftCol += "\n"

	// MOTD status
	leftCol += labelStyle.Render("Message of the Day:") + "\n"
	motdStatus := motd.CheckStatus()
	for _, s := range installedShells {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"

		if motdStatus[s] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		leftCol += fmt.Sprintf("  %s %s: %s\n",
			style.Render(symbol),
			s,
			style.Render(status))
	}

	// --- Right Column ---
	var rightCol string

	// Tool dependencies
	rightCol += labelStyle.Render("Managed Tools:") + "\n"
	deps := shell.CheckDependencies()

	for _, tool := range shell.Tools {
		status := "not installed"
		style := disabledStyle
		symbol := "✗"

		if deps[tool.Binary] {
			status = "installed"
			style = enabledStyle
			symbol = "✓"
		}

		rightCol += fmt.Sprintf("  %s %s: %s\n",
			style.Render(symbol),
			tool.Name,
			style.Render(status))
	}
	rightCol += "\n"

	// Homebrew status
	rightCol += labelStyle.Render("Package Manager:") + "\n"
	if _, err := exec.LookPath("brew"); err == nil {
		rightCol += fmt.Sprintf("  %s Homebrew: %s\n",
			enabledStyle.Render("✓"),
			enabledStyle.Render("installed"))

		// Try to get Homebrew version
		if output, err := exec.Command("brew", "--version").Output(); err == nil {
			version := string(output)
			if len(version) > 0 {
				rightCol += fmt.Sprintf("    %s\n", version[:len(version)-1])
			}
		}
	} else {
		rightCol += fmt.Sprintf("  %s Homebrew: %s\n",
			disabledStyle.Render("✗"),
			disabledStyle.Render("not installed"))
		rightCol += "    Install from: https://brew.sh\n"
	}

	// Combine columns with padding
	formatted := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(40).Render(leftCol),
		string(rightCol),
	)

	fmt.Println(formatted)

	return nil
}
