// Package tui provides terminal UI components including SSH server support via Wish.
package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pong"
	"github.com/vovakirdan/tui-arcade/internal/multiplayer"
	"github.com/vovakirdan/tui-arcade/internal/registry"
	"github.com/vovakirdan/tui-arcade/internal/storage"
)

// SSHServerConfig holds configuration for the SSH server.
type SSHServerConfig struct {
	// Address is the host:port to listen on (e.g., ":23234").
	Address string

	// HostKeyPath is the path to the host key file.
	// If empty, a key will be auto-generated at ~/.arcade/host_key.
	HostKeyPath string

	// DBPath is the path to the scores database.
	DBPath string

	// IdleTimeout is how long to wait before closing idle connections.
	IdleTimeout time.Duration
}

// DefaultSSHServerConfig returns a config with sensible defaults.
func DefaultSSHServerConfig() SSHServerConfig {
	return SSHServerConfig{
		Address:     ":23234",
		DBPath:      "~/.arcade/scores.db",
		IdleTimeout: 30 * time.Minute,
	}
}

// SSHServer wraps a Wish SSH server for the arcade.
type SSHServer struct {
	config      SSHServerConfig
	server      *ssh.Server
	store       *storage.Store
	logger      *log.Logger
	coordinator *multiplayer.Coordinator
	sessions    *multiplayer.SessionRegistry
}

// NewSSHServer creates a new SSH server with the given configuration.
func NewSSHServer(cfg SSHServerConfig) (*SSHServer, error) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		Prefix:          "arcade-ssh",
	})

	// Open storage
	store, err := storage.Open(cfg.DBPath)
	if err != nil {
		logger.Warn("could not open scores database", "error", err)
		// Continue without storage
	}

	// Create session registry
	sessions := multiplayer.NewSessionRegistry()

	// Create coordinator with game factory
	coordCfg := multiplayer.DefaultCoordinatorConfig()
	coordinator := multiplayer.NewCoordinator(coordCfg, createOnlineGame, sessions)

	// Wire up storage for match results
	if store != nil {
		coordinator.SetResultSaver(store)
	}

	srv := &SSHServer{
		config:      cfg,
		store:       store,
		logger:      logger,
		coordinator: coordinator,
		sessions:    sessions,
	}

	// Resolve host key path
	hostKeyPath := cfg.HostKeyPath
	if hostKeyPath == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return nil, fmt.Errorf("cannot get home directory: %w", homeErr)
		}
		hostKeyPath = filepath.Join(home, ".arcade", "host_key")
	}

	// Ensure host key directory exists
	hostKeyDir := filepath.Dir(hostKeyPath)
	if mkdirErr := os.MkdirAll(hostKeyDir, 0o700); mkdirErr != nil {
		return nil, fmt.Errorf("cannot create host key directory: %w", mkdirErr)
	}

	// Create Wish server options
	opts := []ssh.Option{
		wish.WithAddress(cfg.Address),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithIdleTimeout(cfg.IdleTimeout),
		wish.WithMiddleware(
			bubbletea.Middleware(srv.teaHandler),
			srv.loggingMiddleware,
		),
	}

	// Create the server
	server, err := wish.NewServer(opts...)
	if err != nil {
		if store != nil {
			store.Close()
		}
		return nil, fmt.Errorf("cannot create SSH server: %w", err)
	}

	srv.server = server
	return srv, nil
}

// teaHandler creates a Bubble Tea program for each SSH session.
func (s *SSHServer) teaHandler(sshSession ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, ok := sshSession.Pty()
	if !ok {
		s.logger.Warn("no PTY requested", "user", sshSession.User())
		return nil, nil
	}

	// Create runtime config from PTY size
	cfg := core.RuntimeConfig{
		ScreenW:  pty.Window.Width,
		ScreenH:  pty.Window.Height,
		TickRate: 60,
		Seed:     time.Now().UnixNano(),
	}

	// Create session ID and channel session for coordinator communication
	sessionID := multiplayer.SessionID(fmt.Sprintf("%s-%d", sshSession.User(), time.Now().UnixNano()))
	channelSession := multiplayer.NewChannelSession(sessionID, 64)

	// Register session with registry
	s.sessions.Register(channelSession)

	// Create session model that handles menu + game flow
	model := NewSessionModel(s.store, cfg, sshSession.User(), sessionID, channelSession, s.coordinator)

	return model, []tea.ProgramOption{
		tea.WithAltScreen(),
	}
}

