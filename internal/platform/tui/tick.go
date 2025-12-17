// Package tui provides the Bubble Tea integration for the arcade platform.
// It handles the terminal UI loop, input mapping, and game orchestration.
package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent to trigger a game simulation tick.
type TickMsg time.Time

// tickCmd returns a Bubble Tea command that sends tick messages at the specified rate.
func tickCmd(tickRate int) tea.Cmd {
	interval := time.Second / time.Duration(tickRate)
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
