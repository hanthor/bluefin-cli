package cmd

import (
	"bytes"
	"testing"
	"text/template"
)

// TestRootCommand_FallbackHelp covers the root RunE fallback used when the
// interactive menu is unavailable: it must degrade to help output, not panic.
func TestRootCommand_FallbackHelp(t *testing.T) {
	oldMenu := menuCmd
	menuCmd = nil
	t.Cleanup(func() { menuCmd = oldMenu })

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	if err := rootCmd.RunE(rootCmd, nil); err != nil {
		t.Fatalf("root RunE fallback failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected help output from root fallback")
	}
}

// TestRootCommand_VersionTemplate ensures the version string is rendered with
// the project template (guards the fang --version surface).
func TestRootCommand_VersionTemplate(t *testing.T) {
	rootCmd.SetVersionTemplate("bluefin-cli version {{.Version}}\n")
	t.Cleanup(func() { rootCmd.SetVersionTemplate("") })
	buf := new(bytes.Buffer)
	tpl, err := template.New("version").Parse(rootCmd.VersionTemplate())
	if err != nil {
		t.Fatalf("version template parse failed: %v", err)
	}
	if err := tpl.Execute(buf, rootCmd); err != nil {
		t.Fatalf("version template render failed: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("bluefin-cli version")) {
		t.Errorf("version output = %q, want template header", buf.String())
	}
}
