package core

import (
	"fmt"
	"sort"
)

// ValidationError contains details about validation failure.
type ValidationError struct {
	Code    string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ValidateDeck performs comprehensive validation of a deck against a grid.
// Checks:
//   - All shooter colors exist in grid
//   - Ammo per color equals pixel count
//   - Deck is solvable under autoplayer
func ValidateDeck(g *Grid, deck []Shooter, capacity int, maxSteps int) error {
	// Check 1: Colors present
	if err := validateColorsPresent(g, deck); err != nil {
		return err
	}

	// Check 2: Ammo sums
	if err := validateAmmoSums(g, deck); err != nil {
		return err
	}

	// Check 3: Solvability
	if err := validateSolvability(g, deck, capacity, maxSteps); err != nil {
		return err
	}

	return nil
}

// validateColorsPresent checks that all deck colors exist in grid.
func validateColorsPresent(g *Grid, deck []Shooter) error {
	gridColors := make(map[Color]bool)
	for _, c := range g.ColorsPresent() {
		gridColors[c] = true
	}

	for _, s := range deck {
		if !gridColors[s.Color] {
			return ValidationError{
				Code:    "INVALID_COLOR",
				Message: fmt.Sprintf("shooter has color %s not present in grid", s.Color),
			}
		}
	}

	return nil
}

// validateAmmoSums checks that total ammo per color equals pixel count.
func validateAmmoSums(g *Grid, deck []Shooter) error {
	pixelCounts := g.CountByColor()
	ammoCounts := DeckAmmoByColor(deck)

	// Get all colors for deterministic iteration
	allColors := make([]Color, 0)
	for c := range pixelCounts {
		allColors = append(allColors, c)
	}
	for c := range ammoCounts {
		found := false
		for _, existing := range allColors {
			if existing == c {
				found = true
				break
			}
		}
		if !found {
			allColors = append(allColors, c)
		}
	}
	sort.Slice(allColors, func(i, j int) bool {
		return allColors[i] < allColors[j]
	})

	for _, c := range allColors {
		pixels := pixelCounts[c]
		ammo := ammoCounts[c]

		if ammo < pixels {
			return ValidationError{
				Code: "INSUFFICIENT_AMMO",
				Message: fmt.Sprintf("color %s: %d ammo < %d pixels",
					c, ammo, pixels),
			}
		}
		if ammo > pixels {
			return ValidationError{
				Code: "EXCESS_AMMO",
				Message: fmt.Sprintf("color %s: %d ammo > %d pixels",
					c, ammo, pixels),
			}
		}
	}

	return nil
}

// validateSolvability runs simulation to verify deck clears grid.
func validateSolvability(g *Grid, deck []Shooter, capacity int, maxSteps int) error {
	state := NewState(g, deck, capacity)

	_, cleared := state.RunUntilIdle(maxSteps)
	if !cleared {
		remaining := state.Grid.FilledCount()
		return ValidationError{
			Code: "NOT_SOLVABLE",
			Message: fmt.Sprintf("grid not cleared after simulation, %d pixels remain",
				remaining),
		}
	}

	return nil
}

// QuickValidateDeck performs fast validation without simulation.
// Only checks colors and ammo sums.
func QuickValidateDeck(g *Grid, deck []Shooter) error {
	if err := validateColorsPresent(g, deck); err != nil {
		return err
	}
	if err := validateAmmoSums(g, deck); err != nil {
		return err
	}
	return nil
}

// DeckStats returns statistics about a deck.
type DeckStats struct {
	TotalShooters   int
	TotalAmmo       int
	UniqueColors    int
	AmmoByColor     map[Color]int
	ShootersByColor map[Color]int
	MinAmmo         int
	MaxAmmo         int
	AvgAmmo         float64
}

// ComputeDeckStats analyzes a deck and returns statistics.
func ComputeDeckStats(deck []Shooter) DeckStats {
	stats := DeckStats{
		TotalShooters:   len(deck),
		AmmoByColor:     make(map[Color]int),
		ShootersByColor: make(map[Color]int),
		MinAmmo:         -1,
		MaxAmmo:         0,
	}

	if len(deck) == 0 {
		return stats
	}

	totalAmmo := 0
	for _, s := range deck {
		totalAmmo += s.Ammo
		stats.AmmoByColor[s.Color] += s.Ammo
		stats.ShootersByColor[s.Color]++

		if stats.MinAmmo < 0 || s.Ammo < stats.MinAmmo {
			stats.MinAmmo = s.Ammo
		}
		if s.Ammo > stats.MaxAmmo {
			stats.MaxAmmo = s.Ammo
		}
	}

	stats.TotalAmmo = totalAmmo
	stats.UniqueColors = len(stats.AmmoByColor)
	stats.AvgAmmo = float64(totalAmmo) / float64(len(deck))

	return stats
}

// GridStats returns statistics about a grid.
type GridStats struct {
	Width       int
	Height      int
	TotalCells  int
	FilledCells int
	EmptyCells  int
	ColorCounts map[Color]int
	FillRatio   float64
}

// ComputeGridStats analyzes a grid and returns statistics.
func ComputeGridStats(g *Grid) GridStats {
	filled := g.FilledCount()
	total := g.W * g.H

	return GridStats{
		Width:       g.W,
		Height:      g.H,
		TotalCells:  total,
		FilledCells: filled,
		EmptyCells:  total - filled,
		ColorCounts: g.CountByColor(),
		FillRatio:   float64(filled) / float64(total),
	}
}

// CompareDecks checks if two decks are equivalent (same ammo totals per color).
func CompareDecks(a, b []Shooter) bool {
	ammoA := DeckAmmoByColor(a)
	ammoB := DeckAmmoByColor(b)

	if len(ammoA) != len(ammoB) {
		return false
	}

	for c, countA := range ammoA {
		if countB, ok := ammoB[c]; !ok || countA != countB {
			return false
		}
	}

	return true
}
