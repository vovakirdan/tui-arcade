// Package flappy implements a Flappy Bird-style game.
// The player controls a bird that must navigate through gaps in vertical pipes.
package flappy

import (
	"fmt"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/registry"
)

// Physics constants - tuned for playability at 60 FPS
const (
	Gravity       = 0.4   // Downward acceleration per tick
	JumpImpulse   = -6.0  // Upward velocity when jumping (negative = up)
	MaxFallSpeed  = 10.0  // Terminal velocity
	PipeSpeed     = 2     // How fast pipes move left per tick
	PlayerX       = 10    // Fixed horizontal position of player
	PlayerWidth   = 2     // Player hitbox width
	PlayerHeight  = 2     // Player hitbox height
)

// Visual characters for rendering
const (
	PlayerChar    = '▶'
	PipeChar      = '█'
	PipeCapTop    = '▄'
	PipeCapBottom = '▀'
	GroundChar    = '═'
)

// Game implements the Flappy Bird game logic.
type Game struct {
	playerY   float64       // Player vertical position (top of hitbox)
	playerVel float64       // Player vertical velocity
	pipes     *PipeManager  // Obstacle manager
	score     int           // Current score
	gameOver  bool          // Whether game has ended
	paused    bool          // Whether game is paused
	config    core.RuntimeConfig
	tickCount int           // Number of ticks since start
}

// New creates a new Flappy Bird game instance.
func New() *Game {
	return &Game{}
}

// ID returns the unique identifier for this game.
func (g *Game) ID() string {
	return "flappy"
}

// Title returns the display name for this game.
func (g *Game) Title() string {
	return "Flappy Bird"
}

// Reset initializes or restarts the game.
func (g *Game) Reset(cfg core.RuntimeConfig) {
	g.config = cfg
	g.playerY = float64(cfg.ScreenH) / 2.0
	g.playerVel = 0
	g.score = 0
	g.gameOver = false
	g.paused = false
	g.tickCount = 0

	if g.pipes == nil {
		g.pipes = NewPipeManager(cfg.Seed, cfg.ScreenW, cfg.ScreenH)
	} else {
		g.pipes.UpdateScreenSize(cfg.ScreenW, cfg.ScreenH)
		g.pipes.Reset(cfg.Seed)
	}
}

// Step advances the game by one tick.
func (g *Game) Step(in core.InputFrame) core.StepResult {
	if g.gameOver {
		return core.StepResult{State: g.State()}
	}

	// Handle pause toggle
	if in.Has(core.ActionPause) {
		g.paused = !g.paused
	}

	if g.paused {
		return core.StepResult{State: g.State()}
	}

	g.tickCount++

	// Handle jump input
	if in.Has(core.ActionJump) {
		g.playerVel = JumpImpulse
	}

	// Apply physics
	g.playerVel += Gravity
	if g.playerVel > MaxFallSpeed {
		g.playerVel = MaxFallSpeed
	}
	g.playerY += g.playerVel

	// Update pipes and check for scoring
	passed := g.pipes.Update(PlayerX + PlayerWidth)
	g.score += passed

	// Check collisions
	playerRect := g.playerRect()

	// Hit top of screen
	if g.playerY < 0 {
		g.playerY = 0
		g.gameOver = true
	}

	// Hit bottom of screen (ground)
	groundY := g.config.ScreenH - 2 // Leave space for ground line
	if int(g.playerY)+PlayerHeight >= groundY {
		g.playerY = float64(groundY - PlayerHeight)
		g.gameOver = true
	}

	// Hit a pipe
	if g.pipes.CheckCollision(playerRect, g.config.ScreenH) {
		g.gameOver = true
	}

	return core.StepResult{State: g.State()}
}

// playerRect returns the player's collision rectangle.
func (g *Game) playerRect() core.Rect {
	return core.NewRect(PlayerX, int(g.playerY), PlayerWidth, PlayerHeight)
}

// Render draws the current game state to the screen.
func (g *Game) Render(dst *core.Screen) {
	dst.Clear()

	// Draw ground
	groundY := dst.Height() - 1
	dst.DrawHLine(0, groundY, dst.Width(), GroundChar)

	// Draw pipes
	for _, p := range g.pipes.Pipes() {
		g.drawPipe(dst, p)
	}

	// Draw player
	playerY := int(g.playerY)
	for dy := 0; dy < PlayerHeight; dy++ {
		for dx := 0; dx < PlayerWidth; dx++ {
			if dx == PlayerWidth-1 && dy == 0 {
				dst.Set(PlayerX+dx, playerY+dy, PlayerChar)
			} else {
				dst.Set(PlayerX+dx, playerY+dy, '●')
			}
		}
	}

	// Draw HUD
	scoreText := fmt.Sprintf(" Score: %d ", g.score)
	dst.DrawText(2, 0, scoreText)

	if g.paused {
		g.drawCenteredMessage(dst, "PAUSED", "Press P to resume")
	}

	if g.gameOver {
		g.drawCenteredMessage(dst, "GAME OVER", fmt.Sprintf("Score: %d  |  Press R to restart", g.score))
	}
}

// drawPipe renders a single pipe to the screen.
func (g *Game) drawPipe(dst *core.Screen, p Pipe) {
	screenH := dst.Height() - 1 // Account for ground

	// Draw top section (from top of screen to gap)
	for y := 0; y < p.GapY; y++ {
		for x := 0; x < PipeWidth; x++ {
			dst.Set(p.X+x, y, PipeChar)
		}
	}
	// Cap on top section (at bottom of top section)
	if p.GapY > 0 {
		for x := 0; x < PipeWidth; x++ {
			dst.Set(p.X+x, p.GapY-1, PipeCapTop)
		}
	}

	// Draw bottom section (from below gap to ground)
	bottomY := p.GapY + p.GapHeight
	for y := bottomY; y < screenH; y++ {
		for x := 0; x < PipeWidth; x++ {
			dst.Set(p.X+x, y, PipeChar)
		}
	}
	// Cap on bottom section (at top of bottom section)
	if bottomY < screenH {
		for x := 0; x < PipeWidth; x++ {
			dst.Set(p.X+x, bottomY, PipeCapBottom)
		}
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
		Score:    g.score,
		GameOver: g.gameOver,
		Paused:   g.paused,
	}
}

// Register the game with the registry
func init() {
	registry.Register("flappy", func() registry.Game {
		return New()
	})
}
