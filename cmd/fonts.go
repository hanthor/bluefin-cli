package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/tui"
)

// availableFonts maps display name → brew cask name
var availableFonts = []struct {
	Name string
	Cask string
}{
	{"0xProto Nerd Font", "font-0xproto-nerd-font"},
	{"Cascadia Mono Nerd Font", "font-caskaydia-mono-nerd-font"},
	{"Comic Shanns Mono Nerd Font", "font-comic-shanns-mono-nerd-font"},
	{"Droid Sans Mono Nerd Font", "font-droid-sans-mono-nerd-font"},
	{"Fira Code Nerd Font", "font-fira-code-nerd-font"},
	{"Go Mono Nerd Font", "font-go-mono-nerd-font"},
	{"IBM Plex Mono Nerd Font", "font-blex-mono-nerd-font"},
	{"JetBrains Mono Nerd Font", "font-jetbrains-mono-nerd-font"},
	{"Source Code Pro", "font-source-code-pro"},
	{"Source Code Pro Nerd Font", "font-sauce-code-pro-nerd-font"},
	{"Ubuntu Nerd Font", "font-ubuntu-nerd-font"},
}

var fontsCmd = &cobra.Command{
	Use:   "fonts",
	Short: "Install individual development fonts",
	Long:  `Select and install individual development fonts from a curated list.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchFlow(fontsFlow())
	},
}

// installFontCasks installs each font cask with brew, printing progress
// (captured by the native RunnerScreen when run from the menu).
func installFontCasks(casks []string) {
	for _, cask := range casks {
		fmt.Println(tui.InfoStyle.Render("Installing " + cask + "..."))
		cmd := exec.Command("brew", "install", "--cask", cask)
		cmd.Stdout = nil
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Println(tui.ErrorStyle.Render("Failed: " + string(out)))
		} else {
			fmt.Println(tui.SuccessStyle.Render("✓ " + cask))
		}
	}
}

func init() {
	rootCmd.AddCommand(fontsCmd)
}
