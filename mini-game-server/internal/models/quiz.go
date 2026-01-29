package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuizType은 퀴즈 타입을 나타냅니다
type QuizType string

const (
	QuizTypeOX      QuizType = "ox"      // OX 퀴즈 (참/거짓)
	QuizTypeGeneral QuizType = "general" // 일반 상식 퀴즈 (객관식)
	QuizTypeGuess   QuizType = "guess"   // 그림 맞추기 퀴즈
)

// IsValid는 QuizType이 유효한 값인지 검증합니다
func (q QuizType) IsValid() bool {
	switch q {
	case QuizTypeOX, QuizTypeGeneral, QuizTypeGuess:
		return true
	}
	return false
}

// Quiz는 게임에서 사용되는 퀴즈/문제 정보입니다
type Quiz struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type       QuizType           `db:"type" json:"type"`             // 퀴즈 타입 (ox, guess)
	Category   string             `db:"category" json:"category"`     // 카테고리 (동물, 과일, 음식 등)
	Difficulty int                `db:"difficulty" json:"difficulty"` // 난이도 (1-5)
	Question   string             `db:"question" json:"question"`     // 문제/주제
	//TODO 보기 선택지 ox문제는 o랑 x가 들어가면 되고 4지선다는 4지선다용 보기가 필요
	Answer      interface{} `db:"answer" json:"answer"`           // 정답 (OX는 bool, Guess는 string)
	Explanation string      `db:"explanation" json:"explanation"` // 정답 설명
	UsageCount  int         `db:"usage_count" json:"usageCount"`  // 사용 횟수
	IsActive    bool        `db:"is_active" json:"isActive"`      // 활성화 여부
	CreatedAt   time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updatedAt"`
}

// OXQuiz는 OX 퀴즈 전용 구조체입니다
type OXQuiz struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Category    string             `bson:"category" json:"category"`
	Difficulty  int                `bson:"difficulty" json:"difficulty"`   // 1-5 (1=Easy, 2=Normal, 3=Hard, 4=VeryHard, 5=Extreme)
	Question    string             `bson:"question" json:"question"`       // 예: "사과는 과일이다"
	Answer      bool               `bson:"answer" json:"answer"`           // true (O) 또는 false (X)
	Explanation string             `bson:"explanation" json:"explanation"` // 정답 설명
	UsageCount  int                `bson:"usage_count" json:"usageCount"`  // 사용 횟수
	IsActive    bool               `bson:"is_active" json:"isActive"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

// GuessQuiz는 그림 맞추기 퀴즈 전용 구조체입니다
type GuessQuiz struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Category     string             `bson:"category" json:"category"`         // 동물, 과일, 음식 등
	Difficulty   int                `bson:"difficulty" json:"difficulty"`     // 1-5 (1=Easy, 2=Normal, 3=Hard, 4=VeryHard, 5=Extreme)
	Topic        string             `bson:"topic" json:"topic"`               // 정답 주제 (예: "사과")
	Translations map[string]string  `bson:"translations" json:"translations"` // 다국어 번역
	Hint         string             `bson:"hint" json:"hint"`                 // 힌트 (예: "과일 (2자)")
	ImageURL     string             `bson:"image_url" json:"imageUrl"`        // 참고 이미지 URL
	UsageCount   int                `bson:"usage_count" json:"usageCount"`
	IsActive     bool               `bson:"is_active" json:"isActive"`
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updatedAt"`
}

// GeneralQuiz는 일반 상식 퀴즈(객관식) 전용 구조체입니다
type GeneralQuiz struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Category    string             `bson:"category" json:"category"`            // 역사, 과학, 스포츠 등
	Difficulty  int                `bson:"difficulty" json:"difficulty"`        // 1-5 (1=Easy, 2=Normal, 3=Hard, 4=VeryHard, 5=Extreme)
	Question    string             `bson:"question" json:"question"`            // 문제 (예: "대한민국의 수도는?")
	Options     []string           `bson:"options" json:"options"`              // 선택지 (4개)
	Answer      int                `bson:"answer" json:"answer"`                // 정답 인덱스 (0-3)
	Explanation string             `bson:"explanation" json:"explanation"`      // 정답 설명
	ImageURL    string             `bson:"image_url,omitempty" json:"imageUrl"` // 문제 이미지 URL (선택)
	UsageCount  int                `bson:"usage_count" json:"usageCount"`
	IsActive    bool               `bson:"is_active" json:"isActive"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

// QuizDifficulty 상수 정의 (1-5)
const (
	DifficultyEasy     = 1 // 쉬움
	DifficultyNormal   = 2 // 보통
	DifficultyHard     = 3 // 어려움
	DifficultyVeryHard = 4 // 매우 어려움
	DifficultyExtreme  = 5 // 극악
)

// IsValidDifficulty는 난이도 값이 유효한지 검증합니다 (1-5)
func IsValidDifficulty(difficulty int) bool {
	return difficulty >= DifficultyEasy && difficulty <= DifficultyExtreme
}

// GetDifficultyName은 난이도 숫자를 문자열로 변환합니다
func GetDifficultyName(difficulty int) string {
	switch difficulty {
	case DifficultyEasy:
		return "Easy"
	case DifficultyNormal:
		return "Normal"
	case DifficultyHard:
		return "Hard"
	case DifficultyVeryHard:
		return "Very Hard"
	case DifficultyExtreme:
		return "Extreme"
	default:
		return "Unknown"
	}
}
