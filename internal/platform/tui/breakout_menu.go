package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/games/breakout"
)

// BreakoutMode represents the selected game mode.
type BreakoutMode int

const (
	BreakoutModeCampaign BreakoutMode = iota
	BreakoutModeEndless
)

// BreakoutSelection holds the user's selection from the Breakout menu.
type BreakoutSelection struct {
	Mode  BreakoutMode
	Level int // 0 = start from beginning, 1-10 = specific level
}

// BreakoutModeModel lets users choose game mode and starting level for Breakout.
type BreakoutModeModel struct {
	cursor        int
	levelCursor   int
	inLevelSelect bool
	width         int
	height        int
	keyMapper     *KeyMapper
	selection     BreakoutSelection
	choosing      bool
	quitting      bool
	back          bool
}

// Level names for display
var levelNames = []string{
	"Classic",
	"Pyramid",
	"Checkerboard",
	"Diamond",
	"Fortress",
	"Striped",
	"Invaders",
	"Heart",
	"Castle",
	"Final Boss",
}

// NewBreakoutModeModel creates a new Breakout mode selection model.
func NewBreakoutModeModel(width, height int) BreakoutModeModel {
	return BreakoutModeModel{
		cursor:    0,
		width:     width,
		height:    height,
		keyMapper: NewKeyMapper(),
		choosing:  true,
	}
}

// Init initializes the model.
func (m BreakoutModeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m BreakoutModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m BreakoutModeModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := m.keyMapper.MapKeyToMenuAction(msg)

	if m.inLevelSelect {
		return m.handleLevelSelectKey(action)
	}
	return m.handleModeSelectKey(action)
}

func (m BreakoutModeModel) handleModeSelectKey(action MenuAction) (tea.Model, tea.Cmd) {
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
			m.selection = BreakoutSelection{Mode: BreakoutModeCampaign, Level: 0}
			return m, nil
		case 1: // Endless
			m.choosing = false
			m.selection = BreakoutSelection{Mode: BreakoutModeEndless, Level: 0}
			return m, nil
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

func (m BreakoutModeModel) handleLevelSelectKey(action MenuAction) (tea.Model, tea.Cmd) {
	levelCount := breakout.LevelCount()

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
		m.selection = BreakoutSelection{
			Mode:  BreakoutModeCampaign,
			Level: m.levelCursor + 1, // 1-indexed
		}
		return m, nil
	case MenuActionBack:
		m.inLevelSelect = false
	}

	return m, nil
}

// View renders the mode/level selection.
func (m BreakoutModeModel) View() string {
	if m.quitting {
		return ""
	}

	if m.inLevelSelect {
		return m.viewLevelSelect()
	}
	return m.viewModeSelect()
}

func (m BreakoutModeModel) viewModeSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("B R E A K O U T", m.width))
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

func (m BreakoutModeModel) viewLevelSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("SELECT LEVEL", m.width))
	b.WriteString("\n\n")

	levelCount := breakout.LevelCount()
	for i := range levelCount {
		cursor := "  "
		if i == m.levelCursor {
			cursor = "> "
		}

		name := "Level"
		if i < len(levelNames) {
			name = levelNames[i]
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
func (m BreakoutModeModel) Selected() *BreakoutSelection {
	if m.choosing {
		return nil
	}
	return &m.selection
}

// IsChoosing returns true if still in selection mode.
func (m BreakoutModeModel) IsChoosing() bool {
	return m.choosing
}

// IsQuitting returns true if user wants to quit.
func (m BreakoutModeModel) IsQuitting() bool {
	return m.quitting
}

// WantsBack returns true if user pressed back.
func (m BreakoutModeModel) WantsBack() bool {
	return m.back
}

// RunBreakoutModeSelector runs the Breakout mode selection and returns the selection.
func RunBreakoutModeSelector(cfg core.RuntimeConfig) (*BreakoutSelection, core.RuntimeConfig, error) {
	model := NewBreakoutModeModel(cfg.ScreenW, cfg.ScreenH)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, cfg, err
	}

	m, ok := finalModel.(BreakoutModeModel)
	if !ok {
		return nil, cfg, nil
	}

	if m.IsQuitting() || m.WantsBack() {
		return nil, cfg, nil
	}

	return m.Selected(), cfg, nil
}
