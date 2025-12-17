// Package pong implements a classic Pong game supporting both CPU and online PvP modes.
// Player 1 controls the left paddle, Player 2 (CPU or human) controls the right paddle.
package pong

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/multiplayer"
	"github.com/vovakirdan/tui-arcade/internal/registry"
)

// Visual characters for rendering
const (
	PaddleChar = '█'
	BallChar   = '●'
	NetChar    = '│'
)

// Default game settings
const (
	DefaultPaddleHeight   = 5
	DefaultPaddleWidth    = 1
	DefaultPaddleOffset   = 2 // Distance from edge
	DefaultBallSpeed      = 0.5
	DefaultPaddleSpeed    = 1.0
	DefaultWinScore       = 5
	DefaultCPUReactionMin = 0.6  // CPU reaction time (0-1, 1 = perfect)
	DefaultCPUReactionMax = 0.85 // Max CPU skill
)

// GameMode indicates the type of opponent.
type GameMode int

const (
	// ModeVsCPU is player vs AI.
	ModeVsCPU GameMode = iota
	// ModeOnline is player vs player over network.
	ModeOnline
)

// Game implements the Pong game logic.
type Game struct {
	// Paddles
	paddle1Y float64 // Player 1 (left) paddle Y position
	paddle2Y float64 // Player 2/CPU (right) paddle Y position

	// Ball
	ballX  float64
	ballY  float64
	ballVX float64 // Ball velocity X
	ballVY float64 // Ball velocity Y

	// Scores
	score1 int // Player 1 score
	score2 int // Player 2/CPU score

	// Game state
	gameOver   bool
	paused     bool
	winner     int  // 1 or 2
	serving    bool // True when waiting to serve
	serveDelay int  // Ticks to wait before serving

	// Mode and settings
	mode         GameMode
	runtime      core.RuntimeConfig
	paddleHeight int
	paddleWidth  int
	paddleOffset int
	ballSpeed    float64
	paddleSpeed  float64
	winScore     int
	cpuSkill     float64 // CPU reaction skill (0-1), only used in ModeVsCPU
	rng          *rand.Rand
	tickCount    int
}

// New creates a new Pong game instance (vs CPU mode).
func New() *Game {
	return &Game{
		mode:         ModeVsCPU,
		paddleHeight: DefaultPaddleHeight,
		paddleWidth:  DefaultPaddleWidth,
		paddleOffset: DefaultPaddleOffset,
		ballSpeed:    DefaultBallSpeed,
		paddleSpeed:  DefaultPaddleSpeed,
		winScore:     DefaultWinScore,
		cpuSkill:     DefaultCPUReactionMin,
	}
}

// NewOnline creates a new Pong game instance for online multiplayer.
func NewOnline() *Game {
	return &Game{
		mode:         ModeOnline,
		paddleHeight: DefaultPaddleHeight,
		paddleWidth:  DefaultPaddleWidth,
		paddleOffset: DefaultPaddleOffset,
		ballSpeed:    DefaultBallSpeed,
		paddleSpeed:  DefaultPaddleSpeed,
		winScore:     DefaultWinScore,
	}
}

// SetMode changes the game mode.
func (g *Game) SetMode(mode GameMode) {
	g.mode = mode
}

// ID returns the unique identifier for this game.
func (g *Game) ID() string {
	return "pong"
}

// Title returns the display name for this game.
func (g *Game) Title() string {
	return "Pong"
}

// Reset initializes or restarts the game.
func (g *Game) Reset(runtime core.RuntimeConfig) {
	g.runtime = runtime
	g.rng = rand.New(rand.NewSource(runtime.Seed))

	// Adjust paddle height based on screen size
	g.paddleHeight = core.Clamp(runtime.ScreenH/5, 3, 7)

	// Center paddles vertically
	centerY := float64(runtime.ScreenH) / 2.0
	g.paddle1Y = centerY - float64(g.paddleHeight)/2.0
	g.paddle2Y = centerY - float64(g.paddleHeight)/2.0

	// Reset scores
	g.score1 = 0
	g.score2 = 0
	g.gameOver = false
	g.paused = false
	g.winner = 0
	g.tickCount = 0

	// Start with serve
	g.startServe(1)
}

// startServe prepares to serve the ball.
func (g *Game) startServe(server int) {
	g.serving = true
	g.serveDelay = 60 // 1 second at 60fps

	// Center ball
	g.ballX = float64(g.runtime.ScreenW) / 2.0
	g.ballY = float64(g.runtime.ScreenH) / 2.0

	// Ball velocity towards the player who was scored against
	speed := g.ballSpeed
	if server == 1 {
		g.ballVX = -speed
	} else {
		g.ballVX = speed
	}

	// Random vertical angle
	angle := (g.rng.Float64() - 0.5) * 0.6 // -0.3 to 0.3
	g.ballVY = speed * angle
}

