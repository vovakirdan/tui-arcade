package core_test

import (
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
)

func TestShootHitsMatchingColor(t *testing.T) {
	// Create a 5x5 grid with a red pixel at (2,2)
	pixels := map[core.Coord]core.Color{
		core.C(2, 2): core.ColorRed,
	}
	g := core.NewGrid(5, 5, pixels)

	// Create a red shooter from the top, shooting down
	shooter := core.MakeShooterAt(0, core.C(2, -1), core.DirDown, core.ColorRed)

	result := core.Shoot(g, shooter)

	// Should hit and remove the pixel
	if !result.Hit {
		t.Error("expected shot to hit")
	}
	if !result.Removed {
		t.Error("expected pixel to be removed")
	}
	if result.HitAt != core.C(2, 2) {
		t.Errorf("expected hit at (2,2), got %v", result.HitAt)
	}
	if result.RemovedColor != core.ColorRed {
		t.Errorf("expected removed color red, got %v", result.RemovedColor)
	}

	// Verify pixel is actually removed from grid
	cell := g.Get(core.C(2, 2))
	if cell.Filled {
		t.Error("pixel should be removed from grid")
	}
}

func TestShootHitsNonMatchingColor(t *testing.T) {
	// Create a 5x5 grid with a blue pixel at (2,2)
	pixels := map[core.Coord]core.Color{
		core.C(2, 2): core.ColorBlue,
	}
	g := core.NewGrid(5, 5, pixels)

	// Create a red shooter from the top, shooting down
	shooter := core.MakeShooterAt(0, core.C(2, -1), core.DirDown, core.ColorRed)

	result := core.Shoot(g, shooter)

	// Should hit but NOT remove (blocked)
	if !result.Hit {
		t.Error("expected shot to hit")
	}
	if result.Removed {
		t.Error("expected pixel NOT to be removed (color mismatch)")
	}
	if !result.Blocked {
		t.Error("expected shot to be blocked")
	}
	if result.BlockedColor != core.ColorBlue {
		t.Errorf("expected blocked by blue, got %v", result.BlockedColor)
	}

	// Verify pixel is still there
	cell := g.Get(core.C(2, 2))
	if !cell.Filled {
		t.Error("pixel should still be in grid")
	}
}

func TestShootPassesThroughEmpty(t *testing.T) {
	// Create a 5x5 grid with a red pixel at (2,3) - not at (2,0), (2,1), (2,2)
	pixels := map[core.Coord]core.Color{
		core.C(2, 3): core.ColorRed,
	}
	g := core.NewGrid(5, 5, pixels)

	// Shooter from top
	shooter := core.MakeShooterAt(0, core.C(2, -1), core.DirDown, core.ColorRed)

	result := core.Shoot(g, shooter)

	// Should pass through empty cells and hit at (2,3)
	if !result.Hit {
		t.Error("expected shot to hit")
	}
	if result.HitAt != core.C(2, 3) {
		t.Errorf("expected hit at (2,3), got %v", result.HitAt)
	}
	if !result.Removed {
		t.Error("expected pixel to be removed")
	}

	// Path should include (2,0), (2,1), (2,2), (2,3)
	if len(result.Path) < 4 {
		t.Errorf("expected path length >= 4, got %d", len(result.Path))
	}
}

func TestShootExitsBounds(t *testing.T) {
	// Empty grid
	g := core.NewEmptyGrid(5, 5)

	// Shooter from top
	shooter := core.MakeShooterAt(0, core.C(2, -1), core.DirDown, core.ColorRed)

	result := core.Shoot(g, shooter)

	// Should exit bounds without hitting anything
	if result.Hit {
		t.Error("expected no hit on empty grid")
	}
	if result.Removed {
		t.Error("expected no removal on empty grid")
	}
	if !result.OutOfBounds {
		t.Error("expected OutOfBounds to be true")
	}
}

func TestShootAllDirections(t *testing.T) {
	// Test shooting from all four directions
	testCases := []struct {
		name       string
		pixelPos   core.Coord
		shooterPos core.Coord
		dir        core.Dir
	}{
		{"from_top", core.C(2, 2), core.C(2, -1), core.DirDown},
		{"from_bottom", core.C(2, 2), core.C(2, 5), core.DirUp},
		{"from_left", core.C(2, 2), core.C(-1, 2), core.DirRight},
		{"from_right", core.C(2, 2), core.C(5, 2), core.DirLeft},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pixels := map[core.Coord]core.Color{
				tc.pixelPos: core.ColorGreen,
			}
			g := core.NewGrid(5, 5, pixels)
			shooter := core.MakeShooterAt(0, tc.shooterPos, tc.dir, core.ColorGreen)

			result := core.Shoot(g, shooter)

			if !result.Hit {
				t.Errorf("%s: expected hit", tc.name)
			}
			if !result.Removed {
				t.Errorf("%s: expected removal", tc.name)
			}
			if result.HitAt != tc.pixelPos {
				t.Errorf("%s: expected hit at %v, got %v", tc.name, tc.pixelPos, result.HitAt)
			}
		})
	}
}

