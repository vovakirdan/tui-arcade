package core

import "fmt"

// Coord represents a 2D coordinate on the grid.
// X increases to the right, Y increases downward (screen coordinates).
type Coord struct {
	X int
	Y int
}

// C is a convenience constructor for Coord.
func C(x, y int) Coord {
	return Coord{X: x, Y: y}
}

// String returns a string representation of the coordinate.
func (c Coord) String() string {
	return fmt.Sprintf("(%d,%d)", c.X, c.Y)
}

// Add returns a new Coord offset by (dx, dy).
func (c Coord) Add(dx, dy int) Coord {
	return Coord{X: c.X + dx, Y: c.Y + dy}
}

// AddCoord returns the sum of two coordinates.
func (c Coord) AddCoord(other Coord) Coord {
	return Coord{X: c.X + other.X, Y: c.Y + other.Y}
}

// Step returns a new Coord one step in the given direction.
func (c Coord) Step(d Dir) Coord {
	dx, dy := d.Delta()
	return c.Add(dx, dy)
}

// Equal returns true if two coordinates are the same.
func (c Coord) Equal(other Coord) bool {
	return c.X == other.X && c.Y == other.Y
}

// Manhattan returns the Manhattan distance to another coordinate.
func (c Coord) Manhattan(other Coord) int {
	dx := c.X - other.X
	dy := c.Y - other.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}
