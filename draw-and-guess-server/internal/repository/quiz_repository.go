package repository

import (
	"context"
	"fmt"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// QuizRepository defines the interface for quiz data access
type QuizRepository interface {
	IsValidWord(ctx context.Context, word string) (bool, error)
	GetRandomOXQuiz(ctx context.Context, excludedIds []string) (*models.OXQuiz, error)
	GetRandomQAQuiz(ctx context.Context, excludedIds []string) (*models.GeneralQuiz, error)
}

type quizRepository struct {
	wordsCollection     *mongo.Collection
	oxQuizzesCollection *mongo.Collection
	qaQuizzesCollection *mongo.Collection
}

// NewQuizRepository creates a new quiz repository
func NewQuizRepository(wordsCollection, oxQuizzesCollection, qaQuizzesCollection *mongo.Collection) QuizRepository {
	return &quizRepository{
		wordsCollection:     wordsCollection,
		oxQuizzesCollection: oxQuizzesCollection,
		qaQuizzesCollection: qaQuizzesCollection,
	}
}

// IsValidWord checks if a word exists in the dictionary (now uses in-memory dictionary)
// This method is kept for interface compatibility but should not be used
// Use pkg/dictionary directly instead
func (r *quizRepository) IsValidWord(ctx context.Context, word string) (bool, error) {
	// This is deprecated - word validation is now done in-memory
	// Kept for interface compatibility only
	return false, common.NewInternalError("IsValidWord is deprecated, use dictionary package", nil)
}

// GetRandomOXQuiz returns a random OX quiz, excluding already used quiz IDs
func (r *quizRepository) GetRandomOXQuiz(ctx context.Context, excludedIds []string) (*models.OXQuiz, error) {
	filter := bson.M{"is_active": true}

	// Exclude already used IDs
	if len(excludedIds) > 0 {
		var excludeIDList []int64
		for _, id := range excludedIds {
			var numID int64
			if _, err := fmt.Sscanf(id, "%d", &numID); err == nil {
				excludeIDList = append(excludeIDList, numID)
			}
		}
		if len(excludeIDList) > 0 {
			filter["id"] = bson.M{"$nin": excludeIDList}
		}
	}

	// MongoDB aggregation pipeline for random selection
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
	}

	cursor, err := r.oxQuizzesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, common.NewInternalError("failed to get OX quiz", err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, common.NewNotFoundError("no OX quizzes found")
	}

	var quiz models.OXQuiz
	if err := cursor.Decode(&quiz); err != nil {
		return nil, common.NewInternalError("failed to decode OX quiz", err)
	}

	return &quiz, nil
}

// GetRandomQAQuiz returns a random QA (general) quiz, excluding already used quiz IDs
func (r *quizRepository) GetRandomQAQuiz(ctx context.Context, excludedIds []string) (*models.GeneralQuiz, error) {
	filter := bson.M{"is_active": true}

	// Exclude already used IDs
	if len(excludedIds) > 0 {
		var excludeIDList []interface{}
		for _, id := range excludedIds {
			var numID int64
			if _, err := fmt.Sscanf(id, "%d", &numID); err == nil {
				excludeIDList = append(excludeIDList, numID)
			} else {
				// Try as ObjectID
				if oid, err := primitive.ObjectIDFromHex(id); err == nil {
					excludeIDList = append(excludeIDList, oid)
				}
			}
		}
		if len(excludeIDList) > 0 {
			filter["_id"] = bson.M{"$nin": excludeIDList}
		}
	}

	// MongoDB aggregation pipeline for random selection
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
	}

	cursor, err := r.qaQuizzesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, common.NewInternalError("failed to get QA quiz", err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, common.NewNotFoundError("no QA quizzes found")
	}

	var quiz models.GeneralQuiz
	if err := cursor.Decode(&quiz); err != nil {
		return nil, common.NewInternalError("failed to decode QA quiz", err)
	}

	return &quiz, nil
}
