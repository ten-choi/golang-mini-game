package graph

import (
	"context"
	"draw-and-guess-server/internal/graph/model"
	"time"
)

// ========================================
// PlayerStats Queries
// ========================================

// PlayerStats is the resolver for the playerStats field.
func (r *queryResolver) PlayerStats(ctx context.Context, username string, gameType *string) ([]*model.PlayerStats, error) {
	// For now, return mock data. In production, fetch from database.
	now := time.Now()

	stats := []*model.PlayerStats{}

	gameTypes := []string{"WORDCHAIN", "OX", "QA"}
	if gameType != nil {
		gameTypes = []string{*gameType}
	}

	for _, gt := range gameTypes {
		stats = append(stats, &model.PlayerStats{
			Username:   username,
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
func (r *queryResolver) Leaderboard(ctx context.Context, gameType string, limit *int32) ([]*model.PlayerStats, error) {
	// For now, return empty leaderboard. In production, fetch from database.
	return []*model.PlayerStats{}, nil
}
