package breakout

import (
	"fmt"

	"github.com/vovakirdan/tui-arcade/internal/config"
	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/registry"
)

// Visual characters for rendering
const (
	PaddleChar  = '='
	BallChar    = '●'
	BorderVert  = '│'
	BorderHoriz = '─'
	BorderTL    = '┌'
	BorderTR    = '┐'
	BorderBL    = '└'
	BorderBR    = '┘'
)

// Brick glyphs by row (cycling through)
var BrickGlyphs = []rune{'█', '▓', '▒', '░', '#', '+', '*', '='}

// Hard brick glyph
const HardBrickGlyph = '▓'

// Solid brick glyph
const SolidBrickGlyph = '█'

// GameState constants
const (
	StateServe    = "serve"    // Ball on paddle, waiting for launch
	StatePlaying  = "playing"  // Ball in play
	StateGameOver = "gameover" // No lives left
	StateWin      = "win"      // All levels completed (campaign only)
	StatePaused   = "paused"   // Game paused
)

// GameMode represents the game mode.
type GameMode int

const (
	ModeCampaign GameMode = iota // Play through levels, win at end
	ModeEndless                  // Play forever, score until game over
)

// configPath stores the custom config path set via CLI
var configPath string

// difficultyPreset stores the difficulty preset set via CLI
var difficultyPreset config.DifficultyPreset

// SetConfigPath sets the custom config path for loading.
func SetConfigPath(path string) {
	configPath = path
}

// SetDifficultyPreset sets the difficulty preset.
func SetDifficultyPreset(preset string) {
	switch preset {
	case "easy":
		difficultyPreset = config.DifficultyEasy
	case "normal":
		difficultyPreset = config.DifficultyNormal
	case "hard":
		difficultyPreset = config.DifficultyHard
	case "fixed":
		difficultyPreset = config.DifficultyFixed
	default:
		difficultyPreset = ""
	}
}

// Game implements the Breakout game logic.
type Game struct {
	// Game mode
	mode GameMode

	// Game objects
	paddle *Paddle
	balls  []*Ball // Multiple balls supported
	level  *Level

	// Power-up system
	powerups *PowerUpManager

	// Game state
	state            string
	score            int
	lives            int
	levelIndex       int
	tickCount        int
	serveDelay       int   // Countdown before allowing serve after miss
	bricksTotal      int   // Total bricks at level start
	endlessCycle     int   // Number of times levels have cycled (endless mode)
	basePaddleWidth  int   // Original paddle width before effects
	currentBallSpeed Fixed // Current base ball speed (affected by power-ups)

	// Configuration
	runtime    core.RuntimeConfig
	cfg        config.BreakoutConfig
	difficulty *config.DifficultyManager

	// Layout (computed from screen size)
	brickAreaTop   int // Y position where bricks start
	brickHeight    int // Height of each brick in cells
	brickWidth     int // Width of each brick in cells
	paddleY        int // Y position of paddle
	minScreenW     int // Minimum required screen width
	minScreenH     int // Minimum required screen height
	screenTooSmall bool
}

// New creates a new Breakout game instance (campaign mode).
func New() *Game {
	return &Game{mode: ModeCampaign}
}

// NewEndless creates a new Breakout game instance in endless mode.
func NewEndless() *Game {
	return &Game{mode: ModeEndless}
}

// ID returns the unique identifier for this game.
func (g *Game) ID() string {
	if g.mode == ModeEndless {
		return "breakout_endless"
	}
	return "breakout"
}

// Title returns the display name for this game.
func (g *Game) Title() string {
	if g.mode == ModeEndless {
		return "Breakout (Endless)"
	}
	return "Breakout"
}

