package service

import (
	"context"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"
)

// PlayerStatsService defines the interface for player stats business logic
type PlayerStatsService interface {
	GetPlayerStats(ctx context.Context, username string, gameType *string) ([]*models.PlayerStats, error)
	GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.PlayerStats, error)
	RecordGameResult(ctx context.Context, username, gameType string, won bool, score int) error
}

type playerStatsService struct {
	repo repository.PlayerStatsRepository
}

// NewPlayerStatsService creates a new player stats service
func NewPlayerStatsService(repo repository.PlayerStatsRepository) PlayerStatsService {
	return &playerStatsService{repo: repo}
}

func (s *playerStatsService) GetPlayerStats(ctx context.Context, username string, gameType *string) ([]*models.PlayerStats, error) {
	return s.repo.GetByUsername(ctx, username, gameType)
}

func (s *playerStatsService) GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.PlayerStats, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	return s.repo.GetLeaderboard(ctx, gameType, limit)
}

func (s *playerStatsService) RecordGameResult(ctx context.Context, username, gameType string, won bool, score int) error {
	wins := 0
	if won {
		wins = 1
	}

	stats := &models.PlayerStats{
		Username:   username,
		GameType:   gameType,
		TotalGames: 1,
		TotalWins:  wins,
		TotalScore: score,
	}

	return s.repo.UpsertStats(ctx, stats)
}
