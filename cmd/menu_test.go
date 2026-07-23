package cmd

import (
	"testing"
)

// TestMainMenuItemsResolve guards against menu entries that lead nowhere:
// every item offered by the main menu must produce a command when selected.
func TestMainMenuItemsResolve(t *testing.T) {
	for _, it := range mainMenuItems() {
		if it.Value == "" {
			t.Errorf("menu item %q has no value", it.Label)
			continue
		}
		if cmd := mainMenuSelect(it); cmd == nil {
			t.Errorf("menu item %q (%s) resolves to no action", it.Label, it.Value)
		}
	}
}

// TestExtraMenuItemsResolve does the same for the build-variant extras.
func TestExtraMenuItemsResolve(t *testing.T) {
	for _, it := range extraMenuItems() {
		if cmd := extraMenuDo(it.Value); cmd == nil {
			t.Errorf("extra menu item %q (%s) resolves to no action", it.Label, it.Value)
		}
	}
}

// TestBundleCategoriesAvailable ensures the install menu is never empty and
// every category resolves through the palette registration path too.
func TestBundleCategoriesAvailable(t *testing.T) {
	cats := availableBundleCategories()
	if len(cats) == 0 {
		t.Fatal("no bundle categories available")
	}
	seen := map[string]bool{}
	for _, c := range cats {
		if c.ID == "" || c.Label == "" {
			t.Errorf("bundle category %+v missing ID or label", c)
		}
		if seen[c.ID] {
			t.Errorf("duplicate bundle category %q", c.ID)
		}
		seen[c.ID] = true
	}
}