// Reset initializes or restarts the game.
func (g *Game) Reset(runtime core.RuntimeConfig) {
	g.runtime = runtime

	// Load game config
	cfg, err := config.LoadBreakout(configPath)
	if err != nil {
		cfg = config.DefaultBreakoutConfig()
	}

	// Apply difficulty preset if set
	if difficultyPreset != "" {
		config.ApplyBreakoutPreset(&cfg, difficultyPreset)
	}

	g.cfg = cfg

	// Initialize difficulty manager
	g.difficulty = config.NewDifficultyManager(cfg.Difficulty)

	// Calculate layout
	g.calculateLayout()

	// Check screen size
	g.minScreenW = 30
	g.minScreenH = 15
	g.screenTooSmall = runtime.ScreenW < g.minScreenW || runtime.ScreenH < g.minScreenH

	// Initialize game state
	g.score = 0
	g.lives = cfg.Gameplay.Lives
	g.levelIndex = 0
	g.tickCount = 0
	g.serveDelay = 0
	g.endlessCycle = 0
	g.basePaddleWidth = cfg.Paddle.Width
	g.currentBallSpeed = Fixed(cfg.Physics.BallSpeed)

	// Initialize power-up manager
	g.powerups = NewPowerUpManager(runtime.Seed, DefaultPowerUpConfig())

	// Load level
	g.loadLevel(g.levelIndex)

	// Initialize paddle
	g.paddle = &Paddle{
		X:     ToFixed((runtime.ScreenW - cfg.Paddle.Width) / 2),
		Y:     g.paddleY,
		Width: cfg.Paddle.Width,
	}

	// Initialize balls slice with one ball on paddle
	g.balls = make([]*Ball, 0, 8)
	g.placeBallOnPaddle()
	g.state = StateServe
}

// calculateLayout computes brick and paddle positions based on screen size.
func (g *Game) calculateLayout() {
	// HUD takes top 2 rows
	g.brickAreaTop = 2

	// Brick dimensions
	g.brickHeight = 1
	g.brickWidth = g.runtime.ScreenW / 20 // 20 columns of bricks
	if g.brickWidth < 2 {
		g.brickWidth = 2
	}

	// Paddle at bottom (leave 2 rows for bottom border and ball space)
	g.paddleY = g.runtime.ScreenH - 3
}

// loadLevel loads a level by index.
func (g *Game) loadLevel(index int) {
	g.level = GetLevel(index)
	g.bricksTotal = g.level.CountAlive()
}

// placeBallOnPaddle creates a new ball on the paddle.
func (g *Game) placeBallOnPaddle() {
	ball := &Ball{
		X:      g.paddle.CenterX(),
		Y:      ToFixed(g.paddle.Y - 1),
		VX:     0,
		VY:     0,
		Stuck:  true,
		Active: true,
	}
	g.balls = append(g.balls, ball)
}

// launchBalls launches all stuck balls.
func (g *Game) launchBalls() {
	speed := g.currentBallSpeed

	for _, ball := range g.balls {
		if ball.Stuck && ball.Active {
			// Launch upward with slight horizontal bias
			ball.VX = speed / 4
			ball.VY = -speed
			ball.Stuck = false
		}
	}

	g.state = StatePlaying
}

// countActiveBalls returns the number of active balls.
func (g *Game) countActiveBalls() int {
	count := 0
	for _, ball := range g.balls {
		if ball.Active {
			count++
		}
	}
	return count
}

// countStuckBalls returns the number of stuck balls.
func (g *Game) countStuckBalls() int {
	count := 0
	for _, ball := range g.balls {
		if ball.Active && ball.Stuck {
			count++
		}
	}
	return count
}

// Step advances the game by one tick.
func (g *Game) Step(in core.InputFrame) core.StepResult {
	if g.screenTooSmall {
		return core.StepResult{State: g.State()}
	}

	// Handle restart
	if in.Has(core.ActionRestart) && (g.state == StateGameOver || g.state == StateWin) {
		g.Reset(g.runtime)
		return core.StepResult{State: g.State()}
	}

	// Handle pause toggle
	if in.Has(core.ActionPause) {
		if g.state == StatePaused {
			g.state = StatePlaying
		} else if g.state == StatePlaying {
			g.state = StatePaused
		}
	}

	// Don't update if paused or game over
	if g.state == StatePaused || g.state == StateGameOver || g.state == StateWin {
		return core.StepResult{State: g.State()}
	}

	g.tickCount++

	// Handle serve delay countdown
	if g.serveDelay > 0 {
		g.serveDelay--
		return core.StepResult{State: g.State()}
	}

	// Expire power-up effects
	expired := g.powerups.ExpireEffects(g.tickCount)
	for _, effectType := range expired {
		g.onEffectExpired(effectType)
	}

	// Handle paddle movement
	g.updatePaddle(in)

	// Update pickups
	g.powerups.Update(g.runtime.ScreenH)

	// Check pickup collection
	collected := g.powerups.CheckPaddleCollision(g.paddle)
	if collected >= 0 {
		g.activatePickup(collected)
	}

	// Handle ball launch in serve state
	if g.state == StateServe {
		// Update stuck balls to follow paddle
		for _, ball := range g.balls {
			if ball.Active && ball.Stuck {
				ball.X = g.paddle.CenterX()
				ball.Y = ToFixed(g.paddle.Y - 1)
			}
		}

		if in.Has(core.ActionJump) { // Space to launch
			g.launchBalls()
		}
		return core.StepResult{State: g.State()}
	}

	// Update all balls
	g.updateBalls()

	return core.StepResult{State: g.State()}
}

