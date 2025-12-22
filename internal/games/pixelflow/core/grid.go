package core

import "sort"

// Cell represents a single cell in the grid.
type Cell struct {
	Filled bool  // Whether the cell contains a pixel
	Color  Color // Valid only when Filled is true
}

// Empty returns an empty cell.
func Empty() Cell {
	return Cell{Filled: false}
}

// FilledCell returns a filled cell with the given color.
func FilledCell(c Color) Cell {
	return Cell{Filled: true, Color: c}
}

// Grid represents the game board as a rectangular grid of cells.
// Cells are stored in row-major order: index = y*W + x.
type Grid struct {
	W     int    // Width
	H     int    // Height
	Cells []Cell // Flat array, length W*H
}

// NewGrid creates a new grid with given dimensions and initial pixels.
func NewGrid(w, h int, pixels map[Coord]Color) *Grid {
	g := &Grid{
		W:     w,
		H:     h,
		Cells: make([]Cell, w*h),
	}
	for coord, color := range pixels {
		if g.InBounds(coord) {
			g.SetPixel(coord, color)
		}
	}
	return g
}

// NewEmptyGrid creates an empty grid.
func NewEmptyGrid(w, h int) *Grid {
	return NewGrid(w, h, nil)
}

// index converts coordinate to flat array index.
func (g *Grid) index(c Coord) int {
	return c.Y*g.W + c.X
}

// InBounds returns true if coordinate is within grid.
func (g *Grid) InBounds(c Coord) bool {
	return c.X >= 0 && c.X < g.W && c.Y >= 0 && c.Y < g.H
}

// Get returns the cell at coordinate. Returns empty if out of bounds.
func (g *Grid) Get(c Coord) Cell {
	if !g.InBounds(c) {
		return Empty()
	}
	return g.Cells[g.index(c)]
}

// SetEmpty clears the cell at coordinate.
func (g *Grid) SetEmpty(c Coord) {
	if g.InBounds(c) {
		g.Cells[g.index(c)] = Empty()
	}
}

// SetPixel fills the cell with a color.
func (g *Grid) SetPixel(c Coord, color Color) {
	if g.InBounds(c) {
		g.Cells[g.index(c)] = FilledCell(color)
	}
}

// Clone returns a deep copy of the grid.
func (g *Grid) Clone() *Grid {
	cells := make([]Cell, len(g.Cells))
	copy(cells, g.Cells)
	return &Grid{W: g.W, H: g.H, Cells: cells}
}

// FilledCount returns number of filled cells.
func (g *Grid) FilledCount() int {
	count := 0
	for _, cell := range g.Cells {
		if cell.Filled {
			count++
		}
	}
	return count
}

// IsEmpty returns true if no filled cells remain.
func (g *Grid) IsEmpty() bool {
	return g.FilledCount() == 0
}

// CountByColor returns a map of color -> count for all filled pixels.
func (g *Grid) CountByColor() map[Color]int {
	counts := make(map[Color]int)
	for _, cell := range g.Cells {
		if cell.Filled {
			counts[cell.Color]++
		}
	}
	return counts
}

// ColorsPresent returns a sorted slice of colors present in the grid.
func (g *Grid) ColorsPresent() []Color {
	counts := g.CountByColor()
	colors := make([]Color, 0, len(counts))
	for c := range counts {
		colors = append(colors, c)
	}
	// Sort for determinism
	sort.Slice(colors, func(i, j int) bool {
		return colors[i] < colors[j]
	})
	return colors
}

// FilledCoords returns all filled coordinates in deterministic order (row-major).
func (g *Grid) FilledCoords() []Coord {
	coords := make([]Coord, 0)
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := C(x, y)
			if g.Get(c).Filled {
				coords = append(coords, c)
			}
		}
	}
	return coords
}

// Equal returns true if two grids are identical.
func (g *Grid) Equal(other *Grid) bool {
	if g.W != other.W || g.H != other.H {
		return false
	}
	for i, cell := range g.Cells {
		if cell != other.Cells[i] {
			return false
		}
	}
	return true
}

// Hash returns a simple hash for snapshot comparison.
func (g *Grid) Hash() uint64 {
	var h uint64 = 17
	h = h*31 + uint64(g.W)
	h = h*31 + uint64(g.H)
	for _, cell := range g.Cells {
		if cell.Filled {
			h = h*31 + uint64(cell.Color) + 1
		} else {
			h = h * 31
		}
	}
	return h
}
