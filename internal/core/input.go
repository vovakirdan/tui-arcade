package core

import "github.com/vovakirdan/tui-arcade/internal/multiplayer"

// Action represents a semantic game action, abstracted from physical key presses.
// This allows games to work with high-level intents rather than raw input.
type Action int

const (
	ActionNone    Action = iota
	ActionUp             // W, Up arrow - move up (Pong paddle)
	ActionDown           // S, Down arrow - move down (Pong paddle)
	ActionJump           // Space, W, Up - primary action (jump, flap) - also used as ActionUp for runners
	ActionDuck           // S, Down - secondary action (duck, dive)
	ActionConfirm        // Enter - confirm selection in menu
	ActionBack           // B, Escape - go back to menu
	ActionRestart        // R key - restart game after game over
	ActionQuit           // Q, Ctrl+C - exit game/session
	ActionPause          // P, Escape - pause/unpause game
)

// String returns a human-readable name for the action.
func (a Action) String() string {
	switch a {
	case ActionNone:
		return "None"
	case ActionUp:
		return "Up"
	case ActionDown:
		return "Down"
	case ActionJump:
		return "Jump"
	case ActionDuck:
		return "Duck"
	case ActionConfirm:
		return "Confirm"
	case ActionBack:
		return "Back"
	case ActionRestart:
		return "Restart"
	case ActionQuit:
		return "Quit"
	case ActionPause:
		return "Pause"
	default:
		return "Unknown"
	}
}

// InputFrame represents the input state for a single player during one simulation tick.
// It contains all actions that were triggered during this frame.
type InputFrame struct {
	// Actions maps action types to whether they were triggered this frame.
	// Using a map allows checking multiple actions without order dependency.
	Actions map[Action]bool
}

// NewInputFrame creates an empty input frame.
func NewInputFrame() InputFrame {
	return InputFrame{
		Actions: make(map[Action]bool),
	}
}

// Set marks an action as triggered for this frame.
func (f *InputFrame) Set(a Action) {
	if f.Actions == nil {
		f.Actions = make(map[Action]bool)
	}
	f.Actions[a] = true
}

// Has returns true if the given action was triggered this frame.
func (f InputFrame) Has(a Action) bool {
	if f.Actions == nil {
		return false
	}
	return f.Actions[a]
}

// Clear resets all actions for the next frame.
func (f *InputFrame) Clear() {
	for k := range f.Actions {
		delete(f.Actions, k)
	}
}

// Clone creates a copy of this input frame.
func (f InputFrame) Clone() InputFrame {
	clone := NewInputFrame()
	for k, v := range f.Actions {
		clone.Actions[k] = v
	}
	return clone
}

// MultiInputFrame contains input from all players for a single tick.
// Platform builds this from keyboard input (Player1) and AI (Player2 for CPU games).
// Games consume this interface without knowing input source.
type MultiInputFrame struct {
	// ByPlayer maps player IDs to their input frames.
	ByPlayer map[multiplayer.PlayerID]InputFrame
}

// NewMultiInputFrame creates an empty multi-input frame.
func NewMultiInputFrame() MultiInputFrame {
	return MultiInputFrame{
		ByPlayer: make(map[multiplayer.PlayerID]InputFrame),
	}
}

// Player returns the input frame for a specific player.
// Returns an empty frame if player has no input.
func (m MultiInputFrame) Player(id multiplayer.PlayerID) InputFrame {
	if m.ByPlayer == nil {
		return NewInputFrame()
	}
	if frame, ok := m.ByPlayer[id]; ok {
		return frame
	}
	return NewInputFrame()
}

// SetPlayer sets the input frame for a specific player.
func (m *MultiInputFrame) SetPlayer(id multiplayer.PlayerID, frame InputFrame) {
	if m.ByPlayer == nil {
		m.ByPlayer = make(map[multiplayer.PlayerID]InputFrame)
	}
	m.ByPlayer[id] = frame
}

// Player1 returns the input frame for Player 1 (convenience method).
func (m MultiInputFrame) Player1() InputFrame {
	return m.Player(multiplayer.Player1)
}

// Player2 returns the input frame for Player 2 (convenience method).
func (m MultiInputFrame) Player2() InputFrame {
	return m.Player(multiplayer.Player2)
}

// Clear resets all player inputs for the next frame.
func (m *MultiInputFrame) Clear() {
	for id := range m.ByPlayer {
		frame := m.ByPlayer[id]
		frame.Clear()
		m.ByPlayer[id] = frame
	}
}

// Clone creates a deep copy of this multi-input frame.
func (m MultiInputFrame) Clone() MultiInputFrame {
	clone := NewMultiInputFrame()
	for id, frame := range m.ByPlayer {
		clone.ByPlayer[id] = frame.Clone()
	}
	return clone
}
