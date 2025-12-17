package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/registry"
	"github.com/vovakirdan/tui-arcade/internal/storage"
)

// Model is the Bubble Tea model for running arcade games.
type Model struct {
	game       registry.Game
	screen     *core.Screen
	store      *storage.Store
	config     core.RuntimeConfig
	inputFrame core.InputFrame
	gameState  core.GameState
	quitting   bool
	scoreSaved bool // Whether score has been saved for current game over
}

// NewModel creates a new Bubble Tea model for the given game.
func NewModel(game registry.Game, store *storage.Store, cfg core.RuntimeConfig) Model {
	// Use time-based seed if not specified
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}

	return Model{
		game:       game,
		screen:     core.NewScreen(cfg.ScreenW, cfg.ScreenH),
		store:      store,
		config:     cfg,
		inputFrame: core.NewInputFrame(),
	}
}

// Init initializes the model and starts the game.
func (m Model) Init() tea.Cmd {
	// Initialize the game
	m.game.Reset(m.config)
	m.gameState = m.game.State()

	// Start the tick loop
	return tickCmd(m.config.TickRate)
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		return m.handleResize(msg)

	case TickMsg:
		return m.handleTick()
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit keys
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	}

	// Map key to action
	switch msg.String() {
	case " ", "up", "w":
		m.inputFrame.Set(core.ActionJump)
	case "down", "s":
		m.inputFrame.Set(core.ActionDuck)
	case "p", "esc":
		m.inputFrame.Set(core.ActionPause)
	case "r":
		if m.gameState.GameOver {
			m.inputFrame.Set(core.ActionRestart)
		}
	}

	return m, nil
}

// handleResize processes window resize events.
func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	// Update screen size
	m.config.ScreenW = msg.Width
	m.config.ScreenH = msg.Height
	m.screen.Resize(msg.Width, msg.Height)

	// Reinitialize game with new dimensions if needed
	// Note: This resets the game - could be improved to preserve state
	if !m.gameState.GameOver {
		m.game.Reset(m.config)
	}

	return m, nil
}

// handleTick processes simulation ticks.
func (m Model) handleTick() (tea.Model, tea.Cmd) {
	// Check for restart
	if m.inputFrame.Has(core.ActionRestart) && m.gameState.GameOver {
		// Reset seed for new game
		m.config.Seed = time.Now().UnixNano()
		m.game.Reset(m.config)
		m.gameState = m.game.State()
		m.scoreSaved = false
		m.inputFrame.Clear()
		return m, tickCmd(m.config.TickRate)
	}

	// Run game simulation
	result := m.game.Step(m.inputFrame)
	m.gameState = result.State

	// Save score on game over (once)
	if m.gameState.GameOver && !m.scoreSaved && m.gameState.Score > 0 {
		if m.store != nil {
			m.store.SaveScore(m.game.ID(), m.gameState.Score)
		}
		m.scoreSaved = true
	}

	// Clear input for next frame
	m.inputFrame.Clear()

	// Continue ticking
	return m, tickCmd(m.config.TickRate)
}

// View renders the current state to a string for display.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Render game to screen buffer
	m.game.Render(m.screen)

	// Convert screen to string
	return RenderScreen(m.screen)
}

// Run starts the Bubble Tea program with the given model.
func Run(game registry.Game, store *storage.Store, cfg core.RuntimeConfig) error {
	model := NewModel(game, store, cfg)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse (for future use)
	)

	_, err := p.Run()
	return err
}
