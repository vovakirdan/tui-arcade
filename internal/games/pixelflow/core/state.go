package core

import (
	"fmt"
	"hash/fnv"
)

// Deck represents multiple queues of shooters.
// Each queue is independent; player selects which queue to launch from.
type Deck struct {
	Queues [][]Shooter // Multiple queues, each is FIFO (index 0 is top)
}

// NewDeck creates a deck with the specified number of queues.
func NewDeck(numQueues int) Deck {
	if numQueues < 1 {
		numQueues = 2
	}
	queues := make([][]Shooter, numQueues)
	for i := range queues {
		queues[i] = make([]Shooter, 0)
	}
	return Deck{Queues: queues}
}

// NumQueues returns the number of queues.
func (d *Deck) NumQueues() int {
	return len(d.Queues)
}

// QueueLen returns the length of a specific queue.
func (d *Deck) QueueLen(queueIndex int) int {
	if queueIndex < 0 || queueIndex >= len(d.Queues) {
		return 0
	}
	return len(d.Queues[queueIndex])
}

// TopOfQueue returns the top shooter of a queue, or nil if empty.
func (d *Deck) TopOfQueue(queueIndex int) *Shooter {
	if queueIndex < 0 || queueIndex >= len(d.Queues) {
		return nil
	}
	if len(d.Queues[queueIndex]) == 0 {
		return nil
	}
	return &d.Queues[queueIndex][0]
}

// PopFromQueue removes and returns the top shooter from a queue.
// Returns nil if queue is empty or invalid.
func (d *Deck) PopFromQueue(queueIndex int) *Shooter {
	if queueIndex < 0 || queueIndex >= len(d.Queues) {
		return nil
	}
	if len(d.Queues[queueIndex]) == 0 {
		return nil
	}
	shooter := d.Queues[queueIndex][0]
	d.Queues[queueIndex] = d.Queues[queueIndex][1:]
	return &shooter
}

// PushToQueue adds a shooter to the bottom of a queue.
func (d *Deck) PushToQueue(queueIndex int, s Shooter) {
	if queueIndex < 0 || queueIndex >= len(d.Queues) {
		return
	}
	d.Queues[queueIndex] = append(d.Queues[queueIndex], s)
}

// IsEmpty returns true if all queues are empty.
func (d *Deck) IsEmpty() bool {
	for _, q := range d.Queues {
		if len(q) > 0 {
			return false
		}
	}
	return true
}

// TotalShooters returns total shooters across all queues.
func (d *Deck) TotalShooters() int {
	total := 0
	for _, q := range d.Queues {
		total += len(q)
	}
	return total
}

// TotalAmmo returns total ammo across all queues.
func (d *Deck) TotalAmmo() int {
	total := 0
	for _, q := range d.Queues {
		for _, s := range q {
			total += s.Ammo
		}
	}
	return total
}

// Clone creates a deep copy of the deck.
func (d *Deck) Clone() Deck {
	queues := make([][]Shooter, len(d.Queues))
	for i, q := range d.Queues {
		queues[i] = CloneDeck(q)
	}
	return Deck{Queues: queues}
}

// WaitingSlots represents fixed-size waiting slots.
// Shooters that complete a lap with remaining ammo go here.
type WaitingSlots struct {
	Slots    []*Shooter // Fixed size array; nil means empty slot
	Capacity int        // Number of slots
}

// NewWaitingSlots creates waiting slots with given capacity.
func NewWaitingSlots(capacity int) WaitingSlots {
	if capacity < 1 {
		capacity = 5
	}
	return WaitingSlots{
		Slots:    make([]*Shooter, capacity),
		Capacity: capacity,
	}
}

// FindFreeSlot returns the index of the first free slot, or -1 if none.
func (w *WaitingSlots) FindFreeSlot() int {
	for i, s := range w.Slots {
		if s == nil {
			return i
		}
	}
	return -1
}

// Put places a shooter in a specific slot.
// Returns false if slot is occupied or invalid.
func (w *WaitingSlots) Put(slotIndex int, s Shooter) bool {
	if slotIndex < 0 || slotIndex >= w.Capacity {
		return false
	}
	if w.Slots[slotIndex] != nil {
		return false
	}
	shooter := s.Clone()
	w.Slots[slotIndex] = &shooter
	return true
}

