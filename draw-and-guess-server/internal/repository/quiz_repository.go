package repository

import (
	"context"
	"database/sql"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"
)

// QuizRepository defines the interface for quiz data access
type QuizRepository interface {
	IsValidWord(ctx context.Context, word string) (bool, error)
	GetRandomOXQuiz(ctx context.Context) (*models.OXQuiz, error)
	GetRandomQAQuiz(ctx context.Context) (*models.GeneralQuiz, error)
}

type quizRepository struct {
	db *sql.DB
}

// NewQuizRepository creates a new quiz repository
func NewQuizRepository(db *sql.DB) QuizRepository {
	return &quizRepository{db: db}
}

// IsValidWord checks if a word exists in the Korean words dictionary
func (r *quizRepository) IsValidWord(ctx context.Context, word string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM korean_words WHERE word = $1)"
	err := r.db.QueryRowContext(ctx, query, word).Scan(&exists)
	if err != nil {
		return false, common.NewInternalError("failed to check word validity", err)
	}
	return exists, nil
}

// GetRandomOXQuiz returns a random OX quiz
func (r *quizRepository) GetRandomOXQuiz(ctx context.Context) (*models.OXQuiz, error) {
	query := `
		SELECT id, category, difficulty, question, answer, explanation, usage_count, is_active, created_at, updated_at
		FROM ox_quizzes
		WHERE is_active = true
		ORDER BY RANDOM()
		LIMIT 1
	`

	quiz := &models.OXQuiz{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&quiz.ID, &quiz.Category, &quiz.Difficulty, &quiz.Question,
		&quiz.Answer, &quiz.Explanation, &quiz.UsageCount, &quiz.IsActive,
		&quiz.CreatedAt, &quiz.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, common.NewNotFoundError("no OX quizzes found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get OX quiz", err)
	}

	return quiz, nil
}

// GetRandomQAQuiz returns a random QA (general) quiz
func (r *quizRepository) GetRandomQAQuiz(ctx context.Context) (*models.GeneralQuiz, error) {
	query := `
		SELECT id, category, difficulty, question, options, answer, explanation, image_url, usage_count, is_active, created_at, updated_at
		FROM qa_quizzes
		WHERE is_active = true
		ORDER BY RANDOM()
		LIMIT 1
	`

	quiz := &models.GeneralQuiz{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&quiz.ID, &quiz.Category, &quiz.Difficulty, &quiz.Question,
		&quiz.Options, &quiz.Answer, &quiz.Explanation, &quiz.ImageURL,
		&quiz.UsageCount, &quiz.IsActive, &quiz.CreatedAt, &quiz.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, common.NewNotFoundError("no QA quizzes found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get QA quiz", err)
	}

	return quiz, nil
}
