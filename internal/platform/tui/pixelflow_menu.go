package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow"
)

// PixelFlowSelection holds the user's selection from the PixelFlow menu.
type PixelFlowSelection struct {
	Level int // 0 = start from beginning, 1-N = specific level
}

// PixelFlowMenuModel is the level picker for PixelFlow.
type PixelFlowMenuModel struct {
	cursor       int
	width        int
	height       int
	keyMapper    *KeyMapper
	levelNames   []string
	selection    PixelFlowSelection
	choosing     bool
	quitting     bool
	back         bool
	scrollOffset int
	theme        PixelFlowTheme
}

// NewPixelFlowMenuModel creates a new PixelFlow level selection model.
func NewPixelFlowMenuModel(width, height int) PixelFlowMenuModel {
	levelNames := pixelflow.LevelNames()
	if len(levelNames) == 0 {
		levelNames = []string{"No levels found"}
	}

	return PixelFlowMenuModel{
		cursor:     0,
		width:      width,
		height:     height,
		keyMapper:  NewKeyMapper(),
		levelNames: levelNames,
		choosing:   true,
		theme:      GetPixelFlowTheme(),
	}
}

// Init initializes the model.
func (m PixelFlowMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m PixelFlowMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m PixelFlowMenuModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := m.keyMapper.MapKeyToMenuAction(msg)

	switch action {
	case MenuActionQuit:
		m.quitting = true
		return m, tea.Quit
	case MenuActionUp:
		if m.cursor > 0 {
			m.cursor--
			m.updateScroll()
		}
	case MenuActionDown:
		if m.cursor < len(m.levelNames) {
			m.cursor++
			m.updateScroll()
		}
	case MenuActionSelect:
		m.choosing = false
		if m.cursor == 0 {
			m.selection = PixelFlowSelection{Level: 0}
		} else {
			m.selection = PixelFlowSelection{Level: m.cursor}
		}
		return m, tea.Quit
	case MenuActionBack:
		m.back = true
		return m, nil
	}

	return m, nil
}

// updateScroll adjusts scroll offset to keep cursor visible.
func (m *PixelFlowMenuModel) updateScroll() {
	visibleItems := m.height - 10 // Account for header and footer
	if visibleItems < 3 {
		visibleItems = 3
	}

	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+visibleItems {
		m.scrollOffset = m.cursor - visibleItems + 1
	}
}

// View renders the level selection.
func (m PixelFlowMenuModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString("\n")
	title := m.theme.MenuTitle.Render("P I X E L F L O W")
	b.WriteString(centerText(title, m.width))
	b.WriteString("\n\n")

	// Subtitle
	subtitle := m.theme.MenuDescription.Render("Select a level:")
	b.WriteString(centerText(subtitle, m.width))
	b.WriteString("\n\n")

	// Calculate visible range
	visibleItems := m.height - 10
	if visibleItems < 3 {
		visibleItems = 3
	}

	// "Start from Beginning" option
	if m.scrollOffset == 0 {
		cursor := "  "
		style := m.theme.MenuItemNormal
		if m.cursor == 0 {
			cursor = "> "
			style = m.theme.MenuItemActive
		}
		line := style.Render(cursor + "Start from Beginning")
		b.WriteString(centerText(line, m.width))
		b.WriteString("\n")
	}

	// Level list
	startIdx := m.scrollOffset
	if startIdx == 0 {
		startIdx = 0
	}
	endIdx := startIdx + visibleItems
	if endIdx > len(m.levelNames) {
		endIdx = len(m.levelNames)
	}

	for i := startIdx; i < endIdx; i++ {
		actualIdx := i + 1 // Account for "Start from Beginning" option
		cursor := "  "
		style := m.theme.MenuItemNormal
		if actualIdx == m.cursor {
			cursor = "> "
			style = m.theme.MenuItemActive
		}

		levelNum := fmt.Sprintf("%2d. ", i+1)
		line := style.Render(cursor + levelNum + m.levelNames[i])
		b.WriteString(centerText(line, m.width))
		b.WriteString("\n")
	}

	// Scroll indicators
	if m.scrollOffset > 0 {
		b.WriteString(centerText(m.theme.MenuDescription.Render("... more above ..."), m.width))
		b.WriteString("\n")
	}
	if endIdx < len(m.levelNames) {
		b.WriteString(centerText(m.theme.MenuDescription.Render("... more below ..."), m.width))
		b.WriteString("\n")
	}

	// Footer with controls
	b.WriteString("\n")
	controls := m.theme.HUDControls.Render("Up/Down: Navigate  |  Enter: Select  |  Esc: Back  |  Q: Quit")
	b.WriteString(centerText(controls, m.width))
	b.WriteString("\n")

	return b.String()
}

// Selected returns the selection, or nil if still choosing.
func (m PixelFlowMenuModel) Selected() *PixelFlowSelection {
	if m.choosing {
		return nil
	}
	return &m.selection
}

// IsChoosing returns true if still in selection mode.
func (m PixelFlowMenuModel) IsChoosing() bool {
	return m.choosing
}

// IsQuitting returns true if user wants to quit.
func (m PixelFlowMenuModel) IsQuitting() bool {
	return m.quitting
}

// WantsBack returns true if user pressed back.
func (m PixelFlowMenuModel) WantsBack() bool {
	return m.back
}

// RunPixelFlowLevelSelector runs the level selection and returns the selection.
func RunPixelFlowLevelSelector(cfg core.RuntimeConfig) (*PixelFlowSelection, core.RuntimeConfig, error) {
	model := NewPixelFlowMenuModel(cfg.ScreenW, cfg.ScreenH)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, cfg, err
	}

	m, ok := finalModel.(PixelFlowMenuModel)
	if !ok {
		return nil, cfg, nil
	}

	if m.IsQuitting() || m.WantsBack() {
		return nil, cfg, nil
	}

	return m.Selected(), cfg, nil
}
