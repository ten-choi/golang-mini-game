package graph

import (
	"context"
	"time"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ========================================
// UserStats Queries
// ========================================

// UserStats is the resolver for the userStats field.
func (r *queryResolver) UserStats(ctx context.Context, userID string, gameType *string) ([]*model.UserStats, error) {
	// For now, return mock data. In production, fetch from database.
	now := time.Now()

	stats := []*model.UserStats{}

	gameTypes := []string{common.GameTypeWordchain, common.GameTypeOX, common.GameTypeQA, common.GameTypeMafia}
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

// ========================================
// UserStats Mutations
// ========================================

// SaveUserStats is the resolver for the saveUserStats field.
func (r *mutationResolver) SaveUserStats(ctx context.Context, input model.UpdateUserStatsInput) (bool, error) {
	// This is an alias for CreateUserStats for backward compatibility
	return r.CreateUserStats(ctx, input)
}

// CreateUserStats is the resolver for the createUserStats field.
func (r *mutationResolver) CreateUserStats(ctx context.Context, input model.UpdateUserStatsInput) (bool, error) {
	// Convert input to models.UserStats
	userID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		return false, common.NewBadRequestError("invalid user ID format", err)
	}

	// Build game stats for the specific game type
	gameStats := make(map[string]models.DetailedRecord)

	detailedRecord := models.DetailedRecord{}

	// Update solo record if provided
	if input.SoloWins != nil || input.SoloLoses != nil || input.SoloBestScore != nil {
		soloRecord := models.SoloRecord{}
		if input.SoloWins != nil {
			soloRecord.Wins = int(*input.SoloWins)
		}
		if input.SoloLoses != nil {
			soloRecord.Loses = int(*input.SoloLoses)
		}
		if input.SoloBestScore != nil {
			soloRecord.BestScore = int64(*input.SoloBestScore)
		}
		detailedRecord.Solo = soloRecord
	}

	// Update team record if provided
	if input.TeamWins != nil || input.TeamLoses != nil || input.TeamMVPCnt != nil {
		teamRecord := models.TeamRecord{}
		if input.TeamWins != nil {
			teamRecord.Wins = int(*input.TeamWins)
		}
		if input.TeamLoses != nil {
			teamRecord.Loses = int(*input.TeamLoses)
		}
		if input.TeamMVPCnt != nil {
			teamRecord.MVPCnt = int(*input.TeamMVPCnt)
		}
		detailedRecord.Team = teamRecord
	}

	gameStats[input.GameType] = detailedRecord

	stats := &models.UserStats{
		UserID:     userID,
		TotalGames: int(input.TotalGames),
		TotalScore: int(input.TotalScore),
		GameStats:  gameStats,
	}

	err = r.UserStatsService.UpsertStats(ctx, stats)
	if err != nil {
		return false, err
	}

	return true, nil
}
