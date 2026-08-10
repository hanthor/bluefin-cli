package terminal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceKeyLine(t *testing.T) {
	tests := []struct {
		name    string
		content string
		key     string
		newLine string
		want    string
	}{
		{
			name:    "replaces single occurrence",
			content: "font_family Fira Code\nother stuff\n",
			key:     "font_family",
			newLine: "font_family JetBrainsMono Nerd Font",
			want:    "font_family JetBrainsMono Nerd Font\nother stuff\n",
		},
		{
			name:    "preserves indentation",
			content: "  font_family Fira Code\nother stuff\n",
			key:     "font_family",
			newLine: "font_family JetBrainsMono Nerd Font",
			want:    "  font_family JetBrainsMono Nerd Font\nother stuff\n",
		},
		{
			name:    "no match returns original",
			content: "font_size 12\nother stuff\n",
			key:     "font_family",
			newLine: "font_family JetBrainsMono Nerd Font",
			want:    "font_size 12\nother stuff\n",
		},
		{
			name:    "replaces multiple occurrences",
			content: "font_family A\nfont_family B\n",
			key:     "font_family",
			newLine: "font_family X",
			want:    "font_family X\nfont_family X\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceKeyLine(tt.content, tt.key, tt.newLine)
			if got != tt.want {
				t.Errorf("replaceKeyLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReplaceOrInsertTOMLKey(t *testing.T) {
	tests := []struct {
		name    string
		content string
		section string
		key     string
		line    string
		want    string
	}{
		{
			name:    "replaces key under section",
			content: "[font]\nfamily = \"A\"\n",
			section: "[font]",
			key:     "family",
			line:    `family = "B"`,
			want:    "[font]\nfamily = \"B\"\n",
		},
		{
			name:    "inserts key when missing under section",
			content: "[font]\nsize = 12\n",
			section: "[font]",
			key:     "family",
			line:    `family = "B"`,
			want:    "[font]\nfamily = \"B\"\nsize = 12\n",
		},
		{
			name:    "inserts before next section",
			content: "[font]\nsize = 12\n\n[colors]\nbg = \"#000\"\n",
			section: "[font]",
			key:     "family",
			line:    `family = "B"`,
			want:    "[font]\nfamily = \"B\"\nsize = 12\n\n[colors]\nbg = \"#000\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceOrInsertTOMLKey(tt.content, tt.section, tt.key, tt.line)
			if got != tt.want {
				t.Errorf("replaceOrInsertTOMLKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdateAlacrittyFont(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "alacritty.toml")
	initial := "[font]\nfamily = \"OldFont\"\n"
	if err := os.WriteFile(cfgPath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := updateAlacrittyFont(dir, "JetBrainsMono Nerd Font"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `family = "JetBrainsMono Nerd Font"`) {
		t.Errorf("expected font family update, got: %s", got)
	}
}

func TestUpdateKittyFont(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "kitty.conf")
	initial := "font_family Fira Code\n"
	if err := os.WriteFile(cfgPath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := updateKittyFont(dir, "JetBrainsMono Nerd Font"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "font_family JetBrainsMono Nerd Font") {
		t.Errorf("expected font family update, got: %s", got)
	}
}

func TestUpdateWezTermFont(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".wezterm.lua")
	initial := "local wezterm = require 'wezterm'\nlocal config = wezterm.config_builder()\nconfig.font = wezterm.font 'Fira Code Nerd Font'\nreturn config\n"
	if err := os.WriteFile(cfgPath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := updateWezTermFont(dir, "JetBrainsMono Nerd Font"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "config.font = wezterm.font 'JetBrainsMono Nerd Font'") {
		t.Errorf("expected font update, got: %s", got)
	}
}

func TestUpdateGhosttyFont(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// os.UserHomeDir() reads USERPROFILE on Windows and HOME on unix; set both
	// so ghosttyConfigDir resolves to the temp dir on every platform.
	t.Setenv("USERPROFILE", dir)
	// ghosttyConfigDir uses os.UserHomeDir(); override by creating the config dir structure
	ghosttyDir := filepath.Join(dir, ".config", "ghostty")
	if err := os.MkdirAll(ghosttyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initial := "some-setting = value\n"
	if err := os.WriteFile(filepath.Join(ghosttyDir, "config"), []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := updateGhosttyFont("JetBrainsMono Nerd Font"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(ghosttyDir, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "font-family = JetBrainsMono Nerd Font") {
		t.Errorf("expected font-family in config, got: %s", got)
	}
	if !strings.Contains(string(got), markerStart) {
		t.Errorf("expected managed block start marker, got: %s", got)
	}
}
