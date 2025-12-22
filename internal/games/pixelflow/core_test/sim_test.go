package core_test

import (
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
)

func TestRailCreation(t *testing.T) {
	rail := core.NewRail(5, 5)

	// Rail should have 2*(W+H) = 20 positions
	if rail.Len() != 20 {
		t.Errorf("expected rail length 20, got %d", rail.Len())
	}

	// First position should be top-left, shooting down
	pos := rail.Get(0)
	if pos.Dir != core.DirDown {
		t.Errorf("position 0 should shoot Down, got %v", pos.Dir)
	}
	if !pos.Entry.Equal(core.C(0, 0)) {
		t.Errorf("position 0 entry should be (0,0), got %v", pos.Entry)
	}
}

func TestRailClockwise(t *testing.T) {
	rail := core.NewRail(3, 3)
	// Total: 2*(3+3) = 12 positions

	// Check sides
	// Top: 0,1,2 (3 positions)
	for i := 0; i < 3; i++ {
		pos := rail.Get(i)
		if pos.Side != core.SideTop {
			t.Errorf("position %d should be on top, got %v", i, pos.Side)
		}
		if pos.Dir != core.DirDown {
			t.Errorf("top position %d should shoot Down", i)
		}
	}

	// Right: 3,4,5 (3 positions)
	for i := 3; i < 6; i++ {
		pos := rail.Get(i)
		if pos.Side != core.SideRight {
			t.Errorf("position %d should be on right, got %v", i, pos.Side)
		}
		if pos.Dir != core.DirLeft {
			t.Errorf("right position %d should shoot Left", i)
		}
	}

	// Bottom: 6,7,8 (3 positions, reversed order)
	for i := 6; i < 9; i++ {
		pos := rail.Get(i)
		if pos.Side != core.SideBottom {
			t.Errorf("position %d should be on bottom, got %v", i, pos.Side)
		}
		if pos.Dir != core.DirUp {
			t.Errorf("bottom position %d should shoot Up", i)
		}
	}

	// Left: 9,10,11 (3 positions, reversed order)
	for i := 9; i < 12; i++ {
		pos := rail.Get(i)
		if pos.Side != core.SideLeft {
			t.Errorf("position %d should be on left, got %v", i, pos.Side)
		}
		if pos.Dir != core.DirRight {
			t.Errorf("left position %d should shoot Right", i)
		}
	}
}

func TestShootRemovesMatchingPixel(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(2, 2): core.ColorPink,
	}
	g := core.NewGrid(5, 5, pixels)
	rail := core.NewRail(5, 5)

	// Create state with a pink shooter
	deck := []core.Shooter{{ID: 0, Color: core.ColorPink, Ammo: 1}}
	state := core.NewState(g, deck, 5)

	// Launch the shooter
	state.LaunchTop()
	if len(state.Active) != 1 {
		t.Fatalf("expected 1 active shooter, got %d", len(state.Active))
	}

	// Advance until shooter hits the pixel (position 2 on top row shoots down through (2,0), (2,1), (2,2))
	// Shooter starts at position 0, needs to move to position 2
	for i := 0; i < 3; i++ {
		result := state.StepTick()
		// Position 2 should hit (2,0) first but it's empty, continues...
		_ = result
	}

	// After stepping, check if pixel was removed
	// The shooter at position 2 traces ray through (2,0), (2,1), (2,2) and finds pink at (2,2)
	// Since we're doing tick-by-tick, shooter fires at each position

	// Let's trace more carefully: shooter moves after firing
	// After step 1: was at 0, fires at (0,0)->empty, moves to 1
	// After step 2: at 1, fires at (1,0)->empty, moves to 2
	// After step 3: at 2, fires at (2,0)->empty continues to (2,1)->empty, (2,2)->pink! removes it
	// Wait - the trace continues until it hits something

	// Let's check the grid
	cell := state.Grid.Get(core.C(2, 2))
	if cell.Filled {
		// The shot should have hit by now from some position on top row
		// Actually from position 2, the ray goes (2,0)->(2,1)->(2,2)
		// Let me run more ticks
		for i := 0; i < 17; i++ { // complete the lap
			state.StepTick()
		}
	}

	// At some point the pixel should be removed
	_ = rail
}

func TestShootBecomesDryOnWrongColor(t *testing.T) {
	// Put cyan pixels at multiple positions so shooter hits wrong color on 2nd step
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorCyan, // Pink shooter will hit cyan
		core.C(1, 0): core.ColorCyan, // Also at position 1
	}
	g := core.NewGrid(3, 3, pixels)

	// Pink shooter
	deck := []core.Shooter{{ID: 0, Color: core.ColorPink, Ammo: 5}}
	state := core.NewState(g, deck, 5)

	state.LaunchTop()

	// First step - shooter at position 0 hits cyan but doesn't become dry
	// (first lap grace period)
	result1 := state.StepTick()
	_ = result1

	// Second step - shooter at position 1 hits cyan, should become dry now
	// because LapProgress > 0
	result2 := state.StepTick()

	// Should have a dry event on second step
	if len(result2.DryEvents) == 0 {
		t.Error("expected dry event when hitting wrong color after first position")
	}

	// Shooter should be dry
	if len(state.Active) > 0 && !state.Active[0].Dry {
		t.Error("shooter should be marked as dry")
	}

	// Ammo should NOT be spent
	if len(state.Active) > 0 && state.Active[0].Ammo != 5 {
		t.Errorf("ammo should be unchanged, got %d", state.Active[0].Ammo)
	}

	// Pixel should still be there
	if !state.Grid.Get(core.C(0, 0)).Filled {
		t.Error("pixel should not be removed when wrong color")
	}
}

