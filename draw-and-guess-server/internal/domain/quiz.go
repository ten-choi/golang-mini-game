package domain

import "time"

// QuizType represents the type of quiz
type QuizType string

const (
	QuizTypeOX      QuizType = "ox"      // True/False quiz
	QuizTypeGeneral QuizType = "general" // Multiple choice quiz
	QuizTypeGuess   QuizType = "guess"   // Drawing guess game
)

// IsValid checks if the QuizType is valid
func (q QuizType) IsValid() bool {
	switch q {
	case QuizTypeOX, QuizTypeGeneral, QuizTypeGuess:
		return true
	}
	return false
}

// Quiz represents a game quiz/question
type Quiz struct {
	ID          int64       `db:"id" json:"id"`
	Type        QuizType    `db:"type" json:"type"`
	Category    string      `db:"category" json:"category"`
	Difficulty  string      `db:"difficulty" json:"difficulty"`
	Question    string      `db:"question" json:"question"`
	Answer      interface{} `db:"answer" json:"answer"`
	Hint        string      `db:"hint" json:"hint"`
	Explanation string      `db:"explanation" json:"explanation"`
	ImageURL    string      `db:"image_url" json:"image_url"`
	UsageCount  int         `db:"usage_count" json:"usage_count"`
	IsActive    bool        `db:"is_active" json:"is_active"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
}

// OXQuiz represents a True/False quiz
type OXQuiz struct {
	ID          int64     `db:"id" json:"id"`
	Category    string    `db:"category" json:"category"`
	Difficulty  string    `db:"difficulty" json:"difficulty"`
	Question    string    `db:"question" json:"question"`
	Answer      bool      `db:"answer" json:"answer"`
	Explanation string    `db:"explanation" json:"explanation"`
	UsageCount  int       `db:"usage_count" json:"usage_count"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// GuessQuiz represents a drawing guess quiz
type GuessQuiz struct {
	ID           int64             `db:"id" json:"id"`
	Category     string            `db:"category" json:"category"`
	Difficulty   string            `db:"difficulty" json:"difficulty"`
	Topic        string            `db:"topic" json:"topic"`
	Translations map[string]string `db:"translations" json:"translations"`
	Hint         string            `db:"hint" json:"hint"`
	ImageURL     string            `db:"image_url" json:"image_url"`
	UsageCount   int               `db:"usage_count" json:"usage_count"`
	IsActive     bool              `db:"is_active" json:"is_active"`
	CreatedAt    time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `db:"updated_at" json:"updated_at"`
}

// GeneralQuiz represents a multiple choice quiz
type GeneralQuiz struct {
	ID          int64     `db:"id" json:"id"`
	Category    string    `db:"category" json:"category"`
	Difficulty  string    `db:"difficulty" json:"difficulty"`
	Question    string    `db:"question" json:"question"`
	Options     []string  `db:"options" json:"options"`
	Answer      int       `db:"answer" json:"answer"`
	Explanation string    `db:"explanation" json:"explanation"`
	ImageURL    string    `db:"image_url" json:"image_url"`
	UsageCount  int       `db:"usage_count" json:"usage_count"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// QuizDifficulty represents quiz difficulty level
type QuizDifficulty string

const (
	DifficultyEasy   QuizDifficulty = "easy"
	DifficultyMedium QuizDifficulty = "medium"
	DifficultyHard   QuizDifficulty = "hard"
)

// IsValid checks if the QuizDifficulty is valid
func (d QuizDifficulty) IsValid() bool {
	switch d {
	case DifficultyEasy, DifficultyMedium, DifficultyHard:
		return true
	}
	return false
}
