package core

import (
	"strings"
	"testing"
)

func TestNewScreen(t *testing.T) {
	s := NewScreen(80, 24)

	if s.Width() != 80 {
		t.Errorf("Width() = %d, expected 80", s.Width())
	}
	if s.Height() != 24 {
		t.Errorf("Height() = %d, expected 24", s.Height())
	}

	// Check that it's initialized with spaces
	for y := 0; y < s.Height(); y++ {
		for x := 0; x < s.Width(); x++ {
			if s.Get(x, y) != ' ' {
				t.Errorf("New screen should be filled with spaces, got %q at (%d, %d)", s.Get(x, y), x, y)
			}
		}
	}
}

func TestScreenSetGet(t *testing.T) {
	s := NewScreen(10, 10)

	s.Set(5, 5, 'X')
	if s.Get(5, 5) != 'X' {
		t.Errorf("Get(5, 5) = %q, expected 'X'", s.Get(5, 5))
	}

	// Out of bounds should be silent
	s.Set(-1, 0, 'A')  // Should not panic
	s.Set(100, 0, 'A') // Should not panic
	s.Set(0, -1, 'A')  // Should not panic
	s.Set(0, 100, 'A') // Should not panic

	// Out of bounds get should return space
	if s.Get(-1, 0) != ' ' {
		t.Error("Out of bounds Get should return space")
	}
	if s.Get(100, 0) != ' ' {
		t.Error("Out of bounds Get should return space")
	}
}

func TestScreenClear(t *testing.T) {
	s := NewScreen(10, 10)

	// Fill with some characters
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			s.Set(x, y, 'X')
		}
	}

	s.Clear()

	// Should all be spaces now
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if s.Get(x, y) != ' ' {
				t.Errorf("After Clear, expected space at (%d, %d), got %q", x, y, s.Get(x, y))
			}
		}
	}
}

func TestScreenFill(t *testing.T) {
	s := NewScreen(5, 5)
	s.Fill('#')

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if s.Get(x, y) != '#' {
				t.Errorf("After Fill, expected '#' at (%d, %d), got %q", x, y, s.Get(x, y))
			}
		}
	}
}

func TestScreenDrawText(t *testing.T) {
	s := NewScreen(20, 5)
	s.DrawText(2, 1, "Hello")

	expected := "Hello"
	for i, ch := range expected {
		if s.Get(2+i, 1) != ch {
			t.Errorf("DrawText: expected %q at (%d, 1), got %q", ch, 2+i, s.Get(2+i, 1))
		}
	}

	// Text should be clipped at boundaries
	s.DrawText(18, 0, "Hello") // Only "He" should fit
	if s.Get(18, 0) != 'H' || s.Get(19, 0) != 'e' {
		t.Error("Text should be clipped at right boundary")
	}
}

func TestScreenDrawTextCentered(t *testing.T) {
	s := NewScreen(20, 5)
	text := "Hi"
	s.DrawTextCentered(2, text)

	// "Hi" is 2 chars, centered in 20 chars should start at position 9
	x := (20 - 2) / 2
	if s.Get(x, 2) != 'H' || s.Get(x+1, 2) != 'i' {
		t.Errorf("DrawTextCentered failed, text not at expected position")
	}
}

func TestScreenDrawRect(t *testing.T) {
	s := NewScreen(10, 10)
	r := NewRect(2, 2, 3, 3)
	s.DrawRect(r, '#')

	// Check filled area
	for y := 2; y < 5; y++ {
		for x := 2; x < 5; x++ {
			if s.Get(x, y) != '#' {
				t.Errorf("DrawRect: expected '#' at (%d, %d), got %q", x, y, s.Get(x, y))
			}
		}
	}

	// Check outside is still space
	if s.Get(1, 1) != ' ' {
		t.Error("DrawRect should not affect outside area")
	}
	if s.Get(5, 5) != ' ' {
		t.Error("DrawRect should not affect outside area")
	}
}

