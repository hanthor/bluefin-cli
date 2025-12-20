package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/bling"
)

var initCmd = &cobra.Command{
	Use:   "init [bash|zsh|fish]",
	Short: "Generate shell initialization script",
	Long:  `Generate the shell initialization script for bluefin-cli.
Add the following to your shell configuration file:

Bash (~/.bashrc):
  eval "$(bluefin-cli init bash)"

Zsh (~/.zshrc):
  eval "$(bluefin-cli init zsh)"

Fish (~/.config/fish/config.fish):
  bluefin-cli init fish | source
`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		
		// Get Bling script (exports + aliases)
		script, err := bling.Init(shell)
		if err != nil {
			return err
		}
		
		// Print the bling script
		fmt.Println(script)
		fmt.Println()

		// Add MOTD hook
		// We append this here because it's simple enough not to need a separate Init function in motd package
		switch shell {
		case "bash", "zsh":
			// Only run MOTD if interactive
			fmt.Println(`# bluefin-cli motd hook
if [ -n "$PS1" ] && [ -t 1 ]; then
    bluefin-cli motd show
fi`)
		case "fish":
			fmt.Println(`# bluefin-cli motd hook
if status is-interactive
    bluefin-cli motd show
end`)
		default:
			return fmt.Errorf("unsupported shell: %s", shell)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
