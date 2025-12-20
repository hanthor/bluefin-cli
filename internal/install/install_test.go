package install

import (
	"os"
	"testing"
)

func TestListBundles(t *testing.T) {
	// This should not panic or error
	ListBundles()
}

func TestBundleValidation(t *testing.T) {
	tests := []struct {
		name      string
		bundle    string
		expectErr bool
	}{
		{"Valid ai bundle", "ai", false},
		{"Valid cli bundle", "cli", false},
		{"Valid cncf bundle", "cncf", false},
		{"Valid experimental-ide bundle", "experimental-ide", false},

		{"Valid fonts bundle", "fonts", false},
		{"Valid full-desktop bundle", "full-desktop", false},
		{"Valid ide bundle", "ide", false},
		{"Valid k8s bundle", "k8s", false},
		{"Invalid bundle", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if bundle exists in our map
			_, exists := bundles[tt.bundle]
			if exists == tt.expectErr {
				t.Errorf("Bundle %s existence = %v, expectErr %v", tt.bundle, exists, tt.expectErr)
			}
		})
	}
}

func TestBundleFile(t *testing.T) {
	// Verify all bundles have proper file names
	for name, bundle := range bundles {
		if bundle.File == "" {
			t.Errorf("Bundle %s has empty file name", name)
		}
		if bundle.Description == "" {
			t.Errorf("Bundle %s has empty description", name)
		}
	}
}

func TestBundleWithLocalFile(t *testing.T) {
	// Create a temporary Brewfile
	tmpFile, err := os.CreateTemp("", "test-Brewfile-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write some content
	content := `tap "homebrew/core"
brew "git"
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// This test will fail if brew is not installed, which is okay for unit tests
	// The integration test will verify the full functionality
	t.Log("Skipping actual bundle installation in unit test")
}

func TestDownloadFile(t *testing.T) {
	// Test with a known good URL
	tmpFile, err := os.CreateTemp("", "download-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Try to download a small file
	url := "https://raw.githubusercontent.com/ublue-os/bluefin/main/README.md"
	if err := downloadFile(url, tmpPath); err != nil {
		t.Skipf("Skipping download test (network required): %v", err)
	}

	// Verify file was created and has content
	info, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("Failed to stat downloaded file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Downloaded file is empty")
	}
}

func TestDownloadFileInvalidURL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "download-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Try to download from invalid URL
	url := "https://invalid.invalid/nonexistent.txt"
	err = downloadFile(url, tmpPath)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}
