package core

// Action represents a semantic game action, abstracted from physical key presses.
// This allows games to work with high-level intents rather than raw input.
type Action int

const (
	ActionNone    Action = iota
	ActionJump           // Space, Up arrow - primary action (jump, flap)
	ActionDuck           // Down arrow - secondary action (duck, dive) - for future use
	ActionRestart        // R key - restart game after game over
	ActionQuit           // Q, Ctrl+C - exit game
	ActionPause          // P, Escape - pause/unpause game
)

// String returns a human-readable name for the action.
func (a Action) String() string {
	switch a {
	case ActionNone:
		return "None"
	case ActionJump:
		return "Jump"
	case ActionDuck:
		return "Duck"
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

// InputFrame represents the input state for a single simulation tick.
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
