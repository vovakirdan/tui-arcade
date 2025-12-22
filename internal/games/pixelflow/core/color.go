package core

import "strings"

// Color represents a pixel color in the game.
type Color uint8

const (
	ColorRed Color = iota
	ColorGreen
	ColorBlue
	ColorYellow
	ColorPurple
	ColorCount // Sentinel value for iteration
)

// String returns the string representation of a color.
func (c Color) String() string {
	switch c {
	case ColorRed:
		return "red"
	case ColorGreen:
		return "green"
	case ColorBlue:
		return "blue"
	case ColorYellow:
		return "yellow"
	case ColorPurple:
		return "purple"
	default:
		return "unknown"
	}
}

// Char returns a single character representation of the color for ASCII rendering.
func (c Color) Char() rune {
	switch c {
	case ColorRed:
		return 'R'
	case ColorGreen:
		return 'G'
	case ColorBlue:
		return 'B'
	case ColorYellow:
		return 'Y'
	case ColorPurple:
		return 'P'
	default:
		return '?'
	}
}

// ParseColor converts a string to a Color.
// Returns ColorRed and false if the string is not recognized.
func ParseColor(s string) (Color, bool) {
	switch strings.ToLower(s) {
	case "red", "r":
		return ColorRed, true
	case "green", "g":
		return ColorGreen, true
	case "blue", "b":
		return ColorBlue, true
	case "yellow", "y":
		return ColorYellow, true
	case "purple", "p":
		return ColorPurple, true
	default:
		return ColorRed, false
	}
}

// AllColors returns a slice of all valid colors.
func AllColors() []Color {
	return []Color{ColorRed, ColorGreen, ColorBlue, ColorYellow, ColorPurple}
}
