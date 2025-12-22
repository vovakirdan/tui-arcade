# PixelFlow Core Library

A deterministic core library for a conveyor-style pixel puzzle game. This package provides all game logic without any UI dependencies.

## Game Mechanics

### Grid
- Rectangular grid of cells (W x H)
- Cells are either empty or filled with a colored pixel
- Colors: Pink, Cyan, Green, Yellow, Purple

### Rail/Conveyor Loop
- A closed loop running clockwise around the outside of the grid
- Rail has `2*(W+H)` positions
- Order: Top edge → Right edge → Bottom edge → Left edge → back to Top
- Each rail position has an inward shooting direction

### Shooters
- Shooters have a **color** and **ammo** count
- Shooters only remove pixels matching their color
- When ammo reaches 0, shooter disappears

### Active Shooter Behavior
Each simulation tick, every active shooter:
1. **Fires** (if not dry and has ammo):
   - Traces a ray inward from current rail position
   - If ray hits a matching-color pixel: removes pixel, decrements ammo
   - If ray hits a different-color pixel: becomes "dry" (stops firing this lap)
   - If ray hits nothing: no effect
2. **Moves** to next rail position (clockwise)
3. **Lap Completion** check:
   - If ammo > 0: moves to waiting slots
   - If ammo = 0: disappears

### Deck & Capacity
- **Deck**: Queue of shooters waiting to be launched (top = index 0)
- **Capacity**: Maximum shooters allowed on rail simultaneously (e.g., 5)
- **Waiting Slots**: Shooters that complete a lap with remaining ammo park here
- Launch policy: deck first, then waiting slots (when deck empty)

## Package Structure

```
internal/pixelflow/
├── core/                 # Core game logic (no UI deps)
│   ├── color.go         # Color enum (Pink, Cyan, Green, Yellow, Purple)
│   ├── coord.go         # Coordinate and direction types
│   ├── grid.go          # Grid construction and manipulation
│   ├── rail.go          # Rail/conveyor loop
│   ├── shooter.go       # Shooter and ActiveShooter types
│   ├── state.go         # Complete game state
│   ├── sim.go           # Simulation logic
│   ├── generator.go     # Deck generator with style constraints
│   ├── validate.go      # Deck validation
│   └── render_ascii.go  # ASCII renderer for debugging
├── levels/
│   ├── loader.go        # Level file loader
│   └── formats/
│       └── yaml.go      # YAML format parser
├── testdata/
│   └── levels/          # Sample level files
└── core_test/           # Unit tests
```

## Level File Format (YAML)

```yaml
id: "lvl01"
name: "Level Name"
size:
  w: 10
  h: 10
capacity: 5  # Optional, default 5
pixels:
  - { x: 0, y: 0, c: "pink" }
  - { x: 1, y: 0, c: "cyan" }
  - { x: 2, y: 0, c: "green" }
metadata:  # Optional
  difficulty: "easy"
  author: "name"
```

## Usage

### Loading a Level

```go
loader := levels.NewLoader("path/to/levels")
lvl, err := loader.LoadByID("lvl01")
if err != nil {
    log.Fatal(err)
}

grid := lvl.ToGrid()
rail := lvl.ToRail()
```

### Generating a Solvable Deck

```go
params := core.DefaultGenParams()
params.Seed = 12345
params.MaxAmmoPerShooter = 40

deck, err := core.GenerateDeckForGrid(grid, rail, params)
if err != nil {
    log.Fatal(err)
}
```

### Running Simulation

```go
state := core.NewState(grid, deck, capacity)

// Manual control
state.LaunchTop()
result := state.StepTick()

// Or auto-play
steps, cleared := state.RunUntilIdle(10000)
```

### Validating a Deck

```go
err := core.ValidateDeck(grid, deck, capacity, maxSteps)
if err != nil {
    log.Printf("Invalid deck: %v", err)
}
```

## Generator Style Constraints

The generator supports style parameters for "beautiful" queue generation:

```go
params := core.GenParams{
    Capacity:           5,
    Seed:               12345,
    MaxAmmoPerShooter:  40,
    MinAmmoPerShooter:  1,
    TargetLargeAmmo:    40,   // Target for "large" shooters
    TargetMediumAmmo:   20,   // Target for "medium" shooters
    TargetSmallAmmo:    10,   // Target for "small" shooters
    LargeProbability:   0.4,  // 40% chance of large
    MediumProbability:  0.3,  // 30% chance of medium
    PreferAlternating:  true, // Try to alternate colors
    MaxConsecutiveSame: 3,    // Max same-color in a row
}
```

This produces decks with patterns like `40/40/20/10/40/...` instead of uniform `1/1/1/...`.

## Guarantees

1. **Exact Ammo Match**: Total ammo per color equals exact pixel count
2. **Solvability**: Generated decks are validated to clear the grid
3. **Determinism**: Same seed produces identical results
4. **No External Dependencies**: Core package has zero UI dependencies

## Running Tests

```bash
go test ./internal/pixelflow/...
```

## Design Principles

- **Determinism**: No `time.Now()` or unseeded randomness in core
- **Isolation**: Core has no dependency on levels package
- **Testability**: All logic can be unit tested in isolation
- **Simplicity**: Minimal abstractions, clear data flow