// updatePaddle handles paddle movement.
func (g *Game) updatePaddle(in core.InputFrame) {
	speed := Fixed(g.cfg.Physics.PaddleSpeed) // Already scaled by 1000 in config

	// A/Left = move left, D/Right = move right
	if in.Has(core.ActionLeft) {
		g.paddle.X = g.paddle.X.Sub(speed)
	}
	if in.Has(core.ActionRight) {
		g.paddle.X = g.paddle.X.Add(speed)
	}

	// Clamp paddle position
	minX := ToFixed(1)
	maxX := ToFixed(g.runtime.ScreenW - g.paddle.Width - 1)
	g.paddle.X = ClampFixed(g.paddle.X, minX, maxX)
}

// updateBalls handles all ball movements and collisions.
func (g *Game) updateBalls() {
	isSticky := g.powerups.HasEffect(EffectSticky)

	for _, ball := range g.balls {
		if !ball.Active || ball.Stuck {
			continue
		}

		// Move ball
		ball.Move()

		// Check wall collisions
		side, fellOff := CheckWallCollision(ball, g.runtime.ScreenW, g.runtime.ScreenH)
		if fellOff {
			ball.Active = false
			continue
		}
		if side != CollisionNone {
			ApplyCollisionBounce(ball, side)
		}

		// Check paddle collision
		if CheckPaddleCollision(ball, g.paddle, g.currentBallSpeed) {
			// If sticky effect active, stick to paddle
			if isSticky {
				ball.Stuck = true
				ball.VX = 0
				ball.VY = 0
				ball.X = g.paddle.CenterX()
				ball.Y = ToFixed(g.paddle.Y - 1)
			}
			continue
		}

		// Check brick collisions
		row, col, brickSide := CheckBrickCollision(ball, g.level, g.brickAreaTop, g.brickHeight, g.brickWidth)
		if brickSide != CollisionNone && row >= 0 && col >= 0 {
			brick := &g.level.Bricks[row][col]
			if brick.Alive {
				// Handle brick hit
				g.hitBrick(brick, row, col)

				// Bounce ball
				ApplyCollisionBounce(ball, brickSide)
			}
		}
	}

	// Check if all balls are lost
	if g.countActiveBalls() == 0 {
		g.handleMiss()
	} else if g.countStuckBalls() > 0 && g.countStuckBalls() == g.countActiveBalls() {
		// All remaining balls are stuck, go to serve state
		g.state = StateServe
	}
}

// hitBrick handles hitting a brick.
func (g *Game) hitBrick(brick *Brick, row, col int) {
	// Solid bricks cannot be destroyed
	if brick.Type == BrickSolid {
		return
	}

	brick.HP--
	if brick.HP <= 0 {
		// Destroy brick
		brick.Alive = false
		g.score += brick.Points

		// Try to spawn power-up
		brickCenterX := col*g.brickWidth + g.brickWidth/2
		brickCenterY := g.brickAreaTop + row*g.brickHeight
		g.powerups.TrySpawnPickup(brickCenterX, brickCenterY)

		// Check win condition
		if g.level.CountAlive() == 0 {
			g.handleLevelClear()
		}
	}
}

