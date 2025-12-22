package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// PixelFlowTheme contains all configurable visual styles for PixelFlow.
type PixelFlowTheme struct {
	// Grid cell colors
	PinkPixel   lipgloss.Style
	CyanPixel   lipgloss.Style
	GreenPixel  lipgloss.Style
	YellowPixel lipgloss.Style
	PurplePixel lipgloss.Style
	EmptyCell   lipgloss.Style

	// Rail styles
	RailEmpty    lipgloss.Style
	RailCorner   lipgloss.Style
	RailShooter  lipgloss.Style
	RailShooterD lipgloss.Style // Dry shooter

	// HUD styles
	HUDTitle     lipgloss.Style
	HUDValue     lipgloss.Style
	HUDSeparator lipgloss.Style
	HUDControls  lipgloss.Style

	// Deck preview styles
	DeckLabel   lipgloss.Style
	DeckShooter lipgloss.Style
	DeckAmmo    lipgloss.Style

	// Overlay styles
	OverlayBorder lipgloss.Style
	OverlayTitle  lipgloss.Style
	OverlayText   lipgloss.Style

	// Level picker styles
	MenuTitle       lipgloss.Style
	MenuItemNormal  lipgloss.Style
	MenuItemActive  lipgloss.Style
	MenuDescription lipgloss.Style
}

// DefaultPixelFlowTheme returns the default visual theme.
func DefaultPixelFlowTheme() PixelFlowTheme {
	return PixelFlowTheme{
		// Pixel colors - vibrant and distinct
		PinkPixel:   lipgloss.NewStyle().Foreground(lipgloss.Color("205")), // Hot pink
		CyanPixel:   lipgloss.NewStyle().Foreground(lipgloss.Color("51")),  // Bright cyan
		GreenPixel:  lipgloss.NewStyle().Foreground(lipgloss.Color("46")),  // Lime green
		YellowPixel: lipgloss.NewStyle().Foreground(lipgloss.Color("226")), // Bright yellow
		PurplePixel: lipgloss.NewStyle().Foreground(lipgloss.Color("135")), // Medium purple
		EmptyCell:   lipgloss.NewStyle().Foreground(lipgloss.Color("238")), // Dark gray

		// Rail styles
		RailEmpty:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")), // Dim gray
		RailCorner:   lipgloss.NewStyle().Foreground(lipgloss.Color("245")), // Medium gray
		RailShooter:  lipgloss.NewStyle().Bold(true),                        // Inherits color from pixel
		RailShooterD: lipgloss.NewStyle().Foreground(lipgloss.Color("88")),  // Dark red (dry)

		// HUD styles
		HUDTitle:     lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true),
		HUDValue:     lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		HUDSeparator: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		HUDControls:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")),

		// Deck preview
		DeckLabel:   lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		DeckShooter: lipgloss.NewStyle().Bold(true),
		DeckAmmo:    lipgloss.NewStyle().Foreground(lipgloss.Color("250")),

		// Overlay styles
		OverlayBorder: lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		OverlayTitle:  lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true),
		OverlayText:   lipgloss.NewStyle().Foreground(lipgloss.Color("255")),

		// Level picker
		MenuTitle:       lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true),
		MenuItemNormal:  lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		MenuItemActive:  lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true),
		MenuDescription: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
	}
}

// NeonPixelFlowTheme returns a neon-style theme.
func NeonPixelFlowTheme() PixelFlowTheme {
	theme := DefaultPixelFlowTheme()
	theme.PinkPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("199"))   // Neon pink
	theme.CyanPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))    // Neon cyan
	theme.GreenPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("118"))  // Neon green
	theme.YellowPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("227")) // Neon yellow
	theme.PurplePixel = lipgloss.NewStyle().Foreground(lipgloss.Color("171")) // Neon purple
	return theme
}

// PastelPixelFlowTheme returns a softer pastel theme.
func PastelPixelFlowTheme() PixelFlowTheme {
	theme := DefaultPixelFlowTheme()
	theme.PinkPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("218"))   // Pastel pink
	theme.CyanPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("123"))   // Pastel cyan
	theme.GreenPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("157"))  // Pastel green
	theme.YellowPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("229")) // Pastel yellow
	theme.PurplePixel = lipgloss.NewStyle().Foreground(lipgloss.Color("183")) // Pastel purple
	return theme
}

// MonochromePixelFlowTheme returns a grayscale theme.
func MonochromePixelFlowTheme() PixelFlowTheme {
	theme := DefaultPixelFlowTheme()
	theme.PinkPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	theme.CyanPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	theme.GreenPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	theme.YellowPixel = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	theme.PurplePixel = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
	return theme
}

// Global theme variable (can be changed at runtime)
var pixelflowTheme = DefaultPixelFlowTheme()

// SetPixelFlowTheme sets the global theme.
func SetPixelFlowTheme(theme PixelFlowTheme) {
	pixelflowTheme = theme
}

// GetPixelFlowTheme returns the current global theme.
func GetPixelFlowTheme() PixelFlowTheme {
	return pixelflowTheme
}
