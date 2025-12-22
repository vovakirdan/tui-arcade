package core

import (
	"errors"
	"sort"
)

// GenParams configures the deck generator behavior.
type GenParams struct {
	Capacity  int    // Rail capacity (max active shooters)
	NumQueues int    // Number of deck queues (default 2)
	Seed      uint64 // RNG seed for deterministic variety

	// Ammo constraints
	MaxAmmoPerShooter int // Cap per shooter (e.g., 40)
	MinAmmoPerShooter int // Minimum ammo per shooter (e.g., 1)

	// Style constraints for "beautiful" queue generation
	TargetLargeAmmo    int     // Target ammo for "large" shooters (e.g., 40)
	TargetMediumAmmo   int     // Target ammo for "medium" shooters (e.g., 20)
	TargetSmallAmmo    int     // Target ammo for "small" shooters (e.g., 10)
	LargeProbability   float64 // Probability of choosing large ammo (0-1)
	MediumProbability  float64 // Probability of choosing medium ammo (0-1)
	PreferAlternating  bool    // Try to alternate colors when possible
	MaxConsecutiveSame int     // Max consecutive shooters of same color (0 = no limit)

	// Generation limits
	MaxShooters int // Maximum total shooters in deck
	MaxAttempts int // Retry limit for validation
	MaxSimSteps int // Max simulation steps for validation
}

// DefaultGenParams returns sensible defaults for deck generation.
func DefaultGenParams() GenParams {
	return GenParams{
		Capacity:           5,
		NumQueues:          2,
		Seed:               0,
		MaxAmmoPerShooter:  40,
		MinAmmoPerShooter:  1,
		TargetLargeAmmo:    40,
		TargetMediumAmmo:   20,
		TargetSmallAmmo:    10,
		LargeProbability:   0.4,
		MediumProbability:  0.3,
		PreferAlternating:  true,
		MaxConsecutiveSame: 3,
		MaxShooters:        100,
		MaxAttempts:        10,
		MaxSimSteps:        100000,
	}
}

// SimpleRNG is a deterministic pseudo-random number generator (xorshift64).
type SimpleRNG struct {
	state uint64
}

// NewRNG creates a new RNG with the given seed.
func NewRNG(seed uint64) *SimpleRNG {
	if seed == 0 {
		seed = 88172645463325252 // Default seed
	}
	return &SimpleRNG{state: seed}
}

// Next returns the next random uint64.
func (r *SimpleRNG) Next() uint64 {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return r.state
}

// Float returns a random float64 in [0, 1).
func (r *SimpleRNG) Float() float64 {
	return float64(r.Next()&0x7FFFFFFFFFFFFFFF) / float64(0x8000000000000000)
}

// Intn returns a random int in [0, n).
func (r *SimpleRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.Next() % uint64(n))
}

// GenerateDeckForGrid generates a solvable deck for the given grid.
// The generated deck guarantees:
//   - Only colors present in grid
//   - Ammo per color equals exact pixel count
//   - Level is solvable under deterministic autoplayer
//
// Uses constructive generation with style constraints.
func GenerateDeckForGrid(g *Grid, rail Rail, p GenParams) ([]Shooter, error) {
	if g.IsEmpty() {
		return []Shooter{}, nil
	}

	rng := NewRNG(p.Seed)

	// Get color distribution from grid
	colorCounts := g.CountByColor()
	if len(colorCounts) == 0 {
		return []Shooter{}, nil
	}

	// Track remaining ammo needed per color
	remaining := make(map[Color]int)
	for c, count := range colorCounts {
		remaining[c] = count
	}

	// Build the deck
	deck := make([]Shooter, 0)
	shooterID := 0
	lastColor := Color(255) // Invalid, forces first pick to be "alternating"

	for !allZero(remaining) {
		if len(deck) >= p.MaxShooters {
			return nil, errors.New("exceeded max shooters before clearing all pixels")
		}

		// Pick a color to use next
		color := pickNextColor(remaining, lastColor, p, rng)
		if remaining[color] <= 0 {
			// Should not happen, but safety check
			continue
		}

		// Determine ammo amount based on style constraints
		ammoNeeded := remaining[color]
		ammo := determineAmmo(ammoNeeded, p, rng)

		// Create shooter
		deck = append(deck, Shooter{
			ID:    shooterID,
			Color: color,
			Ammo:  ammo,
		})
		shooterID++

		remaining[color] -= ammo
		lastColor = color
	}

	// Validate that the deck is actually solvable
	if err := validateDeckSolvability(g, rail, deck, p); err != nil {
		// Try again with different seed if attempts remain
		if p.MaxAttempts > 1 {
			newParams := p
			newParams.Seed = p.Seed + 12345
			newParams.MaxAttempts = p.MaxAttempts - 1
			return GenerateDeckForGrid(g, rail, newParams)
		}
		return nil, err
	}

	return deck, nil
}

