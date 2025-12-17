package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/multiplayer"
)

// OnlineState represents the current state of the online matchmaking flow.
type OnlineState int

const (
	OnlineStateChooseMode    OnlineState = iota // Choose Host or Join
	OnlineStateHostWaiting                      // Hosting, waiting for joiner
	OnlineStateJoinEnterCode                    // Entering join code
	OnlineStateJoinWaiting                      // Waiting to connect to host
	OnlineStateMatchStarting                    // Match is starting
	OnlineStateInMatch                          // In active match
	OnlineStateMatchEnded                       // Match has ended
)

// OnlineLobbyModel handles the online matchmaking flow.
type OnlineLobbyModel struct {
	state       OnlineState
	width       int
	height      int
	keyMapper   *KeyMapper
	gameID      string
	sessionID   multiplayer.SessionID
	coordinator *multiplayer.Coordinator

	// Host state
	lobbyCode string

	// Join state
	joinCodeInput string
	joinError     string

	// Match state
	matchID    multiplayer.MatchID
	side       core.PlayerID
	opponentID multiplayer.SessionID

	// Result state
	backToMenu bool
	cancelled  bool
	quitting   bool

	// For receiving events from coordinator
	eventChan <-chan multiplayer.SessionEvent
}

// NewOnlineLobbyModel creates a new online lobby model.
func NewOnlineLobbyModel(
	gameID string,
	sessionID multiplayer.SessionID,
	coordinator *multiplayer.Coordinator,
	eventChan <-chan multiplayer.SessionEvent,
	width, height int,
) OnlineLobbyModel {
	return OnlineLobbyModel{
		state:       OnlineStateChooseMode,
		width:       width,
		height:      height,
		keyMapper:   NewKeyMapper(),
		gameID:      gameID,
		sessionID:   sessionID,
		coordinator: coordinator,
		eventChan:   eventChan,
	}
}

// Init initializes the lobby model.
func (m OnlineLobbyModel) Init() tea.Cmd {
	return m.waitForEvent()
}

// waitForEvent returns a command that waits for coordinator events.
func (m OnlineLobbyModel) waitForEvent() tea.Cmd {
	return func() tea.Msg {
		if m.eventChan == nil {
			return nil
		}
		evt, ok := <-m.eventChan
		if !ok {
			return nil
		}
		return evt
	}
}

// Update handles messages.
func (m OnlineLobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case multiplayer.LobbyCreatedEvent:
		m.lobbyCode = msg.Code
		m.state = OnlineStateHostWaiting
		return m, m.waitForEvent()
	case multiplayer.LobbyJoinedEvent:
		m.side = msg.Side
		m.opponentID = msg.OpponentID
		return m, m.waitForEvent()
	case multiplayer.LobbyErrorEvent:
		m.joinError = msg.Message
		if m.state == OnlineStateJoinWaiting {
			m.state = OnlineStateJoinEnterCode
		}
		return m, m.waitForEvent()
	case multiplayer.LobbyPlayerLeftEvent:
		// If in host waiting state and joiner left, stay waiting
		return m, m.waitForEvent()
	case multiplayer.MatchStartedEvent:
		m.matchID = msg.MatchID
		m.side = msg.Side
		m.state = OnlineStateInMatch
		return m, nil // Exit to start game
	case multiplayer.MatchEndedEvent:
		m.state = OnlineStateMatchEnded
		return m, nil
	}
	return m, nil
}

func (m OnlineLobbyModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit
	if key == "ctrl+c" {
		m.quitting = true
		return m, tea.Quit
	}

	switch m.state {
	case OnlineStateChooseMode:
		return m.handleChooseModeKey(msg)
	case OnlineStateHostWaiting:
		return m.handleHostWaitingKey(msg)
	case OnlineStateJoinEnterCode:
		return m.handleJoinCodeKey(msg)
	case OnlineStateJoinWaiting:
		return m.handleJoinWaitingKey(msg)
	}

	return m, nil
}

