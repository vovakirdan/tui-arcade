package t2048

import (
	"math/rand"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/registry"
)

// Mode represents the game mode.
type Mode string

const (
	ModeCampaign Mode = "campaign"
	ModeEndless  Mode = "endless"
)

// Game implements the 2048 puzzle game.
type Game struct {
	mode Mode
	rng  *rand.Rand
	tick uint64

	score         int
	board         Board
	levelIndex    int // Current level (0-indexed)
	currentTarget int // Current tile target
	spawn4Prob    float64

	// Screen dimensions
	screenW int
	screenH int

	// Game state flags
	gameOver        bool
	levelCleared    bool
	won             bool
	paused          bool
	tooSmall        bool
	moveProcessed   bool // Prevent multiple moves per tick
	levelClearTicks int  // Animation ticks for level clear
}

// Package-level variables for config
var (
	selectedStartLevel int
)

// SetStartLevel sets the starting level (1-10). 0 means start from beginning.
func SetStartLevel(level int) {
	selectedStartLevel = level
}

// GetStartLevel returns the currently selected start level.
func GetStartLevel() int {
	return selectedStartLevel
}

// New creates a new campaign mode 2048 game.
func New() *Game {
	return &Game{
		mode: ModeCampaign,
	}
}

// NewEndless creates a new endless mode 2048 game.
func NewEndless() *Game {
	return &Game{
		mode: ModeEndless,
	}
}

func init() {
	registry.Register("2048", func() registry.Game {
		return New()
	})
	registry.Register("2048_endless", func() registry.Game {
		return NewEndless()
	})
}

// ID returns the game identifier.
func (g *Game) ID() string {
	if g.mode == ModeEndless {
		return "2048_endless"
	}
	return "2048"
}

// Title returns the display name.
func (g *Game) Title() string {
	if g.mode == ModeEndless {
		return "2048 (Endless)"
	}
	return "2048"
}

// Reset initializes/restarts the game.
func (g *Game) Reset(cfg core.RuntimeConfig) {
	g.rng = rand.New(rand.NewSource(cfg.Seed))
	g.tick = 0
	g.score = 0
	g.screenW = cfg.ScreenW
	g.screenH = cfg.ScreenH
	g.gameOver = false
	g.levelCleared = false
	g.won = false
	g.paused = false
	g.moveProcessed = false
	g.levelClearTicks = 0

	// Clear board
	g.board = Board{}

	// Apply selected start level (campaign only)
	if g.mode == ModeCampaign && selectedStartLevel > 0 && selectedStartLevel <= LevelCount() {
		g.levelIndex = selectedStartLevel - 1
		selectedStartLevel = 0 // Reset after use
	} else {
		g.levelIndex = 0
	}

	// Set up level
	g.loadLevel()

	// Spawn initial tiles (2 tiles)
	g.spawnTile()
	g.spawnTile()

	// Check screen size
	g.checkScreenSize()
}

// loadLevel sets up the current level parameters.
func (g *Game) loadLevel() {
	if g.mode == ModeEndless {
		g.currentTarget = 0 // No target in endless
		g.spawn4Prob = 0.10
		return
	}

	level := GetLevel(g.levelIndex)
	if level == nil {
		// Shouldn't happen, but default to last level
		level = GetLevel(LevelCount() - 1)
	}

	g.currentTarget = level.Target
	g.spawn4Prob = level.Spawn4
}

// spawnTile spawns a new tile (2 or 4) in a random empty cell.
func (g *Game) spawnTile() {
	emptyCells := EmptyCells(g.board)
	if len(emptyCells) == 0 {
		return
	}

	// Pick random empty cell
	cell := emptyCells[g.rng.Intn(len(emptyCells))]

	// Determine value (90% 2, 10% 4 by default)
	value := 2
	if g.rng.Float64() < g.spawn4Prob {
		value = 4
	}

	g.board[cell.Y][cell.X] = value
}

// checkScreenSize checks if the screen is large enough.
func (g *Game) checkScreenSize() {
	// Minimum size: board (21 wide, 9 tall) + HUD (2 lines)
	minW := 25
	minH := 12
	g.tooSmall = g.screenW < minW || g.screenH < minH
}

// Step advances the game by one tick.
func (g *Game) Step(in core.InputFrame) core.StepResult {
	g.tick++
	g.moveProcessed = false

	// Handle window size check
	if g.tooSmall {
		return core.StepResult{State: g.State()}
	}

	// Handle pause
	if in.Has(core.ActionPause) {
		g.paused = !g.paused
	}

	if g.paused {
		return core.StepResult{State: g.State()}
	}

	// Handle restart
	if in.Has(core.ActionRestart) && (g.gameOver || g.won) {
		// Will be reset by platform
		return core.StepResult{State: g.State()}
	}

	// Handle level cleared animation
	if g.levelCleared {
		g.levelClearTicks++
		// Auto-advance after 2 seconds (120 ticks at 60fps)
		if g.levelClearTicks >= 120 {
			g.advanceLevel()
		}
		return core.StepResult{State: g.State()}
	}

	// Don't process moves if game over or won
	if g.gameOver || g.won {
		return core.StepResult{State: g.State()}
	}

	// Process move input
	var dir Direction
	moved := false

	switch {
	case in.Has(core.ActionUp):
		dir = DirUp
		moved = true
	case in.Has(core.ActionDown):
		dir = DirDown
		moved = true
	case in.Has(core.ActionLeft):
		dir = DirLeft
		moved = true
	case in.Has(core.ActionRight):
		dir = DirRight
		moved = true
	}

	if moved && !g.moveProcessed {
		g.processMove(dir)
		g.moveProcessed = true
	}

	return core.StepResult{State: g.State()}
}

// processMove handles a move in the given direction.
func (g *Game) processMove(dir Direction) {
	newBoard, scoreGained, changed := Slide(g.board, dir)

	if !changed {
		// Board didn't change - don't spawn new tile
		return
	}

	g.board = newBoard
	g.score += scoreGained

	// Check for level target (campaign only)
	if g.mode == ModeCampaign && g.currentTarget > 0 {
		if MaxTile(g.board) >= g.currentTarget {
			g.levelCleared = true
			g.levelClearTicks = 0
			return
		}
	}

	// Spawn new tile
	g.spawnTile()

	// Check for game over
	if IsGameOver(g.board) {
		g.gameOver = true
	}
}

// advanceLevel moves to the next level.
func (g *Game) advanceLevel() {
	g.levelCleared = false
	g.levelClearTicks = 0

	if g.levelIndex >= LevelCount()-1 {
		// Completed all levels
		g.won = true
		return
	}

	g.levelIndex++
	g.loadLevel()
	// Keep current board and score - just update target
}

// State returns the current game state.
func (g *Game) State() core.GameState {
	return core.GameState{
		Score:    g.score,
		GameOver: g.gameOver || g.won,
		Paused:   g.paused || g.tooSmall || g.levelCleared,
	}
}
