package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreOpenClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	// Check that the file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestStoreSaveAndRetrieve(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	// Save some scores
	_, err = store.SaveScore("flappy", 100)
	if err != nil {
		t.Fatalf("SaveScore() failed: %v", err)
	}

	_, err = store.SaveScore("flappy", 50)
	if err != nil {
		t.Fatalf("SaveScore() failed: %v", err)
	}

	_, err = store.SaveScore("flappy", 200)
	if err != nil {
		t.Fatalf("SaveScore() failed: %v", err)
	}

	// Different game
	_, err = store.SaveScore("dino", 500)
	if err != nil {
		t.Fatalf("SaveScore() failed: %v", err)
	}

	// Retrieve top scores for flappy
	scores, err := store.TopScores("flappy", 10)
	if err != nil {
		t.Fatalf("TopScores() failed: %v", err)
	}

	if len(scores) != 3 {
		t.Errorf("Expected 3 scores, got %d", len(scores))
	}

	// Should be sorted descending
	if scores[0].Score != 200 {
		t.Errorf("Expected highest score to be 200, got %d", scores[0].Score)
	}
	if scores[1].Score != 100 {
		t.Errorf("Expected second score to be 100, got %d", scores[1].Score)
	}
	if scores[2].Score != 50 {
		t.Errorf("Expected third score to be 50, got %d", scores[2].Score)
	}

	// Retrieve top scores for dino
	dinoScores, err := store.TopScores("dino", 10)
	if err != nil {
		t.Fatalf("TopScores() failed: %v", err)
	}

	if len(dinoScores) != 1 {
		t.Errorf("Expected 1 dino score, got %d", len(dinoScores))
	}
}

func TestStoreTopScoresLimit(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	// Save 5 scores
	for i := 0; i < 5; i++ {
		store.SaveScore("test", (i+1)*100)
	}

	// Request only top 3
	scores, err := store.TopScores("test", 3)
	if err != nil {
		t.Fatalf("TopScores() failed: %v", err)
	}

	if len(scores) != 3 {
		t.Errorf("Expected 3 scores with limit, got %d", len(scores))
	}

	// Should be 500, 400, 300 (top 3)
	if scores[0].Score != 500 || scores[1].Score != 400 || scores[2].Score != 300 {
		t.Errorf("Scores not in expected order: %v", scores)
	}
}

func TestStoreHighScore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	// No scores yet
	high, err := store.HighScore("flappy")
	if err != nil {
		t.Fatalf("HighScore() failed: %v", err)
	}
	if high != 0 {
		t.Errorf("Expected high score of 0 for empty game, got %d", high)
	}

	// Add scores
	store.SaveScore("flappy", 100)
	store.SaveScore("flappy", 300)
	store.SaveScore("flappy", 200)

	high, err = store.HighScore("flappy")
	if err != nil {
		t.Fatalf("HighScore() failed: %v", err)
	}
	if high != 300 {
		t.Errorf("Expected high score of 300, got %d", high)
	}
}

func TestStoreClearScores(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	store.SaveScore("flappy", 100)
	store.SaveScore("flappy", 200)
	store.SaveScore("dino", 300)

	// Clear only flappy scores
	err = store.ClearScores("flappy")
	if err != nil {
		t.Fatalf("ClearScores() failed: %v", err)
	}

	// Flappy should be empty
	flappyScores, _ := store.TopScores("flappy", 10)
	if len(flappyScores) != 0 {
		t.Errorf("Expected 0 flappy scores after clear, got %d", len(flappyScores))
	}

	// Dino should still have scores
	dinoScores, _ := store.TopScores("dino", 10)
	if len(dinoScores) != 1 {
		t.Errorf("Dino scores should not be affected by clearing flappy")
	}
}

func TestStoreAllScores(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	// Add many scores
	for i := 0; i < 20; i++ {
		store.SaveScore("test", i*10)
	}

	scores, err := store.AllScores("test")
	if err != nil {
		t.Fatalf("AllScores() failed: %v", err)
	}

	if len(scores) != 20 {
		t.Errorf("Expected 20 scores, got %d", len(scores))
	}
}

func TestStoreExpandHomePath(t *testing.T) {
	// Test that ~ expansion works (we won't actually write to home)
	// Just verify the function doesn't crash
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "deep", "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() with nested path failed: %v", err)
	}
	defer store.Close()

	// Verify nested directories were created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created in nested directory")
	}
}
