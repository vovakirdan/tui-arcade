// Package config provides YAML-based game configuration loading and
// difficulty management for the arcade platform.
package config

// FlappyConfig contains all configuration for the Flappy Bird game.
type FlappyConfig struct {
	Physics    FlappyPhysics    `yaml:"physics"`
	Obstacles  FlappyObstacles  `yaml:"obstacles"`
	Player     FlappyPlayer     `yaml:"player"`
	Difficulty DifficultyConfig `yaml:"difficulty"`
}

// FlappyPhysics defines physics parameters for Flappy Bird.
type FlappyPhysics struct {
	Gravity      float64 `yaml:"gravity"`
	JumpImpulse  float64 `yaml:"jump_impulse"`
	MaxFallSpeed float64 `yaml:"max_fall_speed"`
	BaseSpeed    float64 `yaml:"base_speed"`
}

// FlappyObstacles defines obstacle parameters for Flappy Bird.
type FlappyObstacles struct {
	PipeWidth    int `yaml:"pipe_width"`
	PipeSpacing  int `yaml:"pipe_spacing"`
	MinGapSize   int `yaml:"min_gap_size"`
	MaxGapSize   int `yaml:"max_gap_size"`
	TopMargin    int `yaml:"top_margin"`
	BottomMargin int `yaml:"bottom_margin"`
}

// FlappyPlayer defines player parameters for Flappy Bird.
type FlappyPlayer struct {
	X      int `yaml:"x"`
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

// DinoConfig contains all configuration for the Dino Runner game.
type DinoConfig struct {
	Physics    DinoPhysics      `yaml:"physics"`
	Obstacles  DinoObstacles    `yaml:"obstacles"`
	Player     DinoPlayer       `yaml:"player"`
	Difficulty DifficultyConfig `yaml:"difficulty"`
}

// DinoPhysics defines physics parameters for Dino Runner.
type DinoPhysics struct {
	Gravity      float64 `yaml:"gravity"`
	JumpImpulse  float64 `yaml:"jump_impulse"`
	MaxFallSpeed float64 `yaml:"max_fall_speed"`
	BaseSpeed    float64 `yaml:"base_speed"`
}

// DinoObstacles defines obstacle parameters for Dino Runner.
type DinoObstacles struct {
	MinWidth   int `yaml:"min_width"`
	MaxWidth   int `yaml:"max_width"`
	MinHeight  int `yaml:"min_height"`
	MaxHeight  int `yaml:"max_height"`
	MinSpacing int `yaml:"min_spacing"`
	MaxSpacing int `yaml:"max_spacing"`
}

// DinoPlayer defines player parameters for Dino Runner.
type DinoPlayer struct {
	X            int `yaml:"x"`
	Width        int `yaml:"width"`
	Height       int `yaml:"height"`
	GroundOffset int `yaml:"ground_offset"`
}

// DifficultyConfig defines the difficulty progression system.
type DifficultyConfig struct {
	Enabled      bool              `yaml:"enabled"`
	InitialLevel float64           `yaml:"initial_level"` // 0.0 = easy, 1.0 = hard
	Progression  ProgressionConfig `yaml:"progression"`
	Scaling      ScalingConfig     `yaml:"scaling"`
}

// ProgressionConfig defines how difficulty increases over time.
type ProgressionConfig struct {
	Type  string `yaml:"type"`   // "score", "time", or "none"
	MaxAt int    `yaml:"max_at"` // Score/ticks at which max difficulty is reached
}

// ScalingConfig defines the magnitude of difficulty changes.
type ScalingConfig struct {
	SpeedMultiplier  float64 `yaml:"speed_multiplier"`  // Multiplier added to speed at max difficulty
	GapReduction     int     `yaml:"gap_reduction"`     // Gap size reduction at max difficulty
	SpacingReduction int     `yaml:"spacing_reduction"` // Spacing reduction at max difficulty
}

// DifficultyPreset represents a named difficulty level.
type DifficultyPreset string

const (
	DifficultyEasy   DifficultyPreset = "easy"
	DifficultyNormal DifficultyPreset = "normal"
	DifficultyHard   DifficultyPreset = "hard"
	DifficultyFixed  DifficultyPreset = "fixed"
)

// InitialLevelForPreset returns the initial_level for a difficulty preset.
func InitialLevelForPreset(preset DifficultyPreset) float64 {
	switch preset {
	case DifficultyEasy:
		return 0.0
	case DifficultyNormal:
		return 0.3
	case DifficultyHard:
		return 0.7
	default:
		return 0.0
	}
}

// IsFixedPreset returns true if the preset disables progression.
func IsFixedPreset(preset DifficultyPreset) bool {
	return preset == DifficultyFixed
}
