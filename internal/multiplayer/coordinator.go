package multiplayer

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

// Lobby represents a waiting room for a match.
type Lobby struct {
	Code      string
	GameID    string
	Host      SessionHandle
	Joiner    SessionHandle
	CreatedAt time.Time
}

// CoordinatorConfig holds configuration for the coordinator.
type CoordinatorConfig struct {
	LobbyTimeout  time.Duration // How long before an empty lobby expires
	TickRate      int           // Game tick rate (Hz)
	CleanupPeriod time.Duration // How often to clean up expired lobbies
}

// DefaultCoordinatorConfig returns sensible defaults.
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		LobbyTimeout:  2 * time.Minute,
		TickRate:      60,
		CleanupPeriod: 30 * time.Second,
	}
}

// GameFactory creates game instances for matches.
type GameFactory func(gameID string, cfg core.RuntimeConfig) (OnlineGame, error)

// MatchResultSaver is an interface for saving match results.
// This allows the coordinator to save results without depending on the storage package.
type MatchResultSaver interface {
	SaveMatchResult(result MatchResultData) error
}

// MatchResultData contains match result data for persistence.
type MatchResultData struct {
	MatchID        string
	GameID         string
	Player1Session string
	Player2Session string
	Score1         int
	Score2         int
	WinnerSession  string
	EndReason      string
	DurationSecs   int
}

// Coordinator manages lobbies and active matches.
type Coordinator struct {
	config      CoordinatorConfig
	gameFactory GameFactory
	sessions    *SessionRegistry
	resultSaver MatchResultSaver // Optional, can be nil

	mu      sync.RWMutex
	lobbies map[string]*Lobby        // code -> lobby
	matches map[MatchID]*OnlineMatch // matchID -> match

	// Track which session is in which lobby/match
	sessionLobby map[SessionID]string  // sessionID -> lobby code
	sessionMatch map[SessionID]MatchID // sessionID -> matchID

	// Message channel for async processing
	msgChan chan CoordinatorMessage
	done    chan struct{}
}

// NewCoordinator creates a new coordinator.
func NewCoordinator(cfg CoordinatorConfig, factory GameFactory, sessions *SessionRegistry) *Coordinator {
	c := &Coordinator{
		config:       cfg,
		gameFactory:  factory,
		sessions:     sessions,
		lobbies:      make(map[string]*Lobby),
		matches:      make(map[MatchID]*OnlineMatch),
		sessionLobby: make(map[SessionID]string),
		sessionMatch: make(map[SessionID]MatchID),
		msgChan:      make(chan CoordinatorMessage, 256),
		done:         make(chan struct{}),
	}
	return c
}

// SetResultSaver sets the optional match result saver.
func (c *Coordinator) SetResultSaver(saver MatchResultSaver) {
	c.resultSaver = saver
}

// Start begins the coordinator's background processing.
func (c *Coordinator) Start() {
	go c.processMessages()
	go c.cleanupLoop()
}

// Stop shuts down the coordinator.
func (c *Coordinator) Stop() {
	close(c.done)
}

// Send sends a message to the coordinator for async processing.
func (c *Coordinator) Send(msg CoordinatorMessage) {
	select {
	case c.msgChan <- msg:
	case <-c.done:
	}
}

// processMessages handles incoming messages.
func (c *Coordinator) processMessages() {
	for {
		select {
		case msg := <-c.msgChan:
			c.handleMessage(msg)
		case <-c.done:
			return
		}
	}
}

func (c *Coordinator) handleMessage(msg CoordinatorMessage) {
	switch m := msg.(type) {
	case CreateLobbyMsg:
		c.handleCreateLobby(m)
	case JoinLobbyMsg:
		c.handleJoinLobby(m)
	case CancelLobbyMsg:
		c.handleCancelLobby(m)
	case LeaveLobbyMsg:
		c.handleLeaveLobby(m)
	case LeaveMatchMsg:
		c.handleLeaveMatch(m)
	case PlayerInputMsg:
		c.handlePlayerInput(m)
	case SessionDisconnectedMsg:
		c.handleSessionDisconnected(m)
	case ReadyForRematchMsg:
		// TODO: Implement rematch logic
	}
}

func (c *Coordinator) handleCreateLobby(msg CreateLobbyMsg) {
	session, ok := c.sessions.Get(msg.SessionID)
	if !ok {
		return
	}

	// Check if session is already in a lobby
	c.mu.Lock()
	if _, inLobby := c.sessionLobby[msg.SessionID]; inLobby {
		c.mu.Unlock()
		session.Send(LobbyErrorEvent{Message: "Already in a lobby"})
		return
	}

	// Generate unique code
	code := c.generateUniqueCode()

	lobby := &Lobby{
		Code:      code,
		GameID:    msg.GameID,
		Host:      session,
		CreatedAt: time.Now(),
	}

	c.lobbies[code] = lobby
	c.sessionLobby[msg.SessionID] = code
	c.mu.Unlock()

	session.Send(LobbyCreatedEvent{Code: code, GameID: msg.GameID})
}

