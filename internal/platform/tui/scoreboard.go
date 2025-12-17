package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/vovakirdan/tui-arcade/internal/registry"
	"github.com/vovakirdan/tui-arcade/internal/storage"
)

// Scoreboard layout constants
const (
	minWidthForSidebar = 80  // Minimum width to show game list sidebar
	sidebarWidth       = 20  // Width of game list sidebar
	tableMinWidth      = 50  // Minimum table width
	maxScores          = 100 // Max scores to load
)

// ScoreboardKeyMap defines the key bindings for the scoreboard.
type ScoreboardKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Select   key.Binding
	Back     key.Binding
	Quit     key.Binding
	NextGame key.Binding
	PrevGame key.Binding
}

// ShortHelp returns key bindings for the short help view.
func (k ScoreboardKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.NextGame, k.PrevGame, k.Back}
}

// FullHelp returns key bindings for the full help view.
func (k ScoreboardKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.NextGame, k.PrevGame},
		{k.Back, k.Quit},
	}
}

// DefaultScoreboardKeyMap returns default key bindings.
func DefaultScoreboardKeyMap() ScoreboardKeyMap {
	return ScoreboardKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "scroll down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "prev game"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "next game"),
		),
		NextGame: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next game"),
		),
		PrevGame: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("S-tab", "prev game"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "b"),
			key.WithHelp("esc/b", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ScoreboardModel is the Bubble Tea model for the scoreboard screen.
type ScoreboardModel struct {
	games       []registry.GameInfo // List of available games
	gameCursor  int                 // Currently selected game index
	store       *storage.Store      // Score storage
	scores      []storage.ScoreEntry
	table       table.Model
	help        help.Model
	keys        ScoreboardKeyMap
	width       int
	height      int
	quitting    bool
	goingBack   bool // True if user pressed back (not quit)
	showSidebar bool // Whether to show game list sidebar
}

// NewScoreboardModel creates a new scoreboard model.
func NewScoreboardModel(store *storage.Store, width, height int) ScoreboardModel {
	games := registry.List()

	// Filter out endless modes for cleaner display
	filteredGames := make([]registry.GameInfo, 0, len(games))
	for _, g := range games {
		if strings.HasSuffix(g.ID, "_endless") {
			continue
		}
		filteredGames = append(filteredGames, g)
	}

	keys := DefaultScoreboardKeyMap()
	h := help.New()
	h.ShowAll = false

	m := ScoreboardModel{
		games:       filteredGames,
		gameCursor:  0,
		store:       store,
		keys:        keys,
		help:        h,
		width:       width,
		height:      height,
		showSidebar: width >= minWidthForSidebar,
	}

	// Initialize table
	m.table = m.createTable()

	// Load scores for first game
	if len(m.games) > 0 {
		m.loadScores(m.games[0].ID)
	}

	return m
}

// createTable creates a new table with appropriate columns.
func (m *ScoreboardModel) createTable() table.Model {
	columns := []table.Column{
		{Title: "Rank", Width: 6},
		{Title: "Score", Width: 10},
		{Title: "Date", Width: 18},
	}

	// Calculate available width for table
	tableWidth := m.width - 4 // Margins
	if m.showSidebar {
		tableWidth -= sidebarWidth + 3 // Sidebar + border + gap
	}

	// Adjust column widths if we have more space
	if tableWidth > 40 {
		columns[0].Width = 6
		columns[1].Width = 12
		columns[2].Width = tableWidth - 22
		if columns[2].Width > 20 {
			columns[2].Width = 20
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(m.height-8), // Leave room for header, help, and margins
	)

	// Table styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t
}

// loadScores loads scores for the given game ID.
func (m *ScoreboardModel) loadScores(gameID string) {
	if m.store == nil {
		m.scores = nil
		m.updateTableRows()
		return
	}

	scores, err := m.store.TopScores(gameID, maxScores)
	if err != nil {
		m.scores = nil
	} else {
		m.scores = scores
	}
	m.updateTableRows()
}

// updateTableRows updates the table with current scores.
func (m *ScoreboardModel) updateTableRows() {
	rows := make([]table.Row, len(m.scores))
	for i, s := range m.scores {
		rows[i] = table.Row{
			fmt.Sprintf("#%d", i+1),
			fmt.Sprintf("%d", s.Score),
			s.CreatedAt.Format("Jan 02 15:04"),
		}
	}
	m.table.SetRows(rows)

	// Reset cursor to top
	m.table.GotoTop()
}

// Init initializes the scoreboard model.
func (m ScoreboardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the scoreboard.
func (m ScoreboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Back):
			m.goingBack = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.NextGame), key.Matches(msg, m.keys.Right):
			if len(m.games) > 0 {
				m.gameCursor = (m.gameCursor + 1) % len(m.games)
				m.loadScores(m.games[m.gameCursor].ID)
			}
			return m, nil

		case key.Matches(msg, m.keys.PrevGame), key.Matches(msg, m.keys.Left):
			if len(m.games) > 0 {
				m.gameCursor--
				if m.gameCursor < 0 {
					m.gameCursor = len(m.games) - 1
				}
				m.loadScores(m.games[m.gameCursor].ID)
			}
			return m, nil

		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
			// Pass to table for scrolling
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.showSidebar = m.width >= minWidthForSidebar
		m.table = m.createTable()
		m.updateTableRows()
		m.help.Width = msg.Width
		return m, nil
	}

	// Pass other messages to table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the scoreboard.
func (m ScoreboardModel) View() string {
	if m.quitting || m.goingBack {
		return ""
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("229")).
		MarginBottom(1)

	title := "HIGH SCORES"
	if len(m.games) > 0 {
		title = fmt.Sprintf("HIGH SCORES - %s", m.games[m.gameCursor].Title)
	}

	b.WriteString(titleStyle.Render(centerText(title, m.width)))
	b.WriteString("\n\n")

	if m.showSidebar {
		// Wide layout: sidebar + table
		b.WriteString(m.renderWideLayout())
	} else {
		// Narrow layout: game tabs + table
		b.WriteString(m.renderNarrowLayout())
	}

	// Help bar
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	b.WriteString(helpStyle.Render(m.help.View(m.keys)))

	return b.String()
}

// renderWideLayout renders the scoreboard with sidebar for game selection.
func (m ScoreboardModel) renderWideLayout() string {
	// Sidebar (game list)
	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(sidebarWidth).
		Padding(0, 1)

	var sidebar strings.Builder
	sidebar.WriteString("Games\n")
	sidebar.WriteString(strings.Repeat("-", sidebarWidth-4))
	sidebar.WriteString("\n")

	for i, g := range m.games {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.gameCursor {
			cursor = "> "
			style = style.Bold(true).Foreground(lipgloss.Color("229"))
		}

		name := g.Title
		maxLen := sidebarWidth - 6
		if len(name) > maxLen {
			name = name[:maxLen-1] + "."
		}
		sidebar.WriteString(style.Render(cursor + name))
		sidebar.WriteString("\n")
	}

	sidebarRendered := sidebarStyle.Render(sidebar.String())

	// Table
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	tableContent := m.renderTableContent()
	tableRendered := tableStyle.Render(tableContent)

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarRendered, "  ", tableRendered)
}

