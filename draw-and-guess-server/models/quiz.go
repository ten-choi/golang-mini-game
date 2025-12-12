package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuizType은 퀴즈 타입을 나타냅니다
type QuizType string

const (
	QuizTypeOX    QuizType = "ox"    // OX 퀴즈 (참/거짓)
	QuizTypeGuess QuizType = "guess" // 그림 맞추기 퀴즈
)

// IsValid는 QuizType이 유효한 값인지 검증합니다
func (q QuizType) IsValid() bool {
	switch q {
	case QuizTypeOX, QuizTypeGuess:
		return true
	}
	return false
}

// Quiz는 게임에서 사용되는 퀴즈/문제 정보입니다
type Quiz struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type        QuizType           `bson:"type" json:"type"`                         // 퀴즈 타입 (ox, guess)
	Category    string             `bson:"category,omitempty" json:"category"`       // 카테고리 (동물, 과일, 음식 등)
	Difficulty  string             `bson:"difficulty" json:"difficulty"`             // 난이도 (easy, medium, hard)
	Question    string             `bson:"question" json:"question"`                 // 문제/주제
	Answer      interface{}        `bson:"answer" json:"answer"`                     // 정답 (OX는 bool, Guess는 string)
	Hint        string             `bson:"hint,omitempty" json:"hint"`               // 힌트 (선택사항)
	Explanation string             `bson:"explanation,omitempty" json:"explanation"` // 정답 설명 (선택사항)
	ImageURL    string             `bson:"image_url,omitempty" json:"image_url"`     // 이미지 URL (Guess 타입용)
	UsageCount  int                `bson:"usage_count" json:"usage_count"`           // 사용 횟수
	IsActive    bool               `bson:"is_active" json:"is_active"`               // 활성화 여부
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// OXQuiz는 OX 퀴즈 전용 구조체입니다
type OXQuiz struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Category    string             `bson:"category,omitempty" json:"category"`
	Difficulty  string             `bson:"difficulty" json:"difficulty"`             // easy, medium, hard
	Question    string             `bson:"question" json:"question"`                 // 예: "사과는 과일이다"
	Answer      bool               `bson:"answer" json:"answer"`                     // true (O) 또는 false (X)
	Explanation string             `bson:"explanation,omitempty" json:"explanation"` // 정답 설명
	UsageCount  int                `bson:"usage_count" json:"usage_count"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// GuessQuiz는 그림 맞추기 퀴즈 전용 구조체입니다
type GuessQuiz struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Category     string             `bson:"category,omitempty" json:"category"`         // 동물, 과일, 음식 등
	Difficulty   string             `bson:"difficulty" json:"difficulty"`               // easy, medium, hard
	Topic        string             `bson:"topic" json:"topic"`                         // 정답 주제 (예: "사과")
	Translations map[string]string  `bson:"translations,omitempty" json:"translations"` // 다국어 번역
	Hint         string             `bson:"hint,omitempty" json:"hint"`                 // 힌트 (예: "과일 (2자)")
	ImageURL     string             `bson:"image_url,omitempty" json:"image_url"`       // 참고 이미지 URL
	UsageCount   int                `bson:"usage_count" json:"usage_count"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// QuizDifficulty는 퀴즈 난이도를 나타냅니다
type QuizDifficulty string

const (
	DifficultyEasy   QuizDifficulty = "easy"
	DifficultyMedium QuizDifficulty = "medium"
	DifficultyHard   QuizDifficulty = "hard"
)

// IsValid는 QuizDifficulty가 유효한 값인지 검증합니다
func (d QuizDifficulty) IsValid() bool {
	switch d {
	case DifficultyEasy, DifficultyMedium, DifficultyHard:
		return true
	}
	return false
}
