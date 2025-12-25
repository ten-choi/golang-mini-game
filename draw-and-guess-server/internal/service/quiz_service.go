package service

import (
	"context"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"
)

// QuizService defines the interface for quiz business logic
type QuizService interface {
	GetRandomOXQuiz(ctx context.Context) (*models.OXQuiz, error)
	GetRandomQAQuiz(ctx context.Context) (*models.GeneralQuiz, error)
	IsValidWord(ctx context.Context, word string) (bool, error)
}

type quizService struct {
	repo repository.QuizRepository
}

// NewQuizService creates a new quiz service
func NewQuizService(repo repository.QuizRepository) QuizService {
	return &quizService{repo: repo}
}

func (s *quizService) GetRandomOXQuiz(ctx context.Context) (*models.OXQuiz, error) {
	return s.repo.GetRandomOXQuiz(ctx)
}

func (s *quizService) GetRandomQAQuiz(ctx context.Context) (*models.GeneralQuiz, error) {
	return s.repo.GetRandomQAQuiz(ctx)
}

func (s *quizService) IsValidWord(ctx context.Context, word string) (bool, error) {
	return s.repo.IsValidWord(ctx, word)
}