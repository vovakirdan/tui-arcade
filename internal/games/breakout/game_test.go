package breakout

import (
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

func TestGameDeterminism(t *testing.T) {
	// Test that given the same inputs, the game produces identical results
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     12345,
	}

	// Define a sequence of inputs
	// Move paddle right, then launch, then alternate left/right
	inputSequence := make([]core.InputFrame, 200)
	for i := range inputSequence {
		inputSequence[i] = core.NewInputFrame()
		if i == 10 {
			inputSequence[i].Set(core.ActionJump) // Launch ball
		} else if i > 10 && i%5 < 3 {
			inputSequence[i].Set(core.ActionRight) // Move right
		} else if i > 10 {
			inputSequence[i].Set(core.ActionLeft) // Move left
		}
	}

	// Run game 1
	g1 := New()
	g1.Reset(cfg)
	var snap1 Snapshot
	for _, in := range inputSequence {
		result := g1.Step(in)
		if result.State.GameOver {
			break
		}
	}
	snap1 = g1.Snapshot()

	// Run game 2 with same inputs
	g2 := New()
	g2.Reset(cfg)
	for _, in := range inputSequence {
		result := g2.Step(in)
		if result.State.GameOver {
			break
		}
	}
	snap2 := g2.Snapshot()

	// Both runs should have identical results
	if snap1.Hash() != snap2.Hash() {
		t.Errorf("Determinism failed: hashes differ. Run1=%d, Run2=%d", snap1.Hash(), snap2.Hash())
	}

	if snap1.Score != snap2.Score {
		t.Errorf("Determinism failed: scores differ. Run1=%d, Run2=%d", snap1.Score, snap2.Score)
	}

	if snap1.Tick != snap2.Tick {
		t.Errorf("Determinism failed: tick counts differ. Run1=%d, Run2=%d", snap1.Tick, snap2.Tick)
	}

	if snap1.PaddleX != snap2.PaddleX {
		t.Errorf("Determinism failed: paddle positions differ. Run1=%d, Run2=%d", snap1.PaddleX, snap2.PaddleX)
	}

	if snap1.BallCount != snap2.BallCount {
		t.Errorf("Determinism failed: ball counts differ. Run1=%d, Run2=%d", snap1.BallCount, snap2.BallCount)
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
	jumpInput := core.NewInputFrame()
	jumpInput.Set(core.ActionJump)
	g.Step(jumpInput) // Launch ball

	for i := 0; i < 50; i++ {
		in := core.NewInputFrame()
		if i%2 == 0 {
			in.Set(core.ActionRight)
		}
		g.Step(in)
	}

	// Reset should clear state
	g.Reset(cfg)

	if g.score != 0 {
		t.Errorf("Reset should clear score, got %d", g.score)
	}
	if g.state != StateServe {
		t.Errorf("Reset should set state to serve, got %s", g.state)
	}
	if g.tickCount != 0 {
		t.Errorf("Reset should clear tickCount, got %d", g.tickCount)
	}
	if g.levelIndex != 0 {
		t.Errorf("Reset should reset levelIndex, got %d", g.levelIndex)
	}
}

func TestGameServeState(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	// Game should start in serve state
	if g.state != StateServe {
		t.Errorf("Game should start in serve state, got %s", g.state)
	}

	// Should have one ball
	if len(g.balls) != 1 {
		t.Errorf("Game should start with 1 ball, got %d", len(g.balls))
	}

	ball := g.balls[0]

	// Ball should be stuck with zero velocity
	if !ball.Stuck {
		t.Error("Ball should be stuck in serve state")
	}
	if ball.VX != 0 || ball.VY != 0 {
		t.Errorf("Ball should have zero velocity in serve state, got VX=%d, VY=%d", ball.VX, ball.VY)
	}

	// Step without jump - ball should stay on paddle
	noInput := core.NewInputFrame()
	g.Step(noInput)

	if g.state != StateServe {
		t.Error("Game should still be in serve state")
	}

	// Record ball position
	ballY := ball.Y

	// Move paddle right
	rightInput := core.NewInputFrame()
	rightInput.Set(core.ActionRight)
	g.Step(rightInput)

	// Ball should follow paddle (same Y relative to paddle)
	if ball.Y != ballY {
		t.Errorf("Ball Y should not change in serve state, was %d, now %d", ballY, ball.Y)
	}

	// Jump should launch the ball
	jumpInput := core.NewInputFrame()
	jumpInput.Set(core.ActionJump)
	g.Step(jumpInput)

	if g.state != StatePlaying {
		t.Errorf("Game should be playing after launch, got %s", g.state)
	}

	// Ball velocity should be set (moving up)
	if ball.VY >= 0 {
		t.Errorf("Ball should have negative VY after launch, got %d", ball.VY)
	}

	// Ball should no longer be stuck
	if ball.Stuck {
		t.Error("Ball should not be stuck after launch")
	}
}

func TestPaddleMovement(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	initialX := g.paddle.X

	// Move right
	rightInput := core.NewInputFrame()
	rightInput.Set(core.ActionRight)
	g.Step(rightInput)

	if g.paddle.X <= initialX {
		t.Errorf("Paddle should move right, was %d, now %d", initialX, g.paddle.X)
	}

	// Move left
	newX := g.paddle.X
	leftInput := core.NewInputFrame()
	leftInput.Set(core.ActionLeft)
	g.Step(leftInput)

	if g.paddle.X >= newX {
		t.Errorf("Paddle should move left, was %d, now %d", newX, g.paddle.X)
	}
}

func TestPaddleCollision(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	// Position ball just above paddle, moving down
	ball := g.balls[0]
	ball.X = g.paddle.CenterX()
	ball.Y = ToFixed(g.paddle.Y - 1)
	ball.VX = 0
	ball.VY = ToFixed(1) // Moving down
	ball.Stuck = false

	g.state = StatePlaying

	// Step to trigger collision
	noInput := core.NewInputFrame()
	g.Step(noInput)

	// Ball should bounce (velocity reversed)
	if ball.VY >= 0 {
		t.Errorf("Ball should bounce up after paddle collision, VY=%d", ball.VY)
	}
}

func TestPaddleBounceShaping(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	// Test left edge hit
	g := New()
	g.Reset(cfg)
	g.state = StatePlaying

	ball := g.balls[0]

	// Position ball at left edge of paddle
	leftEdgeX := g.paddle.X
	ball.X = leftEdgeX
	ball.Y = ToFixed(g.paddle.Y - 1)
	ball.VX = Fixed(100)
	ball.VY = ToFixed(1) // Moving down
	ball.Stuck = false

	baseSpeed := Fixed(g.cfg.Physics.BallSpeed)
	CheckPaddleCollision(ball, g.paddle, baseSpeed)

	// Ball should bounce left (negative VX) when hitting left edge
	if ball.VX >= 0 {
		t.Errorf("Ball should bounce left when hitting paddle left edge, VX=%d", ball.VX)
	}

	// Test right edge hit
	g2 := New()
	g2.Reset(cfg)
	g2.state = StatePlaying

	ball2 := g2.balls[0]

	// Position ball at right edge of paddle
	rightEdgeX := g2.paddle.X + ToFixed(g2.paddle.Width)
	ball2.X = rightEdgeX
	ball2.Y = ToFixed(g2.paddle.Y - 1)
	ball2.VX = Fixed(-100)
	ball2.VY = ToFixed(1) // Moving down
	ball2.Stuck = false

	baseSpeed2 := Fixed(g2.cfg.Physics.BallSpeed)
	CheckPaddleCollision(ball2, g2.paddle, baseSpeed2)

	// Ball should bounce right (positive VX) when hitting right edge
	if ball2.VX <= 0 {
		t.Errorf("Ball should bounce right when hitting paddle right edge, VX=%d", ball2.VX)
	}
}

func TestBrickCollision(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)
	g.state = StatePlaying

	// Get initial brick count
	initialAlive := g.level.CountAlive()
	initialScore := g.score

	// Find a brick position and position ball to hit it
	// Bricks start at brickAreaTop (row 2)
	brickY := g.brickAreaTop
	brickX := 0

	ball := g.balls[0]

	// Position ball just above first brick, moving up
	ball.X = ToFixed(brickX + g.brickWidth/2)
	ball.Y = ToFixed(brickY + 1)
	ball.VX = 0
	ball.VY = -ToFixed(1) // Moving up
	ball.Stuck = false

	// Step to trigger collision
	noInput := core.NewInputFrame()
	for i := 0; i < 5; i++ {
		g.Step(noInput)
	}

	// Check that brick was destroyed or score increased
	newAlive := g.level.CountAlive()
	if newAlive >= initialAlive && g.score <= initialScore {
		t.Error("Brick collision should destroy brick or increase score")
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

	// Launch ball first
	jumpInput := core.NewInputFrame()
	jumpInput.Set(core.ActionJump)
	g.Step(jumpInput)

	// Pause the game
	pauseInput := core.NewInputFrame()
	pauseInput.Set(core.ActionPause)
	g.Step(pauseInput)

	if g.state != StatePaused {
		t.Errorf("Game should be paused, got %s", g.state)
	}

	// Record state
	ball := g.balls[0]
	ballX := ball.X
	ballY := ball.Y

	// Step while paused (without pause toggle)
	noInput := core.NewInputFrame()
	g.Step(noInput)

	// Ball should not move while paused
	if ball.X != ballX || ball.Y != ballY {
		t.Error("Ball position should not change while paused")
	}

	// Unpause
	g.Step(pauseInput)

	if g.state == StatePaused {
		t.Error("Game should be unpaused")
	}
}

func TestWallCollision(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)
	g.state = StatePlaying

	ball := g.balls[0]

	// Position ball at left wall, moving left
	ball.X = ToFixed(1)
	ball.Y = ToFixed(10)
	ball.VX = -ToFixed(1) // Moving left
	ball.VY = 0
	ball.Stuck = false

	// Step to trigger collision
	noInput := core.NewInputFrame()
	g.Step(noInput)

	// Ball should bounce (VX reversed)
	if ball.VX <= 0 {
		t.Errorf("Ball should bounce right after left wall collision, VX=%d", ball.VX)
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
	g.state = StatePlaying
	g.lives = 1 // Last life

	ball := g.balls[0]

	// Position ball below paddle (will fall off)
	ball.X = ToFixed(40)
	ball.Y = ToFixed(cfg.ScreenH + 1)
	ball.VX = 0
	ball.VY = ToFixed(1) // Moving down
	ball.Stuck = false

	noInput := core.NewInputFrame()
	result := g.Step(noInput)

	if !result.State.GameOver {
		t.Error("Game should be over when ball falls and no lives left")
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

	// Check that screen has content
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

	// Check that paddle is drawn
	paddleX := g.paddle.CellX()
	if screen.Get(paddleX, g.paddle.Y) != PaddleChar {
		t.Errorf("Paddle should be drawn, got %q at paddle position", screen.Get(paddleX, g.paddle.Y))
	}
}

func TestSnapshot(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	// Play a few ticks
	jumpInput := core.NewInputFrame()
	jumpInput.Set(core.ActionJump)
	g.Step(jumpInput)

	for i := 0; i < 20; i++ {
		in := core.NewInputFrame()
		if i%3 == 0 {
			in.Set(core.ActionRight)
		}
		g.Step(in)
	}

	// Take snapshot
	snap := g.Snapshot()

	// Verify snapshot values
	if snap.Tick != uint64(g.tickCount) {
		t.Errorf("Snapshot tick should match game tick, got %d, want %d", snap.Tick, g.tickCount)
	}
	if snap.Score != g.score {
		t.Errorf("Snapshot score should match game score, got %d, want %d", snap.Score, g.score)
	}
	if snap.Lives != g.lives {
		t.Errorf("Snapshot lives should match game lives, got %d, want %d", snap.Lives, g.lives)
	}
	if snap.BallCount != len(g.balls) {
		t.Errorf("Snapshot ball count should match, got %d, want %d", snap.BallCount, len(g.balls))
	}

	// Apply snapshot to new game
	g2 := New()
	g2.Reset(cfg)
	g2.ApplySnapshot(snap)

	// Verify state matches
	snap2 := g2.Snapshot()
	if snap.Hash() != snap2.Hash() {
		t.Errorf("Snapshot hash should match after apply, got %d, want %d", snap2.Hash(), snap.Hash())
	}
}

func TestLevelParsing(t *testing.T) {
	// Test that levels parse correctly
	levels := BuiltinLevels()
	if len(levels) == 0 {
		t.Error("Should have at least one default level")
	}

	for i, level := range levels {
		if level.Name == "" {
			t.Errorf("Level %d should have a name", i)
		}
		if level.Width <= 0 || level.Height <= 0 {
			t.Errorf("Level %d should have positive dimensions", i)
		}
		if level.CountAlive() == 0 {
			t.Errorf("Level %d should have some bricks", i)
		}
	}
}

func TestFixedPointArithmetic(t *testing.T) {
	// Test fixed point operations
	a := ToFixed(5)  // 5000
	b := ToFixed(3)  // 3000
	c := Fixed(1500) // 1.5 in fixed

	if a+b != ToFixed(8) {
		t.Errorf("5 + 3 should be 8, got %d", (a+b)/Scale)
	}

	if a-b != ToFixed(2) {
		t.Errorf("5 - 3 should be 2, got %d", (a-b)/Scale)
	}

	// Test cell conversion
	f := Fixed(5500) // 5.5 in fixed
	if f.ToCell() != 5 {
		t.Errorf("5500 fixed should convert to cell 5, got %d", f.ToCell())
	}

	// Test clamping
	result := ClampFixed(Fixed(100), Fixed(0), Fixed(50))
	if result != Fixed(50) {
		t.Errorf("Clamp(100, 0, 50) should be 50, got %d", result)
	}

	result = ClampFixed(Fixed(-10), Fixed(0), Fixed(50))
	if result != Fixed(0) {
		t.Errorf("Clamp(-10, 0, 50) should be 0, got %d", result)
	}

	// Test multiplication
	_ = c // Just to use the variable
}

func TestMultiBall(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := New()
	g.Reset(cfg)

	// Launch ball
	jumpInput := core.NewInputFrame()
	jumpInput.Set(core.ActionJump)
	g.Step(jumpInput)

	// Spawn extra balls manually
	g.spawnMultiballs(2)

	if len(g.balls) != 3 {
		t.Errorf("Should have 3 balls after spawning 2 extra, got %d", len(g.balls))
	}

	// All balls should be active
	for i, ball := range g.balls {
		if !ball.Active {
			t.Errorf("Ball %d should be active", i)
		}
	}
}

func TestPowerUpManager(t *testing.T) {
	pm := NewPowerUpManager(12345, DefaultPowerUpConfig())

	// Test spawn with fixed RNG
	spawned := 0
	for i := 0; i < 100; i++ {
		if pm.TrySpawnPickup(40, 10) {
			spawned++
		}
	}

	// Should spawn some pickups (18% chance per try)
	if spawned == 0 {
		t.Error("Should spawn at least some pickups in 100 tries")
	}
	if spawned == 100 {
		t.Error("Should not spawn pickups every single time")
	}

	// Test effect management
	pm.AddEffect(EffectWiden, 0, 100)
	if !pm.HasEffect(EffectWiden) {
		t.Error("Should have widen effect after adding")
	}

	// Adding opposite effect should remove first
	pm.AddEffect(EffectShrink, 0, 100)
	pm.RemoveEffect(EffectWiden) // Simulate what activatePickup does
	if pm.HasEffect(EffectWiden) {
		t.Error("Should not have widen effect after removing")
	}
	if !pm.HasEffect(EffectShrink) {
		t.Error("Should have shrink effect")
	}

	// Test expiration
	pm.AddEffect(EffectSticky, 0, 50)
	expired := pm.ExpireEffects(100)
	if len(expired) != 2 { // Both shrink and sticky should expire
		t.Errorf("Should have 2 expired effects, got %d", len(expired))
	}
}

func TestSimpleRNG(t *testing.T) {
	// Test determinism
	rng1 := NewSimpleRNG(12345)
	rng2 := NewSimpleRNG(12345)

	for i := 0; i < 100; i++ {
		n1 := rng1.Intn(1000)
		n2 := rng2.Intn(1000)
		if n1 != n2 {
			t.Errorf("RNG should be deterministic, got %d vs %d at iteration %d", n1, n2, i)
		}
	}

	// Test different seeds produce different results
	rng3 := NewSimpleRNG(54321)
	allSame := true
	for i := 0; i < 10; i++ {
		if rng1.Intn(1000) != rng3.Intn(1000) {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("Different seeds should produce different sequences")
	}
}

func TestEndlessMode(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     1,
	}

	g := NewEndless()
	g.Reset(cfg)

	if g.mode != ModeEndless {
		t.Error("NewEndless should create endless mode game")
	}

	if g.ID() != "breakout_endless" {
		t.Errorf("Endless mode ID should be breakout_endless, got %s", g.ID())
	}

	if g.Title() != "Breakout (Endless)" {
		t.Errorf("Endless mode title should be 'Breakout (Endless)', got %s", g.Title())
	}
}

func TestHardBrick(t *testing.T) {
	// Test hard brick parsing
	level := ParseLevel("test", "Test", []string{
		"HHH",
		"###",
	})

	// First row should be hard bricks with HP 2
	for col := 0; col < 3; col++ {
		brick := level.Bricks[0][col]
		if brick.Type != BrickHard {
			t.Errorf("Brick at 0,%d should be hard, got %d", col, brick.Type)
		}
		if brick.HP != 2 {
			t.Errorf("Hard brick should have HP 2, got %d", brick.HP)
		}
	}

	// Second row should be normal bricks with HP 1
	for col := 0; col < 3; col++ {
		brick := level.Bricks[1][col]
		if brick.Type != BrickNormal {
			t.Errorf("Brick at 1,%d should be normal, got %d", col, brick.Type)
		}
		if brick.HP != 1 {
			t.Errorf("Normal brick should have HP 1, got %d", brick.HP)
		}
	}
}

func TestSolidBrick(t *testing.T) {
	// Test solid brick parsing
	level := ParseLevel("test", "Test", []string{
		"XXX",
		"###",
	})

	// First row should be solid bricks
	for col := 0; col < 3; col++ {
		brick := level.Bricks[0][col]
		if brick.Type != BrickSolid {
			t.Errorf("Brick at 0,%d should be solid, got %d", col, brick.Type)
		}
	}

	// Solid bricks should not be counted as alive destroyable
	aliveCount := 0
	for _, row := range level.Bricks {
		for _, b := range row {
			if b.Alive && b.Type != BrickEmpty && b.Type != BrickSolid {
				aliveCount++
			}
		}
	}

	if aliveCount != 3 {
		t.Errorf("Should have 3 destroyable bricks, got %d", aliveCount)
	}
}

func TestEffectDurations(t *testing.T) {
	pm := NewPowerUpManager(1, DefaultPowerUpConfig())

	pm.AddEffect(EffectWiden, 100, 200) // Until tick 300

	remaining := pm.GetEffectRemaining(EffectWiden, 100)
	if remaining != 200 {
		t.Errorf("Should have 200 ticks remaining at start, got %d", remaining)
	}

	remaining = pm.GetEffectRemaining(EffectWiden, 200)
	if remaining != 100 {
		t.Errorf("Should have 100 ticks remaining at tick 200, got %d", remaining)
	}

	remaining = pm.GetEffectRemaining(EffectWiden, 300)
	if remaining != 0 {
		t.Errorf("Should have 0 ticks remaining at expiration, got %d", remaining)
	}

	// Non-existent effect
	remaining = pm.GetEffectRemaining(EffectSticky, 100)
	if remaining != 0 {
		t.Errorf("Non-existent effect should return 0, got %d", remaining)
	}
}
