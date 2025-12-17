package snake

import (
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

func TestDeterminism(t *testing.T) {
	// Two games with the same seed should produce identical snapshots
	cfg := core.RuntimeConfig{
		Seed:    12345,
		ScreenW: 80,
		ScreenH: 24,
	}

	g1 := New()
	g1.Reset(cfg)

	g2 := New()
	g2.Reset(cfg)

	// Run both games with same inputs for N ticks
	input := core.NewInputFrame()
	for i := 0; i < 100; i++ {
		// Add some inputs at specific ticks
		input.Clear()
		if i == 20 {
			input.Set(core.ActionDown)
		}
		if i == 40 {
			input.Set(core.ActionLeft)
		}

		g1.Step(input)
		g2.Step(input)
	}

	snap1 := g1.Snapshot()
	snap2 := g2.Snapshot()

	if snap1.Tick != snap2.Tick {
		t.Errorf("Tick mismatch: %d vs %d", snap1.Tick, snap2.Tick)
	}
	if snap1.Score != snap2.Score {
		t.Errorf("Score mismatch: %d vs %d", snap1.Score, snap2.Score)
	}
	if snap1.HeadX != snap2.HeadX || snap1.HeadY != snap2.HeadY {
		t.Errorf("Head position mismatch: (%d,%d) vs (%d,%d)",
			snap1.HeadX, snap1.HeadY, snap2.HeadX, snap2.HeadY)
	}
	if snap1.Dir != snap2.Dir {
		t.Errorf("Direction mismatch: %v vs %v", snap1.Dir, snap2.Dir)
	}
	if snap1.FoodX != snap2.FoodX || snap1.FoodY != snap2.FoodY {
		t.Errorf("Food position mismatch: (%d,%d) vs (%d,%d)",
			snap1.FoodX, snap1.FoodY, snap2.FoodX, snap2.FoodY)
	}
}

func TestNoImmediateReversal(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    42,
		ScreenW: 80,
		ScreenH: 24,
	}

	g := New()
	g.Reset(cfg)

	// Initial direction is right
	if g.direction != DirRight {
		t.Fatalf("Expected initial direction Right, got %v", g.direction)
	}

	// Try to go left (opposite) - should be ignored
	input := core.NewInputFrame()
	input.Set(core.ActionLeft)
	g.Step(input)

	// Direction should still be right (nextDir may be set but actual direction unchanged until move)
	if g.nextDir == DirLeft {
		t.Error("Should not allow immediate reversal from Right to Left")
	}

	// Now try valid direction change: down
	input.Clear()
	input.Set(core.ActionDown)
	g.Step(input)

	if g.nextDir != DirDown {
		t.Errorf("Expected nextDir to be Down, got %v", g.nextDir)
	}
}

func TestFoodSpawnValidity(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    999,
		ScreenW: 80,
		ScreenH: 24,
	}

	g := New()
	g.Reset(cfg)

	// Spawn food multiple times and verify it never lands on snake or walls
	for i := 0; i < 100; i++ {
		g.spawnFood()

		// Check food is not on a wall
		if g.walls[g.food] {
			t.Errorf("Food spawned on wall at (%d, %d)", g.food.X, g.food.Y)
		}

		// Check food is not on snake
		if g.isSnakeAt(g.food) {
			t.Errorf("Food spawned on snake at (%d, %d)", g.food.X, g.food.Y)
		}

		// Check food is within bounds
		if g.food.X < 0 || g.food.X >= g.mapWidth || g.food.Y < 0 || g.food.Y >= g.mapHeight {
			t.Errorf("Food spawned out of bounds at (%d, %d)", g.food.X, g.food.Y)
		}
	}
}

func TestLevelCompletion(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    123,
		ScreenW: 80,
		ScreenH: 24,
	}

	g := New()
	g.Reset(cfg)

	level := GetLevel(0)
	if level == nil {
		t.Fatal("Level 0 not found")
	}

	// Simulate eating enough food to complete level
	initialLevel := g.levelIndex
	for i := 0; i < level.TargetFood; i++ {
		g.foodEaten = i
		g.checkLevelCompletion()
	}
	g.foodEaten = level.TargetFood
	g.checkLevelCompletion()

	if !g.levelCleared {
		t.Error("Level should be cleared after eating TargetFood")
	}

	// Simulate level clear animation completing
	g.levelClearTicks = 90
	g.advanceLevel()

	if g.levelIndex != initialLevel+1 {
		t.Errorf("Expected level %d, got %d", initialLevel+1, g.levelIndex)
	}
}

func TestEndlessProgression(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    456,
		ScreenW: 80,
		ScreenH: 24,
	}

	g := NewEndless()
	g.Reset(cfg)

	initialSpeed := g.moveEveryTicks
	initialLevel := g.levelIndex

	// In endless mode, after 10 food, level should advance
	g.foodEaten = 10
	g.checkLevelCompletion()

	if g.levelIndex != initialLevel+1 {
		t.Errorf("Expected level %d, got %d after 10 food in endless", initialLevel+1, g.levelIndex)
	}

	// After cycling through all levels, speed should increase
	for i := 0; i < LevelCount(); i++ {
		g.foodEaten = 10
		g.checkLevelCompletion()
	}

	// Speed should have increased (moveEveryTicks decreased)
	if g.moveEveryTicks >= initialSpeed {
		t.Errorf("Expected speed increase (lower moveEveryTicks), got %d vs initial %d",
			g.moveEveryTicks, initialSpeed)
	}
}

