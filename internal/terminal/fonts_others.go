//go:build !darwin && !windows

package terminal

func detectPlatformTerminals() []ConfiguredTerminal {
	// Linux: no platform-specific terminals beyond Ghostty/Alacritty/Kitty
	// which are handled in the cross-platform DetectTerminals().
	return nil
}

func setPlatformTerminalFont(term ConfiguredTerminal, fontFamily string) error {
	return nil
}
