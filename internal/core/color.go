package core

// Color represents a foreground color for a screen cell.
// Uses ANSI 256-color codes for terminal compatibility.
type Color uint8

// Predefined colors for game elements.
const (
	ColorDefault Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
	ColorBrightRed
	ColorBrightGreen
	ColorBrightYellow
	ColorBrightBlue
	ColorBrightMagenta
	ColorBrightCyan
	ColorBrightWhite
	ColorOrange
	ColorGray
)