func TestCollisionDetection(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    789,
		ScreenW: 80,
		ScreenH: 24,
	}

	g := New()
	g.Reset(cfg)

	// Store initial state
	initialGameOver := g.gameOver

	if initialGameOver {
		t.Fatal("Game should not start in game over state")
	}

	// Move snake into a wall
	// First, position snake near a wall
	g.snake = []Point{
		{X: 1, Y: 1}, // Head near top-left corner
		{X: 2, Y: 1},
		{X: 3, Y: 1},
	}
	g.direction = DirUp
	g.nextDir = DirUp

	// Execute move
	g.moveSnake()

	if !g.gameOver {
		t.Error("Game should be over after hitting wall")
	}
}

func TestSelfCollision(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    111,
		ScreenW: 80,
		ScreenH: 24,
	}

	g := New()
	g.Reset(cfg)

	// Create a snake that will collide with itself
	// Shape like a spiral that will hit itself
	g.snake = []Point{
		{X: 5, Y: 5}, // Head
		{X: 5, Y: 6},
		{X: 6, Y: 6},
		{X: 6, Y: 5},
		{X: 6, Y: 4},
	}
	g.direction = DirRight
	g.nextDir = DirRight

	// Move right would put head at (6, 5) which is occupied
	g.moveSnake()

	if !g.gameOver {
		t.Error("Game should be over after self collision")
	}
}

func TestSnakeGrowth(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    222,
		ScreenW: 80,
		ScreenH: 24,
	}

	g := New()
	g.Reset(cfg)

	initialLen := len(g.snake)

	// Place food directly in front of snake
	if len(g.snake) > 0 {
		head := g.snake[0]
		g.food = Point{X: head.X + 1, Y: head.Y}
		g.direction = DirRight
		g.nextDir = DirRight

		// Move to eat food
		g.moveSnake()

		if len(g.snake) != initialLen+1 {
			t.Errorf("Snake should grow by 1 after eating food, got %d vs %d",
				len(g.snake), initialLen+1)
		}

		if g.score != 1 {
			t.Errorf("Score should be 1 after eating food, got %d", g.score)
		}
	}
}

func TestLevelCount(t *testing.T) {
	if LevelCount() != 10 {
		t.Errorf("Expected 10 levels, got %d", LevelCount())
	}
}

func TestAllLevelsValid(t *testing.T) {
	for i := 0; i < LevelCount(); i++ {
		level := GetLevel(i)
		if level == nil {
			t.Errorf("Level %d is nil", i)
			continue
		}

		if level.ID != i+1 {
			t.Errorf("Level %d has wrong ID: %d", i, level.ID)
		}

		if level.Name == "" {
			t.Errorf("Level %d has empty name", i)
		}

		if level.TargetFood <= 0 {
			t.Errorf("Level %d has invalid TargetFood: %d", i, level.TargetFood)
		}

		if level.MoveEveryTicks <= 0 {
			t.Errorf("Level %d has invalid MoveEveryTicks: %d", i, level.MoveEveryTicks)
		}

		if len(level.Layout) == 0 {
			t.Errorf("Level %d has empty layout", i)
		}
	}
}

func TestGameIDs(t *testing.T) {
	campaign := New()
	if campaign.ID() != "snake" {
		t.Errorf("Campaign ID should be 'snake', got %s", campaign.ID())
	}

	endless := NewEndless()
	if endless.ID() != "snake_endless" {
		t.Errorf("Endless ID should be 'snake_endless', got %s", endless.ID())
	}
}

func TestTitles(t *testing.T) {
	campaign := New()
	if campaign.Title() != "Snake" {
		t.Errorf("Campaign title should be 'Snake', got %s", campaign.Title())
	}

	endless := NewEndless()
	if endless.Title() != "Snake (Endless)" {
		t.Errorf("Endless title should be 'Snake (Endless)', got %s", endless.Title())
	}
}

func TestWindowTooSmall(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    333,
		ScreenW: 10, // Too small
		ScreenH: 5,  // Too small
	}

	g := New()
	g.Reset(cfg)

	if !g.tooSmall {
		t.Error("Game should detect window is too small")
	}

	snap := g.Snapshot()
	if snap.State != StatePausedSmall {
		t.Errorf("State should be paused_small_window, got %s", snap.State)
	}
}

func TestRender(t *testing.T) {
	cfg := core.RuntimeConfig{
		Seed:    444,
		ScreenW: 80,
		ScreenH: 24,
	}

	g := New()
	g.Reset(cfg)

	screen := core.NewScreen(cfg.ScreenW, cfg.ScreenH)
	g.Render(screen)

	// Verify screen is not empty
	content := screen.String()
	if len(content) == 0 {
		t.Error("Rendered screen should not be empty")
	}

	// Verify HUD is present
	if !containsString(content, "Snake") {
		t.Error("HUD should contain 'Snake'")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
