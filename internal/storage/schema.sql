-- Arcade scores database schema

CREATE TABLE IF NOT EXISTS scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id TEXT NOT NULL,
    score INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast lookup by game_id
CREATE INDEX IF NOT EXISTS idx_scores_game_id ON scores(game_id);

-- Composite index for top scores query (game_id + score DESC)
CREATE INDEX IF NOT EXISTS idx_scores_top ON scores(game_id, score DESC);
