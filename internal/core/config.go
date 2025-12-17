package core

// RuntimeConfig contains configuration passed to games at initialization.
// Games use this to adapt to screen size and for deterministic simulation.
type RuntimeConfig struct {
	ScreenW  int   // Screen width in characters
	ScreenH  int   // Screen height in characters
	TickRate int   // Simulation ticks per second (default 60)
	Seed     int64 // RNG seed for deterministic gameplay
}

// DefaultConfig returns a RuntimeConfig with sensible defaults.
func DefaultConfig() RuntimeConfig {
	return RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     0, // 0 means use current time in platform layer
	}
}

// GameState represents the current state of a game.
// Returned by Game.State() to communicate status to the platform.
type GameState struct {
	Score    int  // Current score
	GameOver bool // Whether the game has ended
	Paused   bool // Whether the game is paused
}

// StepResult is returned by Game.Step() after each simulation tick.
// Contains the updated game state and any events that occurred.
type StepResult struct {
	State GameState
}