// renderNarrowLayout renders the scoreboard with game tabs above the table.
func (m ScoreboardModel) renderNarrowLayout() string {
	var b strings.Builder

	// Game tabs (horizontal)
	tabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Padding(0, 1)

	tabs := make([]string, len(m.games))
	for i, g := range m.games {
		shortName := g.Title
		if len(shortName) > 10 {
			shortName = shortName[:9] + "."
		}
		if i == m.gameCursor {
			tabs[i] = activeTabStyle.Render(shortName)
		} else {
			tabs[i] = tabStyle.Render(" " + shortName + " ")
		}
	}

	// Wrap tabs if needed
	tabLine := strings.Join(tabs, " ")
	if len(tabLine) > m.width-4 {
		// Just show current game with arrows
		current := m.games[m.gameCursor].Title
		tabLine = fmt.Sprintf("< %s >", current)
	}
	b.WriteString(centerText(tabLine, m.width))
	b.WriteString("\n\n")

	// Table
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	b.WriteString(centerText(tableStyle.Render(m.renderTableContent()), m.width))

	return b.String()
}

// renderTableContent renders the table or empty message.
func (m ScoreboardModel) renderTableContent() string {
	if len(m.scores) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Padding(2, 4)
		return emptyStyle.Render("No scores recorded yet.\nPlay a game to set a high score!")
	}

	return m.table.View()
}

// IsGoingBack returns true if user wants to go back to menu.
func (m ScoreboardModel) IsGoingBack() bool {
	return m.goingBack
}

// IsQuitting returns true if user wants to quit entirely.
func (m ScoreboardModel) IsQuitting() bool {
	return m.quitting
}

// RunScoreboard runs the scoreboard screen.
// Returns true if user wants to go back to menu, false if quitting.
func RunScoreboard(store *storage.Store, width, height int) (goBack bool, err error) {
	model := NewScoreboardModel(store, width, height)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	m, ok := finalModel.(ScoreboardModel)
	if !ok {
		return false, nil
	}

	return m.IsGoingBack(), nil
}
