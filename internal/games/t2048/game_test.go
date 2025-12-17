package t2048

import (
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

func TestSlideRowMerge(t *testing.T) {
	tests := []struct {
		name     string
		input    [4]int
		expected [4]int
		score    int
	}{
		{
			name:     "simple merge",
			input:    [4]int{2, 2, 0, 0},
			expected: [4]int{4, 0, 0, 0},
			score:    4,
		},
		{
			name:     "merge with trailing tile",
			input:    [4]int{2, 2, 2, 0},
			expected: [4]int{4, 2, 0, 0},
			score:    4,
		},
		{
			name:     "double merge",
			input:    [4]int{2, 2, 2, 2},
			expected: [4]int{4, 4, 0, 0},
			score:    8,
		},
		{
			name:     "no merge possible",
			input:    [4]int{2, 4, 8, 16},
			expected: [4]int{2, 4, 8, 16},
			score:    0,
		},
		{
			name:     "slide with gap",
			input:    [4]int{0, 0, 2, 2},
			expected: [4]int{4, 0, 0, 0},
			score:    4,
		},
		{
			name:     "slide with multiple gaps",
			input:    [4]int{2, 0, 0, 2},
			expected: [4]int{4, 0, 0, 0},
			score:    4,
		},
		{
			name:     "no change needed",
			input:    [4]int{4, 2, 0, 0},
			expected: [4]int{4, 2, 0, 0},
			score:    0,
		},
		{
			name:     "empty row",
			input:    [4]int{0, 0, 0, 0},
			expected: [4]int{0, 0, 0, 0},
			score:    0,
		},
		{
			name:     "single tile",
			input:    [4]int{0, 4, 0, 0},
			expected: [4]int{4, 0, 0, 0},
			score:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, score := slideRow(tt.input)
			if result != tt.expected {
				t.Errorf("slideRow(%v) = %v, want %v", tt.input, result, tt.expected)
			}
			if score != tt.score {
				t.Errorf("slideRow(%v) score = %d, want %d", tt.input, score, tt.score)
			}
		})
	}
}

func TestSlideLeft(t *testing.T) {
	board := Board{
		{2, 2, 0, 0},
		{4, 0, 4, 0},
		{2, 2, 2, 2},
		{0, 0, 0, 2},
	}

	expected := Board{
		{4, 0, 0, 0},
		{8, 0, 0, 0},
		{4, 4, 0, 0},
		{2, 0, 0, 0},
	}

	result, score, changed := SlideLeft(board)

	if result != expected {
		t.Errorf("SlideLeft: got\n%v\nwant\n%v", result, expected)
	}

	if !changed {
		t.Error("SlideLeft should indicate board changed")
	}

	expectedScore := 4 + 8 + 8 // 4+8+4+4 = 20
	if score != expectedScore {
		t.Errorf("SlideLeft score = %d, want %d", score, expectedScore)
	}
}

func TestSlideRight(t *testing.T) {
	board := Board{
		{2, 2, 0, 0},
		{4, 0, 4, 0},
		{2, 2, 2, 2},
		{0, 0, 0, 2},
	}

	expected := Board{
		{0, 0, 0, 4},
		{0, 0, 0, 8},
		{0, 0, 4, 4},
		{0, 0, 0, 2},
	}

	result, _, changed := SlideRight(board)

	if result != expected {
		t.Errorf("SlideRight: got\n%v\nwant\n%v", result, expected)
	}

	if !changed {
		t.Error("SlideRight should indicate board changed")
	}
}

