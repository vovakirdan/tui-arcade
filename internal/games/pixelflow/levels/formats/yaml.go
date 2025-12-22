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
	Capacity int               `yaml:"capacity,omitempty"`
	Pixels   []YAMLPixel       `yaml:"pixels"`
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
	C string `yaml:"c"` // Color as string
}

// Level represents a parsed level ready for use.
type Level struct {
	ID       string
	Name     string
	Width    int
	Height   int
	Capacity int
	Pixels   map[core.Coord]core.Color
	Metadata map[string]string
}

// ParseYAML parses a YAML level file.
func ParseYAML(data []byte) (Level, error) {
	var yl YAMLLevel
	if err := yaml.Unmarshal(data, &yl); err != nil {
		return Level{}, fmt.Errorf("yaml unmarshal: %w", err)
	}

	capacity := yl.Capacity
	if capacity <= 0 {
		capacity = 5 // Default capacity
	}

	level := Level{
		ID:       yl.ID,
		Name:     yl.Name,
		Width:    yl.Size.W,
		Height:   yl.Size.H,
		Capacity: capacity,
		Pixels:   make(map[core.Coord]core.Color),
		Metadata: yl.Metadata,
	}

	// Parse pixels
	for _, p := range yl.Pixels {
		color, ok := core.ParseColor(p.C)
		if !ok {
			continue // Skip invalid colors
		}
		coord := core.C(p.X, p.Y)
		level.Pixels[coord] = color
	}

	return level, nil
}

// FormatExtensions returns supported file extensions.
func FormatExtensions() []string {
	return []string{".yaml", ".yml"}
}

// ToGrid creates a Grid from the level data.
func (l *Level) ToGrid() *core.Grid {
	return core.NewGrid(l.Width, l.Height, l.Pixels)
}

// ToRail creates a Rail from the level dimensions.
func (l *Level) ToRail() core.Rail {
	return core.NewRail(l.Width, l.Height)
}
