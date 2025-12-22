package core

import (
	"fmt"
	"strings"
)

// RenderASCII creates an ASCII representation of the current game state.
// This is used for debugging, testing (golden outputs), and simple visualization.
//
// Format:
//   - Grid cells: empty='.', colors=P/C/G/Y/U (uppercase)
//   - Rail with active shooters shown around the grid
//   - Deck and waiting slots shown below
func RenderASCII(s *State) string {
	var sb strings.Builder

	// Build map of active shooter positions for quick lookup
	activeByRail := make(map[int][]ActiveShooter)
	for _, a := range s.Active {
		activeByRail[a.RailIndex] = append(activeByRail[a.RailIndex], a)
	}

	// Calculate display dimensions
	// Grid is W x H, rail adds 1 cell border on each side
	displayW := s.Grid.W + 2
	displayH := s.Grid.H + 2

	// Header
	sb.WriteString(fmt.Sprintf("Tick: %d | Pixels: %d | Active: %d/%d | Deck: %d | Waiting: %d\n",
		s.Tick, s.Grid.FilledCount(), len(s.Active), s.Capacity, s.Deck.TotalShooters(), s.Waiting.Count()))
	sb.WriteString(strings.Repeat("-", displayW+10) + "\n")

	// Build the display grid
	for dy := 0; dy < displayH; dy++ {
		for dx := 0; dx < displayW; dx++ {
			char := s.getRenderChar(dx, dy, activeByRail)
			sb.WriteRune(char)
		}
		sb.WriteString("\n")
	}

	// Footer: Deck preview (show queues)
	sb.WriteString(strings.Repeat("-", displayW+10) + "\n")
	for qi, queue := range s.Deck.Queues {
		sb.WriteString(fmt.Sprintf("Q%d: ", qi))
		maxShow := 5
		for i, shooter := range queue {
			if i >= maxShow {
				sb.WriteString(fmt.Sprintf("... +%d more", len(queue)-maxShow))
				break
			}
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%c(%d)", shooter.Color.LowerChar(), shooter.Ammo))
		}
		if len(queue) == 0 {
			sb.WriteString("(empty)")
		}
		sb.WriteString("\n")
	}

	// Waiting slots
	sb.WriteString("Wait: ")
	hasAny := false
	for i, shooter := range s.Waiting.Slots {
		if shooter != nil {
			if hasAny {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("[%d]%c(%d)", i, shooter.Color.LowerChar(), shooter.Ammo))
			hasAny = true
		}
	}
	if !hasAny {
		sb.WriteString("(empty)")
	}
	sb.WriteString("\n")

	return sb.String()
}

// getRenderChar returns the character to display at display coordinates (dx, dy).
func (s *State) getRenderChar(dx, dy int, activeByRail map[int][]ActiveShooter) rune {
	// Map display coords to grid coords (offset by 1 for rail border)
	gx := dx - 1
	gy := dy - 1

	// Check if this is a rail position (border)
	if dx == 0 || dx == s.Grid.W+1 || dy == 0 || dy == s.Grid.H+1 {
		return s.getRailChar(dx, dy, activeByRail)
	}

	// Inside grid
	if gx >= 0 && gx < s.Grid.W && gy >= 0 && gy < s.Grid.H {
		cell := s.Grid.Get(C(gx, gy))
		if cell.Filled {
			return cell.Color.Char()
		}
		return '.'
	}

	return ' '
}

// getRailChar returns the character for a rail position.
func (s *State) getRailChar(dx, dy int, activeByRail map[int][]ActiveShooter) rune {
	// Convert display coords to rail index
	railIndex := s.displayToRailIndex(dx, dy)
	if railIndex < 0 {
		return '+' // Corner
	}

	// Check for active shooter at this rail position
	if shooters, ok := activeByRail[railIndex]; ok && len(shooters) > 0 {
		// Show first shooter's color (lowercase indicates shooter on rail)
		return shooters[0].Color.LowerChar()
	}

	// Empty rail position
	return '-'
}

// displayToRailIndex converts display coordinates to rail index.
// Returns -1 for corners or invalid positions.
func (s *State) displayToRailIndex(dx, dy int) int {
	w := s.Grid.W
	h := s.Grid.H

	// Top edge (dy=0, dx=1 to w)
	if dy == 0 && dx >= 1 && dx <= w {
		return dx - 1
	}

	// Right edge (dx=w+1, dy=1 to h)
	if dx == w+1 && dy >= 1 && dy <= h {
		return w + (dy - 1)
	}

	// Bottom edge (dy=h+1, dx=w to 1) - reversed
	if dy == h+1 && dx >= 1 && dx <= w {
		return w + h + (w - dx)
	}

	// Left edge (dx=0, dy=h to 1) - reversed
	if dx == 0 && dy >= 1 && dy <= h {
		return 2*w + h + (h - dy)
	}

	return -1 // Corner
}

// RenderGrid renders just the grid without rail or state info.
func RenderGrid(g *Grid) string {
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

// RenderGridCompact renders the grid as a single line (for hashing/comparison).
func RenderGridCompact(g *Grid) string {
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
	}
	return sb.String()
}

// RenderDeck renders the deck as a string.
func RenderDeck(deck []Shooter) string {
	var sb strings.Builder
	for i, s := range deck {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%c%d", s.Color.LowerChar(), s.Ammo))
	}
	return sb.String()
}

// RenderActive renders active shooters as a string.
func RenderActive(active []ActiveShooter) string {
	var sb strings.Builder
	for i, a := range active {
		if i > 0 {
			sb.WriteString(" ")
		}
		dryMark := ""
		if a.Dry {
			dryMark = "*"
		}
		sb.WriteString(fmt.Sprintf("%c%d@%d%s", a.Color.LowerChar(), a.Ammo, a.RailIndex, dryMark))
	}
	return sb.String()
}
