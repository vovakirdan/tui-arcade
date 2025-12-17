package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/multiplayer"
)

// KeyMapper translates Bubble Tea key messages to game actions.
// This centralizes key bindings and makes them testable.
type KeyMapper struct{}

// NewKeyMapper creates a new key mapper with default bindings.
func NewKeyMapper() *KeyMapper {
	return &KeyMapper{}
}

// MapKey translates a key message to actions for Player1.
// Returns the action (may be ActionNone) and whether it's a quit request.
func (km *KeyMapper) MapKey(msg tea.KeyMsg) (action core.Action, isQuit bool) {
	key := msg.String()

	// Global quit keys
	switch key {
	case "ctrl+c", "q":
		return core.ActionQuit, true
	}

	// Game/menu actions
	switch key {
	case "w", "up":
		return core.ActionUp, false
	case "s", "down":
		return core.ActionDown, false
	case " ": // Space for jump (runners)
		return core.ActionJump, false
	case "enter":
		return core.ActionConfirm, false
	case "b", "esc":
		return core.ActionBack, false
	case "p":
		return core.ActionPause, false
	case "r":
		return core.ActionRestart, false
	}

	return core.ActionNone, false
}

// MapKeyToFrame updates an input frame based on a key message.
// Returns true if the key was a quit request.
func (km *KeyMapper) MapKeyToFrame(msg tea.KeyMsg, frame *core.InputFrame) bool {
	action, isQuit := km.MapKey(msg)
	if action != core.ActionNone {
		frame.Set(action)
	}
	return isQuit
}

// MapKeyToMultiFrame updates a multi-input frame for Player1 based on a key message.
// Returns true if the key was a quit request.
func (km *KeyMapper) MapKeyToMultiFrame(msg tea.KeyMsg, frame *core.MultiInputFrame) bool {
	action, isQuit := km.MapKey(msg)
	if action != core.ActionNone {
		p1 := frame.Player(multiplayer.Player1)
		p1.Set(action)
		frame.SetPlayer(multiplayer.Player1, p1)
	}
	return isQuit
}

// MenuAction represents a menu-specific action derived from input.
type MenuAction int

const (
	MenuActionNone MenuAction = iota
	MenuActionUp
	MenuActionDown
	MenuActionSelect
	MenuActionBack
	MenuActionQuit
)

// MapKeyToMenuAction translates a key to a menu action.
func (km *KeyMapper) MapKeyToMenuAction(msg tea.KeyMsg) MenuAction {
	key := msg.String()

	switch key {
	case "ctrl+c", "q":
		return MenuActionQuit
	case "w", "up", "k": // vim-style k for up
		return MenuActionUp
	case "s", "down", "j": // vim-style j for down
		return MenuActionDown
	case "enter", " ":
		return MenuActionSelect
	case "b", "esc":
		return MenuActionBack
	}

	return MenuActionNone
}
