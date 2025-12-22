// Package core provides the core game logic for PixelFlow puzzle game.
// This package is UI-agnostic and deterministic.
package core

// Dir represents a direction for shooters and movement.
type Dir uint8

const (
	DirUp Dir = iota
	DirRight
	DirDown
	DirLeft
)

// String returns the string representation of a direction.
func (d Dir) String() string {
	switch d {
	case DirUp:
		return "Up"
	case DirRight:
		return "Right"
	case DirDown:
		return "Down"
	case DirLeft:
		return "Left"
	default:
		return "Unknown"
	}
}

// Delta returns the (dx, dy) offset for moving one step in this direction.
// Up decreases Y, Down increases Y (screen coordinates).
func (d Dir) Delta() (dx, dy int) {
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
