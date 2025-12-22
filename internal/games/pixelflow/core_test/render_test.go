package core_test

import (
	"strings"
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
)

func TestRenderCompact(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorRed,
		core.C(2, 0): core.ColorGreen,
		core.C(1, 1): core.ColorBlue,
		core.C(2, 2): core.ColorYellow,
	}
	g := core.NewGrid(3, 3, pixels)

	expected := strings.TrimSpace(`
R.G
.B.
..Y
`)
	result := strings.TrimSpace(core.RenderCompact(g))

	if result != expected {
		t.Errorf("RenderCompact mismatch:\nexpected:\n%s\n\ngot:\n%s", expected, result)
	}
}

func TestRenderWithFrame(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(1, 1): core.ColorPurple,
	}
	g := core.NewGrid(3, 3, pixels)

	result := core.RenderWithFrame(g, "Test")

	// Check it contains frame elements
	if !strings.Contains(result, "|") {
		t.Error("expected frame to contain | characters")
	}
	if !strings.Contains(result, "-") {
		t.Error("expected frame to contain - characters")
	}
	if !strings.Contains(result, "Test") {
		t.Error("expected frame to contain title")
	}
	if !strings.Contains(result, "P") {
		t.Error("expected frame to contain purple pixel (P)")
	}
}

func TestRenderASCII(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(1, 1): core.ColorRed,
		core.C(2, 1): core.ColorGreen,
	}
	g := core.NewGrid(4, 4, pixels)

	opts := core.DefaultRenderOptions()
	result := core.RenderASCII(g, nil, opts)

	// Check that it contains the pixels
	if !strings.Contains(result, "R") {
		t.Error("expected output to contain R")
	}
	if !strings.Contains(result, "G") {
		t.Error("expected output to contain G")
	}
	if !strings.Contains(result, ".") {
		t.Error("expected output to contain empty cells (.)")
	}
}

func TestRenderASCIIWithShooters(t *testing.T) {
	g := core.NewEmptyGrid(3, 3)
	shooters := []core.Shooter{
		core.MakeShooterAt(0, core.C(1, -1), core.DirDown, core.ColorRed),
		core.MakeShooterAt(1, core.C(3, 1), core.DirLeft, core.ColorBlue),
	}

	opts := core.RenderOptions{
		ShowShooters: true,
	}
	result := core.RenderASCII(g, shooters, opts)

	// Should contain shooter arrows
	if !strings.Contains(result, "v") {
		t.Error("expected down arrow (v) for shooter pointing down")
	}
	if !strings.Contains(result, "<") {
		t.Error("expected left arrow (<) for shooter pointing left")
	}
}

func TestRenderASCIIWithLastShot(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(1, 2): core.ColorRed,
	}
	g := core.NewGrid(3, 3, pixels)

	// Simulate a shot path
	shotResult := &core.ShotResult{
		ShooterID: 0,
		Hit:       true,
		HitAt:     core.C(1, 2),
		Removed:   true,
		Path:      []core.Coord{core.C(1, 0), core.C(1, 1), core.C(1, 2)},
	}

	opts := core.RenderOptions{
		LastShot: shotResult,
	}
	result := core.RenderASCII(g, nil, opts)

	// Path should be marked with * or X
	if !strings.Contains(result, "*") && !strings.Contains(result, "X") {
		t.Error("expected shot path to be marked in render output")
	}
}

func TestRenderASCIIWithCoords(t *testing.T) {
	g := core.NewEmptyGrid(3, 3)

	opts := core.RenderOptions{
		ShowCoords: true,
	}
	result := core.RenderASCII(g, nil, opts)

	// Should contain coordinate numbers
	if !strings.Contains(result, "0") {
		t.Error("expected coordinates to include 0")
	}
}

func TestGridToLines(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorRed,
		core.C(1, 1): core.ColorGreen,
		core.C(2, 2): core.ColorBlue,
	}
	g := core.NewGrid(3, 3, pixels)

	lines := core.GridToLines(g)

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	expectedLines := []string{
		"R..",
		".G.",
		"..B",
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("line %d: expected %q, got %q", i, expected, lines[i])
		}
	}
}

func TestGoldenRender(t *testing.T) {
	// Test a known grid renders exactly as expected
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorRed,
		core.C(4, 0): core.ColorBlue,
		core.C(2, 2): core.ColorGreen,
		core.C(0, 4): core.ColorYellow,
		core.C(4, 4): core.ColorPurple,
	}
	g := core.NewGrid(5, 5, pixels)

	expected := `R...B
.....
..G..
.....
Y...P
`

	result := core.RenderCompact(g)
	if result != expected {
		t.Errorf("Golden render mismatch:\nexpected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestRenderEmptyCell(t *testing.T) {
	g := core.NewEmptyGrid(2, 2)

	opts := core.RenderOptions{
		EmptyChar: '_',
	}
	result := core.RenderASCII(g, nil, opts)

	if !strings.Contains(result, "_") {
		t.Error("expected custom empty char '_'")
	}
	if strings.Contains(result, ".") {
		t.Error("should not contain default empty char '.'")
	}
}