func TestShootPure(t *testing.T) {
	// Test that ShootPure doesn't modify the original grid
	pixels := map[core.Coord]core.Color{
		core.C(2, 2): core.ColorRed,
	}
	original := core.NewGrid(5, 5, pixels)
	shooter := core.MakeShooterAt(0, core.C(2, -1), core.DirDown, core.ColorRed)

	// Dereference to get value (ShootPure takes Grid by value)
	newGrid, result := core.ShootPure(*original, shooter)

	// Original should be unchanged
	if !original.Get(core.C(2, 2)).Filled {
		t.Error("original grid should not be modified")
	}

	// New grid should have pixel removed
	if newGrid.Get(core.C(2, 2)).Filled {
		t.Error("new grid should have pixel removed")
	}

	// Result should indicate removal
	if !result.Removed {
		t.Error("result should indicate removal")
	}
}

func TestShootStopsAtFirstHit(t *testing.T) {
	// Two pixels in a row, same color
	pixels := map[core.Coord]core.Color{
		core.C(2, 1): core.ColorRed,
		core.C(2, 3): core.ColorRed,
	}
	g := core.NewGrid(5, 5, pixels)
	shooter := core.MakeShooterAt(0, core.C(2, -1), core.DirDown, core.ColorRed)

	result := core.Shoot(g, shooter)

	// Should hit the first pixel at (2,1), not (2,3)
	if result.HitAt != core.C(2, 1) {
		t.Errorf("expected hit at (2,1), got %v", result.HitAt)
	}

	// Second pixel should still exist
	if !g.Get(core.C(2, 3)).Filled {
		t.Error("second pixel should still exist")
	}
}

func TestMakeShooters(t *testing.T) {
	shooters := core.MakeShooters(4, 0, 5, 5)

	if len(shooters) != 4 {
		t.Errorf("expected 4 shooters, got %d", len(shooters))
	}

	// Each shooter should have unique ID
	ids := make(map[int]bool)
	for _, s := range shooters {
		if ids[s.ID] {
			t.Errorf("duplicate shooter ID: %d", s.ID)
		}
		ids[s.ID] = true
	}

	// Shooters should be deterministic
	shooters2 := core.MakeShooters(4, 0, 5, 5)
	for i := range shooters {
		if shooters[i].Pos != shooters2[i].Pos {
			t.Errorf("shooter %d position not deterministic", i)
		}
		if shooters[i].Dir != shooters2[i].Dir {
			t.Errorf("shooter %d direction not deterministic", i)
		}
	}
}

func TestCanShooterRemove(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(2, 2): core.ColorRed,
		core.C(2, 1): core.ColorBlue, // Blocking pixel
	}
	g := core.NewGrid(5, 5, pixels)

	// Red shooter blocked by blue pixel
	redShooter := core.MakeShooterAt(0, core.C(2, -1), core.DirDown, core.ColorRed)
	if core.CanShooterRemove(g, redShooter) {
		t.Error("red shooter should be blocked by blue pixel")
	}

	// Blue shooter can hit the blue pixel
	blueShooter := core.MakeShooterAt(1, core.C(2, -1), core.DirDown, core.ColorBlue)
	if !core.CanShooterRemove(g, blueShooter) {
		t.Error("blue shooter should be able to hit blue pixel")
	}

	// Green shooter has no target
	greenShooter := core.MakeShooterAt(2, core.C(2, -1), core.DirDown, core.ColorGreen)
	if core.CanShooterRemove(g, greenShooter) {
		t.Error("green shooter has no target")
	}
}

func TestSimulateSequence(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(2, 2): core.ColorRed,
		core.C(3, 2): core.ColorGreen,
	}
	g := core.NewGrid(5, 5, pixels)

	shooters := []core.Shooter{
		core.MakeShooterAt(0, core.C(2, -1), core.DirDown, core.ColorRed),
		core.MakeShooterAt(1, core.C(3, -1), core.DirDown, core.ColorGreen),
	}

	// Fire both shooters in sequence
	results := core.SimulateSequence(g, shooters, []int{0, 1})

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Both should have removed pixels
	for i, r := range results {
		if !r.Removed {
			t.Errorf("shot %d should have removed a pixel", i)
		}
	}

	// Grid should be empty
	if !g.IsCleared() {
		t.Error("grid should be cleared after shooting sequence")
	}
}