func (c *Coordinator) handleJoinLobby(msg JoinLobbyMsg) {
	session, ok := c.sessions.Get(msg.SessionID)
	if !ok {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if session is already in a lobby
	if _, inLobby := c.sessionLobby[msg.SessionID]; inLobby {
		session.Send(LobbyErrorEvent{Message: "Already in a lobby"})
		return
	}

	// Find lobby
	code := strings.ToUpper(msg.Code)
	lobby, exists := c.lobbies[code]
	if !exists {
		session.Send(LobbyErrorEvent{Message: "Lobby not found"})
		return
	}

	// Check if lobby is full
	if lobby.Joiner != nil {
		session.Send(LobbyErrorEvent{Message: "Lobby is full"})
		return
	}

	// Can't join your own lobby
	if lobby.Host.ID() == msg.SessionID {
		session.Send(LobbyErrorEvent{Message: "Cannot join your own lobby"})
		return
	}

	// Add joiner
	lobby.Joiner = session
	c.sessionLobby[msg.SessionID] = code

	// Notify both players
	lobby.Host.Send(LobbyJoinedEvent{
		Code:       code,
		Side:       Player1,
		OpponentID: msg.SessionID,
	})
	session.Send(LobbyJoinedEvent{
		Code:       code,
		Side:       Player2,
		OpponentID: lobby.Host.ID(),
	})

	// Start the match
	c.startMatch(lobby)
}

func (c *Coordinator) startMatch(lobby *Lobby) {
	// Must be called with lock held

	// Create match ID
	matchID := MatchID(fmt.Sprintf("match-%s-%d", lobby.Code, time.Now().UnixNano()))

	// Default runtime config for online games
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: c.config.TickRate,
		Seed:     time.Now().UnixNano(),
	}

	// Create game instance
	game, err := c.gameFactory(lobby.GameID, cfg)
	if err != nil {
		lobby.Host.Send(LobbyErrorEvent{Message: "Failed to create game"})
		lobby.Joiner.Send(LobbyErrorEvent{Message: "Failed to create game"})
		return
	}

	// Create online match
	match := NewOnlineMatch(matchID, lobby.Code, lobby.GameID, game, lobby.Host, lobby.Joiner, c.config.TickRate)

	// Track match
	c.matches[matchID] = match
	hostID := lobby.Host.ID()
	joinerID := lobby.Joiner.ID()

	// Update session tracking
	delete(c.sessionLobby, hostID)
	delete(c.sessionLobby, joinerID)
	c.sessionMatch[hostID] = matchID
	c.sessionMatch[joinerID] = matchID

	// Remove lobby
	delete(c.lobbies, lobby.Code)

	// Notify players
	lobby.Host.Send(MatchStartedEvent{
		MatchID: matchID,
		Side:    Player1,
		Code:    lobby.Code,
	})
	lobby.Joiner.Send(MatchStartedEvent{
		MatchID: matchID,
		Side:    Player2,
		Code:    lobby.Code,
	})

	// Start match loop
	go match.Run(func(result MatchResult) {
		c.handleMatchEnded(matchID, result)
	})
}

func (c *Coordinator) handleMatchEnded(matchID MatchID, result MatchResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	match, exists := c.matches[matchID]
	if !exists {
		return
	}

	// Save match result if saver is configured
	if c.resultSaver != nil {
		winnerSession := ""
		if result.Winner == Player1 {
			winnerSession = string(match.player1Session.ID())
		} else if result.Winner == Player2 {
			winnerSession = string(match.player2Session.ID())
		}

		tickRate := max(1, c.config.TickRate) // Ensure positive tick rate
		resultData := MatchResultData{
			MatchID:        string(matchID),
			GameID:         match.GameID(),
			Player1Session: string(match.player1Session.ID()),
			Player2Session: string(match.player2Session.ID()),
			Score1:         result.Score1,
			Score2:         result.Score2,
			WinnerSession:  winnerSession,
			EndReason:      result.Reason.String(),
			DurationSecs:   int(result.Ticks / uint64(tickRate)), //nolint:gosec // tickRate is clamped positive
		}
		// Best effort save, don't block on error
		go func() {
			_ = c.resultSaver.SaveMatchResult(resultData) //nolint:errcheck // intentional fire-and-forget
		}()
	}

	// Clean up session tracking
	for _, sessionID := range []SessionID{match.player1Session.ID(), match.player2Session.ID()} {
		delete(c.sessionMatch, sessionID)
	}

	// Remove match
	delete(c.matches, matchID)

	// Notify players
	endEvent := MatchEndedEvent{
		MatchID: matchID,
		Reason:  result.Reason,
		Winner:  result.Winner,
		Score1:  result.Score1,
		Score2:  result.Score2,
	}
	match.player1Session.Send(endEvent)
	match.player2Session.Send(endEvent)
}

