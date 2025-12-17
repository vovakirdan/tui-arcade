package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vovakirdan/tui-arcade/internal/core"
)

// colorStyles maps core.Color to lipgloss styles.
var colorStyles = map[core.Color]lipgloss.Style{
	core.ColorDefault:       lipgloss.NewStyle(),
	core.ColorRed:           lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
	core.ColorGreen:         lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
	core.ColorYellow:        lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
	core.ColorBlue:          lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
	core.ColorMagenta:       lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
	core.ColorCyan:          lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
	core.ColorWhite:         lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
	core.ColorBrightRed:     lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
	core.ColorBrightGreen:   lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
	core.ColorBrightYellow:  lipgloss.NewStyle().Foreground(lipgloss.Color("11")),
	core.ColorBrightBlue:    lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
	core.ColorBrightMagenta: lipgloss.NewStyle().Foreground(lipgloss.Color("13")),
	core.ColorBrightCyan:    lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
	core.ColorBrightWhite:   lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
	core.ColorOrange:        lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
	core.ColorGray:          lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
}

// RenderScreen converts a Screen buffer to a styled string for display.
// Groups adjacent cells with the same color to minimize ANSI escape sequences.
func RenderScreen(s *core.Screen) string {
	var sb strings.Builder
	// Pre-allocate with extra space for ANSI codes
	sb.Grow(s.Width()*s.Height()*2 + s.Height())

	for y := range s.Height() {
		if y > 0 {
			sb.WriteRune('\n')
		}

		// Group consecutive cells with the same color for efficiency
		x := 0
		for x < s.Width() {
			cell := s.GetCell(x, y)
			startColor := cell.Color

			// Collect consecutive cells with same color
			var run strings.Builder
			for x < s.Width() {
				cell = s.GetCell(x, y)
				if cell.Color != startColor {
					break
				}
				run.WriteRune(cell.Rune)
				x++
			}

			// Apply style to the run
			style, ok := colorStyles[startColor]
			if !ok {
				style = colorStyles[core.ColorDefault]
			}
			sb.WriteString(style.Render(run.String()))
		}
	}
	return sb.String()
}