// pickNextColor selects the next color to use for a shooter.
func pickNextColor(remaining map[Color]int, lastColor Color, p GenParams, rng *SimpleRNG) Color {
	// Get available colors (those with remaining > 0)
	available := make([]Color, 0)
	for c, count := range remaining {
		if count > 0 {
			available = append(available, c)
		}
	}

	if len(available) == 0 {
		return lastColor // Should not happen
	}

	// Sort for determinism
	sort.Slice(available, func(i, j int) bool {
		return available[i] < available[j]
	})

	// If preferring alternation, try to pick a different color
	if p.PreferAlternating && len(available) > 1 {
		// Filter out last color if MaxConsecutiveSame would be violated
		filtered := make([]Color, 0)
		for _, c := range available {
			if c != lastColor {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) > 0 {
			available = filtered
		}
	}

	// Pick randomly from available (weighted by remaining count)
	totalWeight := 0
	for _, c := range available {
		totalWeight += remaining[c]
	}

	if totalWeight == 0 {
		return available[0]
	}

	pick := rng.Intn(totalWeight)
	cumulative := 0
	for _, c := range available {
		cumulative += remaining[c]
		if pick < cumulative {
			return c
		}
	}

	return available[len(available)-1]
}

// determineAmmo decides how much ammo to give a shooter based on style constraints.
func determineAmmo(needed int, p GenParams, rng *SimpleRNG) int {
	if needed <= 0 {
		return 0
	}

	// Roll to determine target size category
	roll := rng.Float()
	var targetAmmo int

	if roll < p.LargeProbability {
		targetAmmo = p.TargetLargeAmmo
	} else if roll < p.LargeProbability+p.MediumProbability {
		targetAmmo = p.TargetMediumAmmo
	} else {
		targetAmmo = p.TargetSmallAmmo
	}

	// Apply min/max constraints
	ammo := targetAmmo
	if ammo > p.MaxAmmoPerShooter {
		ammo = p.MaxAmmoPerShooter
	}
	if ammo < p.MinAmmoPerShooter {
		ammo = p.MinAmmoPerShooter
	}

	// Don't exceed what's actually needed
	if ammo > needed {
		ammo = needed
	}

	return ammo
}

// validateDeckSolvability runs simulation to verify deck clears the grid.
func validateDeckSolvability(g *Grid, rail Rail, deck []Shooter, p GenParams) error {
	numQueues := p.NumQueues
	if numQueues < 2 {
		numQueues = 2
	}
	state := NewStateWithQueues(g, deck, p.Capacity, numQueues)

	steps, cleared := state.RunUntilIdle(p.MaxSimSteps)
	if !cleared {
		return errors.New("deck does not clear grid under autoplayer")
	}

	_ = steps // Could log this for debugging
	return nil
}

// allZero returns true if all values in the map are zero.
func allZero(m map[Color]int) bool {
	for _, v := range m {
		if v > 0 {
			return false
		}
	}
	return true
}

// GenerateSimpleDeck generates a simple deck with one shooter per pixel.
// Each shooter has ammo=1. This always works but creates many shooters.
func GenerateSimpleDeck(g *Grid) []Shooter {
	deck := make([]Shooter, 0)
	id := 0

	// Process pixels in deterministic order
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			cell := g.Get(C(x, y))
			if cell.Filled {
				deck = append(deck, Shooter{
					ID:    id,
					Color: cell.Color,
					Ammo:  1,
				})
				id++
			}
		}
	}

	return deck
}

