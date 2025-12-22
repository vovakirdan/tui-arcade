package core_test

import (
	"testing"

	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
)

func TestGenerateDeckAmmoSumsMatch(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
		core.C(1, 0): core.ColorPink,
		core.C(2, 0): core.ColorPink,
		core.C(0, 1): core.ColorCyan,
		core.C(1, 1): core.ColorCyan,
		core.C(2, 2): core.ColorGreen,
	}
	g := core.NewGrid(5, 5, pixels)
	rail := core.NewRail(5, 5)

	params := core.DefaultGenParams()
	params.Seed = 12345

	deck, err := core.GenerateDeckForGrid(g, rail, params)
	if err != nil {
		t.Fatalf("GenerateDeckForGrid failed: %v", err)
	}

	// Check ammo sums match pixel counts
	gridCounts := g.CountByColor()
	deckAmmo := core.DeckAmmoByColor(deck)

	for color, pixelCount := range gridCounts {
		if ammo, ok := deckAmmo[color]; !ok || ammo != pixelCount {
			t.Errorf("color %v: expected ammo %d, got %d", color, pixelCount, ammo)
		}
	}

	// Check no extra colors in deck
	for color := range deckAmmo {
		if _, ok := gridCounts[color]; !ok {
			t.Errorf("deck has color %v not in grid", color)
		}
	}
}

func TestGenerateDeckSolvable(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
		core.C(1, 1): core.ColorCyan,
		core.C(2, 2): core.ColorGreen,
	}
	g := core.NewGrid(5, 5, pixels)
	rail := core.NewRail(5, 5)

	params := core.DefaultGenParams()
	params.Seed = 42

	deck, err := core.GenerateDeckForGrid(g, rail, params)
	if err != nil {
		t.Fatalf("GenerateDeckForGrid failed: %v", err)
	}

	// Validate solvability
	err = core.ValidateDeck(g, deck, params.Capacity, params.MaxSimSteps)
	if err != nil {
		t.Errorf("generated deck should be valid: %v", err)
	}
}

func TestGenerateSimpleDeck(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
		core.C(1, 0): core.ColorPink,
		core.C(2, 0): core.ColorCyan,
	}
	g := core.NewGrid(3, 3, pixels)

	deck := core.GenerateSimpleDeck(g)

	// Should have one shooter per pixel
	if len(deck) != 3 {
		t.Errorf("expected 3 shooters, got %d", len(deck))
	}

	// Each shooter should have ammo=1
	for i, s := range deck {
		if s.Ammo != 1 {
			t.Errorf("shooter %d should have ammo 1, got %d", i, s.Ammo)
		}
	}

	// Total ammo should match pixel count
	total := core.DeckTotalAmmo(deck)
	if total != 3 {
		t.Errorf("expected total ammo 3, got %d", total)
	}
}

func TestGenerateBalancedDeck(t *testing.T) {
	pixels := make(map[core.Coord]core.Color)
	// Create 50 pink, 30 cyan pixels
	id := 0
	for i := 0; i < 50; i++ {
		pixels[core.C(id%10, id/10)] = core.ColorPink
		id++
	}
	for i := 0; i < 30; i++ {
		pixels[core.C(id%10, id/10)] = core.ColorCyan
		id++
	}

	g := core.NewGrid(10, 10, pixels)
	deck := core.GenerateBalancedDeck(g, 20, 12345)

	// Check ammo sums
	gridCounts := g.CountByColor()
	deckAmmo := core.DeckAmmoByColor(deck)

	for color, expected := range gridCounts {
		if deckAmmo[color] != expected {
			t.Errorf("color %v: expected %d, got %d", color, expected, deckAmmo[color])
		}
	}
}

func TestGeneratorDeterminism(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
		core.C(1, 1): core.ColorCyan,
		core.C(2, 2): core.ColorGreen,
		core.C(3, 3): core.ColorYellow,
	}
	g := core.NewGrid(5, 5, pixels)
	rail := core.NewRail(5, 5)

	params := core.DefaultGenParams()
	params.Seed = 99999

	// Generate twice with same seed
	deck1, err1 := core.GenerateDeckForGrid(g, rail, params)
	deck2, err2 := core.GenerateDeckForGrid(g, rail, params)

	if err1 != nil || err2 != nil {
		t.Fatalf("generation failed: %v, %v", err1, err2)
	}

	// Should produce identical decks
	if len(deck1) != len(deck2) {
		t.Errorf("deck lengths differ: %d vs %d", len(deck1), len(deck2))
	}

	for i := range deck1 {
		if deck1[i].Color != deck2[i].Color || deck1[i].Ammo != deck2[i].Ammo {
			t.Errorf("deck[%d] differs: %v vs %v", i, deck1[i], deck2[i])
		}
	}
}

