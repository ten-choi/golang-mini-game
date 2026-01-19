package service

import (
	"context"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"
)

// UserStatsService defines the interface for user stats business logic
type UserStatsService interface {
	GetUserStats(ctx context.Context, userID string) (*models.UserStats, error)
	GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.UserStats, error)
	UpsertStats(ctx context.Context, stats *models.UserStats) error
}

type userStatsService struct {
	repo repository.UserStatsRepository
}

// NewUserStatsService creates a new user stats service
func NewUserStatsService(repo repository.UserStatsRepository) UserStatsService {
	return &userStatsService{repo: repo}
}

func (s *userStatsService) GetUserStats(ctx context.Context, userID string) (*models.UserStats, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *userStatsService) GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.UserStats, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	return s.repo.GetLeaderboard(ctx, gameType, limit)
}

func (s *userStatsService) UpsertStats(ctx context.Context, stats *models.UserStats) error {
	return s.repo.UpsertStats(ctx, stats)
}
