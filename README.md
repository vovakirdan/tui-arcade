# TUI Arcade

A terminal-based arcade gaming platform built with Go and Bubble Tea. Play classic-style games directly in your terminal, locally or over SSH with online multiplayer.

## Features

- **Classic Games**: Flappy Bird, Dino Runner, Breakout, Snake, Pong, and 2048
- **SSH Server**: Host an arcade server for remote players
- **Online Multiplayer**: Play Pong against other players over SSH
- **Fixed FPS Simulation**: Deterministic game logic at configurable tick rates
- **Score Persistence**: SQLite-based high score storage (pure Go, no CGO)
- **Cross-Platform**: Single binary, runs anywhere Go compiles

## "Screen" *shots*
> To take a screenshot, press `Ctrl+S` in the game.
 - [Breakout](#breakout-2)
 - [Flappy Bird](#flappy-bird-1)
 - [Dino Runner](#dino-runner-1)
 - [Pong](#pong-1)
 - [Snake](#snake-2)
 - [2048](#2048-2)

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

### Breakout

| Key | Action |
|-----|--------|
| A / Left | Move paddle left |
| D / Right | Move paddle right |
| Space | Launch ball / Release sticky ball |
| P | Pause |
| R | Restart (after game over) |

### Snake

| Key | Action |
|-----|--------|
| Arrow Keys / WASD | Change direction |
| P | Pause |
| R | Restart (after game over) |

### 2048

| Key | Action |
|-----|--------|
| Arrow Keys / WASD | Slide tiles |
| P | Pause |
| R | Restart (after game over) |

## Available Games

### Flappy Bird
Navigate a bird through gaps in vertical pipes. Press Space to flap and stay airborne. Score increases with each pipe passed.

### Dino Runner
Jump over cacti in an endless runner. Speed increases as your score grows. Duck under flying obstacles.

### Pong
Classic two-player pong game. Play against CPU or challenge another player online!

- **Vs CPU**: Play against an AI opponent with adjustable difficulty
- **Online PvP**: Host or join a game to play against another SSH-connected player

### Breakout
Classic brick-breaker game with power-ups and multiple levels!

**Game Modes:**
- **Campaign**: Play through 10 unique levels. Complete all to win!
- **Endless**: Cycle through levels forever with increasing difficulty
- **Level Select**: Start from any level in campaign mode

**Brick Types:**
| Type | Symbol | Description |
|------|--------|-------------|
| Normal | `#` | Destroyed in one hit (10 points) |
| Hard | `H` | Requires 2 hits to destroy (20 points) |
| Solid | `X` | Indestructible |

**Power-Ups:**
Power-ups drop from destroyed bricks (18% chance):

| Pickup | Symbol | Effect | Duration |
|--------|--------|--------|----------|
| Widen | W | Expands paddle | 12 sec |
| Shrink | S | Shrinks paddle | 12 sec |
| Multiball | M | Spawns +2 balls | Instant |
| Sticky | T | Balls stick to paddle | 10 sec |
| Speed Up | + | Increases ball speed | 8 sec |
| Slow Down | - | Decreases ball speed | 8 sec |
| Extra Life | ♥ | +1 life | Instant |

**Levels:**
1. Classic - Simple brick wall
2. Pyramid - Triangle pattern
3. Checkerboard - Alternating bricks
4. Diamond - Diamond shape
5. Fortress - Hard brick walls with normal inside
6. Striped - Horizontal lines
7. Invaders - Space invader patterns
8. Heart - Heart shape
9. Castle - Castle with solid turrets
10. Final Boss - Hard brick fortress

### Snake
Classic snake game - collect food to grow, avoid walls and yourself!

**Game Modes:**
- **Campaign**: Play through 10 unique levels with different layouts
- **Endless**: Cycle through levels forever with increasing speed
- **Level Select**: Start from any level in campaign mode

**Symbols:**
| Symbol | Description |
|--------|-------------|
| `O` | Snake head |
| `o` | Snake body |
| `*` | Food |
| `#` | Wall/obstacle |

**Levels:**
1. Empty Box - Open arena
2. Central Pillar - Obstacle in the center
3. Two Pillars - Two vertical obstacles
4. Zigzag - Horizontal barriers
5. Rooms - Divided spaces with openings
6. Four Corners - Corner obstacles
7. Cross - Cross-shaped barriers
8. Scattered - Multiple small obstacles
9. Spiral - Spiral pattern walls
10. Final Maze - Complex maze layout

### 2048
Classic number puzzle game - slide tiles to combine them and reach the target!

**Game Modes:**
- **Campaign**: Progress through 10 levels with increasing targets (128 → 8192)
- **Endless**: Play forever, no target - game ends only when no moves remain
- **Level Select**: Start from any level in campaign mode

**How to Play:**
- Use arrow keys or WASD to slide all tiles in a direction
- Tiles with the same value merge into one when they collide
- Each merge doubles the tile value and adds to your score
- A new tile (2 or 4) spawns after each valid move
- Game over when no moves are possible

**Scoring:**
- Score increases by the merged tile value (e.g., two 4s merge into 8, +8 points)

**Campaign Levels:**
| Level | Name | Target | 4-spawn % |
|-------|------|--------|-----------|
| 1 | Warm-up | 128 | 10% |
| 2 | Getting Started | 256 | 10% |
| 3 | Building Momentum | 512 | 10% |
| 4 | The Climb | 1024 | 10% |
| 5 | Classic 2048 | 2048 | 10% |
| 6 | Beyond Limits | 4096 | 12% |
| 7 | Master Class | 8192 | 15% |
| 8 | Expert Challenge | 8192 | 18% |
| 9 | Grandmaster | 8192 | 20% |
| 10 | Ultimate Champion | 8192 | 25% |

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
    breakout/     # Breakout implementation (10 levels, power-ups)
    snake/        # Snake implementation (10 levels, campaign & endless)
    pong/         # Pong implementation (CPU & Online modes)
    t2048/        # 2048 implementation (10 levels, campaign & endless)
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

## Screenshots

### Breakout
```bash
 Score: 50                             Lives: 3                              Level: 1
──────────────────────────────────────────────────────────────────────────────────────
████████████████████████████████████████████████████████████████████████████████
▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
################################################################################
    ++++++++++++++++++++                        ++++++++++++++++++++
********************************************************************************
    ====    ====    ================    ====    ================================



    ●












     ========

                                                                                      
```

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

### Snake

```bash
 Snake (Endless) - Score: 2  Level: 1  Speed: 1
──────────────────────────────────────────────────────────────────────────────────────
######################################################################################
#                                                                                    #
#                 ooooO                                  *                           #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
#                                                                                    #
######################################################################################
```

### 2048

```bash
                                        2048
                      Score: 56                         Max: 16
                                       Endless


                      ┌─────────┬─────────┬─────────┬─────────┐
                      │         │         │         │         │
                      │    4    │    2    │         │         │
                      │         │         │         │         │
                      │         │         │         │         │
                      ├─────────┼─────────┼─────────┼─────────┤
                      │         │         │         │         │
                      │    4    │         │         │         │
                      │         │         │         │         │
                      │         │         │         │         │
                      ├─────────┼─────────┼─────────┼─────────┤
                      │         │         │         │         │
                      │   16    │    2    │         │         │
                      │         │         │         │         │
                      │         │         │         │         │
                      ├─────────┼─────────┼─────────┼─────────┤
                      │         │         │         │         │
                      │         │         │    2    │         │
                      │         │         │         │         │
                      │         │         │         │         │
                      └─────────┴─────────┴─────────┴─────────┘


                                                                                      
```

## License

MIT
