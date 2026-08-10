//go:build !windows

package install

import "testing"

func TestAlpinePackageName(t *testing.T) {
	tests := map[string]string{
		"gh":                  "github-cli",
		"kubernetes-cli":      "kubectl",
		"dapr/tap/dapr-cli":   "dapr",
		"buildpacks/tap/pack": "pack-cli",
		"artifacthub/cmd/ah":  "ah",
		"tap/formula":         "formula",
		"ripgrep":             "ripgrep",
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			if got := alpinePackageName(input); got != want {
				t.Fatalf("alpinePackageName(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func TestAlpineInstallerSkipsUnsupportedKinds(t *testing.T) {
	for _, test := range []struct {
		pkg  Package
		want bool
	}{
		{Package{ID: "codex", Kind: "cask"}, false},
		{Package{ID: "org.gnome.Calculator", Kind: "flatpak"}, false},
		{Package{ID: "kubectl", Kind: "brew"}, true},
	} {
		if got := isAlpinePackage(test.pkg); got != test.want {
			t.Errorf("isAlpinePackage(%+v) = %v, want %v", test.pkg, got, test.want)
		}
	}
}