func TestSlideUp(t *testing.T) {
	board := Board{
		{2, 4, 2, 0},
		{2, 0, 2, 0},
		{0, 4, 2, 0},
		{0, 0, 2, 2},
	}

	expected := Board{
		{4, 8, 4, 2},
		{0, 0, 4, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	result, _, changed := SlideUp(board)

	if result != expected {
		t.Errorf("SlideUp: got\n%v\nwant\n%v", result, expected)
	}

	if !changed {
		t.Error("SlideUp should indicate board changed")
	}
}

func TestSlideDown(t *testing.T) {
	board := Board{
		{2, 4, 2, 2},
		{2, 0, 2, 0},
		{0, 4, 2, 0},
		{0, 0, 2, 0},
	}

	expected := Board{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 4, 0},
		{4, 8, 4, 2},
	}

	result, _, changed := SlideDown(board)

	if result != expected {
		t.Errorf("SlideDown: got\n%v\nwant\n%v", result, expected)
	}

	if !changed {
		t.Error("SlideDown should indicate board changed")
	}
}

func TestNoChangeNoSpawn(t *testing.T) {
	board := Board{
		{4, 2, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	// Sliding left when tiles are already left-aligned
	_, _, changed := SlideLeft(board)

	if changed {
		t.Error("SlideLeft should not change already left-aligned tiles")
	}
}

func TestGameOver(t *testing.T) {
	// Board with no empty cells and no possible merges
	board := Board{
		{2, 4, 8, 16},
		{32, 64, 128, 256},
		{512, 1024, 2048, 4096},
		{8192, 16384, 32768, 65536},
	}

	if !IsGameOver(board) {
		t.Error("Board with no moves should be game over")
	}

	// Board with no empty cells but possible merges
	boardWithMerge := Board{
		{2, 2, 8, 16},
		{32, 64, 128, 256},
		{512, 1024, 2048, 4096},
		{8192, 16384, 32768, 65536},
	}

	if IsGameOver(boardWithMerge) {
		t.Error("Board with possible merge should not be game over")
	}

	// Board with empty cells
	boardWithEmpty := Board{
		{2, 4, 8, 16},
		{32, 64, 128, 256},
		{512, 1024, 0, 4096},
		{8192, 16384, 32768, 65536},
	}

	if IsGameOver(boardWithEmpty) {
		t.Error("Board with empty cell should not be game over")
	}
}

func TestDeterministicSpawn(t *testing.T) {
	// Test that the same seed produces the same sequence of spawns
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     12345,
	}

	g1 := New()
	g1.Reset(cfg)
	board1 := g1.board

	g2 := New()
	g2.Reset(cfg)
	board2 := g2.board

	if board1 != board2 {
		t.Errorf("Same seed should produce same initial board:\n%v\nvs\n%v", board1, board2)
	}
}

func TestCampaignProgression(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     42,
	}

	g := New()
	g.Reset(cfg)

	// Set up a board with target tile
	g.board = Board{
		{128, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}
	g.currentTarget = 128

	// Simulate a move (any direction)
	input := core.NewInputFrame()
	input.Set(core.ActionDown)
	g.Step(input)

	// Board already has 128, should trigger level cleared
	if !g.levelCleared {
		t.Error("Should detect level cleared when target tile exists")
	}

	// Advance past animation
	g.levelClearTicks = 120
	g.Step(core.NewInputFrame())

	if g.levelIndex != 1 {
		t.Errorf("Should advance to level 2, got level %d", g.levelIndex+1)
	}
}

func TestEndlessModeNoWin(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     42,
	}

	g := NewEndless()
	g.Reset(cfg)

	// Set up a board with high tile
	g.board = Board{
		{8192, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	// Make a move
	input := core.NewInputFrame()
	input.Set(core.ActionDown)
	g.Step(input)

	// Should not trigger level cleared or win in endless mode
	if g.levelCleared {
		t.Error("Endless mode should not have level cleared")
	}
	if g.won {
		t.Error("Endless mode should not have win state")
	}
}

func TestMaxTile(t *testing.T) {
	board := Board{
		{2, 4, 8, 16},
		{32, 64, 128, 256},
		{512, 1024, 2048, 4},
		{8, 16, 32, 64},
	}

	max := MaxTile(board)
	if max != 2048 {
		t.Errorf("MaxTile = %d, want 2048", max)
	}
}

func TestEmptyCells(t *testing.T) {
	board := Board{
		{2, 0, 8, 0},
		{0, 64, 0, 256},
		{512, 0, 2048, 0},
		{0, 16, 0, 64},
	}

	cells := EmptyCells(board)
	if len(cells) != 8 {
		t.Errorf("EmptyCells count = %d, want 8", len(cells))
	}
}

func TestOneMergePerTilePerMove(t *testing.T) {
	// [4, 4, 4, 4] sliding left should become [8, 8, 0, 0], not [16, 0, 0, 0]
	row := [4]int{4, 4, 4, 4}
	result, score := slideRow(row)

	expected := [4]int{8, 8, 0, 0}
	if result != expected {
		t.Errorf("slideRow(%v) = %v, want %v (one merge per tile per move)", row, result, expected)
	}

	// Score should be 8+8 = 16, not 8+16 = 24
	if score != 16 {
		t.Errorf("slideRow(%v) score = %d, want 16", row, score)
	}
}

func TestSnapshot(t *testing.T) {
	cfg := core.RuntimeConfig{
		ScreenW:  80,
		ScreenH:  24,
		TickRate: 60,
		Seed:     42,
	}

	g := New()
	g.Reset(cfg)

	snap := g.Snapshot()

	if snap.Mode != "campaign" {
		t.Errorf("Snapshot Mode = %s, want campaign", snap.Mode)
	}

	if snap.Level != 1 {
		t.Errorf("Snapshot Level = %d, want 1", snap.Level)
	}

	if snap.Target != 128 {
		t.Errorf("Snapshot Target = %d, want 128", snap.Target)
	}

	if snap.State != StatePlaying {
		t.Errorf("Snapshot State = %s, want playing", snap.State)
	}
}

func TestLevelCount(t *testing.T) {
	if LevelCount() != 10 {
		t.Errorf("LevelCount() = %d, want 10", LevelCount())
	}
}

func TestLevelNames(t *testing.T) {
	names := LevelNames()
	if len(names) != 10 {
		t.Errorf("LevelNames() length = %d, want 10", len(names))
	}

	if names[0] != "Warm-up" {
		t.Errorf("First level name = %s, want Warm-up", names[0])
	}
}
