// Package storage provides SQLite-based persistence for game scores.
// Uses the pure-Go modernc.org/sqlite driver to avoid CGO dependencies.
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Store manages the SQLite database connection for score persistence.
type Store struct {
	db *sql.DB
}

// ScoreEntry represents a single high score record.
type ScoreEntry struct {
	ID        int64
	GameID    string
	Score     int
	CreatedAt time.Time
}

// Open creates or opens a SQLite database at the given path.
// It creates the parent directories if needed and runs migrations.
func Open(dbPath string) (*Store, error) {
	// Expand ~ to home directory
	if len(dbPath) > 0 && dbPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("storage: cannot expand home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[1:])
	}

	// Create parent directories
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("storage: cannot create directory %s: %w", dir, err)
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: cannot connect to database: %w", err)
	}

	store := &Store{db: db}

	// Run migrations
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: migration failed: %w", err)
	}

	return store, nil
}

// migrate creates the database schema if it doesn't exist.
func (s *Store) migrate() error {
	schema := `
		CREATE TABLE IF NOT EXISTS scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id TEXT NOT NULL,
			score INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_scores_game_id ON scores(game_id);
		CREATE INDEX IF NOT EXISTS idx_scores_top ON scores(game_id, score DESC);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SaveScore records a new score for the given game.
// Returns the ID of the inserted record.
func (s *Store) SaveScore(gameID string, score int) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO scores (game_id, score) VALUES (?, ?)",
		gameID, score,
	)
	if err != nil {
		return 0, fmt.Errorf("storage: cannot save score: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("storage: cannot get inserted ID: %w", err)
	}

	return id, nil
}

// TopScores retrieves the top N scores for the given game.
// Results are ordered by score descending.
func (s *Store) TopScores(gameID string, limit int) ([]ScoreEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Query(
		`SELECT id, game_id, score, created_at
		 FROM scores
		 WHERE game_id = ?
		 ORDER BY score DESC
		 LIMIT ?`,
		gameID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot query scores: %w", err)
	}
	defer rows.Close()

	var entries []ScoreEntry
	for rows.Next() {
		var e ScoreEntry
		var createdAt string
		if err := rows.Scan(&e.ID, &e.GameID, &e.Score, &createdAt); err != nil {
			return nil, fmt.Errorf("storage: cannot scan row: %w", err)
		}

		// Parse the datetime string
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: row iteration error: %w", err)
	}

	return entries, nil
}

// AllScores retrieves all scores for the given game (no limit).
func (s *Store) AllScores(gameID string) ([]ScoreEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, game_id, score, created_at
		 FROM scores
		 WHERE game_id = ?
		 ORDER BY score DESC`,
		gameID,
	)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot query scores: %w", err)
	}
	defer rows.Close()

	var entries []ScoreEntry
	for rows.Next() {
		var e ScoreEntry
		var createdAt string
		if err := rows.Scan(&e.ID, &e.GameID, &e.Score, &createdAt); err != nil {
			return nil, fmt.Errorf("storage: cannot scan row: %w", err)
		}

		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		entries = append(entries, e)
	}

	return entries, nil
}

// HighScore returns the highest score for the given game.
// Returns 0 if no scores exist.
func (s *Store) HighScore(gameID string) (int, error) {
	var score sql.NullInt64
	err := s.db.QueryRow(
		"SELECT MAX(score) FROM scores WHERE game_id = ?",
		gameID,
	).Scan(&score)

	if err != nil {
		return 0, fmt.Errorf("storage: cannot query high score: %w", err)
	}

	if !score.Valid {
		return 0, nil
	}

	return int(score.Int64), nil
}

// ClearScores deletes all scores for the given game.
func (s *Store) ClearScores(gameID string) error {
	_, err := s.db.Exec("DELETE FROM scores WHERE game_id = ?", gameID)
	if err != nil {
		return fmt.Errorf("storage: cannot clear scores: %w", err)
	}
	return nil
}
