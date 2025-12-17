package snake

// GameStateType represents the current game state.
type GameStateType string

const (
	StatePlaying      GameStateType = "playing"
	StateLevelCleared GameStateType = "level_cleared"
	StateGameOver     GameStateType = "game_over"
	StateWin          GameStateType = "win"
	StatePausedSmall  GameStateType = "paused_small_window"
)

// Snapshot captures the complete game state for determinism testing and replay.
type Snapshot struct {
	Tick           uint64
	Level          int    // Current level (1-indexed for display)
	Mode           string // "campaign" or "endless"
	Score          int
	FoodEaten      int // Food eaten in current level
	SnakeLen       int
	HeadX          int
	HeadY          int
	Dir            Direction
	FoodX          int
	FoodY          int
	MoveEveryTicks int
	State          GameStateType
}

// Snapshot returns the current game snapshot for determinism verification.
func (g *Game) Snapshot() Snapshot {
	state := StatePlaying
	switch {
	case g.tooSmall:
		state = StatePausedSmall
	case g.won:
		state = StateWin
	case g.gameOver:
		state = StateGameOver
	case g.levelCleared:
		state = StateLevelCleared
	}

	headX, headY := 0, 0
	if len(g.snake) > 0 {
		headX = g.snake[0].X
		headY = g.snake[0].Y
	}

	return Snapshot{
		Tick:           g.tick,
		Level:          g.levelIndex + 1,
		Mode:           string(g.mode),
		Score:          g.score,
		FoodEaten:      g.foodEaten,
		SnakeLen:       len(g.snake),
		HeadX:          headX,
		HeadY:          headY,
		Dir:            g.direction,
		FoodX:          g.food.X,
		FoodY:          g.food.Y,
		MoveEveryTicks: g.moveEveryTicks,
		State:          state,
	}
}