// activatePickup activates a collected pickup.
func (g *Game) activatePickup(pickupType PickupType) {
	cfg := g.powerups.Config

	switch pickupType {
	case PickupWiden:
		g.powerups.AddEffect(EffectWiden, g.tickCount, cfg.DurationWiden)
		g.powerups.RemoveEffect(EffectShrink) // Cancel shrink
		g.applyPaddleWidthEffect()

	case PickupShrink:
		g.powerups.AddEffect(EffectShrink, g.tickCount, cfg.DurationShrink)
		g.powerups.RemoveEffect(EffectWiden) // Cancel widen
		g.applyPaddleWidthEffect()

	case PickupMultiball:
		g.spawnMultiballs(cfg.MultiballCount)

	case PickupSticky:
		g.powerups.AddEffect(EffectSticky, g.tickCount, cfg.DurationSticky)

	case PickupSpeedUp:
		g.powerups.AddEffect(EffectSpeedUp, g.tickCount, cfg.DurationSpeedUp)
		g.powerups.RemoveEffect(EffectSlowDown) // Cancel slow down
		g.applyBallSpeedEffect()

	case PickupSlowDown:
		g.powerups.AddEffect(EffectSlowDown, g.tickCount, cfg.DurationSlowDown)
		g.powerups.RemoveEffect(EffectSpeedUp) // Cancel speed up
		g.applyBallSpeedEffect()

	case PickupExtraLife:
		g.lives++
	}
}

// onEffectExpired handles effect expiration.
func (g *Game) onEffectExpired(effectType EffectType) {
	switch effectType {
	case EffectWiden, EffectShrink:
		g.applyPaddleWidthEffect()
	case EffectSpeedUp, EffectSlowDown:
		g.applyBallSpeedEffect()
	}
}

// applyPaddleWidthEffect applies paddle width based on active effects.
func (g *Game) applyPaddleWidthEffect() {
	cfg := g.powerups.Config
	newWidth := g.basePaddleWidth

	if g.powerups.HasEffect(EffectWiden) {
		newWidth += cfg.WidenAmount
	} else if g.powerups.HasEffect(EffectShrink) {
		newWidth -= cfg.ShrinkAmount
	}

	// Clamp
	if newWidth < cfg.MinPaddleWidth {
		newWidth = cfg.MinPaddleWidth
	}
	if newWidth > cfg.MaxPaddleWidth {
		newWidth = cfg.MaxPaddleWidth
	}

	g.paddle.Width = newWidth
}

// applyBallSpeedEffect applies ball speed based on active effects.
func (g *Game) applyBallSpeedEffect() {
	cfg := g.powerups.Config
	baseSpeed := Fixed(g.cfg.Physics.BallSpeed)

	switch {
	case g.powerups.HasEffect(EffectSpeedUp):
		// Speed up: multiply by 1.5
		g.currentBallSpeed = baseSpeed.Mul(cfg.SpeedMultiplier).Div(100)
	case g.powerups.HasEffect(EffectSlowDown):
		// Slow down: multiply by 0.67 (100/150)
		g.currentBallSpeed = baseSpeed.Mul(100).Div(cfg.SpeedMultiplier)
	default:
		g.currentBallSpeed = baseSpeed
	}

	// Clamp
	minSpeed := Fixed(cfg.MinBallSpeed)
	maxSpeed := Fixed(cfg.MaxBallSpeed)
	g.currentBallSpeed = ClampFixed(g.currentBallSpeed, minSpeed, maxSpeed)
}

// spawnMultiballs spawns additional balls.
func (g *Game) spawnMultiballs(count int) {
	// Find an active ball to clone
	var sourceBall *Ball
	for _, ball := range g.balls {
		if ball.Active && !ball.Stuck {
			sourceBall = ball
			break
		}
	}

	if sourceBall == nil {
		// No active non-stuck ball, use first active ball
		for _, ball := range g.balls {
			if ball.Active {
				sourceBall = ball
				break
			}
		}
	}

	if sourceBall == nil {
		return
	}

	speed := g.currentBallSpeed

	for i := range count {
		// Create ball at same position with different angle
		angleOffset := Fixed((i + 1) * 300) // Spread angles
		if i%2 == 1 {
			angleOffset = -angleOffset
		}

		newBall := &Ball{
			X:      sourceBall.X,
			Y:      sourceBall.Y,
			VX:     sourceBall.VX.Add(angleOffset),
			VY:     sourceBall.VY,
			Stuck:  false,
			Active: true,
		}

		// Normalize speed
		// Simple normalization: ensure total speed magnitude is roughly correct
		if newBall.VX.Abs() > speed {
			newBall.VX = speed.Mul(newBall.VX.Sign())
		}
		if newBall.VY == 0 {
			newBall.VY = -speed
		}

		g.balls = append(g.balls, newBall)
	}
}

