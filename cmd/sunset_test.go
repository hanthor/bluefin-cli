//go:build extra

package cmd

import (
	"testing"
)

// TestSunsetCommand_HasSubcommands verifies the extra-build sunset command
// and its setup subcommand are wired into the root command tree.
func TestSunsetCommand_HasSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"sunset"})
	if err != nil || cmd == rootCmd {
		t.Fatal("sunset command not found")
	}
	sub, _, err := cmd.Find([]string{"setup"})
	if err != nil || sub == cmd || sub == nil {
		t.Error("missing sunset subcommand 'setup'")
	}
}
