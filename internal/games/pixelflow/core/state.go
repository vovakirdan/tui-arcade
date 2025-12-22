package core

import (
	"fmt"
	"hash/fnv"
)

// State represents the complete game state at any point in time.
type State struct {
	Grid     *Grid           // The pixel grid
	Rail     Rail            // Rail/conveyor loop definition
	Deck     []Shooter       // Queue of shooters to launch (index 0 is top)
	Active   []ActiveShooter // Shooters currently on the rail
	Waiting  []Shooter       // Shooters parked in waiting slots (FIFO)
	Capacity int             // Max shooters allowed on rail simultaneously
	Tick     uint64          // Current simulation tick
	NextID   int             // Next shooter ID to assign
}

// NewState creates a new game state from level data.
func NewState(grid *Grid, deck []Shooter, capacity int) *State {
	rail := NewRail(grid.W, grid.H)

	// Assign IDs to deck if not set
	maxID := 0
	for i := range deck {
		if deck[i].ID > maxID {
			maxID = deck[i].ID
		}
	}

	return &State{
		Grid:     grid.Clone(),
		Rail:     rail,
		Deck:     CloneDeck(deck),
		Active:   make([]ActiveShooter, 0),
		Waiting:  make([]Shooter, 0),
		Capacity: capacity,
		Tick:     0,
		NextID:   maxID + 1,
	}
}

// Clone returns a deep copy of the state.
func (s *State) Clone() *State {
	return &State{
		Grid:     s.Grid.Clone(),
		Rail:     s.Rail, // Rail is immutable, no need to clone
		Deck:     CloneDeck(s.Deck),
		Active:   CloneActive(s.Active),
		Waiting:  CloneDeck(s.Waiting),
		Capacity: s.Capacity,
		Tick:     s.Tick,
		NextID:   s.NextID,
	}
}

// CanLaunch returns true if we can launch the top deck shooter.
// Conditions: deck not empty, rail not at capacity.
func (s *State) CanLaunch() bool {
	return len(s.Deck) > 0 && len(s.Active) < s.Capacity
}

// CanRelaunchWaiting returns true if we can move a waiting shooter to rail.
func (s *State) CanRelaunchWaiting() bool {
	return len(s.Waiting) > 0 && len(s.Active) < s.Capacity
}

// LaunchTop launches the top deck shooter onto the rail.
// Returns true if successful.
func (s *State) LaunchTop() bool {
	if !s.CanLaunch() {
		return false
	}

	shooter := s.Deck[0]
	s.Deck = s.Deck[1:]

	spawnIdx := s.Rail.SpawnIndex()
	active := ActiveShooter{
		Shooter:     shooter,
		RailIndex:   spawnIdx,
		StartIndex:  spawnIdx,
		Dry:         false,
		LapProgress: 0,
	}
	s.Active = append(s.Active, active)
	return true
}

// RelaunchWaiting moves the first waiting shooter back to the rail.
// Returns true if successful.
func (s *State) RelaunchWaiting() bool {
	if !s.CanRelaunchWaiting() {
		return false
	}

	shooter := s.Waiting[0]
	s.Waiting = s.Waiting[1:]

	spawnIdx := s.Rail.SpawnIndex()
	active := ActiveShooter{
		Shooter:     shooter,
		RailIndex:   spawnIdx,
		StartIndex:  spawnIdx,
		Dry:         false,
		LapProgress: 0,
	}
	s.Active = append(s.Active, active)
	return true
}

// TopDeckShooter returns the top shooter in the deck, or nil if empty.
func (s *State) TopDeckShooter() *Shooter {
	if len(s.Deck) == 0 {
		return nil
	}
	return &s.Deck[0]
}

// IsWon returns true if the grid is empty (level cleared).
func (s *State) IsWon() bool {
	return s.Grid.IsEmpty()
}

// IsLost returns true if the game cannot be won:
// - Grid not empty AND
// - No active shooters AND
// - No deck AND
// - No waiting shooters
func (s *State) IsLost() bool {
	if s.Grid.IsEmpty() {
		return false
	}
	return len(s.Active) == 0 && len(s.Deck) == 0 && len(s.Waiting) == 0
}

// IsGameOver returns true if won or lost.
func (s *State) IsGameOver() bool {
	return s.IsWon() || s.IsLost()
}

// RemainingPixels returns the count of filled pixels.
func (s *State) RemainingPixels() int {
	return s.Grid.FilledCount()
}

// RemainingAmmo returns total ammo across deck + active + waiting.
func (s *State) RemainingAmmo() int {
	total := 0
	for _, shooter := range s.Deck {
		total += shooter.Ammo
	}
	for _, active := range s.Active {
		total += active.Ammo
	}
	for _, waiting := range s.Waiting {
		total += waiting.Ammo
	}
	return total
}

// Snapshot returns a hash representing the current state.
// Useful for detecting cycles or comparing states.
func (s *State) Snapshot() uint64 {
	h := fnv.New64a()

	// Grid state
	fmt.Fprintf(h, "G:%d;", s.Grid.Hash())

	// Deck
	fmt.Fprintf(h, "D:")
	for _, d := range s.Deck {
		fmt.Fprintf(h, "%d:%d:%d,", d.ID, d.Color, d.Ammo)
	}

	// Active
	fmt.Fprintf(h, ";A:")
	for _, a := range s.Active {
		fmt.Fprintf(h, "%d:%d:%d:%d:%v,", a.ID, a.RailIndex, a.Color, a.Ammo, a.Dry)
	}

	// Waiting
	fmt.Fprintf(h, ";W:")
	for _, w := range s.Waiting {
		fmt.Fprintf(h, "%d:%d:%d,", w.ID, w.Color, w.Ammo)
	}

	fmt.Fprintf(h, ";T:%d", s.Tick)

	return h.Sum64()
}

// ActiveShooterAt returns the active shooter at a given rail index, or nil.
func (s *State) ActiveShooterAt(railIndex int) *ActiveShooter {
	for i := range s.Active {
		if s.Active[i].RailIndex == railIndex {
			return &s.Active[i]
		}
	}
	return nil
}

// ActiveShootersAtIndex returns all active shooters at the given rail index.
// Multiple shooters can be at same position.
func (s *State) ActiveShootersAtIndex(railIndex int) []int {
	indices := make([]int, 0)
	for i, a := range s.Active {
		if a.RailIndex == railIndex {
			indices = append(indices, i)
		}
	}
	return indices
}