// Step advances the game by one tick (single-player interface, used for vs CPU).
func (g *Game) Step(in core.InputFrame) core.StepResult {
	// Convert single input to multi-input for Player 1
	multi := core.NewMultiInputFrame()
	multi.SetPlayer(multiplayer.Player1, in)
	return g.StepMulti(multi)
}

// StepMulti advances the game by one tick using input from multiple players.
// This is the primary step function used for online multiplayer.
func (g *Game) StepMulti(input core.MultiInputFrame) core.StepResult {
	if g.gameOver {
		return core.StepResult{State: g.State()}
	}

	p1Input := input.Player(multiplayer.Player1)
	p2Input := input.Player(multiplayer.Player2)

	// Handle pause toggle (either player can pause)
	if p1Input.Has(core.ActionPause) || p2Input.Has(core.ActionPause) {
		g.paused = !g.paused
	}

	if g.paused {
		return core.StepResult{State: g.State()}
	}

	g.tickCount++

	// Handle serve delay
	if g.serving {
		g.serveDelay--
		if g.serveDelay <= 0 {
			g.serving = false
		}
		// Still process paddle movement during serve
	}

	// Update Player 1 paddle
	if p1Input.Has(core.ActionUp) || p1Input.Has(core.ActionJump) {
		g.paddle1Y -= g.paddleSpeed
	}
	if p1Input.Has(core.ActionDown) || p1Input.Has(core.ActionDuck) {
		g.paddle1Y += g.paddleSpeed
	}

	// Clamp Player 1 paddle
	maxY := float64(g.runtime.ScreenH - g.paddleHeight - 1)
	g.paddle1Y = core.ClampF(g.paddle1Y, 1, maxY)

	// Update Player 2 paddle based on mode
	if g.mode == ModeOnline {
		// Online mode: use actual player input
		if p2Input.Has(core.ActionUp) || p2Input.Has(core.ActionJump) {
			g.paddle2Y -= g.paddleSpeed
		}
		if p2Input.Has(core.ActionDown) || p2Input.Has(core.ActionDuck) {
			g.paddle2Y += g.paddleSpeed
		}
		g.paddle2Y = core.ClampF(g.paddle2Y, 1, maxY)
	} else {
		// CPU mode
		g.updateCPU()
		// Gradually increase CPU skill
		if g.tickCount%600 == 0 && g.cpuSkill < DefaultCPUReactionMax {
			g.cpuSkill += 0.02
		}
	}

	// Update ball if not serving
	if !g.serving {
		g.updateBall()
	}

	return core.StepResult{State: g.State()}
}

// updateCPU handles CPU paddle movement.
func (g *Game) updateCPU() {
	// CPU tracks ball with some imperfection
	targetY := g.ballY - float64(g.paddleHeight)/2.0

	// Add some "reaction time" - CPU doesn't perfectly follow
	diff := targetY - g.paddle2Y

	// Only move if ball is coming towards CPU
	if g.ballVX > 0 {
		// Move towards target with skill-based speed
		moveSpeed := g.paddleSpeed * g.cpuSkill
		if math.Abs(diff) > moveSpeed {
			if diff > 0 {
				g.paddle2Y += moveSpeed
			} else {
				g.paddle2Y -= moveSpeed
			}
		}
	}

	// Clamp CPU paddle
	maxY := float64(g.runtime.ScreenH - g.paddleHeight - 1)
	g.paddle2Y = core.ClampF(g.paddle2Y, 1, maxY)
}