// loggingMiddleware logs SSH session events.
func (s *SSHServer) loggingMiddleware(next ssh.Handler) ssh.Handler {
	return func(sshSession ssh.Session) {
		s.logger.Info("session started",
			"user", sshSession.User(),
			"remote", sshSession.RemoteAddr().String(),
		)
		next(sshSession)
		s.logger.Info("session ended",
			"user", sshSession.User(),
			"remote", sshSession.RemoteAddr().String(),
		)
	}
}

// ListenAndServe starts the SSH server and blocks until shutdown.
func (s *SSHServer) ListenAndServe() error {
	s.logger.Info("starting SSH server", "address", s.config.Address)

	// Start coordinator
	s.coordinator.Start()

	// Setup signal handling for graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			s.logger.Error("server error", "error", err)
		}
	}()

	<-done
	s.logger.Info("shutting down...")
	return s.Shutdown()
}

// Shutdown gracefully stops the server.
func (s *SSHServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop coordinator
	s.coordinator.Stop()

	if s.store != nil {
		s.store.Close()
	}

	return s.server.Shutdown(ctx)
}

// Addr returns the server's listen address string.
func (s *SSHServer) Addr() string {
	return s.config.Address
}

// createOnlineGame is the factory function for creating online game instances.
func createOnlineGame(gameID string, cfg core.RuntimeConfig) (multiplayer.OnlineGame, error) {
	switch gameID {
	case "pong":
		game := pong.NewOnline()
		game.Reset(cfg)
		return game, nil
	default:
		return nil, fmt.Errorf("game %q does not support online multiplayer", gameID)
	}
}

// SessionState represents the current state of the session.
type SessionState int

const (
	SessionStateMenu SessionState = iota
	SessionStatePongMode
	SessionStateOnlineLobby
	SessionStateInGame
	SessionStateOnlineGame
)

// SessionModel manages the full arcade session flow: menu -> game -> menu.
// This is the top-level model used for SSH sessions.
type SessionModel struct {
	store          *storage.Store
	config         core.RuntimeConfig
	username       string
	sessionID      multiplayer.SessionID
	channelSession *multiplayer.ChannelSession
	coordinator    *multiplayer.Coordinator

	state     SessionState
	menu      MenuModel
	pongMode  PongModeModel
	lobby     OnlineLobbyModel
	game      registry.Game
	gameModel *GameModel
	quitting  bool
}

// NewSessionModel creates a new session model.
func NewSessionModel(
	store *storage.Store,
	cfg core.RuntimeConfig,
	username string,
	sessionID multiplayer.SessionID,
	channelSession *multiplayer.ChannelSession,
	coordinator *multiplayer.Coordinator,
) SessionModel {
	return SessionModel{
		store:          store,
		config:         cfg,
		username:       username,
		sessionID:      sessionID,
		channelSession: channelSession,
		coordinator:    coordinator,
		state:          SessionStateMenu,
		menu:           NewMenuModel(store, cfg),
	}
}

// Init initializes the session.
func (m SessionModel) Init() tea.Cmd {
	return m.menu.Init()
}

// Update handles messages for the session.
func (m SessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize globally
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.config.ScreenW = wsm.Width
		m.config.ScreenH = wsm.Height
	}

	switch m.state {
	case SessionStateMenu:
		return m.updateMenu(msg)
	case SessionStatePongMode:
		return m.updatePongMode(msg)
	case SessionStateOnlineLobby:
		return m.updateLobby(msg)
	case SessionStateInGame:
		return m.updateGame(msg)
	case SessionStateOnlineGame:
		return m.updateOnlineGame(msg)
	}
	return m, nil
}

