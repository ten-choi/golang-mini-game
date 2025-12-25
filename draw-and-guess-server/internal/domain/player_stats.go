package domain

import "time"

// PlayerStats represents player statistics
type PlayerStats struct {
	ID         int64     `json:"id" db:"id"`
	Username   string    `json:"username" db:"username"`
	TotalGames int       `json:"total_games" db:"total_games"`
	TotalWins  int       `json:"total_wins" db:"total_wins"`
	TotalScore int       `json:"total_score" db:"total_score"`
	GameType   string    `json:"game_type" db:"game_type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
