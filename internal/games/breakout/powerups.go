package breakout

// PickupType represents different types of power-up pickups.
type PickupType int

const (
	PickupWiden     PickupType = iota // Widen paddle
	PickupShrink                      // Shrink paddle
	PickupMultiball                   // Spawn extra balls
	PickupSticky                      // Sticky paddle
	PickupSpeedUp                     // Speed up ball
	PickupSlowDown                    // Slow down ball
	PickupExtraLife                   // Extra life
	PickupCount                       // Sentinel for counting types
)

// PickupGlyph returns the display character for a pickup type.
func (p PickupType) Glyph() rune {
	switch p {
	case PickupWiden:
		return 'W'
	case PickupShrink:
		return 'S'
	case PickupMultiball:
		return 'M'
	case PickupSticky:
		return 'T'
	case PickupSpeedUp:
		return '+'
	case PickupSlowDown:
		return '-'
	case PickupExtraLife:
		return '♥'
	default:
		return '?'
	}
}

// String returns the name of the pickup type.
func (p PickupType) String() string {
	switch p {
	case PickupWiden:
		return "Widen"
	case PickupShrink:
		return "Shrink"
	case PickupMultiball:
		return "Multi"
	case PickupSticky:
		return "Sticky"
	case PickupSpeedUp:
		return "Fast"
	case PickupSlowDown:
		return "Slow"
	case PickupExtraLife:
		return "Life"
	default:
		return "?"
	}
}

// Pickup represents a falling power-up item.
type Pickup struct {
	Type   PickupType
	X      Fixed // Center X position
	Y      Fixed // Center Y position
	VY     Fixed // Fall speed (positive = down)
	Active bool  // Whether pickup is still in play
}

// CellX returns pickup X in cell coordinates.
func (p *Pickup) CellX() int {
	return p.X.ToCell()
}

// CellY returns pickup Y in cell coordinates.
func (p *Pickup) CellY() int {
	return p.Y.ToCell()
}

// Move updates pickup position.
func (p *Pickup) Move() {
	p.Y = p.Y.Add(p.VY)
}

// EffectType represents active effects on the game.
type EffectType int

const (
	EffectWiden    EffectType = iota // Paddle is widened
	EffectShrink                     // Paddle is shrunk
	EffectSticky                     // Paddle is sticky
	EffectSpeedUp                    // Ball speed increased
	EffectSlowDown                   // Ball speed decreased
	EffectCount                      // Sentinel for counting types
)

// String returns the short name for effect display.
func (e EffectType) String() string {
	switch e {
	case EffectWiden:
		return "W"
	case EffectShrink:
		return "S"
	case EffectSticky:
		return "T"
	case EffectSpeedUp:
		return "+"
	case EffectSlowDown:
		return "-"
	default:
		return "?"
	}
}

// Effect represents an active timed effect.
type Effect struct {
	Type      EffectType
	UntilTick int // Tick at which effect expires
}

// TicksRemaining returns how many ticks until effect expires.
func (e *Effect) TicksRemaining(currentTick int) int {
	remaining := e.UntilTick - currentTick
	if remaining < 0 {
		return 0
	}
	return remaining
}

// PowerUpConfig holds configuration for power-up spawning and effects.
type PowerUpConfig struct {
	// Spawn settings
	SpawnChance int // Percentage chance to spawn on brick destroy (0-100)

	// Spawn weights (relative, higher = more common)
	WeightWiden     int
	WeightShrink    int
	WeightMultiball int
	WeightSticky    int
	WeightSpeedUp   int
	WeightSlowDown  int
	WeightExtraLife int

	// Effect durations in ticks (60 ticks = 1 second at 60 FPS)
	DurationWiden    int
	DurationShrink   int
	DurationSticky   int
	DurationSpeedUp  int
	DurationSlowDown int

	// Pickup physics
	FallSpeed int // Fixed-point fall speed

	// Effect parameters
	WidenAmount     int // Cells to add to paddle width
	ShrinkAmount    int // Cells to remove from paddle width
	MinPaddleWidth  int // Minimum paddle width
	MaxPaddleWidth  int // Maximum paddle width
	SpeedMultiplier int // Speed change in fixed-point percent (e.g., 150 = 1.5x)
	MinBallSpeed    int // Minimum ball speed (fixed-point)
	MaxBallSpeed    int // Maximum ball speed (fixed-point)
	MultiballCount  int // Number of extra balls to spawn
}