// Take removes and returns the shooter from a slot.
// Returns nil if slot is empty or invalid.
func (w *WaitingSlots) Take(slotIndex int) *Shooter {
	if slotIndex < 0 || slotIndex >= w.Capacity {
		return nil
	}
	s := w.Slots[slotIndex]
	w.Slots[slotIndex] = nil
	return s
}

// Get returns the shooter in a slot without removing.
func (w *WaitingSlots) Get(slotIndex int) *Shooter {
	if slotIndex < 0 || slotIndex >= w.Capacity {
		return nil
	}
	return w.Slots[slotIndex]
}

// Count returns number of occupied slots.
func (w *WaitingSlots) Count() int {
	count := 0
	for _, s := range w.Slots {
		if s != nil {
			count++
		}
	}
	return count
}

// IsEmpty returns true if all slots are empty.
func (w *WaitingSlots) IsEmpty() bool {
	return w.Count() == 0
}

// Clone creates a deep copy.
func (w *WaitingSlots) Clone() WaitingSlots {
	slots := make([]*Shooter, w.Capacity)
	for i, s := range w.Slots {
		if s != nil {
			clone := s.Clone()
			slots[i] = &clone
		}
	}
	return WaitingSlots{
		Slots:    slots,
		Capacity: w.Capacity,
	}
}

// TotalAmmo returns total ammo in waiting slots.
func (w *WaitingSlots) TotalAmmo() int {
	total := 0
	for _, s := range w.Slots {
		if s != nil {
			total += s.Ammo
		}
	}
	return total
}

// State represents the complete game state at any point in time.
type State struct {
	Grid       *Grid           // The pixel grid
	Rail       Rail            // Rail/conveyor loop definition
	Deck       Deck            // Multiple queues of shooters
	Active     []ActiveShooter // Shooters currently on the rail
	Waiting    WaitingSlots    // Fixed waiting slots
	Capacity   int             // Max shooters allowed on rail simultaneously
	NumQueues  int             // Number of deck queues
	Tick       uint64          // Current simulation tick
	NextID     int             // Next shooter ID to assign
}

// NewState creates a new game state from level data.
// shooters is a flat list that will be distributed into queues round-robin.
func NewState(grid *Grid, shooters []Shooter, capacity int) *State {
	return NewStateWithQueues(grid, shooters, capacity, 2)
}

// NewStateWithQueues creates a new game state with specified number of queues.
func NewStateWithQueues(grid *Grid, shooters []Shooter, capacity, numQueues int) *State {
	if numQueues < 1 {
		numQueues = 2
	}
	rail := NewRail(grid.W, grid.H)

	// Assign IDs to shooters if not set
	maxID := 0
	for i := range shooters {
		if shooters[i].ID > maxID {
			maxID = shooters[i].ID
		}
	}

	// Distribute shooters round-robin into queues
	deck := NewDeck(numQueues)
	for i, s := range shooters {
		queueIdx := i % numQueues
		deck.PushToQueue(queueIdx, s.Clone())
	}

	return &State{
		Grid:      grid.Clone(),
		Rail:      rail,
		Deck:      deck,
		Active:    make([]ActiveShooter, 0),
		Waiting:   NewWaitingSlots(capacity),
		Capacity:  capacity,
		NumQueues: numQueues,
		Tick:      0,
		NextID:    maxID + 1,
	}
}

// Clone returns a deep copy of the state.
func (s *State) Clone() *State {
	return &State{
		Grid:      s.Grid.Clone(),
		Rail:      s.Rail, // Rail is immutable, no need to clone
		Deck:      s.Deck.Clone(),
		Active:    CloneActive(s.Active),
		Waiting:   s.Waiting.Clone(),
		Capacity:  s.Capacity,
		NumQueues: s.NumQueues,
		Tick:      s.Tick,
		NextID:    s.NextID,
	}
}

// CanLaunchFromQueue returns true if we can launch from specified queue.
func (s *State) CanLaunchFromQueue(queueIndex int) bool {
	return s.Deck.QueueLen(queueIndex) > 0 && len(s.Active) < s.Capacity
}

// CanLaunchFromWaiting returns true if we can launch from specified waiting slot.
func (s *State) CanLaunchFromWaiting(slotIndex int) bool {
	return s.Waiting.Get(slotIndex) != nil && len(s.Active) < s.Capacity
}

