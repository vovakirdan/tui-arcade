// Package t2048 implements the classic 2048 puzzle game with campaign and endless modes.
package t2048

// Level defines a campaign level with a target tile.
type Level struct {
	ID     int
	Name   string
	Target int     // Target tile value to reach
	Spawn4 float64 // Probability of spawning 4 instead of 2 (0.0-1.0)
}

// Levels defines the 10 campaign levels with increasing difficulty.
// Targets are realistic for a 4x4 grid (8192 is very hard but achievable).
// Spawn4 probability increases to make later levels harder.
var Levels = []Level{
	{ID: 1, Name: "Warm-up", Target: 128, Spawn4: 0.10},
	{ID: 2, Name: "Getting Started", Target: 256, Spawn4: 0.10},
	{ID: 3, Name: "Building Momentum", Target: 512, Spawn4: 0.10},
	{ID: 4, Name: "The Climb", Target: 1024, Spawn4: 0.10},
	{ID: 5, Name: "Classic 2048", Target: 2048, Spawn4: 0.10},
	{ID: 6, Name: "Beyond Limits", Target: 4096, Spawn4: 0.12},
	{ID: 7, Name: "Master Class", Target: 8192, Spawn4: 0.15},
	{ID: 8, Name: "Expert Challenge", Target: 8192, Spawn4: 0.18},
	{ID: 9, Name: "Grandmaster", Target: 8192, Spawn4: 0.20},
	{ID: 10, Name: "Ultimate Champion", Target: 8192, Spawn4: 0.25},
}

// LevelCount returns the number of campaign levels.
func LevelCount() int {
	return len(Levels)
}

// GetLevel returns the level at the given index (0-based).
// Returns nil if index is out of range.
func GetLevel(index int) *Level {
	if index < 0 || index >= len(Levels) {
		return nil
	}
	return &Levels[index]
}

// LevelNames returns the names of all levels.
func LevelNames() []string {
	names := make([]string, len(Levels))
	for i, lvl := range Levels {
		names[i] = lvl.Name
	}
	return names
}

// LevelTargets returns the targets of all levels.
func LevelTargets() []int {
	targets := make([]int, len(Levels))
	for i, lvl := range Levels {
		targets[i] = lvl.Target
	}
	return targets
}
