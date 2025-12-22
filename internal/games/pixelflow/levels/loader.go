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

// Level represents a complete level definition ready for gameplay.
type Level struct {
	ID       string
	Name     string
	Width    int
	Height   int
	Pixels   map[core.Coord]core.Color
	Shooters []core.ShooterSpec
	Metadata map[string]string
	FilePath string // Original file path (for debugging)
}

// ToGrid creates a new Grid from this level's pixel data.
func (l *Level) ToGrid() *core.Grid {
	return core.NewGrid(l.Width, l.Height, l.Pixels)
}

// MakeShooters creates shooters for this level.
// If the level has shooter specs, use those; otherwise generate default shooters.
func (l *Level) MakeShooters(defaultCount int) []core.Shooter {
	if len(l.Shooters) > 0 {
		return core.MakeShootersFromSpec(l.Shooters)
	}
	return core.MakeShooters(defaultCount, 0, l.Width, l.Height)
}

// Loader handles loading levels from a directory.
type Loader struct {
	Root string // Root directory for level files
}

// NewLoader creates a new level loader for the given root directory.
func NewLoader(root string) *Loader {
	return &Loader{Root: root}
}

// LoadAll recursively scans the root directory and loads all level files.
// Returns levels sorted by ID for deterministic ordering.
func (l *Loader) LoadAll() ([]Level, error) {
	var levels []Level

	err := filepath.WalkDir(l.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check for supported extensions
		ext := strings.ToLower(filepath.Ext(path))
		if !isSupportedExtension(ext) {
			return nil
		}

		// Load and parse the file
		level, err := l.LoadFile(path)
		if err != nil {
			// Log warning but continue with other files
			// In production, you might want to handle this differently
			return nil
		}

		levels = append(levels, level)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory %s: %w", l.Root, err)
	}

	// Sort by ID for deterministic ordering
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
		Pixels:   parsed.Pixels,
		Shooters: parsed.Shooters,
		Metadata: parsed.Metadata,
		FilePath: path,
	}, nil
}

// isSupportedExtension checks if the file extension is supported.
func isSupportedExtension(ext string) bool {
	supported := formats.FormatExtensions()
	for _, s := range supported {
		if ext == s {
			return true
		}
	}
	return false
}

// parseByExtension parses file data based on extension.
func parseByExtension(data []byte, ext string) (formats.Level, error) {
	switch ext {
	case ".yaml", ".yml":
		return formats.ParseYAML(data)
	default:
		return formats.Level{}, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// LoadByID loads a specific level by its ID.
// Returns an error if not found.
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

// ListIDs returns all available level IDs in sorted order.
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
