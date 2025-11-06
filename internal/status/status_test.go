package status

import (
	"testing"
)

func TestShow(t *testing.T) {
	// This test mainly verifies the function doesn't panic
	// Actual output verification would require capturing stdout
	if err := Show(); err != nil {
		t.Errorf("Show() returned error: %v", err)
	}
}

func TestShowComponents(t *testing.T) {
	// Test that Show() can gather all necessary information
	// without errors (even if tools aren't installed)
	
	// This is more of a smoke test to ensure no panics occur
	err := Show()
	if err != nil {
		t.Fatalf("Show() failed: %v", err)
	}
}