// GenerateBalancedDeck creates a deck with specified ammo per shooter for each color.
// targetAmmo is the desired ammo per shooter; actual values may vary to match totals.
func GenerateBalancedDeck(g *Grid, targetAmmo int, seed uint64) []Shooter {
	if targetAmmo <= 0 {
		targetAmmo = 1
	}

	colorCounts := g.CountByColor()
	colors := make([]Color, 0, len(colorCounts))
	for c := range colorCounts {
		colors = append(colors, c)
	}
	sort.Slice(colors, func(i, j int) bool {
		return colors[i] < colors[j]
	})

	rng := NewRNG(seed)
	deck := make([]Shooter, 0)
	id := 0

	// Create shooters for each color
	for _, color := range colors {
		remaining := colorCounts[color]
		for remaining > 0 {
			ammo := targetAmmo
			if ammo > remaining {
				ammo = remaining
			}
			// Add some variety
			if remaining > targetAmmo && rng.Float() < 0.3 {
				// Occasionally use smaller ammo for variety
				ammo = max(1, targetAmmo/2)
				if ammo > remaining {
					ammo = remaining
				}
			}

			deck = append(deck, Shooter{
				ID:    id,
				Color: color,
				Ammo:  ammo,
			})
			id++
			remaining -= ammo
		}
	}

	// Shuffle deck for variety (deterministic with seed)
	for i := len(deck) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}

	return deck
}

// GenerateStyledDeck creates a deck matching specific style patterns.
// Pattern example: "40/40/10/10/..." creates shooters with those ammo values.
// Colors are distributed to match pixel counts.
func GenerateStyledDeck(g *Grid, pattern []int, seed uint64) ([]Shooter, error) {
	colorCounts := g.CountByColor()
	totalPixels := 0
	for _, c := range colorCounts {
		totalPixels += c
	}

	// Sum of pattern must equal or exceed total pixels
	patternSum := 0
	for _, a := range pattern {
		patternSum += a
	}
	if patternSum < totalPixels {
		return nil, errors.New("pattern sum less than total pixels")
	}

	rng := NewRNG(seed)
	deck := make([]Shooter, 0)

	// Track remaining per color
	remaining := make(map[Color]int)
	for c, count := range colorCounts {
		remaining[c] = count
	}

	// Build colors list for random selection
	colors := make([]Color, 0, len(colorCounts))
	for c := range colorCounts {
		colors = append(colors, c)
	}
	sort.Slice(colors, func(i, j int) bool {
		return colors[i] < colors[j]
	})

	id := 0
	for _, targetAmmo := range pattern {
		if allZero(remaining) {
			break
		}

		// Pick a color that needs this much ammo (or close to it)
		color := pickColorForAmmo(remaining, targetAmmo, colors, rng)
		if remaining[color] <= 0 {
			continue
		}

		ammo := targetAmmo
		if ammo > remaining[color] {
			ammo = remaining[color]
		}

		deck = append(deck, Shooter{
			ID:    id,
			Color: color,
			Ammo:  ammo,
		})
		id++
		remaining[color] -= ammo
	}

	// Handle any remaining pixels
	for !allZero(remaining) {
		for _, c := range colors {
			if remaining[c] > 0 {
				ammo := remaining[c]
				if ammo > 40 {
					ammo = 40
				}
				deck = append(deck, Shooter{
					ID:    id,
					Color: c,
					Ammo:  ammo,
				})
				id++
				remaining[c] -= ammo
			}
		}
	}

	return deck, nil
}

// pickColorForAmmo selects a color that ideally needs close to the target ammo.
func pickColorForAmmo(remaining map[Color]int, targetAmmo int, colors []Color, rng *SimpleRNG) Color {
	// Find colors with enough remaining
	available := make([]Color, 0)
	for _, c := range colors {
		if remaining[c] > 0 {
			available = append(available, c)
		}
	}

	if len(available) == 0 {
		return colors[0]
	}

	// Prefer colors where remaining is close to targetAmmo
	best := available[0]
	bestDiff := abs(remaining[best] - targetAmmo)

	for _, c := range available {
		diff := abs(remaining[c] - targetAmmo)
		if diff < bestDiff {
			best = c
			bestDiff = diff
		}
	}

	// Sometimes pick randomly for variety
	if rng.Float() < 0.3 {
		return available[rng.Intn(len(available))]
	}

	return best
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
