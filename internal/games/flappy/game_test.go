package flappy

import (
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

func TestGameDeterminism(t *testing.T) {
	// Test that given the same seed and inputs, the game produces identical results
	seed := int64(12345)
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     seed,
	}

	// Define a sequence of inputs
	// Jump every 15 ticks to try to stay airborne
	inputSequence := make([]core.InputFrame, 200)
	for i := range inputSequence {
		inputSequence[i] = core.NewInputFrame()
		if i%15 == 0 {
			inputSequence[i].Set(core.ActionJump)
		}
	}

	// Run game 1
	g1 := New()
	g1.Reset(cfg)
	var state1 core.GameState
	for _, in := range inputSequence {
		result := g1.Step(in)
		state1 = result.State
		if state1.GameOver {
			break
		}
	}

	// Run game 2 with same seed and inputs
	g2 := New()
	g2.Reset(cfg)
	var state2 core.GameState
	for _, in := range inputSequence {
		result := g2.Step(in)
		state2 = result.State
		if state2.GameOver {
			break
		}
	}

	// Both runs should have identical results
	if state1.Score != state2.Score {
		t.Errorf("Determinism failed: scores differ. Run1=%d, Run2=%d", state1.Score, state2.Score)
	}

	if g1.tickCount != g2.tickCount {
		t.Errorf("Determinism failed: tick counts differ. Run1=%d, Run2=%d", g1.tickCount, g2.tickCount)
	}
}

func TestGameReset(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     42,
	}

	g := New()
	g.Reset(cfg)

	// Play a few ticks
	for i := 0; i < 50; i++ {
		in := core.NewInputFrame()
		if i%10 == 0 {
			in.Set(core.ActionJump)
		}
		g.Step(in)
	}

	// Reset should clear state
	g.Reset(cfg)

	if g.score != 0 {
		t.Errorf("Reset should clear score, got %d", g.score)
	}
	if g.gameOver {
		t.Error("Reset should clear gameOver flag")
	}
	if g.paused {
		t.Error("Reset should clear paused flag")
	}
	if g.tickCount != 0 {
		t.Errorf("Reset should clear tickCount, got %d", g.tickCount)
	}
}

func TestGameJumpPhysics(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	initialY := g.playerY

	// Jump
	jumpInput := core.NewInputFrame()
	jumpInput.Set(core.ActionJump)
	g.Step(jumpInput)

	// Player should have moved up (negative Y direction)
	if g.playerY >= initialY {
		t.Errorf("Jump should move player up, was %f, now %f", initialY, g.playerY)
	}

	// Velocity should be negative (upward)
	if g.playerVel >= 0 {
		t.Errorf("Jump velocity should be negative, got %f", g.playerVel)
	}
}

func TestGameGravity(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	// Position player in the middle, no velocity
	g.playerY = 10
	g.playerVel = 0

	// Step without input
	noInput := core.NewInputFrame()
	g.Step(noInput)

	// Player should have fallen (positive Y direction due to gravity)
	if g.playerY <= 10 {
		t.Errorf("Gravity should pull player down, Y is still %f", g.playerY)
	}

	// Velocity should now be positive (downward)
	if g.playerVel <= 0 {
		t.Errorf("Velocity should be positive after gravity, got %f", g.playerVel)
	}
}

func TestGamePause(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	// Pause the game
	pauseInput := core.NewInputFrame()
	pauseInput.Set(core.ActionPause)
	g.Step(pauseInput)

	if !g.paused {
		t.Error("Game should be paused")
	}

	// Record state
	yBefore := g.playerY

	// Step while paused (without pause toggle)
	noInput := core.NewInputFrame()
	g.Step(noInput)

	// Physics should not update while paused
	if g.playerY != yBefore {
		t.Errorf("Player position should not change while paused, was %f, now %f", yBefore, g.playerY)
	}

	// Unpause
	g.Step(pauseInput)

	if g.paused {
		t.Error("Game should be unpaused")
	}
}

func TestGameOver(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	// Force player to hit the ground
	g.playerY = float64(cfg.ScreenH - 1)
	g.playerVel = 10 // Moving down fast

	noInput := core.NewInputFrame()
	result := g.Step(noInput)

	if !result.State.GameOver {
		t.Error("Game should be over when player hits ground")
	}
}

func TestGameRender(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	screen := core.NewScreen(cfg.ScreenW, cfg.ScreenH)
	g.Render(screen)

	// Check that screen has content (not all spaces)
	str := screen.String()
	hasContent := false
	for _, ch := range str {
		if ch != ' ' && ch != '\n' {
			hasContent = true
			break
		}
	}

	if !hasContent {
		t.Error("Render should draw something to the screen")
	}

	// Check that ground is drawn
	groundY := cfg.ScreenH - 1
	if screen.Get(0, groundY) != GroundChar {
		t.Errorf("Ground should be drawn at bottom, got %q", screen.Get(0, groundY))
	}
}

func TestPipeCollision(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	// Manually create a pipe right at the player
	g.pipes.pipes = append(g.pipes.pipes, Pipe{
		X:         PlayerX - 1, // Overlapping with player
		GapY:      0,           // Gap at top
		GapHeight: 5,           // Small gap
		Passed:    false,
	})

	// Position player to collide with bottom pipe section
	g.playerY = 15 // Below the gap

	noInput := core.NewInputFrame()
	result := g.Step(noInput)

	if !result.State.GameOver {
		t.Error("Game should be over when player hits pipe")
	}
}
