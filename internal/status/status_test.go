package status

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestShow(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore stdout on exit
	defer func() {
		os.Stdout = oldStdout
	}()

	err := Show()
	if err != nil {
		t.Errorf("Show() returned error: %v", err)
	}

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify expected sections
	expectedStrings := []string{
		"Bluefin CLI Status",
		"Shell Bling:",
		"Message of the Day:",
		"Required Tools:",
		"Optional Tools:",
		"Package Manager:",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(output, s) {
			t.Errorf("Expected output to contain %q", s)
		}
	}
}

func TestShowComponents(t *testing.T) {
	// This test mainly verifies that Show runs without panicking
	// We'll trust TestShow to verify the output content
	err := Show()
	if err != nil {
		t.Fatalf("Show() failed: %v", err)
	}
}
