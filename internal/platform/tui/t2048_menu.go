package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/games/t2048"
)

// T2048Mode represents the selected game mode.
type T2048Mode int

const (
	T2048ModeCampaign T2048Mode = iota
	T2048ModeEndless
)

// T2048Selection holds the user's selection from the 2048 menu.
type T2048Selection struct {
	Mode  T2048Mode
	Level int // 0 = start from beginning, 1-10 = specific level
}

// T2048ModeModel lets users choose game mode and starting level for 2048.
type T2048ModeModel struct {
	cursor        int
	levelCursor   int
	inLevelSelect bool
	width         int
	height        int
	keyMapper     *KeyMapper
	selection     T2048Selection
	choosing      bool
	quitting      bool
	back          bool
}

// NewT2048ModeModel creates a new 2048 mode selection model.
func NewT2048ModeModel(width, height int) T2048ModeModel {
	return T2048ModeModel{
		cursor:    0,
		width:     width,
		height:    height,
		keyMapper: NewKeyMapper(),
		choosing:  true,
	}
}

// Init initializes the model.
func (m T2048ModeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m T2048ModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m T2048ModeModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := m.keyMapper.MapKeyToMenuAction(msg)

	if m.inLevelSelect {
		return m.handleLevelSelectKey(action)
	}
	return m.handleModeSelectKey(action)
}

func (m T2048ModeModel) handleModeSelectKey(action MenuAction) (tea.Model, tea.Cmd) {
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
			m.selection = T2048Selection{Mode: T2048ModeCampaign, Level: 0}
			return m, tea.Quit
		case 1: // Endless
			m.choosing = false
			m.selection = T2048Selection{Mode: T2048ModeEndless, Level: 0}
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

func (m T2048ModeModel) handleLevelSelectKey(action MenuAction) (tea.Model, tea.Cmd) {
	levelCount := t2048.LevelCount()

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
		m.selection = T2048Selection{
			Mode:  T2048ModeCampaign,
			Level: m.levelCursor + 1, // 1-indexed
		}
		return m, tea.Quit
	case MenuActionBack:
		m.inLevelSelect = false
	}

	return m, nil
}

// View renders the mode/level selection.
func (m T2048ModeModel) View() string {
	if m.quitting {
		return ""
	}

	if m.inLevelSelect {
		return m.viewLevelSelect()
	}
	return m.viewModeSelect()
}

func (m T2048ModeModel) viewModeSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("2 0 4 8", m.width))
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

func (m T2048ModeModel) viewLevelSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("SELECT LEVEL", m.width))
	b.WriteString("\n\n")

	levelNames := t2048.LevelNames()
	levelTargets := t2048.LevelTargets()

	for i, name := range levelNames {
		cursor := "  "
		if i == m.levelCursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s%2d. %s (Target: %d)", cursor, i+1, name, levelTargets[i])
		b.WriteString(centerText(line, m.width))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(centerText("Enter: Select  |  Esc: Back  |  Q: Quit", m.width))

	return b.String()
}

// Selected returns the selection, or nil if still choosing.
func (m T2048ModeModel) Selected() *T2048Selection {
	if m.choosing {
		return nil
	}
	return &m.selection
}

// IsChoosing returns true if still in selection mode.
func (m T2048ModeModel) IsChoosing() bool {
	return m.choosing
}

// IsQuitting returns true if user wants to quit.
func (m T2048ModeModel) IsQuitting() bool {
	return m.quitting
}

// WantsBack returns true if user pressed back.
func (m T2048ModeModel) WantsBack() bool {
	return m.back
}

// RunT2048ModeSelector runs the 2048 mode selection and returns the selection.
func RunT2048ModeSelector(cfg core.RuntimeConfig) (*T2048Selection, core.RuntimeConfig, error) {
	model := NewT2048ModeModel(cfg.ScreenW, cfg.ScreenH)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, cfg, err
	}

	m, ok := finalModel.(T2048ModeModel)
	if !ok {
		return nil, cfg, nil
	}

	if m.IsQuitting() || m.WantsBack() {
		return nil, cfg, nil
	}

	return m.Selected(), cfg, nil
}
