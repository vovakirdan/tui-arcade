package core

// RailPos represents a single position on the rail/conveyor loop.
// The rail runs clockwise around the outside of the grid.
type RailPos struct {
	Index int   // Position index on the rail (0 to Len-1)
	Entry Coord // First grid cell that a shot from here would hit
	Dir   Dir   // Inward direction (towards grid center)
	Side  Side  // Which side of the grid this position is on
}

// Side indicates which edge of the grid a rail position is on.
type Side uint8

const (
	SideTop Side = iota
	SideRight
	SideBottom
	SideLeft
)

// Rail represents the closed conveyor loop around the grid.
// Shooters move clockwise: Top -> Right -> Bottom -> Left -> Top...
type Rail struct {
	Positions []RailPos
	GridW     int
	GridH     int
}

// NewRail creates a rail loop for a grid of given dimensions.
// The rail has 2*(W+H) positions going clockwise:
//   - Top edge: indices 0 to W-1, shooting Down
//   - Right edge: indices W to W+H-1, shooting Left
//   - Bottom edge: indices W+H to 2W+H-1, shooting Up
//   - Left edge: indices 2W+H to 2W+2H-1, shooting Right
func NewRail(w, h int) Rail {
	totalLen := 2 * (w + h)
	positions := make([]RailPos, totalLen)
	idx := 0

	// Top edge: x goes 0 to W-1, shoots Down, entry is (x, 0)
	for x := 0; x < w; x++ {
		positions[idx] = RailPos{
			Index: idx,
			Entry: C(x, 0),
			Dir:   DirDown,
			Side:  SideTop,
		}
		idx++
	}

	// Right edge: y goes 0 to H-1, shoots Left, entry is (W-1, y)
	for y := 0; y < h; y++ {
		positions[idx] = RailPos{
			Index: idx,
			Entry: C(w-1, y),
			Dir:   DirLeft,
			Side:  SideRight,
		}
		idx++
	}

	// Bottom edge: x goes W-1 to 0, shoots Up, entry is (x, H-1)
	for x := w - 1; x >= 0; x-- {
		positions[idx] = RailPos{
			Index: idx,
			Entry: C(x, h-1),
			Dir:   DirUp,
			Side:  SideBottom,
		}
		idx++
	}

	// Left edge: y goes H-1 to 0, shoots Right, entry is (0, y)
	for y := h - 1; y >= 0; y-- {
		positions[idx] = RailPos{
			Index: idx,
			Entry: C(0, y),
			Dir:   DirRight,
			Side:  SideLeft,
		}
		idx++
	}

	return Rail{
		Positions: positions,
		GridW:     w,
		GridH:     h,
	}
}

// Len returns the total number of positions on the rail.
func (r Rail) Len() int {
	return len(r.Positions)
}

// Get returns the rail position at the given index.
func (r Rail) Get(index int) RailPos {
	n := r.Len()
	// Normalize to valid range
	index = ((index % n) + n) % n
	return r.Positions[index]
}

// Next returns the next rail index (clockwise movement).
func (r Rail) Next(index int) int {
	return (index + 1) % r.Len()
}

// TraceRay traces a shot from a rail position into the grid.
// Returns the first filled cell coordinate and its color, or (Coord{}, false) if no hit.
func (r Rail) TraceRay(g *Grid, railIndex int) (Coord, Color, bool) {
	pos := r.Get(railIndex)
	current := pos.Entry
	dx, dy := pos.Dir.Delta()

	// Trace until we hit a filled cell or exit bounds
	for g.InBounds(current) {
		cell := g.Get(current)
		if cell.Filled {
			return current, cell.Color, true
		}
		current = current.Add(dx, dy)
	}

	return Coord{}, 0, false
}

// FrontierColors returns a map of railIndex -> frontColor for all positions
// where shooting would hit a filled cell. Useful for analysis.
func (r Rail) FrontierColors(g *Grid) map[int]Color {
	frontier := make(map[int]Color)
	for i := 0; i < r.Len(); i++ {
		if _, color, hit := r.TraceRay(g, i); hit {
			frontier[i] = color
		}
	}
	return frontier
}

// FindRailIndicesForColor returns all rail indices where the frontier pixel
// has the given color. Results are in ascending order.
func (r Rail) FindRailIndicesForColor(g *Grid, c Color) []int {
	indices := make([]int, 0)
	for i := 0; i < r.Len(); i++ {
		if _, color, hit := r.TraceRay(g, i); hit && color == c {
			indices = append(indices, i)
		}
	}
	return indices
}

// CountShootableByColor counts how many rail positions have each color as frontier.
func (r Rail) CountShootableByColor(g *Grid) map[Color]int {
	counts := make(map[Color]int)
	for i := 0; i < r.Len(); i++ {
		if _, color, hit := r.TraceRay(g, i); hit {
			counts[color]++
		}
	}
	return counts
}

// SpawnIndex returns the default spawn point for new shooters.
// This is typically index 0 (top-left of top edge).
func (r Rail) SpawnIndex() int {
	return 0
}

// PositionOnSide returns all rail positions on a given side.
func (r Rail) PositionOnSide(side Side) []RailPos {
	result := make([]RailPos, 0)
	for _, pos := range r.Positions {
		if pos.Side == side {
			result = append(result, pos)
		}
	}
	return result
}
