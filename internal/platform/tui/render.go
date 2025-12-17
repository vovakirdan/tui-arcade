package tui

import (
	"github.com/vovakirdan/tui-arcade/internal/core"
)

// RenderScreen converts a Screen buffer to a string for display.
// This function allows for future enhancements like styling or color support.
func RenderScreen(s *core.Screen) string {
	return s.String()
}
