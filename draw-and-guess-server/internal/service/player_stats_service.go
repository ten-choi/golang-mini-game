package service

import (
	"context"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"
)

// PlayerStatsService defines the interface for player stats business logic
type PlayerStatsService interface {
	GetPlayerStats(ctx context.Context, userID string) (*models.PlayerStats, error)
	GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.PlayerStats, error)
	UpsertStats(ctx context.Context, stats *models.PlayerStats) error
}

type playerStatsService struct {
	repo repository.PlayerStatsRepository
}

// NewPlayerStatsService creates a new player stats service
func NewPlayerStatsService(repo repository.PlayerStatsRepository) PlayerStatsService {
	return &playerStatsService{repo: repo}
}

func (s *playerStatsService) GetPlayerStats(ctx context.Context, userID string) (*models.PlayerStats, error) {
	return s.repo.GetByUserID(ctx, userID)
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

func (s *playerStatsService) UpsertStats(ctx context.Context, stats *models.PlayerStats) error {
	return s.repo.UpsertStats(ctx, stats)
}
