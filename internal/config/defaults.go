package config

import (
	_ "embed"
)

//go:embed defaults/flappy.yaml
var defaultFlappyYAML []byte

//go:embed defaults/dino.yaml
var defaultDinoYAML []byte

//go:embed defaults/pong.yaml
var defaultPongYAML []byte

// DefaultFlappyConfig returns the default Flappy Bird configuration.
func DefaultFlappyConfig() FlappyConfig {
	return FlappyConfig{
		Physics: FlappyPhysics{
			Gravity:      0.25,
			JumpImpulse:  -1.8,
			MaxFallSpeed: 3.0,
			BaseSpeed:    0.8,
		},
		Obstacles: FlappyObstacles{
			PipeWidth:    5,
			PipeSpacing:  40,
			MinGapSize:   8,
			MaxGapSize:   12,
			TopMargin:    3,
			BottomMargin: 3,
		},
		Player: FlappyPlayer{
			X:      10,
			Width:  2,
			Height: 2,
		},
		Difficulty: DifficultyConfig{
			Enabled:      true,
			InitialLevel: 0.0,
			Progression: ProgressionConfig{
				Type:  "score",
				MaxAt: 50,
			},
			Scaling: ScalingConfig{
				SpeedMultiplier:  1.0,
				GapReduction:     4,
				SpacingReduction: 15,
			},
		},
	}
}

// DefaultDinoConfig returns the default Dino Runner configuration.
func DefaultDinoConfig() DinoConfig {
	return DinoConfig{
		Physics: DinoPhysics{
			Gravity:      0.3,
			JumpImpulse:  -2.5,
			MaxFallSpeed: 4.0,
			BaseSpeed:    0.5,
		},
		Obstacles: DinoObstacles{
			MinWidth:   1,
			MaxWidth:   3,
			MinHeight:  2,
			MaxHeight:  4,
			MinSpacing: 30,
			MaxSpacing: 50,
		},
		Player: DinoPlayer{
			X:            8,
			Width:        3,
			Height:       3,
			GroundOffset: 2,
		},
		Difficulty: DifficultyConfig{
			Enabled:      true,
			InitialLevel: 0.0,
			Progression: ProgressionConfig{
				Type:  "score",
				MaxAt: 2000,
			},
			Scaling: ScalingConfig{
				SpeedMultiplier:  2.0,
				GapReduction:     0,
				SpacingReduction: 20,
			},
		},
	}
}

// DefaultPongConfig returns the default Pong configuration.
func DefaultPongConfig() PongConfig {
	return PongConfig{
		Physics: PongPhysics{
			BallSpeed:    0.5,
			PaddleSpeed:  1.0,
			MaxBallSpeed: 3.0,
			SpinFactor:   0.3,
		},
		Paddles: PongPaddles{
			Height: 5,
			Width:  1,
			Offset: 2,
		},
		Gameplay: PongGameplay{
			WinScore:   5,
			ServeDelay: 60,
		},
		CPU: PongCPU{
			MinSkill: 0.6,
			MaxSkill: 0.85,
		},
		Difficulty: DifficultyConfig{
			Enabled:      true,
			InitialLevel: 0.0,
			Progression: ProgressionConfig{
				Type:  "time",
				MaxAt: 36000, // 10 minutes at 60fps
			},
			Scaling: ScalingConfig{
				SpeedMultiplier:  0.5,
				GapReduction:     0,
				SpacingReduction: 0,
			},
		},
	}
}

// GetDefaultYAML returns the embedded default YAML for a game.
func GetDefaultYAML(gameID string) []byte {
	switch gameID {
	case "flappy":
		return defaultFlappyYAML
	case "dino":
		return defaultDinoYAML
	case "pong":
		return defaultPongYAML
	default:
		return nil
	}
}