// updateMenu handles updates when in menu mode.
func (m SessionModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	newMenu, cmd := m.menu.Update(msg)
	if menuModel, ok := newMenu.(MenuModel); ok {
		m.menu = menuModel
	}

	// Check if user quit
	if m.menu.IsQuitting() {
		m.quitting = true
		m.notifyDisconnect()
		return m, tea.Quit
	}

	// Check if game was selected
	if selected := m.menu.Selected(); selected != nil {
		m.config = m.menu.Config()

		// Special handling for Pong - show mode selection
		if selected.GameID == "pong" {
			m.state = SessionStatePongMode
			m.pongMode = NewPongModeModel(m.config.ScreenW, m.config.ScreenH)
			return m, m.pongMode.Init()
		}

		// For other games, start directly
		return m.startLocalGame(selected.GameID, selected.Mode)
	}

	return m, cmd
}

// updatePongMode handles Pong mode selection.
func (m SessionModel) updatePongMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	newModel, cmd := m.pongMode.Update(msg)
	if pongModel, ok := newModel.(PongModeModel); ok {
		m.pongMode = pongModel
	}

	// Check if user quit
	if m.pongMode.IsQuitting() {
		m.quitting = true
		m.notifyDisconnect()
		return m, tea.Quit
	}

	// Check for back
	if m.pongMode.WantsBack() {
		m.state = SessionStateMenu
		m.menu = NewMenuModel(m.store, m.config)
		return m, m.menu.Init()
	}

	// Check if mode was selected
	if !m.pongMode.IsChoosing() {
		mode := m.pongMode.Selected()
		if mode == multiplayer.MatchModeOnlinePvP {
			// Start online lobby
			m.state = SessionStateOnlineLobby
			m.lobby = NewOnlineLobbyModel(
				"pong",
				m.sessionID,
				m.coordinator,
				m.channelSession.Events(),
				m.config.ScreenW,
				m.config.ScreenH,
			)
			return m, m.lobby.Init()
		}
		// Start vs CPU game
		return m.startLocalGame("pong", multiplayer.MatchModeVsCPU)
	}

	return m, cmd
}

// updateLobby handles online lobby updates.
func (m SessionModel) updateLobby(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	newModel, cmd := m.lobby.Update(msg)
	if lobbyModel, ok := newModel.(OnlineLobbyModel); ok {
		m.lobby = lobbyModel
	}

	// Check if user quit
	if m.lobby.IsQuitting() {
		m.quitting = true
		m.notifyDisconnect()
		return m, tea.Quit
	}

	// Check for back to menu
	if m.lobby.BackToMenu() {
		m.state = SessionStateMenu
		m.menu = NewMenuModel(m.store, m.config)
		return m, m.menu.Init()
	}

	// Check if match started
	if m.lobby.State() == OnlineStateInMatch {
		m.state = SessionStateOnlineGame
		// Online game rendering will be handled by snapshot events
		return m, m.waitForEvents()
	}

	return m, cmd
}

// startLocalGame starts a local (solo/vs CPU) game.
func (m SessionModel) startLocalGame(gameID string, mode multiplayer.MatchMode) (tea.Model, tea.Cmd) {
	game, err := registry.Create(gameID)
	if err != nil {
		return m, nil
	}

	m.game = game

	// Create match for this game session
	match := multiplayer.NewMatch(
		multiplayer.MatchID(fmt.Sprintf("match-%d", time.Now().UnixNano())),
		mode,
		m.sessionID,
	)

	// Create game model
	gameModel := NewGameModel(game, m.store, m.config, match)
	m.gameModel = &gameModel
	m.state = SessionStateInGame

	return m, m.gameModel.Init()
}

// waitForEvents returns a command that waits for coordinator events.
func (m SessionModel) waitForEvents() tea.Cmd {
	return func() tea.Msg {
		if m.channelSession == nil {
			return nil
		}
		evt, ok := <-m.channelSession.Events()
		if !ok {
			return nil
		}
		return evt
	}
}

// notifyDisconnect notifies the coordinator that this session is disconnecting.
func (m SessionModel) notifyDisconnect() {
	if m.coordinator != nil {
		m.coordinator.Send(multiplayer.SessionDisconnectedMsg{
			SessionID: m.sessionID,
		})
	}
	if m.channelSession != nil {
		m.channelSession.Close()
	}
}