func (m OnlineLobbyModel) handleChooseModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "h", "H", "1":
		// Host
		m.coordinator.Send(multiplayer.CreateLobbyMsg{
			SessionID: m.sessionID,
			GameID:    m.gameID,
		})
		return m, m.waitForEvent()
	case "j", "J", "2":
		// Join
		m.state = OnlineStateJoinEnterCode
		m.joinCodeInput = ""
		m.joinError = ""
		return m, nil
	case "esc", "b":
		m.backToMenu = true
		return m, nil
	case "q":
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m OnlineLobbyModel) handleHostWaitingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "b":
		// Cancel lobby
		m.coordinator.Send(multiplayer.CancelLobbyMsg{
			SessionID: m.sessionID,
			Code:      m.lobbyCode,
		})
		m.cancelled = true
		m.backToMenu = true
		return m, nil
	case "q":
		// Cancel and quit
		m.coordinator.Send(multiplayer.CancelLobbyMsg{
			SessionID: m.sessionID,
			Code:      m.lobbyCode,
		})
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m OnlineLobbyModel) handleJoinCodeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "b":
		m.backToMenu = true
		return m, nil
	case "enter":
		if m.joinCodeInput != "" {
			m.state = OnlineStateJoinWaiting
			m.joinError = ""
			m.coordinator.Send(multiplayer.JoinLobbyMsg{
				SessionID: m.sessionID,
				Code:      m.joinCodeInput,
			})
			return m, m.waitForEvent()
		}
	case "backspace":
		if m.joinCodeInput != "" {
			m.joinCodeInput = m.joinCodeInput[:len(m.joinCodeInput)-1]
		}
	default:
		// Accept alphanumeric input for code
		if len(key) == 1 && len(m.joinCodeInput) < 6 {
			c := strings.ToUpper(key)
			if (c[0] >= 'A' && c[0] <= 'Z') || (c[0] >= '0' && c[0] <= '9') {
				m.joinCodeInput += c
			}
		}
	}

	return m, nil
}

func (m OnlineLobbyModel) handleJoinWaitingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "b":
		// Leave lobby attempt
		m.coordinator.Send(multiplayer.LeaveLobbyMsg{
			SessionID: m.sessionID,
			Code:      m.joinCodeInput,
		})
		m.state = OnlineStateJoinEnterCode
		return m, nil
	}

	return m, nil
}

// View renders the current state.
func (m OnlineLobbyModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	switch m.state {
	case OnlineStateChooseMode:
		b.WriteString(m.viewChooseMode())
	case OnlineStateHostWaiting:
		b.WriteString(m.viewHostWaiting())
	case OnlineStateJoinEnterCode:
		b.WriteString(m.viewJoinEnterCode())
	case OnlineStateJoinWaiting:
		b.WriteString(m.viewJoinWaiting())
	case OnlineStateMatchStarting:
		b.WriteString(m.viewMatchStarting())
	}

	return b.String()
}

