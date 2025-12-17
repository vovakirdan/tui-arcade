package config

import "math"

// DifficultyManager calculates dynamic game parameters based on score/time.
type DifficultyManager struct {
	cfg          DifficultyConfig
	initialLevel float64
}

// NewDifficultyManager creates a new difficulty manager.
func NewDifficultyManager(cfg DifficultyConfig) *DifficultyManager {
	return &DifficultyManager{
		cfg:          cfg,
		initialLevel: cfg.InitialLevel,
	}
}

// SetInitialLevel overrides the initial difficulty level (0.0 to 1.0).
func (d *DifficultyManager) SetInitialLevel(level float64) {
	d.initialLevel = clampF(level, 0.0, 1.0)
}

// SetEnabled enables or disables difficulty progression.
func (d *DifficultyManager) SetEnabled(enabled bool) {
	d.cfg.Enabled = enabled
}

// IsEnabled returns whether difficulty progression is active.
func (d *DifficultyManager) IsEnabled() bool {
	return d.cfg.Enabled && d.cfg.Progression.Type != "none"
}

// Level returns the current difficulty level (0.0 to 1.0) based on score/ticks.
func (d *DifficultyManager) Level(score int, ticks int) float64 {
	if !d.cfg.Enabled || d.cfg.Progression.Type == "none" {
		return d.initialLevel
	}

	var progress float64
	maxAt := float64(d.cfg.Progression.MaxAt)
	if maxAt <= 0 {
		maxAt = 1 // Prevent division by zero
	}

	switch d.cfg.Progression.Type {
	case "score":
		progress = float64(score) / maxAt
	case "time":
		progress = float64(ticks) / maxAt
	default:
		return d.initialLevel
	}

	// Clamp progress to [0, 1]
	progress = clampF(progress, 0.0, 1.0)

	// Interpolate from initial level to 1.0
	return d.initialLevel + progress*(1.0-d.initialLevel)
}

// Speed returns the current speed multiplier based on difficulty level.
func (d *DifficultyManager) Speed(baseSpeed float64, score int, ticks int) float64 {
	level := d.Level(score, ticks)
	// Speed increases from base to base * (1 + speedMultiplier)
	return baseSpeed * (1.0 + level*d.cfg.Scaling.SpeedMultiplier)
}

// GapSize returns the current gap size based on difficulty level.
func (d *DifficultyManager) GapSize(baseGap int, score int, ticks int) int {
	level := d.Level(score, ticks)
	// Gap decreases as difficulty increases
	reduction := int(level * float64(d.cfg.Scaling.GapReduction))
	result := baseGap - reduction
	if result < 4 { // Minimum playable gap
		result = 4
	}
	return result
}

// Spacing returns the current obstacle spacing based on difficulty level.
func (d *DifficultyManager) Spacing(baseSpacing int, score int, ticks int) int {
	level := d.Level(score, ticks)
	// Spacing decreases as difficulty increases
	reduction := int(level * float64(d.cfg.Scaling.SpacingReduction))
	result := baseSpacing - reduction
	if result < 15 { // Minimum playable spacing
		result = 15
	}
	return result
}

// clampF restricts a float64 to [min, max].
func clampF(val, min, max float64) float64 {
	return math.Max(min, math.Min(max, val))
}
