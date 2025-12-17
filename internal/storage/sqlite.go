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

	"github.com/vovakirdan/tui-arcade/internal/multiplayer"
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

// OnlineMatchResult represents the outcome of an online PvP match.
type OnlineMatchResult struct {
	ID             int64
	MatchID        string
	GameID         string
	Player1Session string
	Player2Session string
	Score1         int
	Score2         int
	WinnerSession  string // Empty if draw or disconnect
	EndReason      string // "completed", "disconnect", "cancelled"
	Duration       int    // Duration in seconds
	CreatedAt      time.Time
}

// Open creates or opens a SQLite database at the given path.
// It creates the parent directories if needed and runs migrations.
func Open(dbPath string) (*Store, error) {
	// Expand ~ to home directory
	if dbPath != "" && dbPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("storage: cannot expand home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[1:])
	}

	// Create parent directories
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
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

		CREATE TABLE IF NOT EXISTS online_matches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			match_id TEXT NOT NULL UNIQUE,
			game_id TEXT NOT NULL,
			player1_session TEXT NOT NULL,
			player2_session TEXT NOT NULL,
			score1 INTEGER NOT NULL DEFAULT 0,
			score2 INTEGER NOT NULL DEFAULT 0,
			winner_session TEXT,
			end_reason TEXT NOT NULL,
			duration_secs INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_online_matches_game_id ON online_matches(game_id);
		CREATE INDEX IF NOT EXISTS idx_online_matches_player1 ON online_matches(player1_session);
		CREATE INDEX IF NOT EXISTS idx_online_matches_player2 ON online_matches(player2_session);
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
		var createdAt any
		if err := rows.Scan(&e.ID, &e.GameID, &e.Score, &createdAt); err != nil {
			return nil, fmt.Errorf("storage: cannot scan row: %w", err)
		}

		// Parse the datetime - handle both time.Time and string
		switch v := createdAt.(type) {
		case time.Time:
			e.CreatedAt = v
		case string:
			if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				e.CreatedAt = parsed
			}
		}
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
		var createdAt any
		if err := rows.Scan(&e.ID, &e.GameID, &e.Score, &createdAt); err != nil {
			return nil, fmt.Errorf("storage: cannot scan row: %w", err)
		}

		// Parse the datetime - handle both time.Time and string
		switch v := createdAt.(type) {
		case time.Time:
			e.CreatedAt = v
		case string:
			if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				e.CreatedAt = parsed
			}
		}
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

// SaveOnlineMatch records the result of an online PvP match.
// Returns the ID of the inserted record.
func (s *Store) SaveOnlineMatch(result OnlineMatchResult) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO online_matches
		 (match_id, game_id, player1_session, player2_session, score1, score2, winner_session, end_reason, duration_secs)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.MatchID,
		result.GameID,
		result.Player1Session,
		result.Player2Session,
		result.Score1,
		result.Score2,
		result.WinnerSession,
		result.EndReason,
		result.Duration,
	)
	if err != nil {
		return 0, fmt.Errorf("storage: cannot save online match: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("storage: cannot get inserted ID: %w", err)
	}

	return id, nil
}

// OnlineMatchByID retrieves an online match by its match ID.
func (s *Store) OnlineMatchByID(matchID string) (*OnlineMatchResult, error) {
	var result OnlineMatchResult
	var createdAt any
	var winnerSession sql.NullString

	err := s.db.QueryRow(
		`SELECT id, match_id, game_id, player1_session, player2_session,
		        score1, score2, winner_session, end_reason, duration_secs, created_at
		 FROM online_matches
		 WHERE match_id = ?`,
		matchID,
	).Scan(
		&result.ID,
		&result.MatchID,
		&result.GameID,
		&result.Player1Session,
		&result.Player2Session,
		&result.Score1,
		&result.Score2,
		&winnerSession,
		&result.EndReason,
		&result.Duration,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("storage: cannot query online match: %w", err)
	}

	if winnerSession.Valid {
		result.WinnerSession = winnerSession.String
	}

	// Parse the datetime
	switch v := createdAt.(type) {
	case time.Time:
		result.CreatedAt = v
	case string:
		if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			result.CreatedAt = parsed
		}
	}

	return &result, nil
}