// updateGame handles updates when in game mode.
func (m SessionModel) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	newModel, cmd := m.gameModel.Update(msg)
	if gameModel, ok := newModel.(GameModel); ok {
		m.gameModel = &gameModel
	}

	// Check if user quit game (back to menu)
	if m.gameModel.BackToMenu() {
		m.state = SessionStateMenu
		m.gameModel = nil
		m.game = nil
		// Reset menu state
		m.menu = NewMenuModel(m.store, m.config)
		return m, m.menu.Init()
	}

	// Check if user quit entirely
	if m.gameModel.IsQuitting() {
		m.quitting = true
		m.notifyDisconnect()
		return m, tea.Quit
	}

	return m, cmd
}

// updateOnlineGame handles updates when in online multiplayer game.
func (m SessionModel) updateOnlineGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleOnlineGameKey(msg)
	case multiplayer.SnapshotEvent:
		// Game state snapshot received from coordinator
		// TODO: Render game state from snapshot
		return m, m.waitForEvents()
	case multiplayer.MatchEndedEvent:
		// Match ended - return to menu
		m.state = SessionStateMenu
		m.menu = NewMenuModel(m.store, m.config)
		return m, m.menu.Init()
	}
	return m, m.waitForEvents()
}

// handleOnlineGameKey handles keyboard input during online game.
func (m SessionModel) handleOnlineGameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit
	if key == "ctrl+c" {
		m.quitting = true
		m.notifyDisconnect()
		return m, tea.Quit
	}

	// Screenshot
	if key == "ctrl+s" {
		m.saveOnlineScreenshot()
		return m, nil
	}

	// Back to menu (Esc)
	if key == "esc" {
		// Send leave match message
		m.coordinator.Send(multiplayer.LeaveMatchMsg{
			SessionID: m.sessionID,
			MatchID:   m.lobby.MatchID(),
		})
		m.state = SessionStateMenu
		m.menu = NewMenuModel(m.store, m.config)
		return m, m.menu.Init()
	}

	// Game input - map keys to input frame and send to coordinator
	input := core.NewInputFrame()
	hasInput := false
	switch key {
	case "up", "w":
		input.Set(core.ActionUp)
		hasInput = true
	case "down", "s":
		input.Set(core.ActionDown)
		hasInput = true
	}

	if hasInput {
		m.coordinator.Send(multiplayer.PlayerInputMsg{
			MatchID: m.lobby.MatchID(),
			Player:  m.lobby.Side(),
			Input:   input,
		})
	}

	return m, nil
}

// View renders the current view.
func (m SessionModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case SessionStateMenu:
		return m.menu.View()
	case SessionStatePongMode:
		return m.pongMode.View()
	case SessionStateOnlineLobby:
		return m.lobby.View()
	case SessionStateInGame:
		if m.gameModel != nil {
			return m.gameModel.View()
		}
	case SessionStateOnlineGame:
		return m.viewOnlineGame()
	}

	return m.menu.View()
}

// viewOnlineGame renders the online game view based on latest snapshot.
func (m SessionModel) viewOnlineGame() string {
	// For now, show a placeholder. Later we'll render from snapshots.
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("ONLINE PONG", m.config.ScreenW))
	b.WriteString("\n\n")

	sideText := "LEFT (P1)"
	if m.lobby.Side() == core.Player2 {
		sideText = "RIGHT (P2)"
	}
	b.WriteString(centerText(fmt.Sprintf("You are: %s", sideText), m.config.ScreenW))
	b.WriteString("\n\n")
	b.WriteString(centerText("Game in progress...", m.config.ScreenW))
	b.WriteString("\n\n")
	b.WriteString(centerText("W/Up: Move up  |  S/Down: Move down  |  Esc: Leave", m.config.ScreenW))

	return b.String()
}

// saveOnlineScreenshot saves a screenshot of the online game view.
func (m *SessionModel) saveOnlineScreenshot() {
	// Create screenshots directory
	dir := filepath.Join(os.Getenv("HOME"), ".arcade", "screenshots")
	//nolint:errcheck // Best-effort directory creation
	os.MkdirAll(dir, 0o755)

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("online_%s_%s.txt", m.username, timestamp)
	path := filepath.Join(dir, filename)

	// Get current view and save it
	content := m.viewOnlineGame()
	//nolint:errcheck // Best-effort save
	os.WriteFile(path, []byte(content), 0o600)
}

