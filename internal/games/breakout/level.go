// Package breakout implements a Breakout/Arkanoid-style brick breaker game.
package breakout

// BrickType represents different types of bricks.
type BrickType int

const (
	BrickEmpty  BrickType = iota // No brick
	BrickNormal                  // Standard brick, destroyed in one hit
	BrickHard                    // Requires 2 hits to destroy
	BrickSolid                   // Indestructible
)

// Brick represents a single brick in the level.
type Brick struct {
	Type   BrickType
	Points int  // Points awarded when destroyed
	Alive  bool // Whether brick is still present
	HP     int  // Hit points remaining (for hard bricks)
}

// Level represents a playable level with brick layout.
type Level struct {
	ID     string
	Name   string
	Width  int       // Number of brick columns
	Height int       // Number of brick rows
	Bricks [][]Brick // 2D grid of bricks [row][col]
}

// Clone creates a deep copy of the level (for reset).
func (l *Level) Clone() *Level {
	clone := &Level{
		ID:     l.ID,
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

// CountAlive returns the number of remaining (alive, destroyable) bricks.
func (l *Level) CountAlive() int {
	count := 0
	for _, row := range l.Bricks {
		for _, b := range row {
			if b.Alive && b.Type != BrickEmpty && b.Type != BrickSolid {
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
//	'H' = hard brick (2 HP, 20 points)
//	'X' = solid/indestructible brick (0 points)
func ParseLevel(id, name string, lines []string) *Level {
	if len(lines) == 0 {
		return &Level{ID: id, Name: name}
	}

	// Find max width
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	level := &Level{
		ID:     id,
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
					HP:     1,
				}
			case ch >= '1' && ch <= '9':
				level.Bricks[row][col] = Brick{
					Type:   BrickNormal,
					Points: int(ch-'0') * 10,
					Alive:  true,
					HP:     1,
				}
			case ch == 'H' || ch == 'h':
				level.Bricks[row][col] = Brick{
					Type:   BrickHard,
					Points: 20,
					Alive:  true,
					HP:     2,
				}
			case ch == 'X' || ch == 'x':
				level.Bricks[row][col] = Brick{
					Type:   BrickSolid,
					Points: 0,
					Alive:  true,
					HP:     999, // Effectively indestructible
				}
			default:
				level.Bricks[row][col] = Brick{
					Type:   BrickEmpty,
					Points: 0,
					Alive:  false,
					HP:     0,
				}
			}
		}
	}

	return level
}

// BuiltinLevels returns all built-in levels.
func BuiltinLevels() []*Level {
	return []*Level{
		// Level 1: Classic
		ParseLevel("classic", "Classic", []string{
			"####################",
			"####################",
			"####################",
			"####################",
			"####################",
		}),

		// Level 2: Pyramid
		ParseLevel("pyramid", "Pyramid", []string{
			"........####........",
			"......########......",
			"....############....",
			"..################..",
			"####################",
		}),

		// Level 3: Checkerboard
		ParseLevel("checker", "Checkerboard", []string{
			"#.#.#.#.#.#.#.#.#.#.",
			".#.#.#.#.#.#.#.#.#.#",
			"#.#.#.#.#.#.#.#.#.#.",
			".#.#.#.#.#.#.#.#.#.#",
			"#.#.#.#.#.#.#.#.#.#.",
			".#.#.#.#.#.#.#.#.#.#",
		}),

		// Level 4: Diamond
		ParseLevel("diamond", "Diamond", []string{
			".........##.........",
			"........####........",
			".......######.......",
			"......########......",
			".....##########.....",
			"......########......",
			".......######.......",
			"........####........",
			".........##.........",
		}),

		// Level 5: Fortress (with hard bricks)
		ParseLevel("fortress", "Fortress", []string{
			"HHHHHHHHHHHHHHHHHHHH",
			"H..................H",
			"H.################.H",
			"H.################.H",
			"H.################.H",
			"H..................H",
			"HHHHHHHHHHHHHHHHHHHH",
		}),

		// Level 6: Striped
		ParseLevel("striped", "Striped", []string{
			"####################",
			"....................",
			"####################",
			"....................",
			"####################",
			"....................",
			"####################",
		}),

		// Level 7: Invaders
		ParseLevel("invaders", "Invaders", []string{
			"..#..........#......",
			".###........###.....",
			"#####......#####....",
			"#.#.#......#.#.#....",
			"#####......#####....",
			"....................",
			"..#..........#......",
			".###........###.....",
			"#####......#####....",
			"#.#.#......#.#.#....",
			"#####......#####....",
		}),

		// Level 8: Heart
		ParseLevel("heart", "Heart", []string{
			"..##....##..........",
			".####..####.........",
			"##############......",
			"##############......",
			".############.......",
			"..##########........",
			"...########.........",
			"....######..........",
			".....####...........",
			"......##............",
		}),

		// Level 9: Castle (with solid blocks)
		ParseLevel("castle", "Castle", []string{
			"X..X....X..X....X..X",
			"XXXX....XXXX....XXXX",
			"X..X....X..X....X..X",
			"....................",
			"####################",
			"####################",
			"####################",
			"####################",
		}),

		// Level 10: Final Boss (hard bricks everywhere)
		ParseLevel("boss", "Final Boss", []string{
			"HHHHHHHHHHHHHHHHHHHH",
			"H##################H",
			"H##################H",
			"H##################H",
			"H##################H",
			"H##################H",
			"HHHHHHHHHHHHHHHHHHHH",
		}),
	}
}

// GetLevelByID returns a level by its ID.
func GetLevelByID(id string) (*Level, bool) {
	for _, level := range BuiltinLevels() {
		if level.ID == id {
			return level.Clone(), true
		}
	}
	return nil, false
}

// GetLevel returns a level by index (wraps around if index >= len).
func GetLevel(index int) *Level {
	levels := BuiltinLevels()
	if len(levels) == 0 {
		return ParseLevel("empty", "Empty", []string{})
	}
	return levels[index%len(levels)].Clone()
}

// LevelCount returns the total number of available levels.
func LevelCount() int {
	return len(BuiltinLevels())
}