// handleMiss handles when all balls are lost.
func (g *Game) handleMiss() {
	g.lives--

	if g.lives <= 0 {
		g.state = StateGameOver
		return
	}

	// Clear all balls and power-ups
	g.balls = g.balls[:0]
	g.powerups.Pickups = g.powerups.Pickups[:0]
	g.powerups.Effects = g.powerups.Effects[:0]

	// Reset paddle width and ball speed
	g.paddle.Width = g.basePaddleWidth
	g.currentBallSpeed = Fixed(g.cfg.Physics.BallSpeed)

	// Place new ball on paddle
	g.placeBallOnPaddle()
	g.state = StateServe
	g.serveDelay = 60 // 1 second delay before player can serve again
}

// handleLevelClear handles when all bricks are destroyed.
func (g *Game) handleLevelClear() {
	g.levelIndex++

	if g.mode == ModeCampaign {
		// Campaign mode: check if all levels completed
		if g.levelIndex >= LevelCount() {
			g.state = StateWin
			return
		}
	} else {
		// Endless mode: cycle through levels
		if g.levelIndex >= LevelCount() {
			g.levelIndex = 0
			g.endlessCycle++
			// Increase difficulty slightly each cycle
			g.currentBallSpeed = g.currentBallSpeed.Add(Fixed(20)) // +0.02 per cycle
		}
	}

	// Load next level
	g.loadLevel(g.levelIndex)

	// Clear pickups but keep effects
	g.powerups.Pickups = g.powerups.Pickups[:0]

	// Reset balls - keep one on paddle
	g.balls = g.balls[:0]
	g.placeBallOnPaddle()
	g.state = StateServe
	g.serveDelay = 60
}

// Render draws the current game state to the screen.
func (g *Game) Render(dst *core.Screen) {
	dst.Clear()

	// Check for screen too small
	if g.screenTooSmall {
		msg := "Window too small"
		hint := fmt.Sprintf("Need %dx%d", g.minScreenW, g.minScreenH)
		dst.DrawTextCentered(dst.Height()/2-1, msg)
		dst.DrawTextCentered(dst.Height()/2+1, hint)
		return
	}

	// Draw HUD
	g.renderHUD(dst)

	// Draw bricks
	g.renderBricks(dst)

	// Draw pickups
	g.renderPickups(dst)

	// Draw paddle
	g.renderPaddle(dst)

	// Draw balls
	g.renderBalls(dst)

	// Draw overlay messages
	g.renderOverlay(dst)
}

// renderHUD draws the score, lives, and level indicator.
func (g *Game) renderHUD(dst *core.Screen) {
	// Score on left
	scoreText := fmt.Sprintf("Score: %d", g.score)
	dst.DrawText(1, 0, scoreText)

	// Lives in center
	livesText := fmt.Sprintf("Lives: %d", g.lives)
	dst.DrawTextCentered(0, livesText)

	// Level on right
	var levelText string
	if g.mode == ModeEndless {
		totalLevel := g.endlessCycle*LevelCount() + g.levelIndex + 1
		levelText = fmt.Sprintf("Level: %d", totalLevel)
	} else {
		levelText = fmt.Sprintf("Level: %d/%d", g.levelIndex+1, LevelCount())
	}
	dst.DrawText(dst.Width()-len(levelText)-1, 0, levelText)

	// Effects display (compact) on row 1
	effectsStr := g.buildEffectsString()
	if effectsStr != "" {
		dst.DrawText(1, 1, effectsStr)
	} else {
		// Separator line if no effects
		for x := range dst.Width() {
			dst.Set(x, 1, BorderHoriz)
		}
	}
}

// buildEffectsString creates a compact effects display.
func (g *Game) buildEffectsString() string {
	if len(g.powerups.Effects) == 0 {
		return ""
	}

	result := ""
	for _, e := range g.powerups.Effects {
		remaining := e.TicksRemaining(g.tickCount)
		secs := remaining / 60
		if result != "" {
			result += " "
		}
		result += fmt.Sprintf("%s(%d)", e.Type.String(), secs)
	}
	return result
}

