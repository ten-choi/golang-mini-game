package repository

import (
	"context"
	"time"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PlayerStatsRepository defines the interface for player stats data access
type PlayerStatsRepository interface {
	GetByUsername(ctx context.Context, username string, gameType *string) ([]*models.PlayerStats, error)
	GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.PlayerStats, error)
	UpsertStats(ctx context.Context, stats *models.PlayerStats) error
}

type playerStatsRepository struct {
	collection *mongo.Collection
}

// NewPlayerStatsRepository creates a new player stats repository
func NewPlayerStatsRepository(collection *mongo.Collection) PlayerStatsRepository {
	return &playerStatsRepository{collection: collection}
}

func (r *playerStatsRepository) GetByUsername(ctx context.Context, username string, gameType *string) ([]*models.PlayerStats, error) {
	filter := bson.M{"username": username}

	if gameType != nil {
		filter["game_type"] = *gameType
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, common.NewInternalError("failed to query player stats", err)
	}
	defer cursor.Close(ctx)

	var stats []*models.PlayerStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, common.NewInternalError("failed to decode player stats", err)
	}

	return stats, nil
}

func (r *playerStatsRepository) GetLeaderboard(ctx context.Context, gameType string, limit int) ([]*models.PlayerStats, error) {
	filter := bson.M{"game_type": gameType}
	opts := options.Find().
		SetSort(bson.D{{Key: "total_score", Value: -1}, {Key: "total_wins", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, common.NewInternalError("failed to query leaderboard", err)
	}
	defer cursor.Close(ctx)

	var stats []*models.PlayerStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, common.NewInternalError("failed to decode leaderboard", err)
	}

	return stats, nil
}

func (r *playerStatsRepository) UpsertStats(ctx context.Context, stats *models.PlayerStats) error {
	filter := bson.M{
		"username":  stats.Username,
		"game_type": stats.GameType,
	}

	update := bson.M{
		"$inc": bson.M{
			"total_games": stats.TotalGames,
			"total_wins":  stats.TotalWins,
			"total_score": stats.TotalScore,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return common.NewInternalError("failed to upsert player stats", err)
	}

	return nil
}
