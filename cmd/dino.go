package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// The easter egg deserves a front door too.
var dinoCmd = &cobra.Command{
	Use:    "dino",
	Short:  "🦕",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run(gameScreen())
	},
}

func init() {
	rootCmd.AddCommand(dinoCmd)
}
