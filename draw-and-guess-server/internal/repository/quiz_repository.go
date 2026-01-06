package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"

	"github.com/lib/pq"
)

// QuizRepository defines the interface for quiz data access
type QuizRepository interface {
	IsValidWord(ctx context.Context, word string) (bool, error)
	GetRandomOXQuiz(ctx context.Context, excludedIds []string) (*models.OXQuiz, error)
	GetRandomQAQuiz(ctx context.Context, excludedIds []string) (*models.GeneralQuiz, error)
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

// GetRandomOXQuiz returns a random OX quiz, excluding already used quiz IDs
func (r *quizRepository) GetRandomOXQuiz(ctx context.Context, excludedIds []string) (*models.OXQuiz, error) {
	query := `
		SELECT id, category, difficulty, question, answer, explanation, usage_count, is_active, created_at, updated_at
		FROM ox_quizzes
		WHERE is_active = true
		AND id NOT IN (SELECT unnest($1::bigint[]))
		ORDER BY RANDOM()
		LIMIT 1
	`

	// Convert string IDs to bigint array format for PostgreSQL
	var idArray []int64
	for _, id := range excludedIds {
		var numID int64
		if _, err := fmt.Sscanf(id, "%d", &numID); err == nil {
			idArray = append(idArray, numID)
		}
	}

	// If no valid excludedIds, use simple query
	if len(idArray) == 0 {
		query = `
			SELECT id, category, difficulty, question, answer, explanation, usage_count, is_active, created_at, updated_at
			FROM ox_quizzes
			WHERE is_active = true
			ORDER BY RANDOM()
			LIMIT 1
		`
	}

	quiz := &models.OXQuiz{}
	var err error

	if len(idArray) == 0 {
		err = r.db.QueryRowContext(ctx, query).Scan(
			&quiz.ID, &quiz.Category, &quiz.Difficulty, &quiz.Question,
			&quiz.Answer, &quiz.Explanation, &quiz.UsageCount, &quiz.IsActive,
			&quiz.CreatedAt, &quiz.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, pq.Array(idArray)).Scan(
			&quiz.ID, &quiz.Category, &quiz.Difficulty, &quiz.Question,
			&quiz.Answer, &quiz.Explanation, &quiz.UsageCount, &quiz.IsActive,
			&quiz.CreatedAt, &quiz.UpdatedAt,
		)
	}

	if err == sql.ErrNoRows {
		return nil, common.NewNotFoundError("no OX quizzes found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get OX quiz", err)
	}

	return quiz, nil
}

// GetRandomQAQuiz returns a random QA (general) quiz, excluding already used quiz IDs
func (r *quizRepository) GetRandomQAQuiz(ctx context.Context, excludedIds []string) (*models.GeneralQuiz, error) {
	query := `
		SELECT id, category, difficulty, question, options, answer, explanation, image_url, usage_count, is_active, created_at, updated_at
		FROM qa_quizzes
		WHERE is_active = true
		AND id NOT IN (SELECT unnest($1::bigint[]))
		ORDER BY RANDOM()
		LIMIT 1
	`

	// Convert string IDs to bigint array format for PostgreSQL
	var idArray []int64
	for _, id := range excludedIds {
		var numID int64
		if _, err := fmt.Sscanf(id, "%d", &numID); err == nil {
			idArray = append(idArray, numID)
		}
	}

	// If no valid excludedIds, use simple query
	if len(idArray) == 0 {
		query = `
			SELECT id, category, difficulty, question, options, answer, explanation, image_url, usage_count, is_active, created_at, updated_at
			FROM qa_quizzes
			WHERE is_active = true
			ORDER BY RANDOM()
			LIMIT 1
		`
	}

	quiz := &models.GeneralQuiz{}
	var optionsJSON []byte
	var explanation, imageURL sql.NullString
	var err error

	if len(idArray) == 0 {
		err = r.db.QueryRowContext(ctx, query).Scan(
			&quiz.ID, &quiz.Category, &quiz.Difficulty, &quiz.Question,
			&optionsJSON, &quiz.Answer, &explanation, &imageURL,
			&quiz.UsageCount, &quiz.IsActive, &quiz.CreatedAt, &quiz.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, pq.Array(idArray)).Scan(
			&quiz.ID, &quiz.Category, &quiz.Difficulty, &quiz.Question,
			&optionsJSON, &quiz.Answer, &explanation, &imageURL,
			&quiz.UsageCount, &quiz.IsActive, &quiz.CreatedAt, &quiz.UpdatedAt,
		)
	}

	if err == sql.ErrNoRows {
		return nil, common.NewNotFoundError("no QA quizzes found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get QA quiz", err)
	}

	// Unmarshal JSON options
	if err := json.Unmarshal(optionsJSON, &quiz.Options); err != nil {
		return nil, common.NewInternalError("failed to unmarshal quiz options", err)
	}

	// Handle nullable fields
	if explanation.Valid {
		quiz.Explanation = explanation.String
	}
	if imageURL.Valid {
		quiz.ImageURL = imageURL.String
	}

	return quiz, nil
}
