package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/wallpaper"
)

var wallpaperCmd = &cobra.Command{
	Use:   "wallpaper [image]",
	Short: "Set the desktop wallpaper (macOS, GNOME)",
	Long: `Set the desktop wallpaper natively — macOS (via the 'wallpaper' brew
formula or AppleScript) and GNOME (via gsettings). With no argument, shows
the current wallpaper.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !wallpaper.Supported() {
			return fmt.Errorf("wallpaper setting is not supported on this platform")
		}
		if len(args) == 0 {
			cur, err := wallpaper.Get()
			if err != nil {
				return err
			}
			fmt.Println(strings.TrimSpace(cur))
			return nil
		}
		if err := wallpaper.Set(args[0]); err != nil {
			return err
		}
		fmt.Println(tui.SuccessStyle.Render("✓ Wallpaper set."))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(wallpaperCmd)
}
