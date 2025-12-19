package motd

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestToggle(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	tests := []struct {
		name    string
		target  string
		enable  bool
		wantErr bool
	}{
		{"Enable for all shells", "all", true, false},
		{"Enable for bash", "bash", true, false},
		{"Enable for zsh", "zsh", true, false},
		{"Enable for fish", "fish", true, false},
		{"Disable for all shells", "all", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Toggle(tt.target, tt.enable)
			if (err != nil) != tt.wantErr {
				t.Errorf("Toggle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetupMOTD(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	if err := setupMOTD(); err != nil {
		t.Fatalf("setupMOTD() failed: %v", err)
	}

	// Verify directories were created
	motdPath := filepath.Join(tmpHome, ".local/share/bluefin-cli/motd")
	tipsPath := filepath.Join(motdPath, "tips")

	if _, err := os.Stat(motdPath); os.IsNotExist(err) {
		t.Error("Expected MOTD directory to exist")
	}

	if _, err := os.Stat(tipsPath); os.IsNotExist(err) {
		t.Error("Expected tips directory to exist")
	}

	// Verify MOTD script was created
	scriptPath := filepath.Join(motdPath, "bluefin-motd.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("Expected MOTD script to exist")
	}

	// Verify script is executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("Failed to stat MOTD script: %v", err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Error("Expected MOTD script to be executable")
	}

	// Verify tips were created
	files, err := filepath.Glob(filepath.Join(tipsPath, "*.md"))
	if err != nil {
		t.Fatalf("Failed to list tips: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected at least one tip file to be created")
	}

	// Verify config was created
	configPath := filepath.Join(motdPath, "motd.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected MOTD config to exist")
	}
}

func TestGetImageInfo(t *testing.T) {
	info := getImageInfo()

	if info.ImageName == "" {
		t.Error("Expected ImageName to be set")
	}

	if info.ImageTag == "" {
		t.Error("Expected ImageTag to be set")
	}

	if info.ImageFlavor != "homebrew" {
		t.Errorf("Expected ImageFlavor to be 'homebrew', got %s", info.ImageFlavor)
	}

	if info.ImageVendor != "bluefin-cli" {
		t.Errorf("Expected ImageVendor to be 'bluefin-cli', got %s", info.ImageVendor)
	}
}

func TestGetRandomTip(t *testing.T) {
	// Create temporary tips directory
	tmpDir := t.TempDir()
	tipsPath := filepath.Join(tmpDir, "tips")
	if err := os.MkdirAll(tipsPath, 0755); err != nil {
		t.Fatalf("Failed to create tips directory: %v", err)
	}

	// Create some test tips
	testTips := []string{
		"This is tip 1",
		"This is tip 2",
		"This is tip 3",
	}

	for i, tip := range testTips {
		tipFile := filepath.Join(tipsPath, string(rune('0'+i))+".md")
		if err := os.WriteFile(tipFile, []byte(tip), 0644); err != nil {
			t.Fatalf("Failed to write tip file: %v", err)
		}
	}

	// Get a random tip
	tip := getRandomTip(tipsPath)

	if tip == "" {
		t.Error("Expected non-empty tip")
	}

	// Verify it starts with the tip prefix
	if !strings.HasPrefix(tip, "💡 **Tip:**") {
		t.Error("Expected tip to start with prefix")
	}

	// Verify the tip content is one of our test tips
	found := false
	for _, testTip := range testTips {
		if strings.Contains(tip, testTip) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected tip to contain one of the test tips")
	}
}

func TestCheckStatus(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	// Enable MOTD for bash
	if err := Toggle("bash", true); err != nil {
		t.Fatalf("Failed to enable MOTD: %v", err)
	}

	status := CheckStatus()

	if !status["bash"] {
		t.Error("Expected bash MOTD to be enabled")
	}
	if status["zsh"] {
		t.Error("Expected zsh MOTD to be disabled")
	}
	if status["fish"] {
		t.Error("Expected fish MOTD to be disabled")
	}
}

func TestToggleIdempotency(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

	// Enable twice for bash
	if err := Toggle("bash", true); err != nil {
		t.Fatalf("First enable failed: %v", err)
	}

	if err := Toggle("bash", true); err != nil {
		t.Fatalf("Second enable failed: %v", err)
	}

	// Check config file doesn't have duplicate entries
	configFile := filepath.Join(tmpHome, ".bashrc")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	markerCount := strings.Count(string(content), motdMarker)
	if markerCount > 1 {
		t.Errorf("Expected at most 1 MOTD marker, found %d", markerCount)
	}
}

func TestShow(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Unsetenv("HOME")

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
	// ReadAll from pipe
	out, _ := io.ReadAll(r)
	output := string(out)

	// Verify expected output
	// We expect the default template content
	expectedStrings := []string{
		"Welcome to Bluefin CLI",
		"Command",
		"Description",
		"bluefin-cli",
		"GitHub Issues",
	}

	strippedOutput := stripAnsi(output)

	for _, s := range expectedStrings {
		if !strings.Contains(strippedOutput, s) {
			t.Errorf("Expected output to contain %q, got:\n%s", s, strippedOutput)
		}
	}
}

func stripAnsi(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
