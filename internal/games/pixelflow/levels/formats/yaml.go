// Package formats provides pluggable level file format parsers.
package formats

import (
	"fmt"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
	"gopkg.in/yaml.v3"
)

// YAMLLevel represents the YAML structure for a level file.
type YAMLLevel struct {
	ID       string            `yaml:"id"`
	Name     string            `yaml:"name"`
	Size     YAMLSize          `yaml:"size"`
	Pixels   []YAMLPixel       `yaml:"pixels"`
	Shooters []YAMLShooter     `yaml:"shooters,omitempty"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

// YAMLSize represents grid dimensions.
type YAMLSize struct {
	W int `yaml:"w"`
	H int `yaml:"h"`
}

// YAMLPixel represents a single pixel in YAML format.
type YAMLPixel struct {
	X int    `yaml:"x"`
	Y int    `yaml:"y"`
	C string `yaml:"c"` // Color as string: "red", "green", etc.
}

// YAMLShooter represents a shooter specification in YAML format.
type YAMLShooter struct {
	X   int    `yaml:"x"`
	Y   int    `yaml:"y"`
	Dir string `yaml:"dir"` // "up", "down", "left", "right"
	C   string `yaml:"c"`   // Color
}

// Level represents a parsed level with all required data.
type Level struct {
	ID       string
	Name     string
	Width    int
	Height   int
	Pixels   map[core.Coord]core.Color
	Shooters []core.ShooterSpec
	Metadata map[string]string
}

// ParseYAML parses a YAML level file and returns a Level struct.
func ParseYAML(data []byte) (Level, error) {
	var yl YAMLLevel
	if err := yaml.Unmarshal(data, &yl); err != nil {
		return Level{}, fmt.Errorf("yaml unmarshal: %w", err)
	}

	level := Level{
		ID:       yl.ID,
		Name:     yl.Name,
		Width:    yl.Size.W,
		Height:   yl.Size.H,
		Pixels:   make(map[core.Coord]core.Color),
		Metadata: yl.Metadata,
	}

	// Parse pixels
	for _, p := range yl.Pixels {
		color, ok := core.ParseColor(p.C)
		if !ok {
			// Skip invalid colors (assume inputs are valid per spec, but be robust)
			continue
		}
		coord := core.C(p.X, p.Y)
		level.Pixels[coord] = color
	}

	// Parse shooters if present
	for _, s := range yl.Shooters {
		dir := parseDir(s.Dir)
		color, _ := core.ParseColor(s.C)
		level.Shooters = append(level.Shooters, core.ShooterSpec{
			X:     s.X,
			Y:     s.Y,
			Dir:   dir,
			Color: color,
		})
	}

	return level, nil
}

// parseDir converts a direction string to Dir enum.
func parseDir(s string) core.Dir {
	switch s {
	case "up", "Up", "UP":
		return core.DirUp
	case "down", "Down", "DOWN":
		return core.DirDown
	case "left", "Left", "LEFT":
		return core.DirLeft
	case "right", "Right", "RIGHT":
		return core.DirRight
	default:
		return core.DirDown // Default direction
	}
}

// FormatExtensions returns the file extensions this parser handles.
func FormatExtensions() []string {
	return []string{".yaml", ".yml"}
}
