package graph

import (
	"context"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/pkg/dictionary"
	"fmt"
)

// ========================================
// Quiz Queries
// ========================================

// RandomOXQuiz is the resolver for the randomOXQuiz field.
func (r *queryResolver) RandomOXQuiz(ctx context.Context, roomID *string) (*model.OXQuiz, error) {
	var excludedIds []string

	// If roomId is provided, get the list of already used quiz IDs
	if roomID != nil {
		room, exists := GetGameRoom(*roomID)
		if exists && room.UsedQuizIds != nil {
			excludedIds = room.UsedQuizIds
		}
	}

	quiz, err := r.QuizService.GetRandomOXQuiz(ctx, excludedIds)
	if err != nil {
		return nil, err
	}

	// If roomId is provided, add this quiz ID to the used list
	if roomID != nil {
		room, exists := GetGameRoom(*roomID)
		if exists {
			quizID := fmt.Sprintf("%d", quiz.ID)
			room.UsedQuizIds = append(room.UsedQuizIds, quizID)
			SetGameRoom(*roomID, room)
		}
	}

	return &model.OXQuiz{
		ID:          fmt.Sprintf("%d", quiz.ID),
		Category:    quiz.Category,
		Difficulty:  quiz.Difficulty,
		Question:    quiz.Question,
		Answer:      quiz.Answer,
		Explanation: &quiz.Explanation,
		UsageCount:  int32(quiz.UsageCount),
		IsActive:    quiz.IsActive,
		CreatedAt:   quiz.CreatedAt,
		UpdatedAt:   quiz.UpdatedAt,
	}, nil
}

// RandomQAQuiz is the resolver for the randomQAQuiz field.
func (r *queryResolver) RandomQAQuiz(ctx context.Context, roomID *string) (*model.GeneralQuiz, error) {
	var excludedIds []string

	// If roomId is provided, get the list of already used quiz IDs
	if roomID != nil {
		room, exists := GetGameRoom(*roomID)
		if exists && room.UsedQuizIds != nil {
			excludedIds = room.UsedQuizIds
		}
	}

	quiz, err := r.QuizService.GetRandomQAQuiz(ctx, excludedIds)
	if err != nil {
		return nil, err
	}

	// If roomId is provided, add this quiz ID to the used list
	if roomID != nil {
		room, exists := GetGameRoom(*roomID)
		if exists {
			quizID := fmt.Sprintf("%d", quiz.ID)
			room.UsedQuizIds = append(room.UsedQuizIds, quizID)
			SetGameRoom(*roomID, room)
		}
	}

	return &model.GeneralQuiz{
		ID:          fmt.Sprintf("%d", quiz.ID),
		Category:    quiz.Category,
		Difficulty:  quiz.Difficulty,
		Question:    quiz.Question,
		Options:     quiz.Options,
		Answer:      int32(quiz.Answer),
		Explanation: &quiz.Explanation,
		ImageURL:    &quiz.ImageURL,
		UsageCount:  int32(quiz.UsageCount),
		IsActive:    quiz.IsActive,
		CreatedAt:   quiz.CreatedAt,
		UpdatedAt:   quiz.UpdatedAt,
	}, nil
}

// IsValidWord is the resolver for the isValidWord field.
func (r *queryResolver) IsValidWord(ctx context.Context, word string) (bool, error) {
	dict := dictionary.GetInstance()
	return dict.IsValidWord(word), nil
}