// GameModel wraps a game with multiplayer support and back-to-menu capability.
type GameModel struct {
	game       registry.Game
	screen     *core.Screen
	store      *storage.Store
	config     core.RuntimeConfig
	match      *multiplayer.Match
	inputFrame core.MultiInputFrame
	gameState  core.GameState
	keyMapper  *KeyMapper
	quitting   bool
	backToMenu bool
	scoreSaved bool
}

// NewGameModel creates a new game model with multiplayer support.
func NewGameModel(game registry.Game, store *storage.Store, cfg core.RuntimeConfig, match *multiplayer.Match) GameModel {
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}

	return GameModel{
		game:       game,
		screen:     core.NewScreen(cfg.ScreenW, cfg.ScreenH),
		store:      store,
		config:     cfg,
		match:      match,
		inputFrame: core.NewMultiInputFrame(),
		keyMapper:  NewKeyMapper(),
	}
}

// Init initializes the game.
func (m GameModel) Init() tea.Cmd {
	m.game.Reset(m.config)
	return tickCmd(m.config.TickRate)
}

// Update handles messages.
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.config.ScreenW = msg.Width
		m.config.ScreenH = msg.Height
		m.screen.Resize(msg.Width, msg.Height)
		if !m.gameState.GameOver {
			m.game.Reset(m.config)
		}
		return m, nil
	case TickMsg:
		return m.handleTick()
	}
	return m, nil
}

// handleKey processes keyboard input.
func (m GameModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check for screenshot
	if msg.String() == "ctrl+s" {
		m.saveScreenshot()
		return m, nil
	}

	// Check for quit
	if m.keyMapper.MapKeyToMultiFrame(msg, &m.inputFrame) {
		m.quitting = true
		return m, tea.Quit
	}

	// Check for back to menu (B or Esc when game over or paused)
	action := m.keyMapper.MapKeyToMenuAction(msg)
	if action == MenuActionBack && (m.gameState.GameOver || m.gameState.Paused) {
		m.backToMenu = true
		return m, nil
	}

	return m, nil
}

// handleTick processes simulation ticks.
func (m GameModel) handleTick() (tea.Model, tea.Cmd) {
	// Check for restart
	p1Input := m.inputFrame.Player1Frame()
	if p1Input.Has(core.ActionRestart) && m.gameState.GameOver {
		m.config.Seed = time.Now().UnixNano()
		m.game.Reset(m.config)
		m.gameState = m.game.State()
		m.scoreSaved = false
		m.inputFrame.Clear()
		return m, tickCmd(m.config.TickRate)
	}

	// Note: For VsCPU mode, AI is handled directly in the game (e.g., Pong handles CPU paddle)

	// Run game simulation with Player1 input (single player games use InputFrame)
	result := m.game.Step(p1Input)
	m.gameState = result.State

	// Save score on game over
	if m.gameState.GameOver && !m.scoreSaved && m.gameState.Score > 0 {
		if m.store != nil {
			//nolint:errcheck // Best-effort save
			m.store.SaveScore(m.game.ID(), m.gameState.Score)
		}
		m.scoreSaved = true
	}

	m.inputFrame.Clear()
	return m, tickCmd(m.config.TickRate)
}

// View renders the game.
func (m GameModel) View() string {
	if m.quitting {
		return ""
	}

	m.game.Render(m.screen)
	return RenderScreen(m.screen)
}

// IsQuitting returns true if user requested to quit entirely.
func (m GameModel) IsQuitting() bool {
	return m.quitting
}

// BackToMenu returns true if user requested to go back to menu.
func (m GameModel) BackToMenu() bool {
	return m.backToMenu
}

// saveScreenshot saves the current screen to a file.
func (m *GameModel) saveScreenshot() {
	// Render current state
	m.game.Render(m.screen)

	// Create screenshots directory
	dir := filepath.Join(os.Getenv("HOME"), ".arcade", "screenshots")
	//nolint:errcheck // Best-effort directory creation
	os.MkdirAll(dir, 0o755)

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.txt", m.game.ID(), timestamp)
	path := filepath.Join(dir, filename)

	// Save screenshot
	//nolint:errcheck // Best-effort save
	os.WriteFile(path, []byte(m.screen.String()), 0o600)
}
