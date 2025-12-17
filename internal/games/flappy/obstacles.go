package flappy

import (
	"math/rand"

	"github.com/vovakirdan/tui-arcade/internal/config"
	"github.com/vovakirdan/tui-arcade/internal/core"
)

// Pipe represents a vertical obstacle with a gap for the player to pass through.
type Pipe struct {
	X         int  // Horizontal position (left edge)
	GapY      int  // Y position where gap starts (top of gap)
	GapHeight int  // Height of the passable gap
	Passed    bool // Whether the player has passed this pipe (for scoring)
}

// TopRect returns the collision rectangle for the top portion of the pipe.
func (p Pipe) TopRect(pipeWidth int) core.Rect {
	return core.NewRect(p.X, 0, pipeWidth, p.GapY)
}

// BottomRect returns the collision rectangle for the bottom portion of the pipe.
func (p Pipe) BottomRect(pipeWidth, screenH int) core.Rect {
	bottomY := p.GapY + p.GapHeight
	return core.NewRect(p.X, bottomY, pipeWidth, screenH-bottomY)
}

// PipeManager handles spawning, movement, and removal of pipes.
type PipeManager struct {
	pipes      []Pipe
	rng        *rand.Rand
	screenW    int
	screenH    int
	nextSpawnX int // X position where next pipe will spawn
	cfg        *config.FlappyConfig
	difficulty *config.DifficultyManager
}

// NewPipeManager creates a new pipe manager with the given RNG seed.
func NewPipeManager(seed int64, screenW, screenH int, cfg *config.FlappyConfig, diff *config.DifficultyManager) *PipeManager {
	pm := &PipeManager{
		pipes:      make([]Pipe, 0, 8),
		rng:        rand.New(rand.NewSource(seed)),
		screenW:    screenW,
		screenH:    screenH,
		cfg:        cfg,
		difficulty: diff,
	}
	pm.Reset(seed)
	return pm
}

// UpdateConfig updates the configuration.
func (pm *PipeManager) UpdateConfig(cfg *config.FlappyConfig, diff *config.DifficultyManager) {
	pm.cfg = cfg
	pm.difficulty = diff
}

// Reset clears all pipes and resets the RNG.
func (pm *PipeManager) Reset(seed int64) {
	pm.pipes = pm.pipes[:0]
	pm.rng = rand.New(rand.NewSource(seed))
	pm.nextSpawnX = pm.screenW + pm.cfg.Obstacles.PipeSpacing // First pipe spawns off-screen
}

// UpdateScreenSize updates the screen dimensions.
func (pm *PipeManager) UpdateScreenSize(screenW, screenH int) {
	pm.screenW = screenW
	pm.screenH = screenH
}

// Update moves pipes left and spawns new ones as needed.
// Returns the number of pipes that were passed this frame (for scoring).
func (pm *PipeManager) Update(playerX, score, ticks int) int {
	passed := 0

	// Calculate current speed based on difficulty
	speed := pm.difficulty.Speed(pm.cfg.Physics.BaseSpeed, score, ticks)
	speedInt := int(speed)
	if speedInt < 1 {
		speedInt = 1
	}

	// Move pipes left
	for i := range pm.pipes {
		pm.pipes[i].X -= speedInt
	}

	pipeWidth := pm.cfg.Obstacles.PipeWidth

	// Check for passed pipes (player passed the right edge of the pipe)
	for i := range pm.pipes {
		if !pm.pipes[i].Passed && pm.pipes[i].X+pipeWidth < playerX {
			pm.pipes[i].Passed = true
			passed++
		}
	}

	// Remove pipes that have moved off the left side
	validPipes := pm.pipes[:0]
	for _, p := range pm.pipes {
		if p.X+pipeWidth > 0 {
			validPipes = append(validPipes, p)
		}
	}
	pm.pipes = validPipes

	// Calculate current spacing based on difficulty
	spacing := pm.difficulty.Spacing(pm.cfg.Obstacles.PipeSpacing, score, ticks)

	// Spawn new pipe if needed
	if len(pm.pipes) == 0 || pm.pipes[len(pm.pipes)-1].X < pm.screenW-spacing {
		pm.spawnPipe(score, ticks)
	}

	return passed
}

// spawnPipe creates a new pipe at the right edge of the screen.
func (pm *PipeManager) spawnPipe(score, ticks int) {
	// Calculate gap size based on difficulty
	maxGap := pm.cfg.Obstacles.MaxGapSize
	currentGap := pm.difficulty.GapSize(maxGap, score, ticks)

	// Ensure gap is at least min
	minGap := pm.cfg.Obstacles.MinGapSize
	if currentGap < minGap {
		currentGap = minGap
	}

	// Random variation in gap size (between minGap and currentGap)
	gapRange := currentGap - minGap
	if gapRange < 0 {
		gapRange = 0
	}
	gapHeight := minGap
	if gapRange > 0 {
		gapHeight = minGap + pm.rng.Intn(gapRange+1)
	}

	// Calculate valid Y range for gap
	topMargin := pm.cfg.Obstacles.TopMargin
	bottomMargin := pm.cfg.Obstacles.BottomMargin
	maxGapY := pm.screenH - bottomMargin - gapHeight
	minGapY := topMargin

	if maxGapY < minGapY {
		maxGapY = minGapY // Edge case for very small screens
	}

	gapY := minGapY
	if maxGapY > minGapY {
		gapY = minGapY + pm.rng.Intn(maxGapY-minGapY+1)
	}

	pipe := Pipe{
		X:         pm.screenW,
		GapY:      gapY,
		GapHeight: gapHeight,
		Passed:    false,
	}

	pm.pipes = append(pm.pipes, pipe)
}

// Pipes returns the current list of pipes.
func (pm *PipeManager) Pipes() []Pipe {
	return pm.pipes
}

// CheckCollision tests if the given rectangle collides with any pipe.
func (pm *PipeManager) CheckCollision(playerRect core.Rect, screenH int) bool {
	pipeWidth := pm.cfg.Obstacles.PipeWidth
	for _, p := range pm.pipes {
		topRect := p.TopRect(pipeWidth)
		bottomRect := p.BottomRect(pipeWidth, screenH)

		if playerRect.Intersects(topRect) || playerRect.Intersects(bottomRect) {
			return true
		}
	}
	return false
}
