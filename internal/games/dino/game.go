// Package dino implements a Chrome Dino-style endless runner game.
// The player must jump over obstacles while running automatically.
package dino

import (
	"fmt"

	"github.com/vovakirdan/tui-arcade/internal/config"
	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/registry"
)

// Visual characters for rendering
const (
	DinoBody   = '█'
	DinoHead   = '◆'
	DinoLeg1   = '╱'
	DinoLeg2   = '╲'
	CactusChar = '▓'
	GroundChar = '═'
)

// Game implements the Dino Runner game logic.
type Game struct {
	playerY    float64           // Player vertical position (relative to ground, negative = up)
	playerVel  float64           // Player vertical velocity
	isGrounded bool              // Whether player is on the ground
	obstacles  *ObstacleManager  // Obstacle manager
	score      int               // Current score (distance traveled)
	gameOver   bool              // Whether game has ended
	paused     bool              // Whether game is paused
	runtime    core.RuntimeConfig // Runtime config (screen size, tick rate)
	cfg        config.DinoConfig // Game-specific config
	difficulty *config.DifficultyManager
	tickCount  int // Number of ticks since start
	groundY    int // Y position of ground line
	legFrame   int // Animation frame for running legs
}

// configPath stores the custom config path set via CLI
var configPath string
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
		difficultyPreset = "" // Use config default
	}
}

// New creates a new Dino Runner game instance.
func New() *Game {
	return &Game{}
}

// ID returns the unique identifier for this game.
func (g *Game) ID() string {
	return "dino"
}

// Title returns the display name for this game.
func (g *Game) Title() string {
	return "Dino Runner"
}

// Reset initializes or restarts the game.
func (g *Game) Reset(runtime core.RuntimeConfig) {
	g.runtime = runtime

	// Load game config
	cfg, err := config.LoadDino(configPath)
	if err != nil {
		cfg = config.DefaultDinoConfig()
	}

	// Apply difficulty preset if set
	if difficultyPreset != "" {
		config.ApplyDinoPreset(&cfg, difficultyPreset)
	}

	g.cfg = cfg

	// Initialize difficulty manager
	g.difficulty = config.NewDifficultyManager(cfg.Difficulty)

	// Initialize game state
	g.groundY = runtime.ScreenH - g.cfg.Player.GroundOffset
	g.playerY = 0 // On ground
	g.playerVel = 0
	g.isGrounded = true
	g.score = 0
	g.gameOver = false
	g.paused = false
	g.tickCount = 0
	g.legFrame = 0

	// Initialize obstacle manager
	if g.obstacles == nil {
		g.obstacles = NewObstacleManager(runtime.Seed, runtime.ScreenW, &cfg, g.difficulty)
	} else {
		g.obstacles.UpdateConfig(&cfg, g.difficulty)
		g.obstacles.UpdateScreenSize(runtime.ScreenW)
		g.obstacles.Reset(runtime.Seed)
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
	g.legFrame = (g.legFrame + 1) % 10 // Animation cycle

	// Handle jump input (only when grounded)
	if in.Has(core.ActionJump) && g.isGrounded {
		g.playerVel = g.cfg.Physics.JumpImpulse
		g.isGrounded = false
	}

	// Apply physics
	if !g.isGrounded {
		g.playerVel += g.cfg.Physics.Gravity
		if g.playerVel > g.cfg.Physics.MaxFallSpeed {
			g.playerVel = g.cfg.Physics.MaxFallSpeed
		}
		g.playerY += g.playerVel

		// Check if landed
		if g.playerY >= 0 {
			g.playerY = 0
			g.playerVel = 0
			g.isGrounded = true
		}
	}

	// Update obstacles with current difficulty
	g.obstacles.Update(g.score, g.tickCount)

	// Increment score (based on distance/time)
	g.score++

	// Check collisions
	playerRect := g.playerRect()
	if g.obstacles.CheckCollision(playerRect, g.groundY) {
		g.gameOver = true
	}

	return core.StepResult{State: g.State()}
}

// playerRect returns the player's collision rectangle in screen coordinates.
func (g *Game) playerRect() core.Rect {
	// Player Y is relative to ground (negative = above ground)
	// Convert to screen coordinates
	screenY := g.groundY - g.cfg.Player.Height - int(-g.playerY)
	return core.NewRect(g.cfg.Player.X, screenY, g.cfg.Player.Width, g.cfg.Player.Height)
}

// Render draws the current game state to the screen.
func (g *Game) Render(dst *core.Screen) {
	dst.Clear()

	// Draw ground
	dst.DrawHLine(0, g.groundY, dst.Width(), GroundChar)

	// Draw obstacles
	for _, c := range g.obstacles.Cacti() {
		g.drawCactus(dst, c)
	}

	// Draw player (dino)
	g.drawDino(dst)

	// Draw HUD
	scoreText := fmt.Sprintf(" Score: %d ", g.score)
	dst.DrawText(2, 0, scoreText)

	// Show difficulty level if progression is enabled
	if g.difficulty.IsEnabled() {
		speed := g.difficulty.Speed(g.cfg.Physics.BaseSpeed, g.score, g.tickCount)
		levelText := fmt.Sprintf(" Spd: %.1f ", speed)
		dst.DrawText(dst.Width()-len(levelText)-2, 0, levelText)
	}

	if g.paused {
		g.drawCenteredMessage(dst, "PAUSED", "Press P to resume")
	}

	if g.gameOver {
		g.drawCenteredMessage(dst, "GAME OVER", fmt.Sprintf("Score: %d  |  Press R to restart", g.score))
	}
}

// drawDino renders the player character.
func (g *Game) drawDino(dst *core.Screen) {
	// Player Y is relative to ground (negative = above ground)
	baseY := g.groundY - g.cfg.Player.Height - int(-g.playerY)
	playerX := g.cfg.Player.X

	// Simple dino sprite (3x3)
	//  ◆█
	// ███
	// ╱╲

	// Head and body
	dst.Set(playerX+1, baseY, DinoHead)
	dst.Set(playerX+2, baseY, DinoBody)

	// Body
	dst.Set(playerX, baseY+1, DinoBody)
	dst.Set(playerX+1, baseY+1, DinoBody)
	dst.Set(playerX+2, baseY+1, DinoBody)

	// Legs (animated when grounded)
	if g.isGrounded {
		if g.legFrame < 5 {
			dst.Set(playerX, baseY+2, DinoLeg1)
			dst.Set(playerX+1, baseY+2, ' ')
			dst.Set(playerX+2, baseY+2, DinoLeg2)
		} else {
			dst.Set(playerX, baseY+2, ' ')
			dst.Set(playerX+1, baseY+2, DinoLeg1)
			dst.Set(playerX+2, baseY+2, DinoLeg2)
		}
	} else {
		// In air - legs tucked
		dst.Set(playerX, baseY+2, DinoLeg1)
		dst.Set(playerX+1, baseY+2, DinoLeg2)
		dst.Set(playerX+2, baseY+2, ' ')
	}
}

// drawCactus renders a single cactus obstacle.
func (g *Game) drawCactus(dst *core.Screen, c Cactus) {
	for dy := 0; dy < c.Height; dy++ {
		for dx := 0; dx < c.Width; dx++ {
			y := g.groundY - c.Height + dy
			dst.Set(c.X+dx, y, CactusChar)
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
	registry.Register("dino", func() registry.Game {
		return New()
	})
}