// DefaultPowerUpConfig returns default power-up configuration.
func DefaultPowerUpConfig() PowerUpConfig {
	return PowerUpConfig{
		// 18% chance to spawn
		SpawnChance: 18,

		// Weights (positive pickups more common)
		WeightWiden:     25,
		WeightShrink:    10, // Less common (negative)
		WeightMultiball: 20,
		WeightSticky:    15,
		WeightSpeedUp:   10, // Less common (can be tricky)
		WeightSlowDown:  15,
		WeightExtraLife: 5, // Rare

		// Durations at 60 FPS
		DurationWiden:    720, // 12 seconds
		DurationShrink:   720, // 12 seconds
		DurationSticky:   600, // 10 seconds
		DurationSpeedUp:  480, // 8 seconds
		DurationSlowDown: 480, // 8 seconds

		// Pickup fall speed
		FallSpeed: 200, // 0.2 cells per tick

		// Effect parameters
		WidenAmount:     4,
		ShrinkAmount:    3,
		MinPaddleWidth:  4,
		MaxPaddleWidth:  16,
		SpeedMultiplier: 150, // 1.5x for speed up, 0.67x for slow down
		MinBallSpeed:    100, // 0.1 cells per tick minimum
		MaxBallSpeed:    800, // 0.8 cells per tick maximum
		MultiballCount:  2,   // Spawn 2 extra balls
	}
}

// PowerUpManager handles pickup spawning, falling, collection, and effects.
type PowerUpManager struct {
	Config  PowerUpConfig
	Pickups []*Pickup  // Active falling pickups
	Effects []*Effect  // Active effects
	RNG     *SimpleRNG // Deterministic RNG
}

// NewPowerUpManager creates a new power-up manager with given seed.
func NewPowerUpManager(seed int64, cfg PowerUpConfig) *PowerUpManager {
	return &PowerUpManager{
		Config:  cfg,
		Pickups: make([]*Pickup, 0),
		Effects: make([]*Effect, 0),
		RNG:     NewSimpleRNG(seed),
	}
}

// Reset clears all pickups and effects.
func (pm *PowerUpManager) Reset(seed int64) {
	pm.Pickups = pm.Pickups[:0]
	pm.Effects = pm.Effects[:0]
	pm.RNG = NewSimpleRNG(seed)
}

// TrySpawnPickup attempts to spawn a pickup at the given brick position.
// Returns true if a pickup was spawned.
func (pm *PowerUpManager) TrySpawnPickup(brickCenterX, brickCenterY int) bool {
	// Roll for spawn chance
	roll := pm.RNG.Intn(100)
	if roll >= pm.Config.SpawnChance {
		return false
	}

	// Roll for pickup type based on weights
	pickupType := pm.rollPickupType()

	pickup := &Pickup{
		Type:   pickupType,
		X:      ToFixed(brickCenterX),
		Y:      ToFixed(brickCenterY),
		VY:     Fixed(pm.Config.FallSpeed),
		Active: true,
	}

	pm.Pickups = append(pm.Pickups, pickup)
	return true
}

// rollPickupType selects a random pickup type based on weights.
func (pm *PowerUpManager) rollPickupType() PickupType {
	totalWeight := pm.Config.WeightWiden + pm.Config.WeightShrink +
		pm.Config.WeightMultiball + pm.Config.WeightSticky +
		pm.Config.WeightSpeedUp + pm.Config.WeightSlowDown +
		pm.Config.WeightExtraLife

	if totalWeight <= 0 {
		return PickupWiden
	}

	roll := pm.RNG.Intn(totalWeight)
	cumulative := 0

	weights := []struct {
		Type   PickupType
		Weight int
	}{
		{PickupWiden, pm.Config.WeightWiden},
		{PickupShrink, pm.Config.WeightShrink},
		{PickupMultiball, pm.Config.WeightMultiball},
		{PickupSticky, pm.Config.WeightSticky},
		{PickupSpeedUp, pm.Config.WeightSpeedUp},
		{PickupSlowDown, pm.Config.WeightSlowDown},
		{PickupExtraLife, pm.Config.WeightExtraLife},
	}

	for _, w := range weights {
		cumulative += w.Weight
		if roll < cumulative {
			return w.Type
		}
	}

	return PickupWiden
}

