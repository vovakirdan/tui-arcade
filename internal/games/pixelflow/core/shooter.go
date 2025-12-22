package core

// Shooter represents a shooter that fires shots to remove pixels.
// Shooters are positioned around the grid perimeter and fire inward.
type Shooter struct {
	ID    int   // Unique identifier
	Pos   Coord // Starting position for the shot (outside grid or on border)
	Dir   Dir   // Direction the shot travels
	Color Color // The shooter only removes pixels of this color
}

// MakeShooters creates n shooters deterministically placed around the grid perimeter.
// The seed parameter is reserved for future randomness; currently ignored for determinism.
// Shooters are placed evenly along the perimeter: top, right, bottom, left.
// Each shooter points inward and cycles through colors.
//
// Placement policy:
// - Shooters are placed one cell outside the grid boundary
// - They fire inward toward the grid
// - Distribution is roughly even across all four sides
func MakeShooters(n int, seed uint64, gridW, gridH int) []Shooter {
	if n <= 0 {
		return nil
	}

	// Generate perimeter positions
	perimeter := generatePerimeterPositions(gridW, gridH)
	if len(perimeter) == 0 {
		return nil
	}

	shooters := make([]Shooter, n)
	colors := AllColors()

	// Distribute shooters evenly across perimeter
	step := len(perimeter) / n
	if step < 1 {
		step = 1
	}

	for i := 0; i < n; i++ {
		idx := (i * step) % len(perimeter)
		pos, dir := perimeter[idx].pos, perimeter[idx].dir
		shooters[i] = Shooter{
			ID:    i,
			Pos:   pos,
			Dir:   dir,
			Color: colors[i%len(colors)],
		}
	}

	return shooters
}

// perimeterEntry holds a position and inward direction for shooter placement.
type perimeterEntry struct {
	pos Coord
	dir Dir
}

// generatePerimeterPositions generates all perimeter positions around a grid.
// Each position is one cell outside the grid, with direction pointing inward.
func generatePerimeterPositions(w, h int) []perimeterEntry {
	entries := make([]perimeterEntry, 0, 2*w+2*h)

	// Top edge: shooters at y=-1, firing down
	for x := 0; x < w; x++ {
		entries = append(entries, perimeterEntry{
			pos: C(x, -1),
			dir: DirDown,
		})
	}

	// Right edge: shooters at x=w, firing left
	for y := 0; y < h; y++ {
		entries = append(entries, perimeterEntry{
			pos: C(w, y),
			dir: DirLeft,
		})
	}

	// Bottom edge: shooters at y=h, firing up (reversed for clockwise order)
	for x := w - 1; x >= 0; x-- {
		entries = append(entries, perimeterEntry{
			pos: C(x, h),
			dir: DirUp,
		})
	}

	// Left edge: shooters at x=-1, firing right (reversed for clockwise order)
	for y := h - 1; y >= 0; y-- {
		entries = append(entries, perimeterEntry{
			pos: C(-1, y),
			dir: DirRight,
		})
	}

	return entries
}

// MakeShooterAt creates a single shooter at a specific position.
func MakeShooterAt(id int, pos Coord, dir Dir, color Color) Shooter {
	return Shooter{
		ID:    id,
		Pos:   pos,
		Dir:   dir,
		Color: color,
	}
}

// MakeShootersFromSpec creates shooters from a specification slice.
// Each spec is (x, y, dir, colorIndex).
// Useful for level-specific shooter configurations.
func MakeShootersFromSpec(specs []ShooterSpec) []Shooter {
	shooters := make([]Shooter, len(specs))
	for i, spec := range specs {
		shooters[i] = Shooter{
			ID:    i,
			Pos:   C(spec.X, spec.Y),
			Dir:   spec.Dir,
			Color: spec.Color,
		}
	}
	return shooters
}

// ShooterSpec describes a shooter for level configuration.
type ShooterSpec struct {
	X     int
	Y     int
	Dir   Dir
	Color Color
}
