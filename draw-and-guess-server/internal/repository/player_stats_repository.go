package repository

import (
	"context"
	"database/sql"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"
)

// PlayerStatsRepository defines the interface for player stats data access
type PlayerStatsRepository interface {
	GetByUsername(ctx context.Context, username string, gameType *string) ([]*models.PlayerStats, error)
	GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.PlayerStats, error)
	UpsertStats(ctx context.Context, stats *models.PlayerStats) error
}

type playerStatsRepository struct {
	db *sql.DB
}

// NewPlayerStatsRepository creates a new player stats repository
func NewPlayerStatsRepository(db *sql.DB) PlayerStatsRepository {
	return &playerStatsRepository{db: db}
}

func (r *playerStatsRepository) GetByUsername(ctx context.Context, username string, gameType *string) ([]*models.PlayerStats, error) {
	var query string
	var rows *sql.Rows
	var err error

	if gameType != nil {
		query = `
			SELECT username, game_type, total_games, total_wins, total_score, created_at, updated_at
			FROM player_stats
			WHERE username = $1 AND game_type = $2
		`
		rows, err = r.db.QueryContext(ctx, query, username, *gameType)
	} else {
		query = `
			SELECT username, game_type, total_games, total_wins, total_score, created_at, updated_at
			FROM player_stats
			WHERE username = $1
		`
		rows, err = r.db.QueryContext(ctx, query, username)
	}

	if err != nil {
		return nil, common.NewInternalError("failed to query player stats", err)
	}
	defer rows.Close()

	var stats []*models.PlayerStats
	for rows.Next() {
		s := &models.PlayerStats{}
		err := rows.Scan(&s.Username, &s.GameType, &s.TotalGames, &s.TotalWins, &s.TotalScore, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, common.NewInternalError("failed to scan player stats", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (r *playerStatsRepository) GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.PlayerStats, error) {
	query := `
		SELECT username, game_type, total_games, total_wins, total_score, created_at, updated_at
		FROM player_stats
		WHERE game_type = $1
		ORDER BY total_score DESC, total_wins DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, gameType, limit)
	if err != nil {
		return nil, common.NewInternalError("failed to query leaderboard", err)
	}
	defer rows.Close()

	var stats []*models.PlayerStats
	for rows.Next() {
		s := &models.PlayerStats{}
		err := rows.Scan(&s.Username, &s.GameType, &s.TotalGames, &s.TotalWins, &s.TotalScore, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, common.NewInternalError("failed to scan leaderboard", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (r *playerStatsRepository) UpsertStats(ctx context.Context, stats *models.PlayerStats) error {
	query := `
		INSERT INTO player_stats (username, game_type, total_games, total_wins, total_score)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (username, game_type) 
		DO UPDATE SET 
			total_games = player_stats.total_games + EXCLUDED.total_games,
			total_wins = player_stats.total_wins + EXCLUDED.total_wins,
			total_score = player_stats.total_score + EXCLUDED.total_score,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, stats.Username, stats.GameType, stats.TotalGames, stats.TotalWins, stats.TotalScore)
	if err != nil {
		return common.NewInternalError("failed to upsert player stats", err)
	}

	return nil
}
