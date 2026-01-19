package graph

import (
	"context"
	"time"

	"draw-and-guess-server/internal/graph/model"
)

// ========================================
// UserStats Queries
// ========================================

// UserStats is the resolver for the userStats field.
func (r *queryResolver) UserStats(ctx context.Context, userID string, gameType *string) ([]*model.UserStats, error) {
	// For now, return mock data. In production, fetch from database.
	now := time.Now()

	stats := []*model.UserStats{}

	gameTypes := []string{"WORDCHAIN", "OX", "QA"}
	if gameType != nil {
		gameTypes = []string{*gameType}
	}

	for _, gt := range gameTypes {
		stats = append(stats, &model.UserStats{
			Username:   userID,
			GameType:   gt,
			TotalGames: 0,
			TotalWins:  0,
			TotalScore: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	return stats, nil
}

// Leaderboard is the resolver for the leaderboard field.
func (r *queryResolver) Leaderboard(ctx context.Context, gameType string, limit *int32) ([]*model.UserStats, error) {
	// For now, return empty leaderboard. In production, fetch from database.
	return []*model.UserStats{}, nil
}
