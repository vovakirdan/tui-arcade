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

// GameState constants
const (
	StateServe    = "serve"    // Ball on paddle, waiting for launch
	StatePlaying  = "playing"  // Ball in play
	StateGameOver = "gameover" // No lives left
	StateWin      = "win"      // All bricks destroyed
	StatePaused   = "paused"   // Game paused
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
	// Game objects
	paddle *Paddle
	ball   *Ball
	level  *Level

	// Game state
	state       string
	score       int
	lives       int
	levelIndex  int
	tickCount   int
	serveDelay  int // Countdown before allowing serve after miss
	bricksTotal int // Total bricks at level start

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

// New creates a new Breakout game instance.
func New() *Game {
	return &Game{}
}

// ID returns the unique identifier for this game.
func (g *Game) ID() string {
	return "breakout"
}

// Title returns the display name for this game.
func (g *Game) Title() string {
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

	// Load level
	g.loadLevel(g.levelIndex)

	// Initialize paddle
	g.paddle = &Paddle{
		X:     ToFixed((runtime.ScreenW - cfg.Paddle.Width) / 2),
		Y:     g.paddleY,
		Width: cfg.Paddle.Width,
	}

	// Initialize ball on paddle
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

// placeBallOnPaddle positions the ball on top of the paddle.
func (g *Game) placeBallOnPaddle() {
	g.ball = &Ball{
		X:  g.paddle.CenterX(),
		Y:  ToFixed(g.paddle.Y - 1),
		VX: 0,
		VY: 0,
	}
}

// launchBall starts the ball moving.
func (g *Game) launchBall() {
	speed := Fixed(g.cfg.Physics.BallSpeed) // Already scaled by 1000 in config

	// Launch upward with slight horizontal bias
	g.ball.VX = speed / 4 // Slight rightward bias
	g.ball.VY = -speed

	g.state = StatePlaying
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

	// Handle paddle movement
	g.updatePaddle(in)

	// Handle ball launch in serve state
	if g.state == StateServe {
		// Ball follows paddle
		g.placeBallOnPaddle()

		if in.Has(core.ActionJump) { // Space to launch
			g.launchBall()
		}
		return core.StepResult{State: g.State()}
	}

	// Update ball position
	g.updateBall()

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

// updateBall handles ball movement and collisions.
func (g *Game) updateBall() {
	// Move ball
	g.ball.Move()

	// Check wall collisions
	side, fellOff := CheckWallCollision(g.ball, g.runtime.ScreenW, g.runtime.ScreenH)
	if fellOff {
		g.handleMiss()
		return
	}
	if side != CollisionNone {
		ApplyCollisionBounce(g.ball, side)
	}

	// Check paddle collision
	baseSpeed := Fixed(g.cfg.Physics.BallSpeed) // Already scaled by 1000 in config
	if CheckPaddleCollision(g.ball, g.paddle, baseSpeed) {
		// Collision handled in CheckPaddleCollision
		return
	}

	// Check brick collisions
	row, col, brickSide := CheckBrickCollision(g.ball, g.level, g.brickAreaTop, g.brickHeight, g.brickWidth)
	if brickSide != CollisionNone && row >= 0 && col >= 0 {
		brick := &g.level.Bricks[row][col]
		if brick.Alive {
			// Destroy brick
			brick.Alive = false
			g.score += brick.Points

			// Bounce ball
			ApplyCollisionBounce(g.ball, brickSide)

			// Check win condition
			if g.level.CountAlive() == 0 {
				g.handleWin()
			}
		}
	}
}

// handleMiss handles when the ball falls below the paddle.
func (g *Game) handleMiss() {
	g.lives--

	if g.lives <= 0 {
		g.state = StateGameOver
		return
	}

	// Reset ball on paddle
	g.placeBallOnPaddle()
	g.state = StateServe
	g.serveDelay = 60 // 1 second delay before player can serve again
}

// handleWin handles when all bricks are destroyed.
func (g *Game) handleWin() {
	g.levelIndex++

	// Check if there are more levels
	if g.levelIndex >= LevelCount() {
		g.state = StateWin
		return
	}

	// Load next level
	g.loadLevel(g.levelIndex)
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

	// Draw paddle
	g.renderPaddle(dst)

	// Draw ball
	g.renderBall(dst)

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
	levelText := fmt.Sprintf("Level: %d", g.levelIndex+1)
	dst.DrawText(dst.Width()-len(levelText)-1, 0, levelText)

	// Separator line
	for x := range dst.Width() {
		dst.Set(x, 1, BorderHoriz)
	}
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

			// Get glyph based on row
			glyph := BrickGlyphs[row%len(BrickGlyphs)]

			// Draw brick
			for dx := range g.brickWidth {
				if screenX+dx < dst.Width() && screenY < dst.Height() {
					dst.Set(screenX+dx, screenY, glyph)
				}
			}
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

// renderBall draws the ball.
func (g *Game) renderBall(dst *core.Screen) {
	ballX := g.ball.CellX()
	ballY := g.ball.CellY()

	if ballX >= 0 && ballX < dst.Width() && ballY >= 0 && ballY < dst.Height() {
		dst.Set(ballX, ballY, BallChar)
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

// Register the game with the registry
func init() {
	registry.Register("breakout", func() registry.Game {
		return New()
	})
}
