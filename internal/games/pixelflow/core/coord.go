package core

import "fmt"

// Coord represents a 2D coordinate on the grid.
// X increases rightward, Y increases downward (screen coordinates).
type Coord struct {
	X int
	Y int
}

// C is a convenience constructor for Coord.
func C(x, y int) Coord {
	return Coord{X: x, Y: y}
}

// String returns a string representation.
func (c Coord) String() string {
	return fmt.Sprintf("(%d,%d)", c.X, c.Y)
}

// Add returns a new Coord offset by (dx, dy).
func (c Coord) Add(dx, dy int) Coord {
	return Coord{X: c.X + dx, Y: c.Y + dy}
}

// Step returns the coordinate one step in the given direction.
func (c Coord) Step(d Dir) Coord {
	dx, dy := d.Delta()
	return c.Add(dx, dy)
}

// Equal returns true if coordinates match.
func (c Coord) Equal(other Coord) bool {
	return c.X == other.X && c.Y == other.Y
}

// Dir represents a direction for movement and shooting.
type Dir uint8

const (
	DirUp Dir = iota
	DirRight
	DirDown
	DirLeft
)

// String returns direction name.
func (d Dir) String() string {
	switch d {
	case DirUp:
		return "up"
	case DirRight:
		return "right"
	case DirDown:
		return "down"
	case DirLeft:
		return "left"
	default:
		return "unknown"
	}
}

// Delta returns (dx, dy) for one step in this direction.
func (d Dir) Delta() (int, int) {
	switch d {
	case DirUp:
		return 0, -1
	case DirRight:
		return 1, 0
	case DirDown:
		return 0, 1
	case DirLeft:
		return -1, 0
	default:
		return 0, 0
	}
}

// Opposite returns the opposite direction.
func (d Dir) Opposite() Dir {
	switch d {
	case DirUp:
		return DirDown
	case DirRight:
		return DirLeft
	case DirDown:
		return DirUp
	case DirLeft:
		return DirRight
	default:
		return d
	}
}