// updateBall handles ball physics and collision.
func (g *Game) updateBall() {
	// Move ball
	g.ballX += g.ballVX
	g.ballY += g.ballVY

	// Bounce off top/bottom walls
	if g.ballY <= 1 {
		g.ballY = 1
		g.ballVY = -g.ballVY
	}
	if g.ballY >= float64(g.runtime.ScreenH-2) {
		g.ballY = float64(g.runtime.ScreenH - 2)
		g.ballVY = -g.ballVY
	}

	// Check paddle collisions
	paddle1X := float64(g.paddleOffset)
	paddle2X := float64(g.runtime.ScreenW - g.paddleOffset - g.paddleWidth)

	// Ball hits left paddle (Player 1)
	if g.ballX <= paddle1X+float64(g.paddleWidth) && g.ballVX < 0 {
		if g.ballY >= g.paddle1Y && g.ballY <= g.paddle1Y+float64(g.paddleHeight) {
			g.ballX = paddle1X + float64(g.paddleWidth)
			g.ballVX = -g.ballVX
			// Add spin based on where ball hit paddle
			hitPos := (g.ballY - g.paddle1Y) / float64(g.paddleHeight)
			g.ballVY += (hitPos - 0.5) * 0.3
			// Slightly increase speed
			g.ballVX *= 1.02
		}
	}

	// Ball hits right paddle (Player 2/CPU)
	if g.ballX >= paddle2X && g.ballVX > 0 {
		if g.ballY >= g.paddle2Y && g.ballY <= g.paddle2Y+float64(g.paddleHeight) {
			g.ballX = paddle2X - 1
			g.ballVX = -g.ballVX
			// Add spin
			hitPos := (g.ballY - g.paddle2Y) / float64(g.paddleHeight)
			g.ballVY += (hitPos - 0.5) * 0.3
			// Slightly increase speed
			g.ballVX *= 1.02
		}
	}

	// Limit ball speed
	maxSpeed := g.ballSpeed * 3
	if math.Abs(g.ballVX) > maxSpeed {
		g.ballVX = maxSpeed * math.Copysign(1, g.ballVX)
	}
	if math.Abs(g.ballVY) > maxSpeed/2 {
		g.ballVY = maxSpeed / 2 * math.Copysign(1, g.ballVY)
	}

	// Check scoring (ball goes past paddle)
	if g.ballX < 0 {
		// Player 2 scores
		g.score2++
		if g.score2 >= g.winScore {
			g.gameOver = true
			g.winner = 2
		} else {
			g.startServe(2)
		}
	}

	if g.ballX > float64(g.runtime.ScreenW) {
		// Player 1 scores
		g.score1++
		if g.score1 >= g.winScore {
			g.gameOver = true
			g.winner = 1
		} else {
			g.startServe(1)
		}
	}
}

// Render draws the current game state to the screen.
func (g *Game) Render(dst *core.Screen) {
	dst.Clear()

	// Draw center line (net)
	centerX := dst.Width() / 2
	for y := 1; y < dst.Height()-1; y += 2 {
		dst.Set(centerX, y, NetChar)
	}

	// Draw paddles
	paddle1X := g.paddleOffset
	paddle2X := dst.Width() - g.paddleOffset - g.paddleWidth

	for i := range g.paddleHeight {
		dst.Set(paddle1X, int(g.paddle1Y)+i, PaddleChar)
		dst.Set(paddle2X, int(g.paddle2Y)+i, PaddleChar)
	}

	// Draw ball
	if !g.serving || (g.serveDelay/10)%2 == 0 { // Blink during serve
		dst.Set(int(g.ballX), int(g.ballY), BallChar)
	}

	// Draw scores
	score1Text := fmt.Sprintf("%d", g.score1)
	score2Text := fmt.Sprintf("%d", g.score2)
	dst.DrawText(centerX-5, 0, score1Text)
	dst.DrawText(centerX+4, 0, score2Text)

	// Draw labels based on mode
	dst.DrawText(1, 0, "P1")
	if g.mode == ModeOnline {
		dst.DrawText(dst.Width()-3, 0, "P2")
	} else {
		dst.DrawText(dst.Width()-4, 0, "CPU")
	}

	if g.paused {
		g.drawCenteredMessage(dst, "PAUSED", "Press P to resume")
	}

	if g.gameOver {
		var msg, subtitle string
		if g.mode == ModeOnline {
			if g.winner == 1 {
				msg = "PLAYER 1 WINS!"
			} else {
				msg = "PLAYER 2 WINS!"
			}
			subtitle = fmt.Sprintf("%d - %d  |  Press Esc to exit", g.score1, g.score2)
		} else {
			if g.winner == 1 {
				msg = "YOU WIN!"
			} else {
				msg = "CPU WINS!"
			}
			subtitle = fmt.Sprintf("%d - %d  |  Press R to restart", g.score1, g.score2)
		}
		g.drawCenteredMessage(dst, msg, subtitle)
	}
}

// drawCenteredMessage draws a message box in the center of the screen.
func (g *Game) drawCenteredMessage(dst *core.Screen, title, subtitle string) {
	w := dst.Width()
	h := dst.Height()

	// Calculate box dimensions
	boxW := core.Max(len(title), len(subtitle)) + 4
	boxH := 5
	boxX := (w - boxW) / 2
	boxY := (h - boxH) / 2

	// Draw box
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
		Score:    g.score1, // Report player's score
		GameOver: g.gameOver,
		Paused:   g.paused,
	}
}

// OnlineGame interface implementation

// IsGameOver returns true if the game has ended.
func (g *Game) IsGameOver() bool {
	return g.gameOver
}

// Winner returns the winning player or 0 if no winner yet.
func (g *Game) Winner() multiplayer.PlayerID {
	if !g.gameOver {
		return 0
	}
	if g.winner == 1 {
		return multiplayer.Player1
	}
	return multiplayer.Player2
}

// Score1 returns Player 1's score.
func (g *Game) Score1() int {
	return g.score1
}

// Score2 returns Player 2's score.
func (g *Game) Score2() int {
	return g.score2
}

// Register the game with the registry
func init() {
	registry.Register("pong", func() registry.Game {
		return New()
	})
}
