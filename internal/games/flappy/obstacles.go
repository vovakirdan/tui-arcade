package flappy

import (
	"math/rand"

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
func (p Pipe) TopRect() core.Rect {
	return core.NewRect(p.X, 0, PipeWidth, p.GapY)
}

// BottomRect returns the collision rectangle for the bottom portion of the pipe.
func (p Pipe) BottomRect(screenH int) core.Rect {
	bottomY := p.GapY + p.GapHeight
	return core.NewRect(p.X, bottomY, PipeWidth, screenH-bottomY)
}

// Constants for pipe configuration
const (
	PipeWidth       = 5      // Width of each pipe in characters
	PipeSpacing     = 30     // Horizontal distance between pipes
	MinGapSize      = 6      // Minimum gap height
	MaxGapSize      = 10     // Maximum gap height
	PipeTopMargin   = 3      // Minimum distance from top of screen
	PipeBottomMargin = 3     // Minimum distance from bottom of screen
)

// PipeManager handles spawning, movement, and removal of pipes.
type PipeManager struct {
	pipes       []Pipe
	rng         *rand.Rand
	screenW     int
	screenH     int
	nextSpawnX  int // X position where next pipe will spawn
}

// NewPipeManager creates a new pipe manager with the given RNG seed.
func NewPipeManager(seed int64, screenW, screenH int) *PipeManager {
	pm := &PipeManager{
		pipes:   make([]Pipe, 0, 8),
		rng:     rand.New(rand.NewSource(seed)),
		screenW: screenW,
		screenH: screenH,
	}
	pm.Reset(seed)
	return pm
}

// Reset clears all pipes and resets the RNG.
func (pm *PipeManager) Reset(seed int64) {
	pm.pipes = pm.pipes[:0]
	pm.rng = rand.New(rand.NewSource(seed))
	pm.nextSpawnX = pm.screenW + PipeSpacing // First pipe spawns off-screen
}

// UpdateScreenSize updates the screen dimensions.
func (pm *PipeManager) UpdateScreenSize(screenW, screenH int) {
	pm.screenW = screenW
	pm.screenH = screenH
}

// Update moves pipes left and spawns new ones as needed.
// Returns the number of pipes that were passed this frame (for scoring).
func (pm *PipeManager) Update(playerX int) int {
	passed := 0

	// Move pipes left
	for i := range pm.pipes {
		pm.pipes[i].X -= PipeSpeed
	}

	// Check for passed pipes (player passed the right edge of the pipe)
	for i := range pm.pipes {
		if !pm.pipes[i].Passed && pm.pipes[i].X+PipeWidth < playerX {
			pm.pipes[i].Passed = true
			passed++
		}
	}

	// Remove pipes that have moved off the left side
	validPipes := pm.pipes[:0]
	for _, p := range pm.pipes {
		if p.X+PipeWidth > 0 {
			validPipes = append(validPipes, p)
		}
	}
	pm.pipes = validPipes

	// Spawn new pipe if needed
	if len(pm.pipes) == 0 || pm.pipes[len(pm.pipes)-1].X < pm.screenW-PipeSpacing {
		pm.spawnPipe()
	}

	return passed
}

// spawnPipe creates a new pipe at the right edge of the screen.
func (pm *PipeManager) spawnPipe() {
	// Calculate gap size (random within bounds)
	gapHeight := MinGapSize + pm.rng.Intn(MaxGapSize-MinGapSize+1)

	// Calculate valid Y range for gap
	maxGapY := pm.screenH - PipeBottomMargin - gapHeight
	minGapY := PipeTopMargin

	if maxGapY < minGapY {
		maxGapY = minGapY // Edge case for very small screens
	}

	gapY := minGapY + pm.rng.Intn(maxGapY-minGapY+1)

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
	for _, p := range pm.pipes {
		topRect := p.TopRect()
		bottomRect := p.BottomRect(screenH)

		if playerRect.Intersects(topRect) || playerRect.Intersects(bottomRect) {
			return true
		}
	}
	return false
}