func TestScreenDrawBox(t *testing.T) {
	s := NewScreen(10, 10)
	r := NewRect(1, 1, 5, 4)
	s.DrawBox(r)

	// Check corners
	if s.Get(1, 1) != '┌' {
		t.Errorf("Top-left corner should be '┌', got %q", s.Get(1, 1))
	}
	if s.Get(5, 1) != '┐' {
		t.Errorf("Top-right corner should be '┐', got %q", s.Get(5, 1))
	}
	if s.Get(1, 4) != '└' {
		t.Errorf("Bottom-left corner should be '└', got %q", s.Get(1, 4))
	}
	if s.Get(5, 4) != '┘' {
		t.Errorf("Bottom-right corner should be '┘', got %q", s.Get(5, 4))
	}

	// Check horizontal edges
	for x := 2; x < 5; x++ {
		if s.Get(x, 1) != '─' {
			t.Errorf("Top edge should be '─' at x=%d, got %q", x, s.Get(x, 1))
		}
		if s.Get(x, 4) != '─' {
			t.Errorf("Bottom edge should be '─' at x=%d, got %q", x, s.Get(x, 4))
		}
	}

	// Check vertical edges
	for y := 2; y < 4; y++ {
		if s.Get(1, y) != '│' {
			t.Errorf("Left edge should be '│' at y=%d, got %q", y, s.Get(1, y))
		}
		if s.Get(5, y) != '│' {
			t.Errorf("Right edge should be '│' at y=%d, got %q", y, s.Get(5, y))
		}
	}
}

func TestScreenDrawHLine(t *testing.T) {
	s := NewScreen(10, 5)
	s.DrawHLine(2, 2, 5, '-')

	for x := 2; x < 7; x++ {
		if s.Get(x, 2) != '-' {
			t.Errorf("DrawHLine: expected '-' at (%d, 2), got %q", x, s.Get(x, 2))
		}
	}
}

func TestScreenDrawVLine(t *testing.T) {
	s := NewScreen(10, 10)
	s.DrawVLine(3, 2, 4, '|')

	for y := 2; y < 6; y++ {
		if s.Get(3, y) != '|' {
			t.Errorf("DrawVLine: expected '|' at (3, %d), got %q", y, s.Get(3, y))
		}
	}
}

func TestScreenString(t *testing.T) {
	s := NewScreen(5, 3)
	s.DrawText(0, 0, "AAAAA")
	s.DrawText(0, 1, "BBBBB")
	s.DrawText(0, 2, "CCCCC")

	result := s.String()
	expected := "AAAAA\nBBBBB\nCCCCC"

	if result != expected {
		t.Errorf("String() = %q, expected %q", result, expected)
	}
}

func TestScreenResize(t *testing.T) {
	s := NewScreen(10, 10)
	s.DrawText(0, 0, "Hello")
	s.DrawText(0, 5, "World")

	// Resize smaller - should preserve top-left content
	s.Resize(8, 4)
	if s.Width() != 8 || s.Height() != 4 {
		t.Errorf("After resize, dimensions should be 8x4, got %dx%d", s.Width(), s.Height())
	}

	row0 := s.Row(0)
	if !strings.HasPrefix(row0, "Hello") {
		t.Errorf("Content should be preserved, row 0 = %q", row0)
	}

	// Resize larger - old content should still be there
	s.Resize(15, 8)
	row0 = s.Row(0)
	if !strings.HasPrefix(row0, "Hello") {
		t.Errorf("Content should be preserved after enlarging, row 0 = %q", row0)
	}
}

func TestScreenRow(t *testing.T) {
	s := NewScreen(10, 5)
	s.DrawText(0, 2, "Test")

	row := s.Row(2)
	if !strings.HasPrefix(row, "Test") {
		t.Errorf("Row(2) should start with 'Test', got %q", row)
	}
	if len(row) != 10 {
		t.Errorf("Row length should be 10, got %d", len(row))
	}

	// Out of bounds row
	outOfBounds := s.Row(-1)
	if outOfBounds != "          " {
		t.Errorf("Out of bounds row should be spaces, got %q", outOfBounds)
	}
}
