package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadFlappy loads Flappy Bird configuration.
// Search order: customPath -> ~/.arcade/configs/flappy.yaml -> ./configs/flappy.yaml -> embedded default
func LoadFlappy(customPath string) (FlappyConfig, error) {
	var cfg FlappyConfig

	// Try custom path first
	if customPath != "" {
		data, err := os.ReadFile(customPath)
		if err != nil {
			return cfg, fmt.Errorf("failed to read config %s: %w", customPath, err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config %s: %w", customPath, err)
		}
		return cfg, nil
	}

	// Try user config directory
	if userCfgPath := userConfigPath("flappy.yaml"); userCfgPath != "" {
		if data, err := os.ReadFile(userCfgPath); err == nil {
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				return cfg, nil
			}
		}
	}

	// Try local configs directory
	if data, err := os.ReadFile("configs/flappy.yaml"); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			return cfg, nil
		}
	}

	// Use embedded default YAML
	if err := yaml.Unmarshal(defaultFlappyYAML, &cfg); err != nil {
		return DefaultFlappyConfig(), nil // Fallback to hardcoded if embed fails
	}
	return cfg, nil
}

// LoadDino loads Dino Runner configuration.
// Search order: customPath -> ~/.arcade/configs/dino.yaml -> ./configs/dino.yaml -> embedded default
func LoadDino(customPath string) (DinoConfig, error) {
	var cfg DinoConfig

	// Try custom path first
	if customPath != "" {
		data, err := os.ReadFile(customPath)
		if err != nil {
			return cfg, fmt.Errorf("failed to read config %s: %w", customPath, err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config %s: %w", customPath, err)
		}
		return cfg, nil
	}

	// Try user config directory
	if userCfgPath := userConfigPath("dino.yaml"); userCfgPath != "" {
		if data, err := os.ReadFile(userCfgPath); err == nil {
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				return cfg, nil
			}
		}
	}

	// Try local configs directory
	if data, err := os.ReadFile("configs/dino.yaml"); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			return cfg, nil
		}
	}

	// Use embedded default YAML
	if err := yaml.Unmarshal(defaultDinoYAML, &cfg); err != nil {
		return DefaultDinoConfig(), nil // Fallback to hardcoded if embed fails
	}
	return cfg, nil
}

// userConfigPath returns the path to user config file, or empty if home is unavailable.
func userConfigPath(filename string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".arcade", "configs", filename)
}

// ApplyDifficultyPreset modifies the config based on a difficulty preset.
func ApplyFlappyPreset(cfg *FlappyConfig, preset DifficultyPreset) {
	if preset == DifficultyFixed {
		cfg.Difficulty.Enabled = false
	} else {
		cfg.Difficulty.Enabled = true
		cfg.Difficulty.InitialLevel = InitialLevelForPreset(preset)
	}
}

// ApplyDinoPreset modifies the config based on a difficulty preset.
func ApplyDinoPreset(cfg *DinoConfig, preset DifficultyPreset) {
	if preset == DifficultyFixed {
		cfg.Difficulty.Enabled = false
	} else {
		cfg.Difficulty.Enabled = true
		cfg.Difficulty.InitialLevel = InitialLevelForPreset(preset)
	}
}

// LoadPong loads Pong configuration.
// Search order: customPath -> ~/.arcade/configs/pong.yaml -> ./configs/pong.yaml -> embedded default
func LoadPong(customPath string) (PongConfig, error) {
	var cfg PongConfig

	// Try custom path first
	if customPath != "" {
		data, err := os.ReadFile(customPath)
		if err != nil {
			return cfg, fmt.Errorf("failed to read config %s: %w", customPath, err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config %s: %w", customPath, err)
		}
		return cfg, nil
	}

	// Try user config directory
	if userCfgPath := userConfigPath("pong.yaml"); userCfgPath != "" {
		if data, err := os.ReadFile(userCfgPath); err == nil {
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				return cfg, nil
			}
		}
	}

	// Try local configs directory
	if data, err := os.ReadFile("configs/pong.yaml"); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			return cfg, nil
		}
	}

	// Use embedded default YAML
	if err := yaml.Unmarshal(defaultPongYAML, &cfg); err != nil {
		return DefaultPongConfig(), nil // Fallback to hardcoded if embed fails
	}
	return cfg, nil
}

// ApplyPongPreset modifies the config based on a difficulty preset.
func ApplyPongPreset(cfg *PongConfig, preset DifficultyPreset) {
	if preset == DifficultyFixed {
		cfg.Difficulty.Enabled = false
	} else {
		cfg.Difficulty.Enabled = true
		cfg.Difficulty.InitialLevel = InitialLevelForPreset(preset)
	}
}

// LoadBreakout loads Breakout configuration.
// Search order: customPath -> ~/.arcade/configs/breakout.yaml -> ./configs/breakout.yaml -> embedded default
func LoadBreakout(customPath string) (BreakoutConfig, error) {
	var cfg BreakoutConfig

	// Try custom path first
	if customPath != "" {
		data, err := os.ReadFile(customPath)
		if err != nil {
			return cfg, fmt.Errorf("failed to read config %s: %w", customPath, err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config %s: %w", customPath, err)
		}
		return cfg, nil
	}

	// Try user config directory
	if userCfgPath := userConfigPath("breakout.yaml"); userCfgPath != "" {
		if data, err := os.ReadFile(userCfgPath); err == nil {
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				return cfg, nil
			}
		}
	}

	// Try local configs directory
	if data, err := os.ReadFile("configs/breakout.yaml"); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			return cfg, nil
		}
	}

	// Use embedded default YAML
	if err := yaml.Unmarshal(defaultBreakoutYAML, &cfg); err != nil {
		return DefaultBreakoutConfig(), nil // Fallback to hardcoded if embed fails
	}
	return cfg, nil
}

// ApplyBreakoutPreset modifies the config based on a difficulty preset.
func ApplyBreakoutPreset(cfg *BreakoutConfig, preset DifficultyPreset) {
	if preset == DifficultyFixed {
		cfg.Difficulty.Enabled = false
	} else {
		cfg.Difficulty.Enabled = true
		cfg.Difficulty.InitialLevel = InitialLevelForPreset(preset)
	}

	// Adjust gameplay based on difficulty
	switch preset {
	case DifficultyEasy:
		cfg.Gameplay.Lives = 5
		cfg.Paddle.Width = 10
		cfg.Physics.BallSpeed = 250
	case DifficultyHard:
		cfg.Gameplay.Lives = 2
		cfg.Paddle.Width = 6
		cfg.Physics.BallSpeed = 400
	}
}
