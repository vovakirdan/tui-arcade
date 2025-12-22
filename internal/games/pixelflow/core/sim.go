package core

// ActionKind represents the type of action in simulation.
type ActionKind int

const (
	ActionNone       ActionKind = iota
	ActionLaunchTop             // Launch top deck shooter
	ActionLaunchWait            // Launch first waiting shooter
	ActionTick                  // Advance simulation by one tick
)

// RemovedEvent records a pixel removal.
type RemovedEvent struct {
	ShooterID int
	Coord     Coord
	Color     Color
}

// DryEvent records when a shooter becomes dry.
type DryEvent struct {
	ShooterID    int
	BlockedBy    Color
	BlockedCoord Coord
	RailIndex    int
}

// ShooterExitEvent records when a shooter leaves the rail.
type ShooterExitEvent struct {
	ShooterID int
	ToWaiting bool // true if moved to waiting, false if removed (ammo=0)
	AmmoLeft  int
}

// StepResult contains information about what happened during a simulation step.
type StepResult struct {
	Tick          uint64
	Removed       []RemovedEvent
	DryEvents     []DryEvent
	ShooterExited []ShooterExitEvent
	GridEmpty     bool
}

// StepTick advances the simulation by one tick.
// Each active shooter attempts to fire and then moves.
//
// Firing rules:
//  1. Trace ray from current rail position inward
//  2. If no filled pixel in ray: nothing happens, not dry
//  3. If first filled pixel matches shooter's color: remove pixel, decrement ammo
//  4. If first filled pixel has different color: shooter becomes Dry (stops firing this lap)
//
// After firing, shooter moves to next rail position.
// If shooter completes a full lap:
//   - If ammo <= 0: shooter disappears
//   - If ammo > 0: shooter moves to waiting slot
func (s *State) StepTick() StepResult {
	result := StepResult{
		Tick:          s.Tick,
		Removed:       make([]RemovedEvent, 0),
		DryEvents:     make([]DryEvent, 0),
		ShooterExited: make([]ShooterExitEvent, 0),
	}

	// Process each active shooter
	// We need to track which shooters to remove after processing
	toRemove := make([]int, 0)
	toWaiting := make([]int, 0)

	for i := range s.Active {
		shooter := &s.Active[i]

		// 1. Attempt to fire (if not dry and has ammo)
		if !shooter.Dry && shooter.Ammo > 0 {
			hitCoord, hitColor, hit := s.Rail.TraceRay(s.Grid, shooter.RailIndex)

			if hit {
				if hitColor == shooter.Color {
					// Color matches: remove pixel, decrement ammo
					s.Grid.SetEmpty(hitCoord)
					shooter.Ammo--
					result.Removed = append(result.Removed, RemovedEvent{
						ShooterID: shooter.ID,
						Coord:     hitCoord,
						Color:     hitColor,
					})
				} else {
					// Color mismatch: become dry
					shooter.Dry = true
					result.DryEvents = append(result.DryEvents, DryEvent{
						ShooterID:    shooter.ID,
						BlockedBy:    hitColor,
						BlockedCoord: hitCoord,
						RailIndex:    shooter.RailIndex,
					})
				}
			}
			// If no hit (empty ray): nothing happens, shooter stays not-dry
		}

		// 2. Move to next position
		shooter.RailIndex = s.Rail.Next(shooter.RailIndex)
		shooter.LapProgress++

		// 3. Check lap completion
		if shooter.HasCompletedLap(s.Rail.Len()) {
			if shooter.Ammo <= 0 {
				// No ammo left: disappear
				toRemove = append(toRemove, i)
				result.ShooterExited = append(result.ShooterExited, ShooterExitEvent{
					ShooterID: shooter.ID,
					ToWaiting: false,
					AmmoLeft:  0,
				})
			} else {
				// Has ammo: move to waiting
				toWaiting = append(toWaiting, i)
				result.ShooterExited = append(result.ShooterExited, ShooterExitEvent{
					ShooterID: shooter.ID,
					ToWaiting: true,
					AmmoLeft:  shooter.Ammo,
				})
			}
		}
	}

	// Process exits (in reverse order to maintain indices)
	// First collect shooters going to waiting
	for _, i := range toWaiting {
		s.Waiting = append(s.Waiting, s.Active[i].Shooter)
	}

	// Build new active list excluding exited shooters
	exitSet := make(map[int]bool)
	for _, i := range toRemove {
		exitSet[i] = true
	}
	for _, i := range toWaiting {
		exitSet[i] = true
	}

	newActive := make([]ActiveShooter, 0, len(s.Active)-len(exitSet))
	for i, a := range s.Active {
		if !exitSet[i] {
			newActive = append(newActive, a)
		}
	}
	s.Active = newActive

	s.Tick++
	result.Tick = s.Tick
	result.GridEmpty = s.Grid.IsEmpty()

	return result
}

// AutoLaunch implements a simple launch policy:
// 1. If can launch from deck, do it
// 2. Else if can relaunch from waiting and deck is empty, do it
// Returns true if any launch was made.
func (s *State) AutoLaunch() bool {
	if s.CanLaunch() {
		return s.LaunchTop()
	}
	// Only relaunch waiting if deck is empty (policy B from spec)
	if len(s.Deck) == 0 && s.CanRelaunchWaiting() {
		return s.RelaunchWaiting()
	}
	return false
}

// RunUntilIdle runs simulation until no active shooters remain.
// Automatically launches when possible using AutoLaunch policy.
// Returns total steps taken and whether grid was cleared.
func (s *State) RunUntilIdle(maxSteps int) (int, bool) {
	steps := 0

	for steps < maxSteps {
		// Try to launch if possible
		s.AutoLaunch()

		// If no active shooters and nothing to launch, we're done
		if len(s.Active) == 0 {
			// Try once more to launch
			if !s.AutoLaunch() {
				break
			}
		}

		// Step simulation
		s.StepTick()
		steps++

		if s.Grid.IsEmpty() {
			return steps, true
		}
	}

	return steps, s.Grid.IsEmpty()
}

// SimulateSingleShooterLap simulates what would happen if a shooter of given color
// started at spawnIndex and did one full lap on the current grid state.
// Returns: pixels it would remove (coords), and whether it would end dry.
// Does NOT modify the state.
func (s *State) SimulateSingleShooterLap(color Color, ammo int) (removed []Coord, endsDry bool, finalAmmo int) {
	// Work on a clone to avoid mutation
	gridClone := s.Grid.Clone()
	removed = make([]Coord, 0)

	railLen := s.Rail.Len()
	startIdx := s.Rail.SpawnIndex()
	currentIdx := startIdx
	dry := false
	currentAmmo := ammo

	for step := 0; step < railLen; step++ {
		if !dry && currentAmmo > 0 {
			hitCoord, hitColor, hit := s.Rail.TraceRay(gridClone, currentIdx)
			if hit {
				if hitColor == color {
					gridClone.SetEmpty(hitCoord)
					currentAmmo--
					removed = append(removed, hitCoord)
				} else {
					dry = true
				}
			}
		}
		currentIdx = s.Rail.Next(currentIdx)
	}

	return removed, dry, currentAmmo
}

// CountPotentialRemovalsForColor counts how many pixels of a given color
// could be removed by a single shooter doing one lap with unlimited ammo.
// This is useful for deck generation.
func (s *State) CountPotentialRemovalsForColor(color Color) int {
	removed, _, _ := s.SimulateSingleShooterLap(color, 1000000)
	return len(removed)
}
