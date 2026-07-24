package profile

import (
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {
	current := &Profile{
		Version:       1,
		Flavor:        "auto",
		Tools:         map[string]bool{"eza": true, "bat": false},
		EnabledShells: []string{"bash", "zsh"},
	}
	want := &Profile{
		Version:       1,
		Flavor:        "mocha",
		Tools:         map[string]bool{"eza": true, "bat": true},
		EnabledShells: []string{"bash", "fish"},
	}

	diff := Diff(current, want)
	joined := strings.Join(diff, "\n")
	for _, expect := range []string{
		"flavor: auto -> mocha",
		"tool bat: off -> on",
		"shell fish: disabled -> enabled",
		"shell zsh: enabled -> disabled",
	} {
		if !strings.Contains(joined, expect) {
			t.Errorf("diff missing %q:\n%s", expect, joined)
		}
	}
	if strings.Contains(joined, "eza") || strings.Contains(joined, "bash") {
		t.Errorf("diff contains unchanged entries:\n%s", joined)
	}

	if d := Diff(want, want); len(d) != 0 {
		t.Errorf("self-diff should be empty, got %v", d)
	}
}
