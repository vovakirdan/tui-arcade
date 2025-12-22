package core

// ShotResult describes the outcome of a single shot.
type ShotResult struct {
	ShooterID    int     // ID of the shooter that fired
	Hit          bool    // True if the shot hit any filled cell
	HitAt        Coord   // Position where the shot hit (valid only if Hit)
	Removed      bool    // True if the hit pixel was removed (color matched)
	RemovedColor Color   // Color of the removed pixel (valid only if Removed)
	Blocked      bool    // True if the shot hit a pixel of wrong color
	BlockedColor Color   // Color that blocked (valid only if Blocked)
	Path         []Coord // All coordinates the shot traveled through
	OutOfBounds  bool    // True if the shot exited the grid without hitting anything
}

// Shoot fires a shot from the given shooter and modifies the grid.
// Returns the result describing what happened.
//
// Shooting behavior:
// 1. The shot starts at shooter.Pos and travels in shooter.Dir
// 2. Empty cells are passed through
// 3. If a filled cell is hit:
//   - If its color matches shooter.Color => remove pixel, stop
//   - If color does NOT match => stop without removing (blocked)
//
// 4. If the shot leaves grid bounds => stop (miss)
func Shoot(g *Grid, s Shooter) ShotResult {
	result := ShotResult{
		ShooterID: s.ID,
		Path:      make([]Coord, 0),
	}

	// Start tracing from shooter position
	pos := s.Pos
	dx, dy := s.Dir.Delta()

	// Maximum iterations to prevent infinite loops (grid diagonal + perimeter)
	maxSteps := g.W + g.H + 4

	for step := 0; step < maxSteps; step++ {
		// Move one step
		pos = pos.Add(dx, dy)
		result.Path = append(result.Path, pos)

		// Check if we're out of bounds
		if !g.InBounds(pos) {
			result.OutOfBounds = true
			return result
		}

		cell := g.Get(pos)

		// Empty cell: continue through
		if !cell.Filled {
			continue
		}

		// Hit a filled cell
		result.Hit = true
		result.HitAt = pos

		if cell.Color == s.Color {
			// Color matches: remove the pixel
			result.Removed = true
			result.RemovedColor = cell.Color
			g.SetEmpty(pos)
		} else {
			// Color doesn't match: blocked
			result.Blocked = true
			result.BlockedColor = cell.Color
		}

		return result
	}

	// Should not reach here, but mark as out of bounds just in case
	result.OutOfBounds = true
	return result
}

// ShootPure is a non-mutating version of Shoot.
// Returns a new grid (copy) and the shot result.
// Useful for testing and preview scenarios.
func ShootPure(g Grid, s Shooter) (Grid, ShotResult) {
	// Create a copy of the grid
	gridCopy := g.Clone()
	result := Shoot(gridCopy, s)
	return *gridCopy, result
}

// SimulateSequence simulates a sequence of shots in order.
// Modifies the grid and returns all results.
func SimulateSequence(g *Grid, shooters []Shooter, shotOrder []int) []ShotResult {
	results := make([]ShotResult, len(shotOrder))
	for i, shooterIdx := range shotOrder {
		if shooterIdx >= 0 && shooterIdx < len(shooters) {
			results[i] = Shoot(g, shooters[shooterIdx])
		}
	}
	return results
}

// CanShooterRemove checks if a shooter can potentially remove any pixel.
// Returns true if there's at least one matching-color pixel in the shooter's line of fire.
func CanShooterRemove(g *Grid, s Shooter) bool {
	pos := s.Pos
	dx, dy := s.Dir.Delta()
	maxSteps := g.W + g.H + 4

	for step := 0; step < maxSteps; step++ {
		pos = pos.Add(dx, dy)

		if !g.InBounds(pos) {
			return false
		}

		cell := g.Get(pos)
		if !cell.Filled {
			continue
		}

		// Hit a filled cell - check if it matches
		return cell.Color == s.Color
	}

	return false
}

// CountRemovableByShooter counts how many pixels a shooter can remove.
// This simulates repeated shooting until the shooter can no longer remove pixels.
// Uses a cloned grid to avoid mutation.
func CountRemovableByShooter(g *Grid, s Shooter) int {
	gridCopy := g.Clone()
	count := 0

	for {
		result := Shoot(gridCopy, s)
		if result.Removed {
			count++
		} else {
			break
		}
	}

	return count
}
