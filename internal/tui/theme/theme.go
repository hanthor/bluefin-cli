// Package theme defines the semantic color tokens for the application.
//
// All UI code styles against tokens (Accent, TextMuted, Success, ...) rather
// than raw colors. Resolve builds the token set from a Catppuccin palette for
// the terminal's detected background (Latte for light, Mocha for dark).
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
	catppuccin "github.com/catppuccin/go"
)

// Theme is the semantic color palette for the application.
type Theme struct {
	// Semantic tokens
	Bg        color.Color
	Surface   color.Color
	Overlay   color.Color
	Accent    color.Color
	AccentAlt color.Color
	TextBase  color.Color
	TextMuted color.Color
	TextFaint color.Color
	Success   color.Color
	Warning   color.Color
	Error     color.Color
	Info      color.Color

	// Legacy aliases still referenced by older styles.
	PrimaryBorder   color.Color
	SecondaryBorder color.Color
	FaintBorder     color.Color
	PrimaryText     color.Color
	SecondaryText   color.Color
	FaintText       color.Color
	InvertedText    color.Color
	SuccessText     color.Color
	WarningText     color.Color
	ErrorText       color.Color
	InfoText        color.Color

	SelectedBackground color.Color

	IsDark bool
}

// Resolve returns the theme for the given terminal background.
func Resolve(isDark bool) Theme {
	f := catppuccin.Latte
	if isDark {
		f = catppuccin.Mocha
	}
	c := func(cc catppuccin.Color) color.Color { return lipgloss.Color(cc.Hex) }

	t := Theme{
		Bg:        c(f.Base()),
		Surface:   c(f.Surface0()),
		Overlay:   c(f.Overlay0()),
		Accent:    c(f.Blue()),
		AccentAlt: c(f.Sapphire()),
		TextBase:  c(f.Text()),
		TextMuted: c(f.Subtext1()),
		TextFaint: c(f.Overlay1()),
		Success:   c(f.Green()),
		Warning:   c(f.Yellow()),
		Error:     c(f.Red()),
		Info:      c(f.Sky()),
		IsDark:    isDark,
	}

	t.SelectedBackground = c(f.Surface1())
	t.PrimaryBorder = t.Accent
	t.SecondaryBorder = t.Overlay
	t.FaintBorder = t.Surface
	t.PrimaryText = t.TextBase
	t.SecondaryText = t.TextMuted
	t.FaintText = t.TextFaint
	t.InvertedText = t.Bg
	t.SuccessText = t.Success
	t.WarningText = t.Warning
	t.ErrorText = t.Error
	t.InfoText = t.Info
	return t
}

// DefaultTheme is the fallback before the terminal background is detected.
// Most terminals are dark, so default to the dark variant.
var DefaultTheme = Resolve(true)
