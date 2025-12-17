package core

import (
	"strings"
)

// Cell represents a single screen cell with character and color.
type Cell struct {
	Rune  rune
	Color Color
}

// Screen is a 2D character buffer for rendering game graphics.
// It decouples game rendering from the terminal, allowing games to draw
// using simple rune operations while the platform handles actual display.
type Screen struct {
	width  int
	height int
	cells  [][]Cell
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
	s.cells = make([][]Cell, s.height)
	for y := range s.cells {
		s.cells[y] = make([]Cell, s.width)
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

	// Copy old content (both rune and color)
	copyW := Min(oldW, width)
	copyH := Min(oldH, height)
	for y := range copyH {
		for x := range copyW {
			s.cells[y][x] = oldCells[y][x]
		}
	}
}

// Clear fills the entire screen with spaces and resets colors.
func (s *Screen) Clear() {
	for y := range s.cells {
		for x := range s.cells[y] {
			s.cells[y][x] = Cell{Rune: ' ', Color: ColorDefault}
		}
	}
}

// Fill fills the entire screen with the given rune (default color).
func (s *Screen) Fill(r rune) {
	for y := range s.cells {
		for x := range s.cells[y] {
			s.cells[y][x] = Cell{Rune: r, Color: ColorDefault}
		}
	}
}

// Set places a rune at the given position (preserves existing color).
// Out-of-bounds coordinates are silently ignored.
func (s *Screen) Set(x, y int, r rune) {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}
	s.cells[y][x].Rune = r
}

// SetWithColor places a rune with a specific color at the given position.
// Out-of-bounds coordinates are silently ignored.
func (s *Screen) SetWithColor(x, y int, r rune, c Color) {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}
	s.cells[y][x] = Cell{Rune: r, Color: c}
}

// SetColor sets the color at the given position without changing the rune.
// Out-of-bounds coordinates are silently ignored.
func (s *Screen) SetColor(x, y int, c Color) {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}
	s.cells[y][x].Color = c
}

// Get returns the rune at the given position.
// Returns space for out-of-bounds coordinates.
func (s *Screen) Get(x, y int) rune {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return ' '
	}
	return s.cells[y][x].Rune
}

// GetCell returns the full cell (rune + color) at the given position.
// Returns empty cell for out-of-bounds coordinates.
func (s *Screen) GetCell(x, y int) Cell {
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return Cell{Rune: ' ', Color: ColorDefault}
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

// DrawTextWithColor writes a colored string horizontally starting at (x, y).
// Characters that extend beyond screen bounds are clipped.
func (s *Screen) DrawTextWithColor(x, y int, text string, c Color) {
	for i, r := range text {
		s.SetWithColor(x+i, y, r, c)
	}
}

// DrawTextCentered draws text centered horizontally at the given y position.
func (s *Screen) DrawTextCentered(y int, text string) {
	x := (s.width - len(text)) / 2
	s.DrawText(x, y, text)
}

// DrawTextCenteredWithColor draws colored text centered horizontally.
func (s *Screen) DrawTextCenteredWithColor(y int, text string, c Color) {
	x := (s.width - len(text)) / 2
	s.DrawTextWithColor(x, y, text, c)
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
	for i := range length {
		s.Set(x+i, y, r)
	}
}

// DrawVLine draws a vertical line from (x, y) with the given length.
func (s *Screen) DrawVLine(x, y, length int, r rune) {
	for i := range length {
		s.Set(x, y+i, r)
	}
}

// String converts the screen buffer to a plain text string (no colors).
// Each row is joined with newlines.
func (s *Screen) String() string {
	var sb strings.Builder
	sb.Grow(s.width*s.height + s.height) // Pre-allocate for efficiency

	for y := range s.height {
		if y > 0 {
			sb.WriteRune('\n')
		}
		for x := range s.width {
			sb.WriteRune(s.cells[y][x].Rune)
		}
	}
	return sb.String()
}

// Row returns a copy of the specified row as a plain text string.
func (s *Screen) Row(y int) string {
	if y < 0 || y >= s.height {
		return strings.Repeat(" ", s.width)
	}
	runes := make([]rune, s.width)
	for x := range s.width {
		runes[x] = s.cells[y][x].Rune
	}
	return string(runes)
}
