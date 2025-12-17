package breakout

// Snapshot contains the complete game state for replay/save/multiplayer.
// Uses primitive types only for stable serialization.
type Snapshot struct {
	Tick            uint64
	PaddleX         int
	BallX           int
	BallY           int
	BallVX          int
	BallVY          int
	Score           int
	Lives           int
	LevelIndex      int
	BricksRemaining int
	State           string
	ServeDelay      int

	// Brick states (flattened: row*width + col = index)
	BrickAlive []bool
}

// Snapshot returns the current game state as a Snapshot.
func (g *Game) Snapshot() Snapshot {
	// Flatten brick alive states
	brickCount := g.level.Width * g.level.Height
	brickAlive := make([]bool, brickCount)

	for row := range g.level.Height {
		for col := range g.level.Width {
			idx := row*g.level.Width + col
			brickAlive[idx] = g.level.Bricks[row][col].Alive
		}
	}

	return Snapshot{
		Tick:            uint64(g.tickCount), //#nosec G115 -- tick count is always positive
		PaddleX:         int(g.paddle.X),
		BallX:           int(g.ball.X),
		BallY:           int(g.ball.Y),
		BallVX:          int(g.ball.VX),
		BallVY:          int(g.ball.VY),
		Score:           g.score,
		Lives:           g.lives,
		LevelIndex:      g.levelIndex,
		BricksRemaining: g.level.CountAlive(),
		State:           g.state,
		ServeDelay:      g.serveDelay,
		BrickAlive:      brickAlive,
	}
}

// ApplySnapshot restores game state from a snapshot.
func (g *Game) ApplySnapshot(snap Snapshot) {
	g.tickCount = int(snap.Tick) //#nosec G115 -- tick count fits in int
	g.paddle.X = Fixed(snap.PaddleX)
	g.ball.X = Fixed(snap.BallX)
	g.ball.Y = Fixed(snap.BallY)
	g.ball.VX = Fixed(snap.BallVX)
	g.ball.VY = Fixed(snap.BallVY)
	g.score = snap.Score
	g.lives = snap.Lives
	g.levelIndex = snap.LevelIndex
	g.state = snap.State
	g.serveDelay = snap.ServeDelay

	// Restore brick states
	if g.level != nil && len(snap.BrickAlive) == g.level.Width*g.level.Height {
		for row := range g.level.Height {
			for col := range g.level.Width {
				idx := row*g.level.Width + col
				g.level.Bricks[row][col].Alive = snap.BrickAlive[idx]
			}
		}
	}
}

// Hash returns a simple hash of the snapshot for determinism testing.
func (snap *Snapshot) Hash() uint64 {
	h := uint64(snap.Tick)
	h = h*31 + uint64(snap.PaddleX)         //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.BallX)           //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.BallY)           //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.BallVX)          //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.BallVY)          //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.Score)           //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.Lives)           //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.BricksRemaining) //#nosec G115 -- hash computation

	for _, alive := range snap.BrickAlive {
		if alive {
			h = h*31 + 1
		} else {
			h = h*31 + 0
		}
	}

	return h
}
