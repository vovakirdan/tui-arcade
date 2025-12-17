package multiplayer

import "github.com/vovakirdan/tui-arcade/internal/core"

// SessionEvent represents an event sent from the coordinator to a session.
type SessionEvent interface {
	sessionEvent()
}

// LobbyCreatedEvent is sent when a lobby is successfully created.
type LobbyCreatedEvent struct {
	Code   string
	GameID string
}

func (LobbyCreatedEvent) sessionEvent() {}

// LobbyErrorEvent is sent when a lobby operation fails.
type LobbyErrorEvent struct {
	Message string
}

func (LobbyErrorEvent) sessionEvent() {}

// LobbyJoinedEvent is sent to both host and joiner when someone joins.
type LobbyJoinedEvent struct {
	Code       string
	Side       PlayerID // Which side this session plays (Player1 or Player2)
	OpponentID SessionID
}

func (LobbyJoinedEvent) sessionEvent() {}

// LobbyPlayerLeftEvent is sent when a player leaves the lobby before match starts.
type LobbyPlayerLeftEvent struct {
	Code string
}

func (LobbyPlayerLeftEvent) sessionEvent() {}

// MatchStartedEvent is sent when the match begins.
type MatchStartedEvent struct {
	MatchID MatchID
	Side    PlayerID
	Code    string // Keep code for display
}

func (MatchStartedEvent) sessionEvent() {}

// MatchEndedEvent is sent when the match ends.
type MatchEndedEvent struct {
	MatchID MatchID
	Reason  MatchEndReason
	Winner  PlayerID // 0 if no winner (disconnect)
	Score1  int
	Score2  int
}

func (MatchEndedEvent) sessionEvent() {}

// MatchEndReason describes why a match ended.
type MatchEndReason int

const (
	MatchEndReasonCompleted  MatchEndReason = iota // Normal game completion
	MatchEndReasonDisconnect                       // Opponent disconnected
	MatchEndReasonCancelled                        // Match was cancelled
	MatchEndReasonHostLeft                         // Host left the lobby
	MatchEndReasonJoinerLeft                       // Joiner left the lobby
)

func (r MatchEndReason) String() string {
	switch r {
	case MatchEndReasonCompleted:
		return "Match completed"
	case MatchEndReasonDisconnect:
		return "Opponent disconnected"
	case MatchEndReasonCancelled:
		return "Match cancelled"
	case MatchEndReasonHostLeft:
		return "Host left"
	case MatchEndReasonJoinerLeft:
		return "Opponent left"
	default:
		return "Unknown"
	}
}

// SnapshotEvent carries a game state snapshot to sessions.
type SnapshotEvent struct {
	MatchID  MatchID
	Tick     uint64
	Snapshot GameSnapshot
}

func (SnapshotEvent) sessionEvent() {}

// GameSnapshot is the interface for game-specific snapshot data.
type GameSnapshot interface {
	IsGameSnapshot() // Marker method for type safety
}

// CoordinatorMessage represents a message from a session to the coordinator.
type CoordinatorMessage interface {
	coordinatorMessage()
}

// CreateLobbyMsg requests creation of a new lobby.
type CreateLobbyMsg struct {
	SessionID SessionID
	GameID    string
}

func (CreateLobbyMsg) coordinatorMessage() {}

// JoinLobbyMsg requests joining an existing lobby.
type JoinLobbyMsg struct {
	SessionID SessionID
	Code      string
}

func (JoinLobbyMsg) coordinatorMessage() {}

// CancelLobbyMsg requests cancellation of a hosted lobby.
type CancelLobbyMsg struct {
	SessionID SessionID
	Code      string
}

func (CancelLobbyMsg) coordinatorMessage() {}

// LeaveLobbyMsg requests leaving a joined lobby.
type LeaveLobbyMsg struct {
	SessionID SessionID
	Code      string
}

func (LeaveLobbyMsg) coordinatorMessage() {}

// LeaveMatchMsg requests leaving an active match.
type LeaveMatchMsg struct {
	SessionID SessionID
	MatchID   MatchID
}

func (LeaveMatchMsg) coordinatorMessage() {}

// PlayerInputMsg sends player input to a match.
type PlayerInputMsg struct {
	MatchID  MatchID
	Player   PlayerID
	TickHint uint64 // Optional client tick counter
	Input    core.InputFrame
}

func (PlayerInputMsg) coordinatorMessage() {}

// ReadyForRematchMsg signals readiness for a rematch.
type ReadyForRematchMsg struct {
	SessionID SessionID
	MatchID   MatchID
}

func (ReadyForRematchMsg) coordinatorMessage() {}

// SessionDisconnectedMsg is sent when a session disconnects.
type SessionDisconnectedMsg struct {
	SessionID SessionID
}

func (SessionDisconnectedMsg) coordinatorMessage() {}
