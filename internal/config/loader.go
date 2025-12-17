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
