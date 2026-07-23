package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/profile"
	"github.com/tuna-os/bluefin-cli/internal/tui"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Export or import your bluefin-cli setup",
	Long: `Capture this machine's setup (enabled shells, tool selection, theme
flavor) as a portable JSON document, and replay it on another machine:

  bluefin-cli profile export > my-setup.json
  bluefin-cli profile import my-setup.json`,
}

var profileExportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Write the current setup as JSON (stdout by default)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := profile.Export(currentShellName())
		if err != nil {
			return err
		}
		path := "-"
		if len(args) > 0 {
			path = args[0]
		}
		if err := p.Save(path); err != nil {
			return err
		}
		if path != "-" {
			fmt.Println(tui.SuccessStyle.Render("✓ Profile exported to " + path))
		}
		return nil
	},
}

var profileImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Apply a previously exported setup to this machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := profile.Load(args[0])
		if err != nil {
			return err
		}
		if err := p.Apply(); err != nil {
			return err
		}
		fmt.Println(tui.SuccessStyle.Render("✓ Profile applied."))
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileExportCmd)
	profileCmd.AddCommand(profileImportCmd)
	rootCmd.AddCommand(profileCmd)
}
