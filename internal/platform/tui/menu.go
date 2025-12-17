package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/multiplayer"
	"github.com/vovakirdan/tui-arcade/internal/registry"
	"github.com/vovakirdan/tui-arcade/internal/storage"
)

// MenuItem represents a selectable game in the menu.
type MenuItem struct {
	GameID string
	Title  string
	Mode   multiplayer.MatchMode
}

// MenuModel is the Bubble Tea model for the game picker menu.
type MenuModel struct {
	items          []MenuItem
	cursor         int
	width          int
	height         int
	store          *storage.Store
	config         core.RuntimeConfig
	keyMapper      *KeyMapper
	quitting       bool
	selected       *MenuItem // Set when user selects a game
	openScoreboard bool      // True if user pressed Tab for scoreboard
}

// NewMenuModel creates a new menu model.
func NewMenuModel(store *storage.Store, cfg core.RuntimeConfig) MenuModel {
	games := registry.List()
	items := make([]MenuItem, 0, len(games))

	for _, g := range games {
		// Skip endless modes - they're accessible via sub-menus
		if g.ID == "breakout_endless" || g.ID == "snake_endless" || g.ID == "2048_endless" {
			continue
		}

		mode := multiplayer.MatchModeSolo
		// Pong is vs CPU mode
		if g.ID == "pong" {
			mode = multiplayer.MatchModeVsCPU
		}
		items = append(items, MenuItem{
			GameID: g.ID,
			Title:  g.Title,
			Mode:   mode,
		})
	}

	return MenuModel{
		items:     items,
		cursor:    0,
		width:     cfg.ScreenW,
		height:    cfg.ScreenH,
		store:     store,
		config:    cfg,
		keyMapper: NewKeyMapper(),
	}
}

// Init initializes the menu model.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the menu.
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.config.ScreenW = msg.Width
		m.config.ScreenH = msg.Height
		return m, nil
	}

	return m, nil
}

// handleKey processes keyboard input for menu navigation.
func (m MenuModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := m.keyMapper.MapKeyToMenuAction(msg)

	switch action {
	case MenuActionQuit:
		m.quitting = true
		return m, tea.Quit

	case MenuActionUp:
		if m.cursor > 0 {
			m.cursor--
		}

	case MenuActionDown:
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}

	case MenuActionSelect:
		if len(m.items) > 0 {
			selected := m.items[m.cursor]
			m.selected = &selected
			return m, tea.Quit // Exit menu to start game
		}

	case MenuActionScoreboard:
		m.openScoreboard = true
		return m, tea.Quit // Exit menu to show scoreboard
	}

	return m, nil
}

// View renders the menu.
func (m MenuModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title
	title := "  A R C A D E  "
	titleLine := centerText(title, m.width)
	b.WriteString("\n")
	b.WriteString(titleLine)
	b.WriteString("\n\n")

	// Subtitle
	subtitle := "Select a game"
	subtitleLine := centerText(subtitle, m.width)
	b.WriteString(subtitleLine)
	b.WriteString("\n\n")

	// Game list
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		modeStr := ""
		if item.Mode == multiplayer.MatchModeVsCPU {
			modeStr = " (CPU)"
		}

		line := fmt.Sprintf("%s%s%s", cursor, item.Title, modeStr)
		b.WriteString(centerText(line, m.width))
		b.WriteString("\n")
	}

	// Footer with controls
	b.WriteString("\n")
	controls := "Up/Down: Navigate  |  Enter: Select  |  Tab: Scores  |  Q: Quit"
	b.WriteString(centerText(controls, m.width))
	b.WriteString("\n")

	return b.String()
}

// Selected returns the selected menu item, or nil if none selected.
func (m MenuModel) Selected() *MenuItem {
	return m.selected
}

// IsQuitting returns true if user requested to quit.
func (m MenuModel) IsQuitting() bool {
	return m.quitting
}

// WantsScoreboard returns true if user requested scoreboard.
func (m MenuModel) WantsScoreboard() bool {
	return m.openScoreboard
}

// Config returns the current runtime config (may have been updated by resize).
func (m MenuModel) Config() core.RuntimeConfig {
	return m.config
}

// centerText centers text within given width.
func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}

// MenuResult holds the result of running the menu.
type MenuResult struct {
	GameID          string
	Mode            multiplayer.MatchMode
	Config          core.RuntimeConfig
	WantsScoreboard bool
	Quit            bool
}

// RunMenu runs the menu and returns the selection result.
func RunMenu(store *storage.Store, cfg core.RuntimeConfig) (MenuResult, error) {
	model := NewMenuModel(store, cfg)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return MenuResult{Config: cfg}, err
	}

	m, ok := finalModel.(MenuModel)
	if !ok {
		return MenuResult{Config: cfg, Quit: true}, nil
	}

	result := MenuResult{
		Config: m.Config(),
	}

	if m.WantsScoreboard() {
		result.WantsScoreboard = true
		return result, nil
	}

	if m.IsQuitting() {
		result.Quit = true
		return result, nil
	}

	if m.Selected() != nil {
		result.GameID = m.Selected().GameID
		result.Mode = m.Selected().Mode
	} else {
		result.Quit = true
	}

	return result, nil
}
