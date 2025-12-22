// Package levels provides level loading functionality for PixelFlow.
// This package depends on core but core does not depend on levels.
package levels

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/levels/formats"
)

// Level represents a complete level definition.
type Level struct {
	ID       string
	Name     string
	Width    int
	Height   int
	Capacity int
	Pixels   map[core.Coord]core.Color
	Metadata map[string]string
	FilePath string
}

// ToGrid creates a Grid from the level.
func (l *Level) ToGrid() *core.Grid {
	return core.NewGrid(l.Width, l.Height, l.Pixels)
}

// ToRail creates a Rail from the level dimensions.
func (l *Level) ToRail() core.Rail {
	return core.NewRail(l.Width, l.Height)
}

// NewState creates a game state from this level with the given deck.
func (l *Level) NewState(deck []core.Shooter) *core.State {
	return core.NewState(l.ToGrid(), deck, l.Capacity)
}

// Loader handles loading levels from a directory.
type Loader struct {
	Root string
}

// NewLoader creates a new level loader.
func NewLoader(root string) *Loader {
	return &Loader{Root: root}
}

// LoadAll recursively scans and loads all level files.
// Returns levels sorted by ID for deterministic ordering.
func (l *Loader) LoadAll() ([]Level, error) {
	var levels []Level

	err := filepath.WalkDir(l.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !isSupportedExtension(ext) {
			return nil
		}

		level, err := l.LoadFile(path)
		if err != nil {
			// Skip invalid files
			return nil
		}

		levels = append(levels, level)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory %s: %w", l.Root, err)
	}

	// Sort by ID for determinism
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].ID < levels[j].ID
	})

	return levels, nil
}

// LoadFile loads a single level file.
func (l *Loader) LoadFile(path string) (Level, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Level{}, fmt.Errorf("reading file %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	parsed, err := parseByExtension(data, ext)
	if err != nil {
		return Level{}, fmt.Errorf("parsing file %s: %w", path, err)
	}

	return Level{
		ID:       parsed.ID,
		Name:     parsed.Name,
		Width:    parsed.Width,
		Height:   parsed.Height,
		Capacity: parsed.Capacity,
		Pixels:   parsed.Pixels,
		Metadata: parsed.Metadata,
		FilePath: path,
	}, nil
}

// LoadByID loads a specific level by ID.
func (l *Loader) LoadByID(id string) (Level, error) {
	levels, err := l.LoadAll()
	if err != nil {
		return Level{}, err
	}

	for _, lvl := range levels {
		if lvl.ID == id {
			return lvl, nil
		}
	}

	return Level{}, fmt.Errorf("level not found: %s", id)
}

// ListIDs returns all level IDs in sorted order.
func (l *Loader) ListIDs() ([]string, error) {
	levels, err := l.LoadAll()
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(levels))
	for i, lvl := range levels {
		ids[i] = lvl.ID
	}
	return ids, nil
}

// isSupportedExtension checks if extension is supported.
func isSupportedExtension(ext string) bool {
	for _, supported := range formats.FormatExtensions() {
		if ext == supported {
			return true
		}
	}
	return false
}

// parseByExtension routes to the correct parser.
func parseByExtension(data []byte, ext string) (formats.Level, error) {
	switch ext {
	case ".yaml", ".yml":
		return formats.ParseYAML(data)
	default:
		return formats.Level{}, fmt.Errorf("unsupported extension: %s", ext)
	}
}
