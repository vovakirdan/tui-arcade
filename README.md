# TUI Arcade

A terminal-based arcade gaming platform built with Go and Bubble Tea. Play classic-style games directly in your terminal.

## Features

- **Fixed FPS Simulation**: Deterministic game logic at configurable tick rates
- **Extensible SDK**: Easy-to-use Game interface for adding new games
- **Score Persistence**: SQLite-based high score storage (pure Go, no CGO)
- **Cross-Platform**: Single binary, runs anywhere Go compiles

## Installation

```bash
go install github.com/vovakirdan/tui-arcade/cmd/arcade@latest
```

Or build from source:

```bash
git clone https://github.com/vovakirdan/tui-arcade.git
cd tui-arcade
go build -o arcade ./cmd/arcade
```

## Usage

```bash
# List available games
arcade list

# Play a game
arcade play flappy
arcade play dino

# View high scores
arcade scores flappy

# Advanced options
arcade play flappy --fps 30      # Custom tick rate
arcade play dino --seed 12345    # Reproducible gameplay
arcade --db ./my.db play flappy  # Custom database path
```

## Controls

| Key | Action |
|-----|--------|
| Space / Up / W | Jump / Flap |
| P / Esc | Pause |
| R | Restart (after game over) |
| Q / Ctrl+C | Quit |

## Available Games

### Flappy Bird
Navigate a bird through gaps in vertical pipes. Press Space to flap and stay airborne.

### Dino Runner
Jump over cacti in an endless runner. Speed increases as your score grows.

## Architecture

The platform follows a clean SDK architecture:

```
internal/
  core/       # Platform-agnostic primitives (Screen, Input, Config)
  registry/   # Game registration and factory
  storage/    # SQLite score persistence
  platform/
    tui/      # Bubble Tea integration
  games/
    flappy/   # Flappy Bird implementation
    dino/     # Dino Runner implementation
```

### Adding a New Game

1. Create a new package under `internal/games/<your-game>/`
2. Implement the `registry.Game` interface:

```go
type Game interface {
    ID() string                    // Unique identifier
    Title() string                 // Display name
    Reset(cfg core.RuntimeConfig)  // Initialize/restart game
    Step(in core.InputFrame) core.StepResult  // Advance simulation
    Render(dst *core.Screen)       // Draw to screen buffer
    State() core.GameState         // Get current state
}
```

3. Register in `init()`:

```go
func init() {
    registry.Register("my-game", func() registry.Game {
        return New()
    })
}
```

4. Add blank import in `cmd/arcade/main.go`:

```go
import _ "github.com/vovakirdan/tui-arcade/internal/games/mygame"
```

### Key Design Principles

- **No Bubble Tea in game logic**: Games depend only on `core` package types
- **Deterministic simulation**: Fixed timestep, seeded RNG, no `time.Now()` in game logic
- **Screen buffer abstraction**: Games render to `*core.Screen`, platform handles display

## Development

```bash
# Run tests
go test ./...

# Build
go build -o arcade ./cmd/arcade

# Run with debug options
./arcade play flappy --seed 42 --fps 30
```

## License

MIT