func (m OnlineLobbyModel) viewChooseMode() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("ONLINE PONG", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Choose an option:", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("[H] Host a game", m.width))
	b.WriteString("\n")
	b.WriteString(centerText("[J] Join a game", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Esc: Back  |  Q: Quit", m.width))

	return b.String()
}

func (m OnlineLobbyModel) viewHostWaiting() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("HOSTING GAME", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Share this code with your opponent:", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText(fmt.Sprintf("[ %s ]", m.lobbyCode), m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Waiting for player to join...", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Esc: Cancel  |  Q: Quit", m.width))

	return b.String()
}

func (m OnlineLobbyModel) viewJoinEnterCode() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("JOIN GAME", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Enter the game code:", m.width))
	b.WriteString("\n\n")

	// Display code input with cursor
	codeDisplay := m.joinCodeInput
	if len(codeDisplay) < 6 {
		codeDisplay += "_"
		codeDisplay += strings.Repeat(" ", 5-len(m.joinCodeInput))
	}
	b.WriteString(centerText(fmt.Sprintf("[ %s ]", codeDisplay), m.width))
	b.WriteString("\n")

	if m.joinError != "" {
		b.WriteString("\n")
		b.WriteString(centerText(fmt.Sprintf("Error: %s", m.joinError), m.width))
	}

	b.WriteString("\n\n")
	b.WriteString(centerText("Enter: Connect  |  Esc: Back", m.width))

	return b.String()
}

func (m OnlineLobbyModel) viewJoinWaiting() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("CONNECTING", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText(fmt.Sprintf("Joining game: %s", m.joinCodeInput), m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Please wait...", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Esc: Cancel", m.width))

	return b.String()
}

func (m OnlineLobbyModel) viewMatchStarting() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("MATCH STARTING", m.width))
	b.WriteString("\n\n")

	sideText := "LEFT (P1)"
	if m.side == core.Player2 {
		sideText = "RIGHT (P2)"
	}
	b.WriteString(centerText(fmt.Sprintf("You are: %s", sideText), m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Get ready!", m.width))

	return b.String()
}

// State returns the current online state.
func (m OnlineLobbyModel) State() OnlineState {
	return m.state
}

// BackToMenu returns true if user wants to go back to menu.
func (m OnlineLobbyModel) BackToMenu() bool {
	return m.backToMenu
}

// IsQuitting returns true if user wants to quit entirely.
func (m OnlineLobbyModel) IsQuitting() bool {
	return m.quitting
}

// MatchID returns the match ID if a match was started.
func (m OnlineLobbyModel) MatchID() multiplayer.MatchID {
	return m.matchID
}

// Side returns which side (P1/P2) this session plays.
func (m OnlineLobbyModel) Side() core.PlayerID {
	return m.side
}

// LobbyCode returns the lobby code.
func (m OnlineLobbyModel) LobbyCode() string {
	return m.lobbyCode
}

// PongModeModel lets users choose between Vs CPU and Online PvP for Pong.
type PongModeModel struct {
	cursor    int
	width     int
	height    int
	keyMapper *KeyMapper
	selected  multiplayer.MatchMode
	choosing  bool
	quitting  bool
	back      bool
}

// NewPongModeModel creates a new pong mode selection model.
func NewPongModeModel(width, height int) PongModeModel {
	return PongModeModel{
		cursor:    0,
		width:     width,
		height:    height,
		keyMapper: NewKeyMapper(),
		choosing:  true,
	}
}

// Init initializes the model.
func (m PongModeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m PongModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m PongModeModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.cursor < 1 { // Only 2 options
			m.cursor++
		}
	case MenuActionSelect:
		m.choosing = false
		if m.cursor == 0 {
			m.selected = multiplayer.MatchModeVsCPU
		} else {
			m.selected = multiplayer.MatchModeOnlinePvP
		}
		return m, nil
	case MenuActionBack:
		m.back = true
		return m, nil
	}

	return m, nil
}

// View renders the mode selection.
func (m PongModeModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText("PONG", m.width))
	b.WriteString("\n\n")
	b.WriteString(centerText("Select game mode:", m.width))
	b.WriteString("\n\n")

	modes := []string{"Vs CPU", "Online PvP"}
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

// Selected returns the selected mode, or -1 if still choosing.
func (m PongModeModel) Selected() multiplayer.MatchMode {
	if m.choosing {
		return -1
	}
	return m.selected
}

// IsChoosing returns true if still in selection mode.
func (m PongModeModel) IsChoosing() bool {
	return m.choosing
}

// IsQuitting returns true if user wants to quit.
func (m PongModeModel) IsQuitting() bool {
	return m.quitting
}

// WantsBack returns true if user pressed back.
func (m PongModeModel) WantsBack() bool {
	return m.back
}
