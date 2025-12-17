package pong

import (
	"math"

	"github.com/vovakirdan/tui-arcade/internal/multiplayer"
)

// PongSnapshot contains the complete state of a Pong game for network transmission.
// Uses primitive types only for stable serialization.
type PongSnapshot struct {
	Tick     uint64
	BallX    int
	BallY    int
	BallVX   int // Velocity scaled by 1000 (for precision)
	BallVY   int // Velocity scaled by 1000
	Paddle1Y int
	Paddle2Y int
	Score1   int
	Score2   int
	GameOver bool
	Winner   int // 0=none, 1=Player1, 2=Player2
	Serving  bool
}

// IsGameSnapshot implements the GameSnapshot interface marker.
func (PongSnapshot) IsGameSnapshot() {}

// Ensure PongSnapshot implements multiplayer.GameSnapshot
var _ multiplayer.GameSnapshot = PongSnapshot{}

// Snapshot returns the current game state as a PongSnapshot.
func (g *Game) Snapshot() multiplayer.GameSnapshot {
	return PongSnapshot{
		Tick:     uint64(max(0, g.tickCount)), //nolint:gosec // tickCount is always non-negative in game logic
		BallX:    int(g.ballX),
		BallY:    int(g.ballY),
		BallVX:   int(g.ballVX * 1000),
		BallVY:   int(g.ballVY * 1000),
		Paddle1Y: int(g.paddle1Y),
		Paddle2Y: int(g.paddle2Y),
		Score1:   g.score1,
		Score2:   g.score2,
		GameOver: g.gameOver,
		Winner:   g.winner,
		Serving:  g.serving,
	}
}

// ApplySnapshot updates the game state from a snapshot.
// Used by clients to sync with server state.
func (g *Game) ApplySnapshot(snap PongSnapshot) {
	g.tickCount = int(min(snap.Tick, math.MaxInt)) //nolint:gosec // clamped to max int
	g.ballX = float64(snap.BallX)
	g.ballY = float64(snap.BallY)
	g.ballVX = float64(snap.BallVX) / 1000.0
	g.ballVY = float64(snap.BallVY) / 1000.0
	g.paddle1Y = float64(snap.Paddle1Y)
	g.paddle2Y = float64(snap.Paddle2Y)
	g.score1 = snap.Score1
	g.score2 = snap.Score2
	g.gameOver = snap.GameOver
	g.winner = snap.Winner
	g.serving = snap.Serving
}
