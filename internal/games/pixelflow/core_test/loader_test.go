package core_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/levels"
)

// getTestdataPath returns the path to testdata/levels relative to this test file.
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

	// Should load at least 2 levels (level1.yaml and level2.yaml)
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

func TestLoaderLoadLevel1(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvl, err := loader.LoadByID("lvl_001")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	// Check basic properties
	if lvl.ID != "lvl_001" {
		t.Errorf("expected ID 'lvl_001', got %q", lvl.ID)
	}
	if lvl.Name != "Warmup" {
		t.Errorf("expected Name 'Warmup', got %q", lvl.Name)
	}
	if lvl.Width != 5 || lvl.Height != 5 {
		t.Errorf("expected 5x5, got %dx%d", lvl.Width, lvl.Height)
	}

	// Check pixels
	expectedPixels := map[core.Coord]core.Color{
		core.C(2, 2): core.ColorRed,
		core.C(3, 2): core.ColorRed,
		core.C(2, 3): core.ColorGreen,
		core.C(1, 1): core.ColorBlue,
	}

	if len(lvl.Pixels) != len(expectedPixels) {
		t.Errorf("expected %d pixels, got %d", len(expectedPixels), len(lvl.Pixels))
	}

	for coord, expectedColor := range expectedPixels {
		if color, ok := lvl.Pixels[coord]; !ok {
			t.Errorf("missing pixel at %v", coord)
		} else if color != expectedColor {
			t.Errorf("at %v: expected %v, got %v", coord, expectedColor, color)
		}
	}
}

func TestLoaderLoadLevel2WithShooters(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	lvl, err := loader.LoadByID("lvl_002")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	// Check basic properties
	if lvl.ID != "lvl_002" {
		t.Errorf("expected ID 'lvl_002', got %q", lvl.ID)
	}
	if lvl.Width != 8 || lvl.Height != 8 {
		t.Errorf("expected 8x8, got %dx%d", lvl.Width, lvl.Height)
	}

	// Check shooters are loaded
	if len(lvl.Shooters) != 3 {
		t.Errorf("expected 3 shooters, got %d", len(lvl.Shooters))
	}

	// Check first shooter
	if len(lvl.Shooters) > 0 {
		s := lvl.Shooters[0]
		if s.X != 1 || s.Y != -1 {
			t.Errorf("shooter 0: expected pos (1,-1), got (%d,%d)", s.X, s.Y)
		}
		if s.Dir != core.DirDown {
			t.Errorf("shooter 0: expected dir Down, got %v", s.Dir)
		}
		if s.Color != core.ColorRed {
			t.Errorf("shooter 0: expected color Red, got %v", s.Color)
		}
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

	lvl, err := loader.LoadByID("lvl_001")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	g := lvl.ToGrid()

	// Check grid dimensions
	if g.W != lvl.Width || g.H != lvl.Height {
		t.Errorf("grid dimensions mismatch: level %dx%d, grid %dx%d",
			lvl.Width, lvl.Height, g.W, g.H)
	}

	// Check filled count matches pixel count
	if g.FilledCount() != len(lvl.Pixels) {
		t.Errorf("filled count mismatch: level %d, grid %d",
			len(lvl.Pixels), g.FilledCount())
	}

	// Check a specific pixel
	cell := g.Get(core.C(2, 2))
	if !cell.Filled || cell.Color != core.ColorRed {
		t.Errorf("expected red pixel at (2,2), got filled=%v color=%v",
			cell.Filled, cell.Color)
	}
}

func TestLevelMakeShootersDefault(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	// Level 1 has no shooters defined, should use defaults
	lvl, err := loader.LoadByID("lvl_001")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	shooters := lvl.MakeShooters(3)
	if len(shooters) != 3 {
		t.Errorf("expected 3 default shooters, got %d", len(shooters))
	}
}

func TestLevelMakeShootersFromSpec(t *testing.T) {
	loader := levels.NewLoader(getTestdataPath())

	// Level 2 has shooters defined
	lvl, err := loader.LoadByID("lvl_002")
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	shooters := lvl.MakeShooters(5) // param ignored when level has specs
	if len(shooters) != 3 {
		t.Errorf("expected 3 shooters from spec, got %d", len(shooters))
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

	// Load multiple times
	lvls1, _ := loader.LoadAll()
	lvls2, _ := loader.LoadAll()
	lvls3, _ := loader.LoadAll()

	// Order should be the same each time
	for i := range lvls1 {
		if lvls1[i].ID != lvls2[i].ID || lvls1[i].ID != lvls3[i].ID {
			t.Errorf("level order not deterministic at index %d", i)
		}
	}
}
