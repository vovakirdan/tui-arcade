package core_test

import (
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
)

func TestNewGrid(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorRed,
		core.C(2, 2): core.ColorBlue,
		core.C(4, 4): core.ColorGreen,
	}

	g := core.NewGrid(5, 5, pixels)

	// Check dimensions
	if g.W != 5 || g.H != 5 {
		t.Errorf("expected 5x5 grid, got %dx%d", g.W, g.H)
	}

	// Check cells array length
	if len(g.Cells) != 25 {
		t.Errorf("expected 25 cells, got %d", len(g.Cells))
	}

	// Check filled pixels
	testCases := []struct {
		coord  core.Coord
		filled bool
		color  core.Color
	}{
		{core.C(0, 0), true, core.ColorRed},
		{core.C(2, 2), true, core.ColorBlue},
		{core.C(4, 4), true, core.ColorGreen},
		{core.C(1, 1), false, 0},
		{core.C(3, 3), false, 0},
	}

	for _, tc := range testCases {
		cell := g.Get(tc.coord)
		if cell.Filled != tc.filled {
			t.Errorf("at %v: expected filled=%v, got %v", tc.coord, tc.filled, cell.Filled)
		}
		if tc.filled && cell.Color != tc.color {
			t.Errorf("at %v: expected color=%v, got %v", tc.coord, tc.color, cell.Color)
		}
	}
}

func TestGridInBounds(t *testing.T) {
	g := core.NewEmptyGrid(5, 5)

	testCases := []struct {
		coord    core.Coord
		expected bool
	}{
		{core.C(0, 0), true},
		{core.C(4, 4), true},
		{core.C(2, 2), true},
		{core.C(-1, 0), false},
		{core.C(0, -1), false},
		{core.C(5, 0), false},
		{core.C(0, 5), false},
		{core.C(5, 5), false},
	}

	for _, tc := range testCases {
		result := g.InBounds(tc.coord)
		if result != tc.expected {
			t.Errorf("InBounds(%v): expected %v, got %v", tc.coord, tc.expected, result)
		}
	}
}

func TestGridSetAndGet(t *testing.T) {
	g := core.NewEmptyGrid(3, 3)

	// Set a pixel
	g.SetPixel(core.C(1, 1), core.ColorYellow)
	cell := g.Get(core.C(1, 1))
	if !cell.Filled || cell.Color != core.ColorYellow {
		t.Errorf("expected yellow pixel at (1,1), got filled=%v color=%v", cell.Filled, cell.Color)
	}

	// Clear the pixel
	g.SetEmpty(core.C(1, 1))
	cell = g.Get(core.C(1, 1))
	if cell.Filled {
		t.Errorf("expected empty cell at (1,1) after SetEmpty")
	}
}

func TestGridClone(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(1, 1): core.ColorRed,
	}
	g := core.NewGrid(3, 3, pixels)

	clone := g.Clone()

	// Verify clone is equal
	if !g.Equal(clone) {
		t.Error("clone should be equal to original")
	}

	// Modify original
	g.SetEmpty(core.C(1, 1))

	// Clone should be unchanged
	cloneCell := clone.Get(core.C(1, 1))
	if !cloneCell.Filled {
		t.Error("clone should not be affected by original modification")
	}
}

func TestGridFilledCount(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorRed,
		core.C(1, 1): core.ColorBlue,
		core.C(2, 2): core.ColorGreen,
	}
	g := core.NewGrid(5, 5, pixels)

	if g.FilledCount() != 3 {
		t.Errorf("expected 3 filled cells, got %d", g.FilledCount())
	}

	g.SetEmpty(core.C(1, 1))
	if g.FilledCount() != 2 {
		t.Errorf("expected 2 filled cells after removal, got %d", g.FilledCount())
	}
}

func TestGridIsCleared(t *testing.T) {
	g := core.NewEmptyGrid(3, 3)
	if !g.IsCleared() {
		t.Error("empty grid should be cleared")
	}

	g.SetPixel(core.C(1, 1), core.ColorRed)
	if g.IsCleared() {
		t.Error("grid with pixel should not be cleared")
	}

	g.SetEmpty(core.C(1, 1))
	if !g.IsCleared() {
		t.Error("grid should be cleared after removing last pixel")
	}
}

func TestCoordOperations(t *testing.T) {
	c := core.C(2, 3)

	// Test Add
	added := c.Add(1, -1)
	if added.X != 3 || added.Y != 2 {
		t.Errorf("Add(1,-1): expected (3,2), got (%d,%d)", added.X, added.Y)
	}

	// Test Step
	stepped := c.Step(core.DirRight)
	if stepped.X != 3 || stepped.Y != 3 {
		t.Errorf("Step(Right): expected (3,3), got (%d,%d)", stepped.X, stepped.Y)
	}

	stepped = c.Step(core.DirUp)
	if stepped.X != 2 || stepped.Y != 2 {
		t.Errorf("Step(Up): expected (2,2), got (%d,%d)", stepped.X, stepped.Y)
	}
}

func TestColorParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected core.Color
		ok       bool
	}{
		{"red", core.ColorRed, true},
		{"RED", core.ColorRed, true},
		{"Red", core.ColorRed, true},
		{"r", core.ColorRed, true},
		{"green", core.ColorGreen, true},
		{"blue", core.ColorBlue, true},
		{"yellow", core.ColorYellow, true},
		{"purple", core.ColorPurple, true},
		{"invalid", core.ColorRed, false},
	}

	for _, tc := range testCases {
		color, ok := core.ParseColor(tc.input)
		if ok != tc.ok {
			t.Errorf("ParseColor(%q): expected ok=%v, got %v", tc.input, tc.ok, ok)
		}
		if tc.ok && color != tc.expected {
			t.Errorf("ParseColor(%q): expected %v, got %v", tc.input, tc.expected, color)
		}
	}
}

func TestColorChar(t *testing.T) {
	testCases := []struct {
		color    core.Color
		expected rune
	}{
		{core.ColorRed, 'R'},
		{core.ColorGreen, 'G'},
		{core.ColorBlue, 'B'},
		{core.ColorYellow, 'Y'},
		{core.ColorPurple, 'P'},
	}

	for _, tc := range testCases {
		if tc.color.Char() != tc.expected {
			t.Errorf("Color %v: expected char '%c', got '%c'", tc.color, tc.expected, tc.color.Char())
		}
	}
}
