# TUI Arcade

A terminal-based arcade gaming platform built with Go and Bubble Tea. Play classic-style games directly in your terminal, locally or over SSH with online multiplayer.

## Features

- **Classic Games**: Flappy Bird, Dino Runner, and Pong
- **SSH Server**: Host an arcade server for remote players
- **Online Multiplayer**: Play Pong against other players over SSH
- **Fixed FPS Simulation**: Deterministic game logic at configurable tick rates
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
make build
# or: go build -o arcade ./cmd/arcade
```

## Usage

### Local Play

```bash
# List available games
arcade list

# Play a game locally
arcade play flappy
arcade play dino
arcade play pong

# View high scores
arcade scores flappy

# Advanced options
arcade play flappy --fps 30      # Custom tick rate
arcade play dino --seed 12345    # Reproducible gameplay
arcade --db ./my.db play flappy  # Custom database path
```

### SSH Server (Multiplayer)

```bash
# Start the SSH arcade server
arcade serve

# With custom options
arcade serve --port 23234        # Custom port (default: 23234)
arcade serve --host-key ./key    # Custom host key path
```

Players can then connect:

```bash
ssh localhost -p 23234
```

### Online Pong (PvP)

When connected to the SSH server:

1. Select **Pong** from the game menu
2. Choose **Online PvP** mode
3. **Host**: Press `H` to create a lobby and get a 6-character join code
4. **Join**: Press `J` and enter the host's join code
5. Play against your opponent in real-time!

## Controls

### General

| Key | Action |
|-----|--------|
| Arrow Keys / WASD | Navigate menu / Move |
| Enter | Select |
| Esc / B | Back to menu |
| P | Pause |
| R | Restart (after game over) |
| Q / Ctrl+C | Quit |

### Flappy Bird / Dino Runner

| Key | Action |
|-----|--------|
| Space / Up / W | Jump / Flap |
| Down / S | Duck (Dino only) |

### Pong

| Key | Action |
|-----|--------|
| W / Up | Move paddle up |
| S / Down | Move paddle down |

## Available Games

### Flappy Bird
Navigate a bird through gaps in vertical pipes. Press Space to flap and stay airborne. Score increases with each pipe passed.

### Dino Runner
Jump over cacti in an endless runner. Speed increases as your score grows. Duck under flying obstacles.

### Pong
Classic two-player pong game. Play against CPU or challenge another player online!

- **Vs CPU**: Play against an AI opponent with adjustable difficulty
- **Online PvP**: Host or join a game to play against another SSH-connected player

## Architecture

The platform follows a clean SDK architecture:

```
internal/
  core/           # Platform-agnostic primitives (Screen, Input, Config)
  registry/       # Game registration and factory
  storage/        # SQLite score persistence
  multiplayer/    # Online multiplayer infrastructure
    coordinator/  # Lobby and match management
    session/      # Player session handling
    events/       # Network event types
  platform/
    tui/          # Bubble Tea integration & SSH server
  games/
    flappy/       # Flappy Bird implementation
    dino/         # Dino Runner implementation
    pong/         # Pong implementation (CPU & Online modes)
```

### Multiplayer Architecture

Online PvP uses an **authoritative server** model:

- **Coordinator**: Manages lobbies and pairs players using 6-character join codes
- **Match Loop**: Server runs the game simulation at a fixed tick rate
- **Snapshots**: Game state is broadcast to both players each tick
- **Input**: Players send inputs to the server, which applies them deterministically

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

### Adding Online Multiplayer Support

To make a game support online PvP, additionally implement:

```go
type OnlineGame interface {
    StepMulti(input core.MultiInputFrame) core.StepResult
    Snapshot() multiplayer.GameSnapshot
    IsGameOver() bool
    Winner() multiplayer.PlayerID
    Score1() int
    Score2() int
}
```

### Key Design Principles

- **No Bubble Tea in game logic**: Games depend only on `core` package types
- **Deterministic simulation**: Fixed timestep, seeded RNG, no `time.Now()` in game logic
- **Screen buffer abstraction**: Games render to `*core.Screen`, platform handles display
- **Transport-neutral multiplayer**: Games don't depend on SSH/network specifics

## Development

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Build
make build

# Run with debug options
./arcade play flappy --seed 42 --fps 30
```

## Configuration

Configuration files can be placed in `~/.arcade/` or `./configs/`:

- `~/.arcade/scores.db` - High scores database
- `~/.arcade/host_key` - SSH server host key (auto-generated)

## Examples

### Flappy Bird
```bash
   Score: 0                                             █████               Lvl: 0%   
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        ▄▄▄▄▄                         
                                                                                      
                                                                                      
          ●▶                                                                          
          ●●                                                                          
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                        ▀▀▀▀▀                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
                                                        █████                         
══════════════════════════════════════════════════════════════════════════════════════
```

### Dino Runner

```bash
   Score: 86                                                                          
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                                                                                      
                              ▓▓                                                      
         ◆█                   ▓▓                              ▓▓                      
        ███                   ▓▓                              ▓▓                      
         ╱╲                   ▓▓                              ▓▓                      
══════════════════════════════════════════════════════════════════════════════════════
```

### Pong

```bash
 P1                                   0        0                                  CPU 
                                           │                                          
                                                                                      
                                           │                                          
                                                                                      
                                           │                                          
                                                                                      
                                           │                                          
                                                                                      
                                           │                                          
                                                                                      
                                           │                                       █  
  █                                                                                █  
  █                                 ●      │                                       █  
  █                                                                                █  
  █                                        │                                       █  
  █                                                                                   
                                           │                                          
                                                                                      
                                           │                                          
                                                                                      
                                           │                                          
                                                                                      
                                           │                                          
                                                                                      
                                           │                                          
                                                                                      
                                           │                                          
```

## License

MIT
