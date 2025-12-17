package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/games/snake"
)

// SnakeMode represents the selected game mode.
type SnakeMode int

const (
	SnakeModeCampaign SnakeMode = iota
	SnakeModeEndless
)

// SnakeSelection holds the user's selection from the Snake menu.
type SnakeSelection struct {
	Mode  SnakeMode
	Level int // 0 = start from beginning, 1-10 = specific level
}

// SnakeModeModel lets users choose game mode and starting level for Snake.
type SnakeModeModel struct {
	cursor        int
	levelCursor   int
	inLevelSelect bool
	width         int
	height        int
	keyMapper     *KeyMapper
	selection     SnakeSelection
	choosing      bool
	quitting      bool
	back          bool
}

// NewSnakeModeModel creates a new Snake mode selection model.
func NewSnakeModeModel(width, height int) SnakeModeModel {
	return SnakeModeModel{
		cursor:    0,
		width:     width,
		height:    height,
		keyMapper: NewKeyMapper(),
		choosing:  true,
	}
}

// Init initializes the model.
func (m SnakeModeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m SnakeModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m SnakeModeModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := m.keyMapper.MapKeyToMenuAction(msg)

	if m.inLevelSelect {
		return m.handleLevelSelectKey(action)
	}
	return m.handleModeSelectKey(action)
}

func (m SnakeModeModel) handleModeSelectKey(action MenuAction) (tea.Model, tea.Cmd) {
	switch action {
	case MenuActionQuit:
		m.quitting = true
		return m, tea.Quit
	case MenuActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case MenuActionDown:
		if m.cursor < 2 { // 3 options: Campaign, Endless, Select Level
			m.cursor++
		}
	case MenuActionSelect:
		switch m.cursor {
		case 0: // Campaign
			m.choosing = false
			m.selection = SnakeSelection{Mode: SnakeModeCampaign, Level: 0}
			return m, tea.Quit
		case 1: // Endless
			m.choosing = false
			m.selection = SnakeSelection{Mode: SnakeModeEndless, Level: 0}
			return m, tea.Quit
		case 2: // Select Level
			m.inLevelSelect = true
			m.levelCursor = 0
		}
	case MenuActionBack:
		m.back = true
		return m, nil
	}

	return m, nil
}

func (m SnakeModeModel) handleLevelSelectKey(action MenuAction) (tea.Model, tea.Cmd) {
	levelCount := snake.LevelCount()

	switch action {
	case MenuActionQuit:
		m.quitting = true
		return m, tea.Quit
	case MenuActionUp:
		if m.levelCursor > 0 {
			m.levelCursor--
		}
	case MenuActionDown:
		if m.levelCursor < levelCount-1 {
			m.levelCursor++
		}
	case MenuActionSelect:
		m.choosing = false
		m.selection = SnakeSelection{
			Mode:  SnakeModeCampaign,
			Level: m.levelCursor + 1, // 1-indexed
		}
		return m, tea.Quit
	case MenuActionBack:
		m.inLevelSelect = false
	}

	return m, nil
}

// View renders the mode/level selection.
func (m SnakeModeModel) View() string {
	if m.quitting {
		return ""
	}

	if m.inLevelSelect {
		return m.viewLevelSelect()
	}
	return m.viewModeSelect()
}

func (m SnakeModeModel) viewModeSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("S N A K E", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Select game mode:", m.width))
	b.WriteString("\n\n")

	modes := []string{
		"Campaign (10 levels)",
		"Endless Mode",
		"Select Level...",
	}

	for i, mode := range modes {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		b.WriteString(centerText(fmt.Sprintf("%s%s", cursor, mode), m.width))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(centerText("Enter: Select  |  Esc: Back  |  Q: Quit", m.width))

	return b.String()
}

func (m SnakeModeModel) viewLevelSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("SELECT LEVEL", m.width))
	b.WriteString("\n\n")

	levelNames := snake.LevelNames()
	for i, name := range levelNames {
		cursor := "  "
		if i == m.levelCursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s%2d. %s", cursor, i+1, name)
		b.WriteString(centerText(line, m.width))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(centerText("Enter: Select  |  Esc: Back  |  Q: Quit", m.width))

	return b.String()
}

// Selected returns the selection, or nil if still choosing.
func (m SnakeModeModel) Selected() *SnakeSelection {
	if m.choosing {
		return nil
	}
	return &m.selection
}

// IsChoosing returns true if still in selection mode.
func (m SnakeModeModel) IsChoosing() bool {
	return m.choosing
}

// IsQuitting returns true if user wants to quit.
func (m SnakeModeModel) IsQuitting() bool {
	return m.quitting
}

// WantsBack returns true if user pressed back.
func (m SnakeModeModel) WantsBack() bool {
	return m.back
}

// RunSnakeModeSelector runs the Snake mode selection and returns the selection.
func RunSnakeModeSelector(cfg core.RuntimeConfig) (*SnakeSelection, core.RuntimeConfig, error) {
	model := NewSnakeModeModel(cfg.ScreenW, cfg.ScreenH)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, cfg, err
	}

	m, ok := finalModel.(SnakeModeModel)
	if !ok {
		return nil, cfg, nil
	}

	if m.IsQuitting() || m.WantsBack() {
		return nil, cfg, nil
	}

	return m.Selected(), cfg, nil
}
