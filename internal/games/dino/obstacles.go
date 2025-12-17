package dino

import (
	"math/rand"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

// Cactus represents a ground obstacle the player must jump over.
type Cactus struct {
	X      int  // Horizontal position (left edge)
	Width  int  // Width in characters
	Height int  // Height in characters
}

// Rect returns the collision rectangle for this cactus.
func (c Cactus) Rect(groundY int) core.Rect {
	return core.NewRect(c.X, groundY-c.Height, c.Width, c.Height)
}

// Constants for obstacle configuration
const (
	MinCactusWidth  = 1
	MaxCactusWidth  = 3
	MinCactusHeight = 2
	MaxCactusHeight = 4
	MinSpacing      = 20  // Minimum distance between obstacles
	MaxSpacing      = 40  // Maximum distance between obstacles
)

// ObstacleManager handles spawning, movement, and removal of cacti.
type ObstacleManager struct {
	cacti      []Cactus
	rng        *rand.Rand
	screenW    int
	nextSpawnX int // X position where next cactus will spawn
}

// NewObstacleManager creates a new obstacle manager with the given RNG seed.
func NewObstacleManager(seed int64, screenW int) *ObstacleManager {
	om := &ObstacleManager{
		cacti:   make([]Cactus, 0, 8),
		rng:     rand.New(rand.NewSource(seed)),
		screenW: screenW,
	}
	om.Reset(seed)
	return om
}

// Reset clears all obstacles and resets the RNG.
func (om *ObstacleManager) Reset(seed int64) {
	om.cacti = om.cacti[:0]
	om.rng = rand.New(rand.NewSource(seed))
	om.nextSpawnX = om.screenW + MinSpacing // First obstacle spawns off-screen
}

// UpdateScreenSize updates the screen width.
func (om *ObstacleManager) UpdateScreenSize(screenW int) {
	om.screenW = screenW
}

// Update moves obstacles left and spawns new ones as needed.
// speed determines how fast obstacles move (increases with score).
func (om *ObstacleManager) Update(speed int) {
	// Move cacti left
	for i := range om.cacti {
		om.cacti[i].X -= speed
	}

	// Remove cacti that have moved off the left side
	validCacti := om.cacti[:0]
	for _, c := range om.cacti {
		if c.X+c.Width > 0 {
			validCacti = append(validCacti, c)
		}
	}
	om.cacti = validCacti

	// Update next spawn position
	om.nextSpawnX -= speed

	// Spawn new cactus if needed
	if om.nextSpawnX <= om.screenW {
		om.spawnCactus()
	}
}

// spawnCactus creates a new cactus at the spawn position.
func (om *ObstacleManager) spawnCactus() {
	width := MinCactusWidth + om.rng.Intn(MaxCactusWidth-MinCactusWidth+1)
	height := MinCactusHeight + om.rng.Intn(MaxCactusHeight-MinCactusHeight+1)

	cactus := Cactus{
		X:      om.nextSpawnX,
		Width:  width,
		Height: height,
	}

	om.cacti = append(om.cacti, cactus)

	// Set next spawn position with random spacing
	spacing := MinSpacing + om.rng.Intn(MaxSpacing-MinSpacing+1)
	om.nextSpawnX += width + spacing
}

// Cacti returns the current list of obstacles.
func (om *ObstacleManager) Cacti() []Cactus {
	return om.cacti
}

// CheckCollision tests if the given rectangle collides with any cactus.
func (om *ObstacleManager) CheckCollision(playerRect core.Rect, groundY int) bool {
	for _, c := range om.cacti {
		if playerRect.Intersects(c.Rect(groundY)) {
			return true
		}
	}
	return false
}

// GetFirstCactusX returns the X position of the leftmost cactus, or -1 if none.
// Used for scoring - player gets points when passing obstacles.
func (om *ObstacleManager) GetFirstCactusX() int {
	if len(om.cacti) == 0 {
		return -1
	}
	return om.cacti[0].X
}