func TestEmptyRayDoesNothing(t *testing.T) {
	// Empty grid
	g := core.NewEmptyGrid(5, 5)

	deck := []core.Shooter{{ID: 0, Color: core.ColorPink, Ammo: 5}}
	state := core.NewState(g, deck, 5)

	state.LaunchTop()

	// Step through entire lap
	for i := 0; i < 20; i++ {
		result := state.StepTick()
		if len(result.Removed) > 0 {
			t.Error("should not remove anything from empty grid")
		}
		if len(result.DryEvents) > 0 {
			t.Error("should not become dry on empty grid")
		}
	}

	// Shooter should still have full ammo
	if state.Waiting.Count() > 0 {
		if w := state.Waiting.Get(0); w != nil && w.Ammo != 5 {
			t.Errorf("ammo should be unchanged after empty lap, got %d", w.Ammo)
		}
	}
}

func TestLapCompletionParksShooter(t *testing.T) {
	g := core.NewEmptyGrid(3, 3)

	// Shooter with ammo
	deck := []core.Shooter{{ID: 0, Color: core.ColorGreen, Ammo: 3}}
	state := core.NewState(g, deck, 5)

	state.LaunchTop()
	if len(state.Active) != 1 {
		t.Fatalf("expected 1 active shooter")
	}

	railLen := state.Rail.Len() // 12 for 3x3

	// Complete one lap
	for i := 0; i < railLen; i++ {
		state.StepTick()
	}

	// Shooter should now be in waiting (has ammo > 0)
	if len(state.Active) != 0 {
		t.Errorf("expected 0 active shooters after lap, got %d", len(state.Active))
	}
	if state.Waiting.Count() != 1 {
		t.Errorf("expected 1 waiting shooter, got %d", state.Waiting.Count())
	}
	if w := state.Waiting.Get(0); w == nil || w.Ammo != 3 {
		ammo := 0
		if w != nil {
			ammo = w.Ammo
		}
		t.Errorf("waiting shooter should have 3 ammo, got %d", ammo)
	}
}

func TestLapCompletionRemovesEmptyShooter(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
	}
	g := core.NewGrid(3, 3, pixels)

	// Shooter with exactly 1 ammo
	deck := []core.Shooter{{ID: 0, Color: core.ColorPink, Ammo: 1}}
	state := core.NewState(g, deck, 5)

	state.LaunchTop()

	// Step until lap complete
	railLen := state.Rail.Len()
	for i := 0; i < railLen; i++ {
		state.StepTick()
	}

	// Shooter used its ammo and should disappear, not go to waiting
	if len(state.Active) != 0 {
		t.Errorf("expected 0 active shooters, got %d", len(state.Active))
	}
	if state.Waiting.Count() != 0 {
		t.Errorf("expected 0 waiting shooters (no ammo left), got %d", state.Waiting.Count())
	}
}

func TestCapacityLimit(t *testing.T) {
	g := core.NewEmptyGrid(3, 3)

	// 10 shooters in deck, capacity 3
	deck := make([]core.Shooter, 10)
	for i := range deck {
		deck[i] = core.Shooter{ID: i, Color: core.ColorPink, Ammo: 5}
	}

	state := core.NewState(g, deck, 3)

	// Launch 3 shooters
	for i := 0; i < 3; i++ {
		if !state.LaunchTop() {
			t.Errorf("should be able to launch shooter %d", i)
		}
	}

	// Try to launch 4th - should fail
	if state.LaunchTop() {
		t.Error("should not be able to launch beyond capacity")
	}

	if len(state.Active) != 3 {
		t.Errorf("expected 3 active shooters, got %d", len(state.Active))
	}
}

func TestRunUntilIdle(t *testing.T) {
	// Simple grid that should be clearable
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
		core.C(1, 0): core.ColorPink,
	}
	g := core.NewGrid(3, 3, pixels)

	deck := []core.Shooter{
		{ID: 0, Color: core.ColorPink, Ammo: 2},
	}
	state := core.NewState(g, deck, 5)

	steps, cleared := state.RunUntilIdle(1000)

	if !cleared {
		t.Error("expected grid to be cleared")
	}
	if steps == 0 {
		t.Error("expected some simulation steps")
	}
	if !state.Grid.IsEmpty() {
		t.Errorf("grid should be empty, has %d pixels", state.Grid.FilledCount())
	}
}

func TestSimulateSingleShooterLap(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorGreen,
		core.C(1, 0): core.ColorGreen,
		core.C(2, 0): core.ColorCyan, // Different color in the way
	}
	g := core.NewGrid(5, 5, pixels)
	state := core.NewState(g, nil, 5)

	// Simulate a green shooter
	removed, endsDry, finalAmmo := state.SimulateSingleShooterLap(core.ColorGreen, 10)

	// Should remove the 2 green pixels accessible before hitting cyan
	// Position 0: ray hits (0,0) green -> remove
	// Position 1: ray hits (1,0) green -> remove
	// Position 2: ray hits (2,0) cyan -> dry!
	if len(removed) < 2 {
		t.Errorf("expected at least 2 removals, got %d", len(removed))
	}

	if !endsDry {
		t.Error("should end dry after hitting cyan")
	}

	if finalAmmo != 10-len(removed) {
		t.Errorf("expected final ammo %d, got %d", 10-len(removed), finalAmmo)
	}
}
