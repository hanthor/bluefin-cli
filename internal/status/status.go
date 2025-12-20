package status

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/hanthor/bluefin-cli/internal/motd"
)

var (
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Underline(true)
	enabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
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
	for _, shell := range []string{"bash", "zsh", "fish"} {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"
		
		if shellStatus[shell] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		leftCol += fmt.Sprintf("  %s %s: %s\n", 
			style.Render(symbol),
			shell,
			style.Render(status))
	}
	leftCol += "\n"

	// MOTD status
	leftCol += labelStyle.Render("Message of the Day:") + "\n"
	motdStatus := motd.CheckStatus()
	for _, shell := range []string{"bash", "zsh", "fish"} {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"
		
		if motdStatus[shell] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		leftCol += fmt.Sprintf("  %s %s: %s\n", 
			style.Render(symbol),
			shell,
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
