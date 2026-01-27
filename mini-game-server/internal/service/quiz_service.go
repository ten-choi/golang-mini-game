package service

import (
	"context"
	"fmt"

	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"
	"draw-and-guess-server/pkg/dictionary"
)

// QuizService defines the interface for quiz business logic
type QuizService interface {
	GetRandomOXQuiz(ctx context.Context, excludedIds []string) (*models.OXQuiz, error)
	GetRandomQAQuiz(ctx context.Context, excludedIds []string) (*models.GeneralQuiz, error)
	GetRandomOXQuizzes(ctx context.Context, excludedIds []string, count int) ([]*models.OXQuiz, error)
	GetRandomQAQuizzes(ctx context.Context, excludedIds []string, count int) ([]*models.GeneralQuiz, error)
	IsValidWord(ctx context.Context, word string) (bool, error)
}

type quizService struct {
	repo repository.QuizRepository
}

// NewQuizService creates a new quiz service
func NewQuizService(repo repository.QuizRepository) QuizService {
	return &quizService{repo: repo}
}

func (s *quizService) GetRandomOXQuiz(ctx context.Context, excludedIds []string) (*models.OXQuiz, error) {
	return s.repo.GetRandomOXQuiz(ctx, excludedIds)
}

func (s *quizService) GetRandomQAQuiz(ctx context.Context, excludedIds []string) (*models.GeneralQuiz, error) {
	return s.repo.GetRandomQAQuiz(ctx, excludedIds)
}

func (s *quizService) GetRandomOXQuizzes(ctx context.Context, excludedIds []string, count int) ([]*models.OXQuiz, error) {
	return s.repo.GetRandomOXQuizzes(ctx, excludedIds, count)
}

func (s *quizService) GetRandomQAQuizzes(ctx context.Context, excludedIds []string, count int) ([]*models.GeneralQuiz, error) {
	return s.repo.GetRandomQAQuizzes(ctx, excludedIds, count)
}

func (s *quizService) IsValidWord(ctx context.Context, word string) (bool, error) {
	// Use in-memory dictionary instead of database
	// Import: "draw-and-guess-server/pkg/dictionary"
	dict := dictionary.GetInstance()
	if !dict.IsLoaded() {
		return false, fmt.Errorf("dictionary not loaded")
	}
	return dict.IsValidWord(word), nil
}