func (c *Coordinator) handleCancelLobby(msg CancelLobbyMsg) {
	c.mu.Lock()
	defer c.mu.Unlock()

	lobby, exists := c.lobbies[msg.Code]
	if !exists {
		return
	}

	// Only host can cancel
	if lobby.Host.ID() != msg.SessionID {
		return
	}

	// Notify joiner if present
	if lobby.Joiner != nil {
		lobby.Joiner.Send(MatchEndedEvent{
			Reason: MatchEndReasonHostLeft,
		})
		delete(c.sessionLobby, lobby.Joiner.ID())
	}

	delete(c.lobbies, msg.Code)
	delete(c.sessionLobby, msg.SessionID)
}

func (c *Coordinator) handleLeaveLobby(msg LeaveLobbyMsg) {
	c.mu.Lock()
	defer c.mu.Unlock()

	lobby, exists := c.lobbies[msg.Code]
	if !exists {
		return
	}

	// If joiner is leaving
	if lobby.Joiner != nil && lobby.Joiner.ID() == msg.SessionID {
		lobby.Joiner = nil
		delete(c.sessionLobby, msg.SessionID)
		lobby.Host.Send(LobbyPlayerLeftEvent{Code: msg.Code})
		return
	}

	// If host is leaving, close lobby
	if lobby.Host.ID() == msg.SessionID {
		if lobby.Joiner != nil {
			lobby.Joiner.Send(MatchEndedEvent{Reason: MatchEndReasonHostLeft})
			delete(c.sessionLobby, lobby.Joiner.ID())
		}
		delete(c.lobbies, msg.Code)
		delete(c.sessionLobby, msg.SessionID)
	}
}

func (c *Coordinator) handleLeaveMatch(msg LeaveMatchMsg) {
	c.mu.Lock()
	match, exists := c.matches[msg.MatchID]
	c.mu.Unlock()

	if !exists {
		return
	}

	// Signal disconnect to match
	match.PlayerDisconnected(msg.SessionID)
}

func (c *Coordinator) handlePlayerInput(msg PlayerInputMsg) {
	c.mu.RLock()
	match, exists := c.matches[msg.MatchID]
	c.mu.RUnlock()

	if !exists {
		return
	}

	match.SendInput(msg.Player, msg.Input)
}

func (c *Coordinator) handleSessionDisconnected(msg SessionDisconnectedMsg) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if in lobby
	if code, inLobby := c.sessionLobby[msg.SessionID]; inLobby {
		if lobby, exists := c.lobbies[code]; exists {
			// If host disconnected
			if lobby.Host.ID() == msg.SessionID {
				if lobby.Joiner != nil {
					lobby.Joiner.Send(MatchEndedEvent{Reason: MatchEndReasonHostLeft})
					delete(c.sessionLobby, lobby.Joiner.ID())
				}
				delete(c.lobbies, code)
			} else if lobby.Joiner != nil && lobby.Joiner.ID() == msg.SessionID {
				// Joiner disconnected
				lobby.Joiner = nil
				lobby.Host.Send(LobbyPlayerLeftEvent{Code: code})
			}
		}
		delete(c.sessionLobby, msg.SessionID)
	}

	// Check if in match
	if matchID, inMatch := c.sessionMatch[msg.SessionID]; inMatch {
		if match, exists := c.matches[matchID]; exists {
			match.PlayerDisconnected(msg.SessionID)
		}
	}
}

func (c *Coordinator) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpiredLobbies()
		case <-c.done:
			return
		}
	}
}

func (c *Coordinator) cleanupExpiredLobbies() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for code, lobby := range c.lobbies {
		// Only expire lobbies without joiners
		if lobby.Joiner == nil && now.Sub(lobby.CreatedAt) > c.config.LobbyTimeout {
			lobby.Host.Send(LobbyErrorEvent{Message: "Lobby expired"})
			delete(c.sessionLobby, lobby.Host.ID())
			delete(c.lobbies, code)
		}
	}
}

func (c *Coordinator) generateUniqueCode() string {
	for {
		code := generateJoinCode()
		if _, exists := c.lobbies[code]; !exists {
			return code
		}
	}
}

// generateJoinCode creates a 6-character uppercase alphanumeric code.
func generateJoinCode() string {
	b := make([]byte, 4) // 4 bytes = 32 bits, base32 encodes to 8 chars, we take 6
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based
		return fmt.Sprintf("%06X", time.Now().UnixNano()&0xFFFFFF)
	}
	// Use base32 encoding (A-Z, 2-7), take first 6 chars
	code := base32.StdEncoding.EncodeToString(b)[:6]
	return strings.ToUpper(code)
}

// GetLobby returns a lobby by code (for testing/debug).
func (c *Coordinator) GetLobby(code string) (*Lobby, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	l, ok := c.lobbies[strings.ToUpper(code)]
	return l, ok
}

// GetMatch returns a match by ID (for testing/debug).
func (c *Coordinator) GetMatch(id MatchID) (*OnlineMatch, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.matches[id]
	return m, ok
}

// LobbyCount returns the number of active lobbies.
func (c *Coordinator) LobbyCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.lobbies)
}

// MatchCount returns the number of active matches.
func (c *Coordinator) MatchCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.matches)
}
