package core_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/levels"
)

// getTestdataPath returns path to testdata/levels.
func getTestdataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "testdata", "levels")
}

func TestLoaderLoadAll(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvls, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(lvls) < 2 {
		t.Errorf("expected at least 2 levels, got %d", len(lvls))
	}

	// Should be sorted by ID
	for i := 1; i < len(lvls); i++ {
		if lvls[i-1].ID >= lvls[i].ID {
			t.Errorf("levels not sorted: %s >= %s", lvls[i-1].ID, lvls[i].ID)
		}
	}
}

func TestLoaderLoadLevel01(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvl, err := loader.LoadByID("lvl01")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	if lvl.ID != "lvl01" {
		t.Errorf("expected ID 'lvl01', got %q", lvl.ID)
	}
	if lvl.Name != "Intro" {
		t.Errorf("expected Name 'Intro', got %q", lvl.Name)
	}
	if lvl.Width != 5 || lvl.Height != 5 {
		t.Errorf("expected 5x5, got %dx%d", lvl.Width, lvl.Height)
	}
	if lvl.Capacity != 5 {
		t.Errorf("expected capacity 5, got %d", lvl.Capacity)
	}

	// Check pixel count
	if len(lvl.Pixels) != 5 {
		t.Errorf("expected 5 pixels, got %d", len(lvl.Pixels))
	}
}

func TestLoaderLoadLevel02(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvl, err := loader.LoadByID("lvl02")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	if lvl.ID != "lvl02" {
		t.Errorf("expected ID 'lvl02', got %q", lvl.ID)
	}
	if lvl.Width != 10 || lvl.Height != 10 {
		t.Errorf("expected 10x10, got %dx%d", lvl.Width, lvl.Height)
	}

	// Should have multiple colors
	grid := lvl.ToGrid()
	colors := grid.ColorsPresent()
	if len(colors) < 4 {
		t.Errorf("expected at least 4 colors, got %d", len(colors))
	}
}

func TestLoaderNotFound(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	_, err := loader.LoadByID("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent level")
	}
}

func TestLevelToGrid(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvl, err := loader.LoadByID("lvl01")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	g := lvl.ToGrid()

	if g.W != lvl.Width || g.H != lvl.Height {
		t.Errorf("grid dimensions mismatch")
	}

	if g.FilledCount() != len(lvl.Pixels) {
		t.Errorf("filled count mismatch: %d vs %d", g.FilledCount(), len(lvl.Pixels))
	}

	// Check specific pixel
	cell := g.Get(core.C(0, 0))
	if !cell.Filled || cell.Color != core.ColorPink {
		t.Errorf("expected pink at (0,0)")
	}
}

func TestLevelToRail(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvl, err := loader.LoadByID("lvl01")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	rail := lvl.ToRail()

	expectedLen := 2 * (lvl.Width + lvl.Height)
	if rail.Len() != expectedLen {
		t.Errorf("expected rail length %d, got %d", expectedLen, rail.Len())
	}
}

func TestLoaderListIDs(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	ids, err := loader.ListIDs()
	if err != nil {
		t.Fatalf("ListIDs failed: %v", err)
	}

	if len(ids) < 2 {
		t.Errorf("expected at least 2 IDs, got %d", len(ids))
	}

	// Should be sorted
	for i := 1; i < len(ids); i++ {
		if ids[i-1] >= ids[i] {
			t.Errorf("IDs not sorted: %s >= %s", ids[i-1], ids[i])
		}
	}
}

func TestLoaderDeterministicOrder(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvls1, _ := loader.LoadAll()
	lvls2, _ := loader.LoadAll()

	for i := range lvls1 {
		if lvls1[i].ID != lvls2[i].ID {
			t.Errorf("order not deterministic at %d", i)
		}
	}
}

func TestLevelNewState(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvl, err := loader.LoadByID("lvl01")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	deck := []core.Shooter{
		{ID: 0, Color: core.ColorPink, Ammo: 2},
	}

	state := lvl.NewState(deck)

	if state.Capacity != lvl.Capacity {
		t.Errorf("state capacity mismatch")
	}
	if state.Deck.TotalShooters() != 1 {
		t.Errorf("expected 1 shooter in deck, got %d", state.Deck.TotalShooters())
	}
	if state.Grid.FilledCount() != len(lvl.Pixels) {
		t.Errorf("grid pixel count mismatch")
	}
}

func TestColorParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected core.Color
		ok       bool
	}{
		{"pink", core.ColorPink, true},
		{"PINK", core.ColorPink, true},
		{"p", core.ColorPink, true},
		{"cyan", core.ColorCyan, true},
		{"green", core.ColorGreen, true},
		{"yellow", core.ColorYellow, true},
		{"purple", core.ColorPurple, true},
		{"invalid", core.ColorPink, false},
	}

	for _, tc := range tests {
		color, ok := core.ParseColor(tc.input)
		if ok != tc.ok {
			t.Errorf("ParseColor(%q): expected ok=%v, got %v", tc.input, tc.ok, ok)
		}
		if tc.ok && color != tc.expected {
			t.Errorf("ParseColor(%q): expected %v, got %v", tc.input, tc.expected, color)
		}
	}
}
