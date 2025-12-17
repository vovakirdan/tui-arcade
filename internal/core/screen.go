package core

import (
	"strings"
)

// Screen is a 2D character buffer for rendering game graphics.
// It decouples game rendering from the terminal, allowing games to draw
// using simple rune operations while the platform handles actual display.
type Screen struct {
	width  int
	height int
	cells  [][]rune
}

// NewScreen creates a new screen buffer with the given dimensions.
func NewScreen(width, height int) *Screen {
	s := &Screen{
		width:  width,
		height: height,
	}
	s.allocate()
	s.Clear()
	return s
}

// allocate creates the underlying cell storage.
func (s *Screen) allocate() {
	s.cells = make([][]rune, s.height)
	for y := range s.cells {
		s.cells[y] = make([]rune, s.width)
	}
}

// Width returns the screen width in characters.
func (s *Screen) Width() int {
	return s.width
}

// Height returns the screen height in characters.
func (s *Screen) Height() int {
	return s.height
}

// Resize changes the screen dimensions, preserving content where possible.
func (s *Screen) Resize(width, height int) {
	if width == s.width && height == s.height {
		return
	}

	oldCells := s.cells
	oldW, oldH := s.width, s.height

	s.width = width
	s.height = height
	s.allocate()
	s.Clear()

	// Copy old content
	copyW := Min(oldW, width)
	copyH := Min(oldH, height)
	for y := 0; y < copyH; y++ {
		for x := 0; x < copyW; x++ {
			s.cells[y][x] = oldCells[y][x]
		}
	}
}

// Clear fills the entire screen with spaces.
func (s *Screen) Clear() {
	for y := range s.cells {
		for x := range s.cells[y] {
			s.cells[y][x] = ' '
		}
	}
}

// Fill fills the entire screen with the given rune.
func (s *Screen) Fill(r rune) {
	for y := range s.cells {
		for x := range s.cells[y] {
			s.cells[y][x] = r
		}
	}
}

// Set places a rune at the given position.
// Out-of-bounds coordinates are silently ignored.
func (s *Screen) Set(x, y int, r rune) {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}
	s.cells[y][x] = r
}

// Get returns the rune at the given position.
// Returns space for out-of-bounds coordinates.
func (s *Screen) Get(x, y int) rune {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return ' '
	}
	return s.cells[y][x]
}

// DrawText writes a string horizontally starting at (x, y).
// Characters that extend beyond screen bounds are clipped.
func (s *Screen) DrawText(x, y int, text string) {
	for i, r := range text {
		s.Set(x+i, y, r)
	}
}

// DrawTextCentered draws text centered horizontally at the given y position.
func (s *Screen) DrawTextCentered(y int, text string) {
	x := (s.width - len(text)) / 2
	s.DrawText(x, y, text)
}

// DrawRect fills a rectangular area with the given rune.
func (s *Screen) DrawRect(r Rect, fill rune) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			s.Set(x, y, fill)
		}
	}
}

// DrawBox draws a box outline using box-drawing characters.
func (s *Screen) DrawBox(r Rect) {
	// Corners
	s.Set(r.X, r.Y, '┌')
	s.Set(r.Right()-1, r.Y, '┐')
	s.Set(r.X, r.Bottom()-1, '└')
	s.Set(r.Right()-1, r.Bottom()-1, '┘')

	// Horizontal edges
	for x := r.X + 1; x < r.Right()-1; x++ {
		s.Set(x, r.Y, '─')
		s.Set(x, r.Bottom()-1, '─')
	}

	// Vertical edges
	for y := r.Y + 1; y < r.Bottom()-1; y++ {
		s.Set(r.X, y, '│')
		s.Set(r.Right()-1, y, '│')
	}
}

// DrawHLine draws a horizontal line from (x, y) with the given length.
func (s *Screen) DrawHLine(x, y, length int, r rune) {
	for i := 0; i < length; i++ {
		s.Set(x+i, y, r)
	}
}

// DrawVLine draws a vertical line from (x, y) with the given length.
func (s *Screen) DrawVLine(x, y, length int, r rune) {
	for i := 0; i < length; i++ {
		s.Set(x, y+i, r)
	}
}

// String converts the screen buffer to a renderable string.
// Each row is joined with newlines.
func (s *Screen) String() string {
	var sb strings.Builder
	sb.Grow(s.width*s.height + s.height) // Pre-allocate for efficiency

	for y := 0; y < s.height; y++ {
		if y > 0 {
			sb.WriteRune('\n')
		}
		for x := 0; x < s.width; x++ {
			sb.WriteRune(s.cells[y][x])
		}
	}
	return sb.String()
}

// Row returns a copy of the specified row as a string.
func (s *Screen) Row(y int) string {
	if y < 0 || y >= s.height {
		return strings.Repeat(" ", s.width)
	}
	return string(s.cells[y])
}