// Update moves all pickups and removes out-of-bounds ones.
func (pm *PowerUpManager) Update(screenH int) {
	maxY := ToFixed(screenH + 1)

	// Filter active pickups
	active := pm.Pickups[:0]
	for _, p := range pm.Pickups {
		if !p.Active {
			continue
		}
		p.Move()
		if p.Y < maxY {
			active = append(active, p)
		}
	}
	pm.Pickups = active
}

// CheckPaddleCollision checks if any pickup hits the paddle.
// Returns the pickup type if collected, or -1 if none.
func (pm *PowerUpManager) CheckPaddleCollision(paddle *Paddle) PickupType {
	paddleLeft := paddle.Left()
	paddleRight := paddle.Right()
	paddleY := paddle.Y

	for _, p := range pm.Pickups {
		if !p.Active {
			continue
		}

		// Check Y proximity (pickup at paddle level)
		pickupY := p.CellY()
		if pickupY != paddleY && pickupY != paddleY-1 {
			continue
		}

		// Check X overlap
		if p.X >= paddleLeft && p.X <= paddleRight {
			p.Active = false
			return p.Type
		}
	}

	return PickupType(-1)
}

// AddEffect adds or extends an effect.
func (pm *PowerUpManager) AddEffect(effectType EffectType, currentTick, duration int) {
	// Check if effect already exists
	for _, e := range pm.Effects {
		if e.Type == effectType {
			// Extend duration
			e.UntilTick = currentTick + duration
			return
		}
	}

	// Add new effect
	pm.Effects = append(pm.Effects, &Effect{
		Type:      effectType,
		UntilTick: currentTick + duration,
	})
}

// RemoveEffect removes an effect by type.
func (pm *PowerUpManager) RemoveEffect(effectType EffectType) {
	for i, e := range pm.Effects {
		if e.Type == effectType {
			pm.Effects = append(pm.Effects[:i], pm.Effects[i+1:]...)
			return
		}
	}
}

// ExpireEffects removes effects that have expired.
func (pm *PowerUpManager) ExpireEffects(currentTick int) []EffectType {
	expired := make([]EffectType, 0)
	active := pm.Effects[:0]

	for _, e := range pm.Effects {
		if e.UntilTick <= currentTick {
			expired = append(expired, e.Type)
		} else {
			active = append(active, e)
		}
	}

	pm.Effects = active
	return expired
}

// HasEffect returns true if the given effect is active.
func (pm *PowerUpManager) HasEffect(effectType EffectType) bool {
	for _, e := range pm.Effects {
		if e.Type == effectType {
			return true
		}
	}
	return false
}

// GetEffectRemaining returns ticks remaining for an effect, or 0 if not active.
func (pm *PowerUpManager) GetEffectRemaining(effectType EffectType, currentTick int) int {
	for _, e := range pm.Effects {
		if e.Type == effectType {
			return e.TicksRemaining(currentTick)
		}
	}
	return 0
}

// SimpleRNG is a deterministic pseudo-random number generator.
// Uses a simple LCG (Linear Congruential Generator).
type SimpleRNG struct {
	state uint64
}

// NewSimpleRNG creates a new RNG with the given seed.
func NewSimpleRNG(seed int64) *SimpleRNG {
	s := uint64(seed) //#nosec G115 -- intentional conversion for RNG seeding
	if s == 0 {
		s = 1
	}
	return &SimpleRNG{state: s}
}

// Next generates the next random uint64.
func (r *SimpleRNG) Next() uint64 {
	// LCG parameters (same as MINSTD)
	r.state = r.state*6364136223846793005 + 1442695040888963407
	return r.state
}

// Intn returns a random int in [0, n).
func (r *SimpleRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.Next() % uint64(n)) //#nosec G115 -- n is always positive
}

// Float64 returns a random float64 in [0, 1).
func (r *SimpleRNG) Float64() float64 {
	return float64(r.Next()) / float64(1<<64)
}
