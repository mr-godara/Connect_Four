package database

import (
	"connect-four/game"
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func NewDB() (*DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("DATABASE_URL not set, database features will be disabled")
		return nil, nil
	}

	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Printf("Database connection error: %v", err)
		return nil, nil
	}

	if err := conn.Ping(); err != nil {
		log.Printf("Database ping error: %v", err)
		return nil, nil
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		log.Printf("Database schema initialization error: %v", err)
		return nil, nil
	}

	log.Println("Database connected successfully")
	return db, nil
}

func (db *DB) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS games (
			id VARCHAR(255) PRIMARY KEY,
			player1 VARCHAR(255) NOT NULL,
			player2 VARCHAR(255) NOT NULL,
			winner VARCHAR(255),
			game_duration INTEGER,
			is_bot_game BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS leaderboard (
			username VARCHAR(255) PRIMARY KEY,
			wins INTEGER DEFAULT 0,
			losses INTEGER DEFAULT 0,
			draws INTEGER DEFAULT 0,
			total_games INTEGER DEFAULT 0,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_leaderboard_wins ON leaderboard(wins DESC)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) SaveGame(id, player1, player2, winner string, duration int, isBot bool) error {
	if db == nil || db.conn == nil {
		return nil
	}

	query := `INSERT INTO games (id, player1, player2, winner, game_duration, is_bot_game, completed_at)
	          VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	_, err := db.conn.Exec(query, id, player1, player2, winner, duration, isBot)
	return err
}

func (db *DB) UpdateLeaderboard(g *game.Game) error {
	if db == nil || db.conn == nil {
		return nil
	}

	player1 := g.Player1.Username
	player2 := g.Player2.Username

	// Update player 1
	if g.Winner == player1 {
		db.incrementPlayerStats(player1, "wins")
	} else if g.Winner == "draw" {
		db.incrementPlayerStats(player1, "draws")
	} else {
		db.incrementPlayerStats(player1, "losses")
	}

	// Update player 2 (if not bot)
	if !g.IsBot {
		if g.Winner == player2 {
			db.incrementPlayerStats(player2, "wins")
		} else if g.Winner == "draw" {
			db.incrementPlayerStats(player2, "draws")
		} else {
			db.incrementPlayerStats(player2, "losses")
		}
	}

	return nil
}

func (db *DB) incrementPlayerStats(username, stat string) error {
	query := `INSERT INTO leaderboard (username, ` + stat + `, total_games, updated_at)
	          VALUES ($1, 1, 1, NOW())
	          ON CONFLICT (username)
	          DO UPDATE SET ` + stat + ` = leaderboard.` + stat + ` + 1,
	                        total_games = leaderboard.total_games + 1,
	                        updated_at = NOW()`

	_, err := db.conn.Exec(query, username)
	return err
}

type LeaderboardEntry struct {
	Username   string `json:"username"`
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Draws      int    `json:"draws"`
	TotalGames int    `json:"total_games"`
}

func (db *DB) GetLeaderboard(limit, offset int) ([]LeaderboardEntry, error) {
	if db == nil || db.conn == nil {
		return []LeaderboardEntry{}, nil
	}

	query := `SELECT username, wins, losses, draws, total_games
	          FROM leaderboard
	          ORDER BY wins DESC, total_games DESC
	          LIMIT $1 OFFSET $2`

	rows, err := db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.Username, &entry.Wins, &entry.Losses, &entry.Draws, &entry.TotalGames); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (db *DB) GetPlayerStats(username string) (*LeaderboardEntry, error) {
	if db == nil || db.conn == nil {
		return &LeaderboardEntry{
			Username:   username,
			Wins:       0,
			Losses:     0,
			Draws:      0,
			TotalGames: 0,
		}, nil
	}

	query := `SELECT username, wins, losses, draws, total_games
	          FROM leaderboard
	          WHERE username = $1`

	var entry LeaderboardEntry
	err := db.conn.QueryRow(query, username).Scan(&entry.Username, &entry.Wins, &entry.Losses, &entry.Draws, &entry.TotalGames)
	if err == sql.ErrNoRows {
		return &LeaderboardEntry{
			Username:   username,
			Wins:       0,
			Losses:     0,
			Draws:      0,
			TotalGames: 0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (db *DB) Close() error {
	if db != nil && db.conn != nil {
		return db.conn.Close()
	}
	return nil
}
