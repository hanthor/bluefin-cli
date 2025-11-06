package status

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/bluefin-cli/internal/bling"
	"github.com/yourusername/bluefin-cli/internal/motd"
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

	// Bling status
	fmt.Println(labelStyle.Render("Shell Bling:"))
	blingStatus := bling.CheckStatus()
	for _, shell := range []string{"bash", "zsh", "fish"} {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"
		
		if blingStatus[shell] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		fmt.Printf("  %s %s: %s\n", 
			style.Render(symbol),
			shell,
			style.Render(status))
	}
	fmt.Println()

	// MOTD status
	fmt.Println(labelStyle.Render("Message of the Day:"))
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

		fmt.Printf("  %s %s: %s\n", 
			style.Render(symbol),
			shell,
			style.Render(status))
	}
	fmt.Println()

	// Tool dependencies
	fmt.Println(labelStyle.Render("Required Tools:"))
	deps := bling.CheckDependencies()
	tools := []string{"eza", "bat", "zoxide", "atuin", "starship", "ugrep"}
	
	for _, tool := range tools {
		status := "not installed"
		style := disabledStyle
		symbol := "✗"
		
		if deps[tool] {
			status = "installed"
			style = enabledStyle
			symbol = "✓"
		}

		fmt.Printf("  %s %s: %s\n", 
			style.Render(symbol),
			tool,
			style.Render(status))
	}
	fmt.Println()

	// Additional tools
	fmt.Println(labelStyle.Render("Optional Tools:"))
	optionalTools := []string{"glow", "fastfetch", "gh", "jq", "fzf"}
	for _, tool := range optionalTools {
		_, err := exec.LookPath(tool)
		status := "not installed"
		style := disabledStyle
		symbol := "✗"
		
		if err == nil {
			status = "installed"
			style = enabledStyle
			symbol = "✓"
		}

		fmt.Printf("  %s %s: %s\n", 
			style.Render(symbol),
			tool,
			style.Render(status))
	}
	fmt.Println()

	// Homebrew status
	fmt.Println(labelStyle.Render("Package Manager:"))
	if _, err := exec.LookPath("brew"); err == nil {
		fmt.Printf("  %s Homebrew: %s\n", 
			enabledStyle.Render("✓"),
			enabledStyle.Render("installed"))
		
		// Try to get Homebrew version
		if output, err := exec.Command("brew", "--version").Output(); err == nil {
			version := string(output)
			if len(version) > 0 {
				fmt.Printf("    %s\n", version[:len(version)-1]) // Remove trailing newline
			}
		}
	} else {
		fmt.Printf("  %s Homebrew: %s\n", 
			disabledStyle.Render("✗"),
			disabledStyle.Render("not installed"))
		fmt.Println("    Install from: https://brew.sh")
	}

	return nil
}