// LaunchFromQueue launches top shooter from specified queue.
// Returns true if successful.
func (s *State) LaunchFromQueue(queueIndex int) bool {
	if !s.CanLaunchFromQueue(queueIndex) {
		return false
	}

	shooter := s.Deck.PopFromQueue(queueIndex)
	if shooter == nil {
		return false
	}

	spawnIdx := s.Rail.SpawnIndex()
	active := ActiveShooter{
		Shooter:     *shooter,
		RailIndex:   spawnIdx,
		StartIndex:  spawnIdx,
		Dry:         false,
		LapProgress: 0,
	}
	s.Active = append(s.Active, active)
	return true
}

// LaunchFromWaiting launches shooter from specified waiting slot.
// Returns true if successful.
func (s *State) LaunchFromWaiting(slotIndex int) bool {
	if !s.CanLaunchFromWaiting(slotIndex) {
		return false
	}

	shooter := s.Waiting.Take(slotIndex)
	if shooter == nil {
		return false
	}

	spawnIdx := s.Rail.SpawnIndex()
	active := ActiveShooter{
		Shooter:     *shooter,
		RailIndex:   spawnIdx,
		StartIndex:  spawnIdx,
		Dry:         false,
		LapProgress: 0,
	}
	s.Active = append(s.Active, active)
	return true
}

// Legacy compatibility methods

// CanLaunch returns true if any queue has shooters and rail has capacity.
func (s *State) CanLaunch() bool {
	if len(s.Active) >= s.Capacity {
		return false
	}
	return !s.Deck.IsEmpty()
}

// CanRelaunchWaiting returns true if any waiting slot has a shooter.
func (s *State) CanRelaunchWaiting() bool {
	if len(s.Active) >= s.Capacity {
		return false
	}
	return !s.Waiting.IsEmpty()
}

// LaunchTop launches from first non-empty queue (for autoplayer compatibility).
func (s *State) LaunchTop() bool {
	for i := 0; i < s.Deck.NumQueues(); i++ {
		if s.LaunchFromQueue(i) {
			return true
		}
	}
	return false
}

// RelaunchWaiting launches from first occupied waiting slot.
func (s *State) RelaunchWaiting() bool {
	for i := 0; i < s.Waiting.Capacity; i++ {
		if s.LaunchFromWaiting(i) {
			return true
		}
	}
	return false
}

// TopDeckShooter returns the top shooter from first non-empty queue.
func (s *State) TopDeckShooter() *Shooter {
	for i := 0; i < s.Deck.NumQueues(); i++ {
		if top := s.Deck.TopOfQueue(i); top != nil {
			return top
		}
	}
	return nil
}

// IsWon returns true if the grid is empty (level cleared).
func (s *State) IsWon() bool {
	return s.Grid.IsEmpty()
}

// IsLost returns true if the game cannot be won:
// - Grid not empty AND
// - No active shooters AND
// - All deck queues empty AND
// - All waiting slots empty
func (s *State) IsLost() bool {
	if s.Grid.IsEmpty() {
		return false
	}
	return len(s.Active) == 0 && s.Deck.IsEmpty() && s.Waiting.IsEmpty()
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
	total := s.Deck.TotalAmmo()
	for _, active := range s.Active {
		total += active.Ammo
	}
	total += s.Waiting.TotalAmmo()
	return total
}

// Snapshot returns a hash representing the current state.
func (s *State) Snapshot() uint64 {
	h := fnv.New64a()

	// Grid state
	fmt.Fprintf(h, "G:%d;", s.Grid.Hash())

	// Deck queues
	fmt.Fprintf(h, "D:")
	for qi, q := range s.Deck.Queues {
		fmt.Fprintf(h, "Q%d:", qi)
		for _, d := range q {
			fmt.Fprintf(h, "%d:%d:%d,", d.ID, d.Color, d.Ammo)
		}
	}

	// Active
	fmt.Fprintf(h, ";A:")
	for _, a := range s.Active {
		fmt.Fprintf(h, "%d:%d:%d:%d:%v,", a.ID, a.RailIndex, a.Color, a.Ammo, a.Dry)
	}

	// Waiting slots
	fmt.Fprintf(h, ";W:")
	for i, w := range s.Waiting.Slots {
		if w != nil {
			fmt.Fprintf(h, "%d:%d:%d:%d,", i, w.ID, w.Color, w.Ammo)
		}
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
func (s *State) ActiveShootersAtIndex(railIndex int) []int {
	indices := make([]int, 0)
	for i, a := range s.Active {
		if a.RailIndex == railIndex {
			indices = append(indices, i)
		}
	}
	return indices
}
