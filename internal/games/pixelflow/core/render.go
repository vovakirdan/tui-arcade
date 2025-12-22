package core

import (
	"fmt"
	"strings"
)

// RenderOptions configures ASCII rendering behavior.
type RenderOptions struct {
	ShowCoords   bool        // Include coordinate axes in output
	ShowShooters bool        // Show shooter positions around the grid
	LastShot     *ShotResult // Highlight the last shot path (optional)
	EmptyChar    rune        // Character for empty cells (default '.')
	BorderChar   rune        // Character for borders (default '+')
}

// DefaultRenderOptions returns sensible default rendering options.
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		ShowCoords:   false,
		ShowShooters: false,
		LastShot:     nil,
		EmptyChar:    '.',
		BorderChar:   '+',
	}
}

// RenderASCII converts the current game state to an ASCII string representation.
// This is UI-library-agnostic and can be used for testing, debugging, or TUI rendering.
func RenderASCII(g *Grid, shooters []Shooter, opt RenderOptions) string {
	if opt.EmptyChar == 0 {
		opt.EmptyChar = '.'
	}
	if opt.BorderChar == 0 {
		opt.BorderChar = '+'
	}

	var sb strings.Builder

	// Build a map of last shot path for highlighting
	pathSet := make(map[Coord]bool)
	if opt.LastShot != nil {
		for _, c := range opt.LastShot.Path {
			pathSet[c] = true
		}
	}

	// Build shooter position map
	shooterMap := make(map[Coord]Shooter)
	if opt.ShowShooters && shooters != nil {
		for _, s := range shooters {
			shooterMap[s.Pos] = s
		}
	}

	// Calculate render boundaries
	minX, maxX := 0, g.W-1
	minY, maxY := 0, g.H-1
	if opt.ShowShooters {
		minX, maxX = -1, g.W
		minY, maxY = -1, g.H
	}

	// Render column headers if showing coords
	if opt.ShowCoords {
		sb.WriteString("  ")
		if opt.ShowShooters {
			sb.WriteString(" ") // extra space for shooter row
		}
		for x := 0; x < g.W; x++ {
			sb.WriteString(fmt.Sprintf("%d", x%10))
		}
		sb.WriteString("\n")
	}

	// Render each row
	for y := minY; y <= maxY; y++ {
		// Row label
		if opt.ShowCoords {
			if y >= 0 && y < g.H {
				sb.WriteString(fmt.Sprintf("%2d", y%100))
			} else {
				sb.WriteString("  ")
			}
		}

		// Render each cell in the row
		for x := minX; x <= maxX; x++ {
			c := C(x, y)

			// Check if this is a shooter position
			if s, ok := shooterMap[c]; ok {
				sb.WriteRune(shooterChar(s))
				continue
			}

			// Check if out of grid bounds
			if !g.InBounds(c) {
				// Border area
				if opt.ShowShooters {
					if pathSet[c] {
						sb.WriteRune('*') // Shot path through border
					} else {
						sb.WriteRune(' ')
					}
				}
				continue
			}

			// Inside grid
			cell := g.Get(c)

			// Check if on shot path
			if pathSet[c] {
				if cell.Filled {
					sb.WriteRune('X') // Hit marker
				} else {
					sb.WriteRune('*') // Path through empty
				}
				continue
			}

			// Normal cell rendering
			if cell.Filled {
				sb.WriteRune(cell.Color.Char())
			} else {
				sb.WriteRune(opt.EmptyChar)
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// shooterChar returns a character representing a shooter based on its direction.
func shooterChar(s Shooter) rune {
	switch s.Dir {
	case DirUp:
		return '^'
	case DirDown:
		return 'v'
	case DirLeft:
		return '<'
	case DirRight:
		return '>'
	default:
		return '?'
	}
}

// RenderCompact returns a minimal string representation without borders or coords.
func RenderCompact(g *Grid) string {
	var sb strings.Builder
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			cell := g.Get(C(x, y))
			if cell.Filled {
				sb.WriteRune(cell.Color.Char())
			} else {
				sb.WriteRune('.')
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// RenderWithFrame renders the grid with a simple border frame.
func RenderWithFrame(g *Grid, title string) string {
	var sb strings.Builder

	// Top border with title
	borderWidth := g.W + 2
	if title != "" {
		padding := (borderWidth - len(title) - 2) / 2
		if padding < 0 {
			padding = 0
		}
		sb.WriteString(strings.Repeat("-", padding))
		sb.WriteString(" ")
		sb.WriteString(title)
		sb.WriteString(" ")
		remaining := borderWidth - padding - len(title) - 2
		if remaining > 0 {
			sb.WriteString(strings.Repeat("-", remaining))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString(strings.Repeat("-", borderWidth))
		sb.WriteString("\n")
	}

	// Grid rows with side borders
	for y := 0; y < g.H; y++ {
		sb.WriteString("|")
		for x := 0; x < g.W; x++ {
			cell := g.Get(C(x, y))
			if cell.Filled {
				sb.WriteRune(cell.Color.Char())
			} else {
				sb.WriteRune('.')
			}
		}
		sb.WriteString("|\n")
	}

	// Bottom border
	sb.WriteString(strings.Repeat("-", borderWidth))
	sb.WriteString("\n")

	return sb.String()
}

// GridToLines converts the grid to a slice of strings (one per row).
// Useful for line-by-line comparisons in tests.
func GridToLines(g *Grid) []string {
	lines := make([]string, g.H)
	for y := 0; y < g.H; y++ {
		var sb strings.Builder
		for x := 0; x < g.W; x++ {
			cell := g.Get(C(x, y))
			if cell.Filled {
				sb.WriteRune(cell.Color.Char())
			} else {
				sb.WriteRune('.')
			}
		}
		lines[y] = sb.String()
	}
	return lines
}
