package multiplayer

import (
	"sync"
	"time"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

// OnlineGame is the interface that games must implement to support online multiplayer.
// It extends the basic game interface with multiplayer-specific methods.
type OnlineGame interface {
	// Reset initializes the game state.
	Reset(cfg core.RuntimeConfig)

	// StepMulti advances the game by one tick using input from multiple players.
	StepMulti(input core.MultiInputFrame) core.StepResult

	// Snapshot returns the current game state as a snapshot for network transmission.
	Snapshot() GameSnapshot

	// IsGameOver returns true if the game has ended.
	IsGameOver() bool

	// Winner returns the winning player (Player1/Player2) or 0 if no winner yet.
	Winner() PlayerID

	// Score1 returns Player 1's score.
	Score1() int

	// Score2 returns Player 2's score.
	Score2() int
}

// MatchResult contains the outcome of a completed match.
type MatchResult struct {
	MatchID MatchID
	Reason  MatchEndReason
	Winner  PlayerID
	Score1  int
	Score2  int
	Ticks   uint64
}

// OnlineMatch represents an active multiplayer game session.
type OnlineMatch struct {
	id     MatchID
	code   string
	gameID string
	game   OnlineGame

	player1Session SessionHandle
	player2Session SessionHandle

	// Input handling
	inputMu    sync.Mutex
	lastInput1 core.InputFrame
	lastInput2 core.InputFrame
	inputChan  chan playerInput

	// Match state
	tick     uint64
	tickRate int
	done     chan struct{}
	doneOnce sync.Once

	// Disconnect handling
	disconnectChan chan SessionID
}

type playerInput struct {
	player PlayerID
	input  core.InputFrame
}

// NewOnlineMatch creates a new online match.
func NewOnlineMatch(
	id MatchID,
	code string,
	gameID string,
	game OnlineGame,
	p1Session, p2Session SessionHandle,
	tickRate int,
) *OnlineMatch {
	return &OnlineMatch{
		id:             id,
		code:           code,
		gameID:         gameID,
		game:           game,
		player1Session: p1Session,
		player2Session: p2Session,
		lastInput1:     core.NewInputFrame(),
		lastInput2:     core.NewInputFrame(),
		inputChan:      make(chan playerInput, 64),
		tick:           0,
		tickRate:       tickRate,
		done:           make(chan struct{}),
		disconnectChan: make(chan SessionID, 2),
	}
}

// ID returns the match identifier.
func (m *OnlineMatch) ID() MatchID {
	return m.id
}

// Code returns the join code used to create this match.
func (m *OnlineMatch) Code() string {
	return m.code
}

// GameID returns the game identifier.
func (m *OnlineMatch) GameID() string {
	return m.gameID
}

// SendInput sends player input to the match.
// Non-blocking, uses a buffered channel.
func (m *OnlineMatch) SendInput(player PlayerID, input core.InputFrame) {
	select {
	case m.inputChan <- playerInput{player: player, input: input}:
	default:
		// Channel full, drop input (rare under normal conditions)
	}
}

// PlayerDisconnected signals that a player has disconnected.
func (m *OnlineMatch) PlayerDisconnected(sessionID SessionID) {
	select {
	case m.disconnectChan <- sessionID:
	default:
	}
}

// Run starts the authoritative match loop.
// The callback is called when the match ends.
func (m *OnlineMatch) Run(onComplete func(MatchResult)) {
	defer func() {
		m.doneOnce.Do(func() {
			close(m.done)
		})
	}()

	tickDuration := time.Second / time.Duration(m.tickRate)
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	// Monitor session disconnects
	go m.monitorSessions()

	for {
		select {
		case <-ticker.C:
			result, done := m.runTick()
			if done {
				if onComplete != nil {
					onComplete(result)
				}
				return
			}

		case sessionID := <-m.disconnectChan:
			result := m.handleDisconnect(sessionID)
			if onComplete != nil {
				onComplete(result)
			}
			return

		case <-m.done:
			return
		}
	}
}

func (m *OnlineMatch) runTick() (MatchResult, bool) {
	// Drain input channel and update last known inputs
	m.drainInputs()

	// Build multi-input frame
	m.inputMu.Lock()
	multiInput := core.NewMultiInputFrame()
	multiInput.SetPlayer(Player1, m.lastInput1.Clone())
	multiInput.SetPlayer(Player2, m.lastInput2.Clone())
	// Clear inputs after use (they're "consumed" this tick)
	m.lastInput1.Clear()
	m.lastInput2.Clear()
	m.inputMu.Unlock()

	// Run game simulation
	m.game.StepMulti(multiInput)
	m.tick++

	// Broadcast snapshot to both sessions
	snapshot := m.game.Snapshot()
	snapshotEvent := SnapshotEvent{
		MatchID:  m.id,
		Tick:     m.tick,
		Snapshot: snapshot,
	}
	m.player1Session.Send(snapshotEvent)
	m.player2Session.Send(snapshotEvent)

	// Check for game over
	if m.game.IsGameOver() {
		return MatchResult{
			MatchID: m.id,
			Reason:  MatchEndReasonCompleted,
			Winner:  m.game.Winner(),
			Score1:  m.game.Score1(),
			Score2:  m.game.Score2(),
			Ticks:   m.tick,
		}, true
	}

	return MatchResult{}, false
}

func (m *OnlineMatch) drainInputs() {
	m.inputMu.Lock()
	defer m.inputMu.Unlock()

	for {
		select {
		case pi := <-m.inputChan:
			if pi.player == Player1 {
				// Merge inputs (OR together actions)
				for action, pressed := range pi.input.Actions {
					if pressed {
						m.lastInput1.Set(action)
					}
				}
			} else {
				for action, pressed := range pi.input.Actions {
					if pressed {
						m.lastInput2.Set(action)
					}
				}
			}
		default:
			return
		}
	}
}

func (m *OnlineMatch) handleDisconnect(sessionID SessionID) MatchResult {
	var winner PlayerID
	var reason MatchEndReason

	if sessionID == m.player1Session.ID() {
		winner = Player2
		reason = MatchEndReasonDisconnect
	} else {
		winner = Player1
		reason = MatchEndReasonDisconnect
	}

	return MatchResult{
		MatchID: m.id,
		Reason:  reason,
		Winner:  winner,
		Score1:  m.game.Score1(),
		Score2:  m.game.Score2(),
		Ticks:   m.tick,
	}
}

func (m *OnlineMatch) monitorSessions() {
	select {
	case <-m.player1Session.Done():
		select {
		case m.disconnectChan <- m.player1Session.ID():
		default:
		}
	case <-m.player2Session.Done():
		select {
		case m.disconnectChan <- m.player2Session.ID():
		default:
		}
	case <-m.done:
	}
}

// Stop gracefully stops the match.
func (m *OnlineMatch) Stop() {
	m.doneOnce.Do(func() {
		close(m.done)
	})
}