// RecentOnlineMatches retrieves the most recent online matches.
func (s *Store) RecentOnlineMatches(limit int) ([]OnlineMatchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(
		`SELECT id, match_id, game_id, player1_session, player2_session,
		        score1, score2, winner_session, end_reason, duration_secs, created_at
		 FROM online_matches
		 ORDER BY created_at DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot query online matches: %w", err)
	}
	defer rows.Close()

	var results []OnlineMatchResult
	for rows.Next() {
		var result OnlineMatchResult
		var createdAt any
		var winnerSession sql.NullString

		if err := rows.Scan(
			&result.ID,
			&result.MatchID,
			&result.GameID,
			&result.Player1Session,
			&result.Player2Session,
			&result.Score1,
			&result.Score2,
			&winnerSession,
			&result.EndReason,
			&result.Duration,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("storage: cannot scan row: %w", err)
		}

		if winnerSession.Valid {
			result.WinnerSession = winnerSession.String
		}

		// Parse the datetime
		switch v := createdAt.(type) {
		case time.Time:
			result.CreatedAt = v
		case string:
			if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				result.CreatedAt = parsed
			}
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: row iteration error: %w", err)
	}

	return results, nil
}

// PlayerMatchHistory retrieves match history for a specific session/player.
func (s *Store) PlayerMatchHistory(sessionID string, limit int) ([]OnlineMatchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(
		`SELECT id, match_id, game_id, player1_session, player2_session,
		        score1, score2, winner_session, end_reason, duration_secs, created_at
		 FROM online_matches
		 WHERE player1_session = ? OR player2_session = ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		sessionID, sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot query player matches: %w", err)
	}
	defer rows.Close()

	var results []OnlineMatchResult
	for rows.Next() {
		var result OnlineMatchResult
		var createdAt any
		var winnerSession sql.NullString

		if err := rows.Scan(
			&result.ID,
			&result.MatchID,
			&result.GameID,
			&result.Player1Session,
			&result.Player2Session,
			&result.Score1,
			&result.Score2,
			&winnerSession,
			&result.EndReason,
			&result.Duration,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("storage: cannot scan row: %w", err)
		}

		if winnerSession.Valid {
			result.WinnerSession = winnerSession.String
		}

		// Parse the datetime
		switch v := createdAt.(type) {
		case time.Time:
			result.CreatedAt = v
		case string:
			if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				result.CreatedAt = parsed
			}
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: row iteration error: %w", err)
	}

	return results, nil
}

// SaveMatchResult implements multiplayer.MatchResultSaver.
// This adapter allows the coordinator to save match results without direct storage dependency.
func (s *Store) SaveMatchResult(data multiplayer.MatchResultData) error {
	result := OnlineMatchResult{
		MatchID:        data.MatchID,
		GameID:         data.GameID,
		Player1Session: data.Player1Session,
		Player2Session: data.Player2Session,
		Score1:         data.Score1,
		Score2:         data.Score2,
		WinnerSession:  data.WinnerSession,
		EndReason:      data.EndReason,
		Duration:       data.DurationSecs,
	}
	_, err := s.SaveOnlineMatch(result)
	return err
}

// Ensure Store implements MatchResultSaver
var _ multiplayer.MatchResultSaver = (*Store)(nil)

// GameStats contains aggregated statistics for a game.
type GameStats struct {
	GameID     string
	GamesCount int
	HighScore  int
	AvgScore   float64
	TotalScore int64
	LastPlayed time.Time
}

// GetGameStats retrieves aggregated statistics for a specific game.
func (s *Store) GetGameStats(gameID string) (*GameStats, error) {
	stats := &GameStats{GameID: gameID}

	// Get count, high, avg, total
	err := s.db.QueryRow(
		`SELECT COUNT(*), COALESCE(MAX(score), 0), COALESCE(AVG(score), 0), COALESCE(SUM(score), 0)
		 FROM scores WHERE game_id = ?`,
		gameID,
	).Scan(&stats.GamesCount, &stats.HighScore, &stats.AvgScore, &stats.TotalScore)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot get game stats: %w", err)
	}

	// Get last played
	var lastPlayed any
	err = s.db.QueryRow(
		`SELECT created_at FROM scores WHERE game_id = ? ORDER BY created_at DESC LIMIT 1`,
		gameID,
	).Scan(&lastPlayed)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("storage: cannot get last played: %w", err)
	}
	if err == nil {
		switch v := lastPlayed.(type) {
		case time.Time:
			stats.LastPlayed = v
		case string:
			if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				stats.LastPlayed = parsed
			}
		}
	}

	return stats, nil
}

// GetAllGamesStats retrieves statistics for all games that have been played.
func (s *Store) GetAllGamesStats() (map[string]*GameStats, error) {
	rows, err := s.db.Query(
		`SELECT game_id, COUNT(*), MAX(score), AVG(score), SUM(score), MAX(created_at)
		 FROM scores
		 GROUP BY game_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot get all games stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]*GameStats)
	for rows.Next() {
		var s GameStats
		var lastPlayed any
		if err := rows.Scan(&s.GameID, &s.GamesCount, &s.HighScore, &s.AvgScore, &s.TotalScore, &lastPlayed); err != nil {
			return nil, fmt.Errorf("storage: cannot scan stats row: %w", err)
		}

		switch v := lastPlayed.(type) {
		case time.Time:
			s.LastPlayed = v
		case string:
			if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				s.LastPlayed = parsed
			}
		}

		stats[s.GameID] = &s
	}

	return stats, nil
}
