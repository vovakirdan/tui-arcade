// Package core provides fundamental types and utilities for the arcade platform.
// It contains no external dependencies (especially no Bubble Tea) to keep game
// logic pure and testable.
package core

// Rect represents an axis-aligned bounding box used for collision detection.
type Rect struct {
	X, Y int // Top-left corner position
	W, H int // Width and height
}

// NewRect creates a new rectangle with the given position and dimensions.
func NewRect(x, y, w, h int) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

// Right returns the x-coordinate of the right edge.
func (r Rect) Right() int {
	return r.X + r.W
}

// Bottom returns the y-coordinate of the bottom edge.
func (r Rect) Bottom() int {
	return r.Y + r.H
}

// Intersects returns true if this rectangle overlaps with another.
// Uses standard AABB collision detection.
func (r Rect) Intersects(other Rect) bool {
	// No overlap if one rect is completely to the left, right, above, or below
	if r.X >= other.Right() || other.X >= r.Right() {
		return false
	}
	if r.Y >= other.Bottom() || other.Y >= r.Bottom() {
		return false
	}
	return true
}

// Contains returns true if the point (x, y) is inside this rectangle.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.Right() && y >= r.Y && y < r.Bottom()
}

// Center returns the center point of the rectangle.
func (r Rect) Center() (int, int) {
	return r.X + r.W/2, r.Y + r.H/2
}

// Clamp restricts a value to be within [min, max].
func Clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// ClampF restricts a float64 value to be within [min, max].
func ClampF(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// Abs returns the absolute value of an integer.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Min returns the smaller of two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
