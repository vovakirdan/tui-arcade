package core

// Shooter represents a shooter in the deck or waiting slots.
// Shooters have a color (only shoots their color) and ammo count.
type Shooter struct {
	ID    int   // Unique identifier
	Color Color // Shooter only removes pixels of this color
	Ammo  int   // Number of shots remaining before disappearing
}

// Clone returns a copy of the shooter.
func (s Shooter) Clone() Shooter {
	return Shooter{ID: s.ID, Color: s.Color, Ammo: s.Ammo}
}

// ActiveShooter represents a shooter currently on the rail.
type ActiveShooter struct {
	Shooter
	RailIndex   int  // Current position on the rail
	StartIndex  int  // Where this shooter entered the rail (for lap detection)
	Dry         bool // When true, shooter won't fire until lap ends
	LapProgress int  // Steps taken since entering rail
}

// Clone returns a copy of the active shooter.
func (a ActiveShooter) Clone() ActiveShooter {
	return ActiveShooter{
		Shooter:     a.Shooter.Clone(),
		RailIndex:   a.RailIndex,
		StartIndex:  a.StartIndex,
		Dry:         a.Dry,
		LapProgress: a.LapProgress,
	}
}

// HasCompletedLap returns true if the shooter has done a full loop.
func (a ActiveShooter) HasCompletedLap(railLen int) bool {
	return a.LapProgress >= railLen
}

// CloneDeck creates a deep copy of a shooter slice.
func CloneDeck(deck []Shooter) []Shooter {
	if deck == nil {
		return nil
	}
	clone := make([]Shooter, len(deck))
	for i, s := range deck {
		clone[i] = s.Clone()
	}
	return clone
}

// CloneActive creates a deep copy of an active shooter slice.
func CloneActive(active []ActiveShooter) []ActiveShooter {
	if active == nil {
		return nil
	}
	clone := make([]ActiveShooter, len(active))
	for i, a := range active {
		clone[i] = a.Clone()
	}
	return clone
}

// DeckTotalAmmo returns total ammo across all shooters in deck.
func DeckTotalAmmo(deck []Shooter) int {
	total := 0
	for _, s := range deck {
		total += s.Ammo
	}
	return total
}

// DeckAmmoByColor returns ammo totals by color for a deck.
func DeckAmmoByColor(deck []Shooter) map[Color]int {
	counts := make(map[Color]int)
	for _, s := range deck {
		counts[s.Color] += s.Ammo
	}
	return counts
}

// DeckColors returns unique colors in deck (sorted).
func DeckColors(deck []Shooter) []Color {
	seen := make(map[Color]bool)
	for _, s := range deck {
		seen[s.Color] = true
	}
	colors := make([]Color, 0, len(seen))
	for c := range seen {
		colors = append(colors, c)
	}
	// Sort for determinism
	for i := 0; i < len(colors)-1; i++ {
		for j := i + 1; j < len(colors); j++ {
			if colors[i] > colors[j] {
				colors[i], colors[j] = colors[j], colors[i]
			}
		}
	}
	return colors
}
