// Package registry provides a global registry for game factories.
// Games register themselves in init() functions, allowing the platform
// to discover and instantiate games without hardcoded dependencies.
package registry

import (
	"fmt"
	"sort"
	"sync"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

// Game is the core interface that all arcade games must implement.
// Games contain pure logic with no external dependencies (especially no Bubble Tea).
// The platform handles input mapping, timing, and rendering.
type Game interface {
	// ID returns a unique identifier for this game (e.g., "flappy", "dino").
	// Used for CLI commands and score storage.
	ID() string

	// Title returns a human-readable name for display (e.g., "Flappy Bird").
	Title() string

	// Reset initializes or resets the game state.
	// Called once at start and again when restarting after game over.
	// The RuntimeConfig provides screen dimensions and RNG seed.
	Reset(cfg core.RuntimeConfig)

	// Step advances the simulation by one fixed tick.
	// Input is abstracted to platform-level actions (Jump, Pause, etc.).
	// Returns the result of this tick including current game state.
	Step(in core.InputFrame) core.StepResult

	// Render draws the current game state into the provided screen buffer.
	// The screen is pre-cleared before this call.
	Render(dst *core.Screen)

	// State returns the current game state (score, game over, paused).
	State() core.GameState
}

// GameInfo contains metadata about a registered game.
type GameInfo struct {
	ID    string
	Title string
}

// Factory is a function that creates a new instance of a game.
type Factory func() Game

var (
	factories = make(map[string]Factory)
	titles    = make(map[string]string)
	mu        sync.RWMutex
)

// Register adds a game factory to the registry.
// Typically called from a game's init() function.
// Panics if a game with the same ID is already registered.
func Register(id string, f Factory) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := factories[id]; exists {
		panic(fmt.Sprintf("registry: game %q already registered", id))
	}

	factories[id] = f

	// Get title by creating a temporary instance
	g := f()
	titles[id] = g.Title()
}

// List returns information about all registered games, sorted by ID.
func List() []GameInfo {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]GameInfo, 0, len(factories))
	for id := range factories {
		result = append(result, GameInfo{
			ID:    id,
			Title: titles[id],
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

// Create instantiates a new game by its ID.
// Returns an error if the game ID is not registered.
func Create(id string) (Game, error) {
	mu.RLock()
	defer mu.RUnlock()

	f, ok := factories[id]
	if !ok {
		return nil, fmt.Errorf("registry: unknown game %q", id)
	}

	return f(), nil
}

// Exists checks if a game with the given ID is registered.
func Exists(id string) bool {
	mu.RLock()
	defer mu.RUnlock()

	_, ok := factories[id]
	return ok
}