// renderBricks draws all alive bricks.
func (g *Game) renderBricks(dst *core.Screen) {
	for row := range g.level.Height {
		for col := range g.level.Width {
			brick := g.level.Bricks[row][col]
			if !brick.Alive || brick.Type == BrickEmpty {
				continue
			}

			// Calculate screen position
			screenY := g.brickAreaTop + row*g.brickHeight
			screenX := col * g.brickWidth

			// Get glyph based on brick type
			var glyph rune
			switch brick.Type {
			case BrickHard:
				if brick.HP > 1 {
					glyph = HardBrickGlyph
				} else {
					glyph = BrickGlyphs[row%len(BrickGlyphs)]
				}
			case BrickSolid:
				glyph = SolidBrickGlyph
			default:
				glyph = BrickGlyphs[row%len(BrickGlyphs)]
			}

			// Draw brick
			for dx := range g.brickWidth {
				if screenX+dx < dst.Width() && screenY < dst.Height() {
					dst.Set(screenX+dx, screenY, glyph)
				}
			}
		}
	}
}

// renderPickups draws falling power-ups.
func (g *Game) renderPickups(dst *core.Screen) {
	for _, pickup := range g.powerups.Pickups {
		if !pickup.Active {
			continue
		}

		x := pickup.CellX()
		y := pickup.CellY()

		if x >= 0 && x < dst.Width() && y >= 0 && y < dst.Height() {
			dst.Set(x, y, pickup.Type.Glyph())
		}
	}
}

// renderPaddle draws the player's paddle.
func (g *Game) renderPaddle(dst *core.Screen) {
	paddleX := g.paddle.CellX()
	for i := range g.paddle.Width {
		if paddleX+i < dst.Width() {
			dst.Set(paddleX+i, g.paddle.Y, PaddleChar)
		}
	}
}

// renderBalls draws all balls.
func (g *Game) renderBalls(dst *core.Screen) {
	for _, ball := range g.balls {
		if !ball.Active {
			continue
		}

		ballX := ball.CellX()
		ballY := ball.CellY()

		if ballX >= 0 && ballX < dst.Width() && ballY >= 0 && ballY < dst.Height() {
			dst.Set(ballX, ballY, BallChar)
		}
	}
}

// renderOverlay draws game state messages.
func (g *Game) renderOverlay(dst *core.Screen) {
	switch g.state {
	case StateServe:
		if g.serveDelay <= 0 {
			dst.DrawTextCentered(dst.Height()-1, "Press SPACE to launch")
		} else {
			dst.DrawTextCentered(dst.Height()-1, "Get ready...")
		}

	case StatePaused:
		g.drawCenteredBox(dst, "PAUSED", "Press P to resume")

	case StateGameOver:
		subtitle := fmt.Sprintf("Score: %d  |  Press R to restart", g.score)
		g.drawCenteredBox(dst, "GAME OVER", subtitle)

	case StateWin:
		subtitle := fmt.Sprintf("Final Score: %d  |  Press R to restart", g.score)
		g.drawCenteredBox(dst, "YOU WIN!", subtitle)
	}
}

// drawCenteredBox draws a centered message box.
func (g *Game) drawCenteredBox(dst *core.Screen, title, subtitle string) {
	w := dst.Width()
	h := dst.Height()

	boxW := core.Max(len(title), len(subtitle)) + 4
	boxH := 5
	boxX := (w - boxW) / 2
	boxY := (h - boxH) / 2

	// Draw box background
	dst.DrawRect(core.NewRect(boxX, boxY, boxW, boxH), ' ')
	dst.DrawBox(core.NewRect(boxX, boxY, boxW, boxH))

	// Draw text
	titleX := boxX + (boxW-len(title))/2
	dst.DrawText(titleX, boxY+1, title)

	subtitleX := boxX + (boxW-len(subtitle))/2
	dst.DrawText(subtitleX, boxY+3, subtitle)
}

// State returns the current game state.
func (g *Game) State() core.GameState {
	return core.GameState{
		Score:    g.score,
		GameOver: g.state == StateGameOver || g.state == StateWin,
		Paused:   g.state == StatePaused,
	}
}

// Register the games with the registry
func init() {
	registry.Register("breakout", func() registry.Game {
		return New()
	})
	registry.Register("breakout_endless", func() registry.Game {
		return NewEndless()
	})
}