func TestValidateDeckInvalidColor(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
	}
	g := core.NewGrid(3, 3, pixels)

	// Deck with color not in grid
	deck := []core.Shooter{{ID: 0, Color: core.ColorCyan, Ammo: 1}}

	err := core.ValidateDeck(g, deck, 5, 1000)
	if err == nil {
		t.Error("expected error for invalid color")
	}
}

func TestValidateDeckInsufficientAmmo(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
		core.C(1, 0): core.ColorPink,
	}
	g := core.NewGrid(3, 3, pixels)

	// Deck with not enough ammo
	deck := []core.Shooter{{ID: 0, Color: core.ColorPink, Ammo: 1}}

	err := core.ValidateDeck(g, deck, 5, 1000)
	if err == nil {
		t.Error("expected error for insufficient ammo")
	}
}

func TestValidateDeckExcessAmmo(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
	}
	g := core.NewGrid(3, 3, pixels)

	// Deck with too much ammo
	deck := []core.Shooter{{ID: 0, Color: core.ColorPink, Ammo: 5}}

	err := core.ValidateDeck(g, deck, 5, 1000)
	if err == nil {
		t.Error("expected error for excess ammo")
	}
}

func TestQuickValidateDeck(t *testing.T) {
	pixels := map[core.Coord]core.Color{
		core.C(0, 0): core.ColorPink,
		core.C(1, 0): core.ColorCyan,
	}
	g := core.NewGrid(3, 3, pixels)

	// Valid deck
	deck := []core.Shooter{
		{ID: 0, Color: core.ColorPink, Ammo: 1},
		{ID: 1, Color: core.ColorCyan, Ammo: 1},
	}

	err := core.QuickValidateDeck(g, deck)
	if err != nil {
		t.Errorf("quick validation failed: %v", err)
	}
}

func TestDeckStats(t *testing.T) {
	deck := []core.Shooter{
		{ID: 0, Color: core.ColorPink, Ammo: 40},
		{ID: 1, Color: core.ColorPink, Ammo: 20},
		{ID: 2, Color: core.ColorCyan, Ammo: 10},
	}

	stats := core.ComputeDeckStats(deck)

	if stats.TotalShooters != 3 {
		t.Errorf("expected 3 shooters, got %d", stats.TotalShooters)
	}
	if stats.TotalAmmo != 70 {
		t.Errorf("expected 70 total ammo, got %d", stats.TotalAmmo)
	}
	if stats.MinAmmo != 10 {
		t.Errorf("expected min ammo 10, got %d", stats.MinAmmo)
	}
	if stats.MaxAmmo != 40 {
		t.Errorf("expected max ammo 40, got %d", stats.MaxAmmo)
	}
	if stats.UniqueColors != 2 {
		t.Errorf("expected 2 unique colors, got %d", stats.UniqueColors)
	}
}

func TestRNGDeterminism(t *testing.T) {
	rng1 := core.NewRNG(12345)
	rng2 := core.NewRNG(12345)

	for i := 0; i < 100; i++ {
		v1 := rng1.Next()
		v2 := rng2.Next()
		if v1 != v2 {
			t.Errorf("RNG not deterministic at step %d: %d vs %d", i, v1, v2)
		}
	}
}

func TestGenerateStyledDeck(t *testing.T) {
	pixels := make(map[core.Coord]core.Color)
	// 80 pink pixels
	for i := 0; i < 80; i++ {
		pixels[core.C(i%10, i/10)] = core.ColorPink
	}
	g := core.NewGrid(10, 10, pixels)

	// Pattern: two 40s
	pattern := []int{40, 40}
	deck, err := core.GenerateStyledDeck(g, pattern, 12345)
	if err != nil {
		t.Fatalf("GenerateStyledDeck failed: %v", err)
	}

	// Check total ammo
	totalAmmo := core.DeckTotalAmmo(deck)
	if totalAmmo != 80 {
		t.Errorf("expected 80 total ammo, got %d", totalAmmo)
	}
}
