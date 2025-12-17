// Package breakout implements a Breakout/Arkanoid-style brick breaker game.
package breakout

// BrickType represents different types of bricks.
type BrickType int

const (
	BrickEmpty  BrickType = iota // No brick
	BrickNormal                  // Standard brick, destroyed in one hit
	// Future: BrickHard, BrickIndestructible, etc.
)

// Brick represents a single brick in the level.
type Brick struct {
	Type   BrickType
	Points int  // Points awarded when destroyed
	Alive  bool // Whether brick is still present
}

// Level represents a playable level with brick layout.
type Level struct {
	Name   string
	Width  int       // Number of brick columns
	Height int       // Number of brick rows
	Bricks [][]Brick // 2D grid of bricks [row][col]
}

// Clone creates a deep copy of the level (for reset).
func (l *Level) Clone() *Level {
	clone := &Level{
		Name:   l.Name,
		Width:  l.Width,
		Height: l.Height,
		Bricks: make([][]Brick, len(l.Bricks)),
	}
	for i, row := range l.Bricks {
		clone.Bricks[i] = make([]Brick, len(row))
		copy(clone.Bricks[i], row)
	}
	return clone
}

// CountAlive returns the number of remaining (alive) bricks.
func (l *Level) CountAlive() int {
	count := 0
	for _, row := range l.Bricks {
		for _, b := range row {
			if b.Alive && b.Type != BrickEmpty {
				count++
			}
		}
	}
	return count
}

// ParseLevel creates a Level from an ASCII map.
// Characters:
//
//	'#' = normal brick (10 points)
//	'.' = empty
//	'1'-'9' = brick with custom points (10 * digit)
func ParseLevel(name string, lines []string) *Level {
	if len(lines) == 0 {
		return &Level{Name: name}
	}

	// Find max width
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	level := &Level{
		Name:   name,
		Width:  maxWidth,
		Height: len(lines),
		Bricks: make([][]Brick, len(lines)),
	}

	for row, line := range lines {
		level.Bricks[row] = make([]Brick, maxWidth)
		for col := range maxWidth {
			var ch byte = '.'
			if col < len(line) {
				ch = line[col]
			}

			switch {
			case ch == '#':
				level.Bricks[row][col] = Brick{
					Type:   BrickNormal,
					Points: 10,
					Alive:  true,
				}
			case ch >= '1' && ch <= '9':
				level.Bricks[row][col] = Brick{
					Type:   BrickNormal,
					Points: int(ch-'0') * 10,
					Alive:  true,
				}
			default:
				level.Bricks[row][col] = Brick{
					Type:   BrickEmpty,
					Points: 0,
					Alive:  false,
				}
			}
		}
	}

	return level
}

// DefaultLevels returns the built-in levels.
func DefaultLevels() []*Level {
	return []*Level{
		ParseLevel("Classic", []string{
			"####################",
			"####################",
			"####################",
			"####################",
			"####################",
			".#####......#####..",
			"####################",
			"####################",
		}),
		ParseLevel("Pyramid", []string{
			"........####........",
			"......########......",
			"....############....",
			"..################..",
			"####################",
		}),
		ParseLevel("Checkerboard", []string{
			"#.#.#.#.#.#.#.#.#.#.",
			".#.#.#.#.#.#.#.#.#.#",
			"#.#.#.#.#.#.#.#.#.#.",
			".#.#.#.#.#.#.#.#.#.#",
			"#.#.#.#.#.#.#.#.#.#.",
			".#.#.#.#.#.#.#.#.#.#",
		}),
	}
}

// GetLevel returns a level by index (wraps around if index > len).
func GetLevel(index int) *Level {
	levels := DefaultLevels()
	if len(levels) == 0 {
		return ParseLevel("Empty", []string{})
	}
	return levels[index%len(levels)].Clone()
}

// LevelCount returns the total number of available levels.
func LevelCount() int {
	return len(DefaultLevels())
}
