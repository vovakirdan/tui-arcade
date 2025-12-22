package core

// Grid represents the game board as a rectangular grid of cells.
// Cells are stored in row-major order: index = y*W + x.
type Grid struct {
	W     int    // Width of the grid
	H     int    // Height of the grid
	Cells []Cell // Flat array of cells, length W*H
}

// NewGrid creates a new grid with the given dimensions and initial pixels.
// All cells start empty; pixels map provides initial filled cells.
func NewGrid(w, h int, pixels map[Coord]Color) *Grid {
	g := &Grid{
		W:     w,
		H:     h,
		Cells: make([]Cell, w*h),
	}
	// Initialize all cells as empty (default zero value is already empty)
	// Set filled pixels from the map
	for coord, color := range pixels {
		if g.InBounds(coord) {
			g.SetPixel(coord, color)
		}
	}
	return g
}

// NewEmptyGrid creates a new grid with all cells empty.
func NewEmptyGrid(w, h int) *Grid {
	return NewGrid(w, h, nil)
}

// index converts a coordinate to a flat array index.
func (g *Grid) index(c Coord) int {
	return c.Y*g.W + c.X
}

// InBounds returns true if the coordinate is within the grid boundaries.
func (g *Grid) InBounds(c Coord) bool {
	return c.X >= 0 && c.X < g.W && c.Y >= 0 && c.Y < g.H
}

// Get returns the cell at the given coordinate.
// Returns an empty cell if out of bounds.
func (g *Grid) Get(c Coord) Cell {
	if !g.InBounds(c) {
		return Empty()
	}
	return g.Cells[g.index(c)]
}

// SetEmpty clears the cell at the given coordinate.
func (g *Grid) SetEmpty(c Coord) {
	if g.InBounds(c) {
		g.Cells[g.index(c)] = Empty()
	}
}

// SetPixel fills the cell at the given coordinate with the specified color.
func (g *Grid) SetPixel(c Coord, color Color) {
	if g.InBounds(c) {
		g.Cells[g.index(c)] = FilledCell(color)
	}
}

// Set sets the cell at the given coordinate.
func (g *Grid) Set(c Coord, cell Cell) {
	if g.InBounds(c) {
		g.Cells[g.index(c)] = cell
	}
}

// Clone returns a deep copy of the grid.
func (g *Grid) Clone() *Grid {
	cells := make([]Cell, len(g.Cells))
	copy(cells, g.Cells)
	return &Grid{
		W:     g.W,
		H:     g.H,
		Cells: cells,
	}
}

// FilledCount returns the number of filled cells in the grid.
func (g *Grid) FilledCount() int {
	count := 0
	for _, cell := range g.Cells {
		if cell.Filled {
			count++
		}
	}
	return count
}

// IsCleared returns true if all cells are empty.
func (g *Grid) IsCleared() bool {
	return g.FilledCount() == 0
}

// AllCoords returns an iterator-like slice of all coordinates in the grid.
// Ordered by row then column.
func (g *Grid) AllCoords() []Coord {
	coords := make([]Coord, 0, g.W*g.H)
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			coords = append(coords, C(x, y))
		}
	}
	return coords
}

// FilledCoords returns all coordinates that contain filled pixels.
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

// Equal returns true if two grids have the same dimensions and contents.
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
