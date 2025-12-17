package t2048

// Animation constants
const (
	slideAnimationDuration = 8 // ~133ms at 60fps
	popAnimationDuration   = 6 // ~100ms at 60fps
)

// TileAnimation represents an animated tile.
type TileAnimation struct {
	Value    int     // Tile value
	FromX    int     // Start position X (in cells)
	FromY    int     // Start position Y (in cells)
	ToX      int     // End position X (in cells)
	ToY      int     // End position Y (in cells)
	Progress float64 // 0.0 → 1.0
	Merged   bool    // Result of a merge (for visual effect)
	IsNew    bool    // New tile (for pop effect)
}

// TileMove represents a tile movement from one position to another.
type TileMove struct {
	FromX  int
	FromY  int
	ToX    int
	ToY    int
	Value  int  // Original value (before merge)
	Merged bool // Whether this tile merged with another
}

// AnimationPhase represents the current phase of animation.
type AnimationPhase int

const (
	PhaseNone AnimationPhase = iota
	PhaseSlide
	PhasePop
)

// startSlideAnimation initializes slide animations from move tracking.
func (g *Game) startSlideAnimation(moves []TileMove) {
	g.animations = nil
	for _, m := range moves {
		g.animations = append(g.animations, TileAnimation{
			Value:    m.Value,
			FromX:    m.FromX,
			FromY:    m.FromY,
			ToX:      m.ToX,
			ToY:      m.ToY,
			Progress: 0,
			Merged:   m.Merged,
			IsNew:    false,
		})
	}
	g.animating = true
	g.animationPhase = PhaseSlide
	g.animationTicks = 0
}

// startPopAnimation initializes pop animation for a new tile.
func (g *Game) startPopAnimation(x, y, value int) {
	g.animations = []TileAnimation{
		{
			Value:    value,
			FromX:    x,
			FromY:    y,
			ToX:      x,
			ToY:      y,
			Progress: 0,
			Merged:   false,
			IsNew:    true,
		},
	}
	g.animating = true
	g.animationPhase = PhasePop
	g.animationTicks = 0
}

// updateAnimation advances the animation state.
// Returns true if animation is still in progress.
func (g *Game) updateAnimation() bool {
	if !g.animating {
		return false
	}

	g.animationTicks++

	// Determine duration based on phase
	var duration int
	switch g.animationPhase {
	case PhaseSlide:
		duration = slideAnimationDuration
	case PhasePop:
		duration = popAnimationDuration
	default:
		g.animating = false
		return false
	}

	// Update progress for all animations
	progress := float64(g.animationTicks) / float64(duration)
	if progress > 1.0 {
		progress = 1.0
	}

	for i := range g.animations {
		g.animations[i].Progress = progress
	}

	// Check if animation is complete
	if g.animationTicks >= duration {
		g.finishAnimation()
		return false
	}

	return true
}

// finishAnimation completes the current animation phase.
func (g *Game) finishAnimation() {
	if g.animationPhase == PhaseSlide && g.pendingNewTile != nil {
		// Start pop animation for new tile
		g.startPopAnimation(g.pendingNewTile.X, g.pendingNewTile.Y, g.pendingNewTile.Value)
		g.pendingNewTile = nil
		return
	}

	// Animation complete
	g.animating = false
	g.animationPhase = PhaseNone
	g.animations = nil
}

// PendingTile stores info about a tile to be animated after slide.
type PendingTile struct {
	X, Y  int
	Value int
}

// easeOutQuad provides smooth deceleration for animation.
func easeOutQuad(t float64) float64 {
	return t * (2 - t)
}

// interpolatePosition calculates the current position during animation.
func (a *TileAnimation) interpolatePosition() (x, y float64) {
	t := easeOutQuad(a.Progress)
	x = float64(a.FromX) + (float64(a.ToX)-float64(a.FromX))*t
	y = float64(a.FromY) + (float64(a.ToY)-float64(a.FromY))*t
	return x, y
}
