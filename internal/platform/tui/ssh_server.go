// Package tui provides terminal UI components including SSH server support via Wish.
package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
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
	config SSHServerConfig
	server *ssh.Server
	store  *storage.Store
	logger *log.Logger
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

	srv := &SSHServer{
		config: cfg,
		store:  store,
		logger: logger,
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

	// Create session model that handles menu + game flow
	model := NewSessionModel(s.store, cfg, sshSession.User())

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

	if s.store != nil {
		s.store.Close()
	}

	return s.server.Shutdown(ctx)
}

// Addr returns the server's listen address string.
func (s *SSHServer) Addr() string {
	return s.config.Address
}

// SessionModel manages the full arcade session flow: menu -> game -> menu.
// This is the top-level model used for SSH sessions.
type SessionModel struct {
	store     *storage.Store
	config    core.RuntimeConfig
	username  string
	sessionID multiplayer.SessionID
	menu      MenuModel
	game      registry.Game
	gameModel *GameModel
	inGame    bool
	quitting  bool
}

// NewSessionModel creates a new session model.
func NewSessionModel(store *storage.Store, cfg core.RuntimeConfig, username string) SessionModel {
	sessionID := multiplayer.SessionID(fmt.Sprintf("%s-%d", username, time.Now().UnixNano()))

	return SessionModel{
		store:     store,
		config:    cfg,
		username:  username,
		sessionID: sessionID,
		menu:      NewMenuModel(store, cfg),
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

	if m.inGame && m.gameModel != nil {
		return m.updateGame(msg)
	}
	return m.updateMenu(msg)
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
		return m, tea.Quit
	}

	// Check if game was selected
	if selected := m.menu.Selected(); selected != nil {
		// Create the game
		game, err := registry.Create(selected.GameID)
		if err != nil {
			// Shouldn't happen since menu only shows registered games
			return m, nil
		}

		m.game = game
		m.config = m.menu.Config() // Get possibly updated config from resize

		// Create match for this game session
		match := multiplayer.NewMatch(
			multiplayer.MatchID(fmt.Sprintf("match-%d", time.Now().UnixNano())),
			selected.Mode,
			m.sessionID,
		)

		// Create game model
		gameModel := NewGameModel(game, m.store, m.config, match)
		m.gameModel = &gameModel
		m.inGame = true

		return m, m.gameModel.Init()
	}

	return m, cmd
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
		m.inGame = false
		m.gameModel = nil
		m.game = nil
		// Reset menu state
		m.menu = NewMenuModel(m.store, m.config)
		return m, m.menu.Init()
	}

	// Check if user quit entirely
	if m.gameModel.IsQuitting() {
		m.quitting = true
		return m, tea.Quit
	}

	return m, cmd
}

// View renders the current view.
func (m SessionModel) View() string {
	if m.quitting {
		return ""
	}

	if m.inGame && m.gameModel != nil {
		return m.gameModel.View()
	}

	return m.menu.View()
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
	p1Input := m.inputFrame.Player1()
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
