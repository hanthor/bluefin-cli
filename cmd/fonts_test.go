package cmd

import (
	"strings"
	"testing"
)

// TestAvailableFonts_DataIntegrity guards the curated font list against
// accidental regressions: every entry must carry a display name and a brew
// cask, and the list must not contain duplicates (brew would otherwise fail
// or double-install casks).
func TestAvailableFonts_DataIntegrity(t *testing.T) {
	if len(availableFonts) == 0 {
		t.Fatal("availableFonts must not be empty")
	}
	seenCasks := map[string]bool{}
	seenNames := map[string]bool{}
	for _, f := range availableFonts {
		if strings.TrimSpace(f.Name) == "" {
			t.Errorf("font entry has empty display name: %+v", f)
		}
		if strings.TrimSpace(f.Cask) == "" {
			t.Errorf("font %q has empty cask", f.Name)
		}
		if !strings.HasPrefix(f.Cask, "font-") {
			t.Errorf("font %q cask %q should start with the brew font- prefix", f.Name, f.Cask)
		}
		if seenCasks[f.Cask] {
			t.Errorf("duplicate cask %q", f.Cask)
		}
		if seenNames[f.Name] {
			t.Errorf("duplicate display name %q", f.Name)
		}
		seenCasks[f.Cask] = true
		seenNames[f.Name] = true
	}
}

// TestFontsCommand_Registered ensures the fonts command is wired into the
// root command tree so `bluefin-cli fonts` is discoverable.
func TestFontsCommand_Registered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"fonts"})
	if err != nil || cmd == rootCmd || cmd == nil {
		t.Fatal("fonts command not found under root")
	}
	if cmd.Use != "fonts" {
		t.Errorf("fonts command Use = %q, want %q", cmd.Use, "fonts")
	}
	if cmd.Short == "" {
		t.Error("fonts command should have a Short description")
	}
	if cmd.RunE == nil {
		t.Error("fonts command should have a RunE (menu flow entry)")
	}
}
