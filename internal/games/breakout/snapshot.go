package breakout

// Snapshot contains the complete game state for replay/save/multiplayer.
// Uses primitive types only for stable serialization.
type Snapshot struct {
	Tick            uint64
	PaddleX         int
	PaddleWidth     int
	Score           int
	Lives           int
	LevelIndex      int
	BricksRemaining int
	State           string
	ServeDelay      int

	// Game mode and endless tracking
	Mode         int // 0=Campaign, 1=Endless
	EndlessCycle int
	BallSpeed    int // Current base ball speed (fixed-point)

	// Multi-ball state (each ball is 6 ints: X, Y, VX, VY, Stuck, Active)
	BallCount int
	BallData  []int

	// Pickup state (each pickup is 5 ints: Type, X, Y, VY, Active)
	PickupCount int
	PickupData  []int

	// Effect state (each effect is 2 ints: Type, UntilTick)
	EffectCount int
	EffectData  []int

	// Brick states (flattened: row*width + col = index)
	// Each brick is 2 ints: Alive, HP
	BrickData []int

	// RNG state for power-up manager
	RNGState uint64
}

// Snapshot returns the current game state as a Snapshot.
func (g *Game) Snapshot() Snapshot {
	// Flatten brick states
	brickCount := g.level.Width * g.level.Height
	brickData := make([]int, brickCount*2)

	for row := range g.level.Height {
		for col := range g.level.Width {
			idx := (row*g.level.Width + col) * 2
			brick := g.level.Bricks[row][col]
			if brick.Alive {
				brickData[idx] = 1
			} else {
				brickData[idx] = 0
			}
			brickData[idx+1] = brick.HP
		}
	}

	// Flatten ball states
	ballData := make([]int, len(g.balls)*6)
	for i, ball := range g.balls {
		idx := i * 6
		ballData[idx] = int(ball.X)
		ballData[idx+1] = int(ball.Y)
		ballData[idx+2] = int(ball.VX)
		ballData[idx+3] = int(ball.VY)
		if ball.Stuck {
			ballData[idx+4] = 1
		}
		if ball.Active {
			ballData[idx+5] = 1
		}
	}

	// Flatten pickup states
	pickupData := make([]int, len(g.powerups.Pickups)*5)
	for i, pickup := range g.powerups.Pickups {
		idx := i * 5
		pickupData[idx] = int(pickup.Type)
		pickupData[idx+1] = int(pickup.X)
		pickupData[idx+2] = int(pickup.Y)
		pickupData[idx+3] = int(pickup.VY)
		if pickup.Active {
			pickupData[idx+4] = 1
		}
	}

	// Flatten effect states
	effectData := make([]int, len(g.powerups.Effects)*2)
	for i, effect := range g.powerups.Effects {
		idx := i * 2
		effectData[idx] = int(effect.Type)
		effectData[idx+1] = effect.UntilTick
	}

	return Snapshot{
		Tick:            uint64(g.tickCount), //#nosec G115 -- tick count is always positive
		PaddleX:         int(g.paddle.X),
		PaddleWidth:     g.paddle.Width,
		Score:           g.score,
		Lives:           g.lives,
		LevelIndex:      g.levelIndex,
		BricksRemaining: g.level.CountAlive(),
		State:           g.state,
		ServeDelay:      g.serveDelay,

		Mode:         int(g.mode),
		EndlessCycle: g.endlessCycle,
		BallSpeed:    int(g.currentBallSpeed),

		BallCount:   len(g.balls),
		BallData:    ballData,
		PickupCount: len(g.powerups.Pickups),
		PickupData:  pickupData,
		EffectCount: len(g.powerups.Effects),
		EffectData:  effectData,

		BrickData: brickData,
		RNGState:  g.powerups.RNG.state,
	}
}

// ApplySnapshot restores game state from a snapshot.
func (g *Game) ApplySnapshot(snap Snapshot) {
	g.tickCount = int(snap.Tick) //#nosec G115 -- tick count fits in int
	g.paddle.X = Fixed(snap.PaddleX)
	g.paddle.Width = snap.PaddleWidth
	g.score = snap.Score
	g.lives = snap.Lives
	g.levelIndex = snap.LevelIndex
	g.state = snap.State
	g.serveDelay = snap.ServeDelay

	g.mode = GameMode(snap.Mode)
	g.endlessCycle = snap.EndlessCycle
	g.currentBallSpeed = Fixed(snap.BallSpeed)

	// Restore brick states
	if g.level != nil && len(snap.BrickData) == g.level.Width*g.level.Height*2 {
		for row := range g.level.Height {
			for col := range g.level.Width {
				idx := (row*g.level.Width + col) * 2
				g.level.Bricks[row][col].Alive = snap.BrickData[idx] == 1
				g.level.Bricks[row][col].HP = snap.BrickData[idx+1]
			}
		}
	}

	// Restore ball states
	g.balls = make([]*Ball, snap.BallCount)
	for i := range snap.BallCount {
		idx := i * 6
		if idx+5 < len(snap.BallData) {
			g.balls[i] = &Ball{
				X:      Fixed(snap.BallData[idx]),
				Y:      Fixed(snap.BallData[idx+1]),
				VX:     Fixed(snap.BallData[idx+2]),
				VY:     Fixed(snap.BallData[idx+3]),
				Stuck:  snap.BallData[idx+4] == 1,
				Active: snap.BallData[idx+5] == 1,
			}
		}
	}

	// Restore pickup states
	g.powerups.Pickups = make([]*Pickup, snap.PickupCount)
	for i := range snap.PickupCount {
		idx := i * 5
		if idx+4 < len(snap.PickupData) {
			g.powerups.Pickups[i] = &Pickup{
				Type:   PickupType(snap.PickupData[idx]),
				X:      Fixed(snap.PickupData[idx+1]),
				Y:      Fixed(snap.PickupData[idx+2]),
				VY:     Fixed(snap.PickupData[idx+3]),
				Active: snap.PickupData[idx+4] == 1,
			}
		}
	}

	// Restore effect states
	g.powerups.Effects = make([]*Effect, snap.EffectCount)
	for i := range snap.EffectCount {
		idx := i * 2
		if idx+1 < len(snap.EffectData) {
			g.powerups.Effects[i] = &Effect{
				Type:      EffectType(snap.EffectData[idx]),
				UntilTick: snap.EffectData[idx+1],
			}
		}
	}

	// Restore RNG state
	g.powerups.RNG.state = snap.RNGState
}

// Hash returns a simple hash of the snapshot for determinism testing.
func (snap *Snapshot) Hash() uint64 {
	h := uint64(snap.Tick)
	h = h*31 + uint64(snap.PaddleX)         //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.PaddleWidth)     //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.Score)           //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.Lives)           //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.BricksRemaining) //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.Mode)            //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.EndlessCycle)    //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.BallSpeed)       //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.BallCount)       //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.PickupCount)     //#nosec G115 -- hash computation
	h = h*31 + uint64(snap.EffectCount)     //#nosec G115 -- hash computation

	for _, v := range snap.BallData {
		h = h*31 + uint64(v) //#nosec G115 -- hash computation
	}

	for _, v := range snap.PickupData {
		h = h*31 + uint64(v) //#nosec G115 -- hash computation
	}

	for _, v := range snap.EffectData {
		h = h*31 + uint64(v) //#nosec G115 -- hash computation
	}

	for _, v := range snap.BrickData {
		h = h*31 + uint64(v) //#nosec G115 -- hash computation
	}

	h = h*31 + snap.RNGState

	return h
}
