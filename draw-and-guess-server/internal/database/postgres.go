package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"draw-and-guess-server/internal/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Connect establishes a connection to PostgreSQL
func Connect() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Println("Successfully connected to PostgreSQL!")
	return nil
}

// Disconnect closes the database connection
func Disconnect() error {
	if DB == nil {
		return nil
	}
	return DB.Close()
}

// InitSchema creates all necessary tables
func InitSchema() error {
	schema := `
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY,
		username VARCHAR(100) UNIQUE NOT NULL,
		display_name VARCHAR(100),
		email VARCHAR(255),
		avatar_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Korean words dictionary for wordchain game
	CREATE TABLE IF NOT EXISTS korean_words (
		id BIGINT PRIMARY KEY,
		word VARCHAR(50) UNIQUE NOT NULL,
		first_char CHAR(1) NOT NULL,
		last_char CHAR(1) NOT NULL,
		length INT NOT NULL,
		category VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- OX Quiz table
	CREATE TABLE IF NOT EXISTS ox_quizzes (
		id BIGINT PRIMARY KEY,
		category VARCHAR(100),
		difficulty VARCHAR(20) NOT NULL,
		question TEXT NOT NULL,
		answer BOOLEAN NOT NULL,
		explanation TEXT,
		usage_count INT DEFAULT 0,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- General Quiz (QA) table
	CREATE TABLE IF NOT EXISTS qa_quizzes (
		id BIGINT PRIMARY KEY,
		category VARCHAR(100),
		difficulty VARCHAR(20) NOT NULL,
		question TEXT NOT NULL,
		options JSONB NOT NULL,
		answer INT NOT NULL,
		explanation TEXT,
		image_url TEXT,
		usage_count INT DEFAULT 0,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Player statistics table
	CREATE TABLE IF NOT EXISTS player_stats (
		id BIGINT PRIMARY KEY,
		username VARCHAR(100) NOT NULL,
		total_games INT DEFAULT 0,
		total_wins INT DEFAULT 0,
		total_score INT DEFAULT 0,
		game_type VARCHAR(20) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(username, game_type)
	);

	-- Indexes for fast queries
	CREATE INDEX IF NOT EXISTS idx_korean_words_word ON korean_words(word);
	CREATE INDEX IF NOT EXISTS idx_korean_words_first_char ON korean_words(first_char);
	CREATE INDEX IF NOT EXISTS idx_korean_words_last_char ON korean_words(last_char);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_player_stats_username ON player_stats(username);
	CREATE INDEX IF NOT EXISTS idx_ox_quizzes_active ON ox_quizzes(is_active);
	CREATE INDEX IF NOT EXISTS idx_qa_quizzes_active ON qa_quizzes(is_active);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := DB.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("Database schema initialized successfully!")
	return nil
}
