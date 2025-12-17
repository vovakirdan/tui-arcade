package dino

import (
	"math/rand"

	"github.com/vovakirdan/tui-arcade/internal/config"
	"github.com/vovakirdan/tui-arcade/internal/core"
)

// Cactus represents a ground obstacle the player must jump over.
type Cactus struct {
	X      int // Horizontal position (left edge)
	Width  int // Width in characters
	Height int // Height in characters
}

// Rect returns the collision rectangle for this cactus.
func (c Cactus) Rect(groundY int) core.Rect {
	return core.NewRect(c.X, groundY-c.Height, c.Width, c.Height)
}

// ObstacleManager handles spawning, movement, and removal of cacti.
type ObstacleManager struct {
	cacti      []Cactus
	rng        *rand.Rand
	screenW    int
	nextSpawnX int // X position where next cactus will spawn
	cfg        *config.DinoConfig
	difficulty *config.DifficultyManager
}

// NewObstacleManager creates a new obstacle manager with the given RNG seed.
func NewObstacleManager(seed int64, screenW int, cfg *config.DinoConfig, diff *config.DifficultyManager) *ObstacleManager {
	om := &ObstacleManager{
		cacti:      make([]Cactus, 0, 8),
		rng:        rand.New(rand.NewSource(seed)),
		screenW:    screenW,
		cfg:        cfg,
		difficulty: diff,
	}
	om.Reset(seed)
	return om
}

// UpdateConfig updates the configuration.
func (om *ObstacleManager) UpdateConfig(cfg *config.DinoConfig, diff *config.DifficultyManager) {
	om.cfg = cfg
	om.difficulty = diff
}

// Reset clears all obstacles and resets the RNG.
func (om *ObstacleManager) Reset(seed int64) {
	om.cacti = om.cacti[:0]
	om.rng = rand.New(rand.NewSource(seed))
	om.nextSpawnX = om.screenW + om.cfg.Obstacles.MinSpacing // First obstacle spawns off-screen
}

// UpdateScreenSize updates the screen width.
func (om *ObstacleManager) UpdateScreenSize(screenW int) {
	om.screenW = screenW
}

// Update moves obstacles left and spawns new ones as needed.
func (om *ObstacleManager) Update(score int, ticks int) {
	// Calculate current speed based on difficulty
	speed := om.difficulty.Speed(om.cfg.Physics.BaseSpeed, score, ticks)
	speedInt := int(speed)
	if speedInt < 1 {
		speedInt = 1
	}

	// Move cacti left
	for i := range om.cacti {
		om.cacti[i].X -= speedInt
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
	om.nextSpawnX -= speedInt

	// Spawn new cactus if needed
	if om.nextSpawnX <= om.screenW {
		om.spawnCactus(score, ticks)
	}
}

// spawnCactus creates a new cactus at the spawn position.
func (om *ObstacleManager) spawnCactus(score int, ticks int) {
	minW := om.cfg.Obstacles.MinWidth
	maxW := om.cfg.Obstacles.MaxWidth
	minH := om.cfg.Obstacles.MinHeight
	maxH := om.cfg.Obstacles.MaxHeight

	width := minW
	if maxW > minW {
		width = minW + om.rng.Intn(maxW-minW+1)
	}

	height := minH
	if maxH > minH {
		height = minH + om.rng.Intn(maxH-minH+1)
	}

	cactus := Cactus{
		X:      om.nextSpawnX,
		Width:  width,
		Height: height,
	}

	om.cacti = append(om.cacti, cactus)

	// Calculate current spacing based on difficulty
	baseSpacing := om.cfg.Obstacles.MaxSpacing
	currentSpacing := om.difficulty.Spacing(baseSpacing, score, ticks)

	// Ensure minimum spacing
	minSpacing := om.cfg.Obstacles.MinSpacing
	if currentSpacing < minSpacing {
		currentSpacing = minSpacing
	}

	// Random variation in spacing
	spacingRange := currentSpacing - minSpacing
	if spacingRange < 0 {
		spacingRange = 0
	}

	spacing := minSpacing
	if spacingRange > 0 {
		spacing = minSpacing + om.rng.Intn(spacingRange+1)
	}

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
