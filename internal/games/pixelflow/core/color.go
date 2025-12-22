// Package core provides the core game logic for PixelFlow conveyor puzzle game.
// This package is UI-agnostic and deterministic.
package core

import "strings"

// Color represents a pixel/shooter color in the game.
type Color uint8

const (
	ColorPink Color = iota
	ColorCyan
	ColorGreen
	ColorYellow
	ColorPurple
	ColorCount // Sentinel for iteration
)

// String returns the lowercase string name of the color.
func (c Color) String() string {
	switch c {
	case ColorPink:
		return "pink"
	case ColorCyan:
		return "cyan"
	case ColorGreen:
		return "green"
	case ColorYellow:
		return "yellow"
	case ColorPurple:
		return "purple"
	default:
		return "unknown"
	}
}

// Char returns a single character for ASCII rendering.
// Pink=P, Cyan=C, Green=G, Yellow=Y, Purple=U
func (c Color) Char() rune {
	switch c {
	case ColorPink:
		return 'P'
	case ColorCyan:
		return 'C'
	case ColorGreen:
		return 'G'
	case ColorYellow:
		return 'Y'
	case ColorPurple:
		return 'U'
	default:
		return '?'
	}
}

// LowerChar returns lowercase char for shooter display.
func (c Color) LowerChar() rune {
	switch c {
	case ColorPink:
		return 'p'
	case ColorCyan:
		return 'c'
	case ColorGreen:
		return 'g'
	case ColorYellow:
		return 'y'
	case ColorPurple:
		return 'u'
	default:
		return '?'
	}
}

// ParseColor converts a string to Color.
// Returns ColorPink and false if unrecognized.
func ParseColor(s string) (Color, bool) {
	switch strings.ToLower(s) {
	case "pink", "p":
		return ColorPink, true
	case "cyan", "c":
		return ColorCyan, true
	case "green", "g":
		return ColorGreen, true
	case "yellow", "y":
		return ColorYellow, true
	case "purple", "u":
		return ColorPurple, true
	default:
		return ColorPink, false
	}
}

// AllColors returns all valid colors in order.
func AllColors() []Color {
	return []Color{ColorPink, ColorCyan, ColorGreen, ColorYellow, ColorPurple}
}
