package t2048

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
	Tick    uint64
	Mode    string // "campaign" or "endless"
	Level   int    // Current level (1-indexed for display), 0 for endless
	Target  int    // Current target tile value
	Score   int
	Board   [BoardSize][BoardSize]int
	MaxTile int // Highest tile on board
	State   GameStateType
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

	return Snapshot{
		Tick:    g.tick,
		Mode:    string(g.mode),
		Level:   g.levelIndex + 1,
		Target:  g.currentTarget,
		Score:   g.score,
		Board:   g.board,
		MaxTile: MaxTile(g.board),
		State:   state,
	}
}
