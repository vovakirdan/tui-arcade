// Package multiplayer provides types and abstractions for multiplayer game support.
// Currently used for Player vs CPU modes, designed for future SSH session-to-session matches.
package multiplayer

import "github.com/vovakirdan/tui-arcade/internal/core"

// PlayerID is an alias to core.PlayerID for convenience.
// Player1 is always the local human player, Player2 can be CPU or remote player.
type PlayerID = core.PlayerID

// Re-export player constants for convenience.
const (
	Player1 = core.Player1
	Player2 = core.Player2
)

// SessionID uniquely identifies a player's session (e.g., SSH connection).
// Used to track individual connections and potentially pair them for matches.
type SessionID string

// MatchID uniquely identifies a game match.
// A match can involve one or more sessions depending on mode.
type MatchID string

// MatchMode defines how a game match is configured.
type MatchMode int

const (
	// MatchModeSolo is a single-player game (runner games like flappy, dino).
	MatchModeSolo MatchMode = iota

	// MatchModeVsCPU is player vs computer (Pong vs AI).
	MatchModeVsCPU

	// MatchModeOnlinePvP is reserved for future player vs player over network.
	// Not implemented in v0.2 but the type exists for API stability.
	MatchModeOnlinePvP
)

// String returns a human-readable name for the match mode.
func (m MatchMode) String() string {
	switch m {
	case MatchModeSolo:
		return "Solo"
	case MatchModeVsCPU:
		return "vs CPU"
	case MatchModeOnlinePvP:
		return "Online PvP"
	default:
		return "Unknown"
	}
}

// MatchHandle provides access to match metadata.
// Games receive this to know their context without managing match lifecycle.
type MatchHandle interface {
	// ID returns the unique identifier for this match.
	ID() MatchID

	// Mode returns how this match is configured.
	Mode() MatchMode
}

// Match is a concrete implementation of MatchHandle.
// Platform creates matches and passes handles to games.
type Match struct {
	id   MatchID
	mode MatchMode

	// SessionIDs tracks which sessions are part of this match.
	// For Solo/VsCPU: one session. For OnlinePvP: two sessions.
	SessionIDs []SessionID
}

// NewMatch creates a new match with the given parameters.
func NewMatch(id MatchID, mode MatchMode, sessions ...SessionID) *Match {
	return &Match{
		id:         id,
		mode:       mode,
		SessionIDs: sessions,
	}
}

// ID returns the match identifier.
func (m *Match) ID() MatchID {
	return m.id
}

// Mode returns the match mode.
func (m *Match) Mode() MatchMode {
	return m.mode
}

// Sessions returns the session IDs participating in this match.
func (m *Match) Sessions() []SessionID {
	return m.SessionIDs
}
