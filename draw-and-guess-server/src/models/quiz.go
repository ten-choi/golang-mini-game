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

// GeneralQuiz는 일반 상식 퀴즈(객관식) 전용 구조체입니다
type GeneralQuiz struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Category    string             `bson:"category,omitempty" json:"category"`       // 역사, 과학, 스포츠 등
	Difficulty  string             `bson:"difficulty" json:"difficulty"`             // easy, medium, hard
	Question    string             `bson:"question" json:"question"`                 // 문제 (예: "대한민국의 수도는?")
	Options     []string           `bson:"options" json:"options"`                   // 선택지 (4개)
	Answer      int                `bson:"answer" json:"answer"`                     // 정답 인덱스 (0-3)
	Explanation string             `bson:"explanation,omitempty" json:"explanation"` // 정답 설명
	ImageURL    string             `bson:"image_url,omitempty" json:"image_url"`     // 문제 이미지 URL (선택)
	UsageCount  int                `bson:"usage_count" json:"usage_count"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
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

// GetTestOXQuizzes는 테스트용 하드코딩된 OX 퀴즈 데이터를 반환합니다
func GetTestOXQuizzes() []OXQuiz {
	now := time.Now()
	return []OXQuiz{
		{
			ID:          primitive.NewObjectID(),
			Category:    "과일",
			Difficulty:  "easy",
			Question:    "사과는 과일이다",
			Answer:      true,
			Explanation: "사과는 과일입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "동물",
			Difficulty:  "easy",
			Question:    "고양이는 포유류이다",
			Answer:      true,
			Explanation: "고양이는 포유류에 속합니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "동물",
			Difficulty:  "easy",
			Question:    "펭귄은 날 수 있다",
			Answer:      false,
			Explanation: "펭귄은 날지 못하는 새입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "과학",
			Difficulty:  "medium",
			Question:    "지구는 태양 주위를 돈다",
			Answer:      true,
			Explanation: "지구는 태양 주위를 공전합니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "과학",
			Difficulty:  "medium",
			Question:    "물은 섭씨 100도에서 끓는다",
			Answer:      true,
			Explanation: "1기압에서 물은 100도에서 끓습니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "지리",
			Difficulty:  "easy",
			Question:    "한국의 수도는 서울이다",
			Answer:      true,
			Explanation: "대한민국의 수도는 서울특별시입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "음식",
			Difficulty:  "easy",
			Question:    "김치는 한국 음식이다",
			Answer:      true,
			Explanation: "김치는 대표적인 한국 전통 음식입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "동물",
			Difficulty:  "medium",
			Question:    "상어는 포유류이다",
			Answer:      false,
			Explanation: "상어는 어류(물고기)입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "과학",
			Difficulty:  "medium",
			Question:    "달은 스스로 빛을 낸다",
			Answer:      false,
			Explanation: "달은 태양빛을 반사할 뿐 스스로 빛을 내지 않습니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "상식",
			Difficulty:  "easy",
			Question:    "하루는 24시간이다",
			Answer:      true,
			Explanation: "하루는 24시간으로 구성됩니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "과일",
			Difficulty:  "easy",
			Question:    "바나나는 노란색이다",
			Answer:      true,
			Explanation: "익은 바나나는 노란색입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "동물",
			Difficulty:  "medium",
			Question:    "거미는 곤충이다",
			Answer:      false,
			Explanation: "거미는 거미강에 속하며 곤충이 아닙니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "지리",
			Difficulty:  "medium",
			Question:    "일본은 섬나라이다",
			Answer:      true,
			Explanation: "일본은 섬으로 이루어진 나라입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "과학",
			Difficulty:  "hard",
			Question:    "빛의 속도는 소리의 속도보다 느리다",
			Answer:      false,
			Explanation: "빛의 속도는 소리의 속도보다 훨씬 빠릅니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "상식",
			Difficulty:  "easy",
			Question:    "불은 뜨겁다",
			Answer:      true,
			Explanation: "불은 높은 온도로 뜨겁습니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// GetTestGuessQuizzes는 테스트용 하드코딩된 그림 맞추기 퀴즈 데이터를 반환합니다
func GetTestGuessQuizzes() []GuessQuiz {
	now := time.Now()
	return []GuessQuiz{
		{
			ID:         primitive.NewObjectID(),
			Category:   "과일",
			Difficulty: "easy",
			Topic:      "사과",
			Translations: map[string]string{
				"ko": "사과",
				"en": "apple",
			},
			Hint:      "과일 (2자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "과일",
			Difficulty: "easy",
			Topic:      "바나나",
			Translations: map[string]string{
				"ko": "바나나",
				"en": "banana",
			},
			Hint:      "노란 과일 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "동물",
			Difficulty: "easy",
			Topic:      "고양이",
			Translations: map[string]string{
				"ko": "고양이",
				"en": "cat",
			},
			Hint:      "애완동물 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "동물",
			Difficulty: "easy",
			Topic:      "강아지",
			Translations: map[string]string{
				"ko": "강아지",
				"en": "dog",
			},
			Hint:      "애완동물 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "교통",
			Difficulty: "easy",
			Topic:      "자동차",
			Translations: map[string]string{
				"ko": "자동차",
				"en": "car",
			},
			Hint:      "교통수단 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "교통",
			Difficulty: "medium",
			Topic:      "비행기",
			Translations: map[string]string{
				"ko": "비행기",
				"en": "airplane",
			},
			Hint:      "하늘을 나는 것 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "건물",
			Difficulty: "easy",
			Topic:      "집",
			Translations: map[string]string{
				"ko": "집",
				"en": "house",
			},
			Hint:      "사는 곳 (1자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "자연",
			Difficulty: "easy",
			Topic:      "나무",
			Translations: map[string]string{
				"ko": "나무",
				"en": "tree",
			},
			Hint:      "식물 (2자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "자연",
			Difficulty: "easy",
			Topic:      "꽃",
			Translations: map[string]string{
				"ko": "꽃",
				"en": "flower",
			},
			Hint:      "예쁜 식물 (1자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "자연",
			Difficulty: "easy",
			Topic:      "해",
			Translations: map[string]string{
				"ko": "해",
				"en": "sun",
			},
			Hint:      "낮에 뜨는 것 (1자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "자연",
			Difficulty: "easy",
			Topic:      "달",
			Translations: map[string]string{
				"ko": "달",
				"en": "moon",
			},
			Hint:      "밤에 뜨는 것 (1자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "자연",
			Difficulty: "easy",
			Topic:      "별",
			Translations: map[string]string{
				"ko": "별",
				"en": "star",
			},
			Hint:      "밤하늘에 반짝이는 것 (1자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "동물",
			Difficulty: "easy",
			Topic:      "물고기",
			Translations: map[string]string{
				"ko": "물고기",
				"en": "fish",
			},
			Hint:      "물속 동물 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "동물",
			Difficulty: "easy",
			Topic:      "새",
			Translations: map[string]string{
				"ko": "새",
				"en": "bird",
			},
			Hint:      "하늘을 나는 동물 (1자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "음식",
			Difficulty: "easy",
			Topic:      "피자",
			Translations: map[string]string{
				"ko": "피자",
				"en": "pizza",
			},
			Hint:      "이탈리아 음식 (2자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "음식",
			Difficulty: "easy",
			Topic:      "햄버거",
			Translations: map[string]string{
				"ko": "햄버거",
				"en": "hamburger",
			},
			Hint:      "패스트푸드 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "음식",
			Difficulty: "easy",
			Topic:      "케이크",
			Translations: map[string]string{
				"ko": "케이크",
				"en": "cake",
			},
			Hint:      "생일에 먹는 것 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "음식",
			Difficulty: "easy",
			Topic:      "아이스크림",
			Translations: map[string]string{
				"ko": "아이스크림",
				"en": "ice cream",
			},
			Hint:      "차가운 디저트 (5자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "전자제품",
			Difficulty: "easy",
			Topic:      "컴퓨터",
			Translations: map[string]string{
				"ko": "컴퓨터",
				"en": "computer",
			},
			Hint:      "전자제품 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "전자제품",
			Difficulty: "easy",
			Topic:      "전화기",
			Translations: map[string]string{
				"ko": "전화기",
				"en": "phone",
			},
			Hint:      "통화하는 것 (3자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "생활용품",
			Difficulty: "easy",
			Topic:      "우산",
			Translations: map[string]string{
				"ko": "우산",
				"en": "umbrella",
			},
			Hint:      "비올 때 쓰는 것 (2자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "생활용품",
			Difficulty: "easy",
			Topic:      "안경",
			Translations: map[string]string{
				"ko": "안경",
				"en": "glasses",
			},
			Hint:      "눈에 쓰는 것 (2자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "생활용품",
			Difficulty: "easy",
			Topic:      "시계",
			Translations: map[string]string{
				"ko": "시계",
				"en": "clock",
			},
			Hint:      "시간 보는 것 (2자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "학용품",
			Difficulty: "easy",
			Topic:      "책",
			Translations: map[string]string{
				"ko": "책",
				"en": "book",
			},
			Hint:      "읽는 것 (1자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         primitive.NewObjectID(),
			Category:   "학용품",
			Difficulty: "easy",
			Topic:      "연필",
			Translations: map[string]string{
				"ko": "연필",
				"en": "pencil",
			},
			Hint:      "쓰는 도구 (2자)",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

// GetTestGeneralQuizzes는 테스트용 하드코딩된 일반 상식 퀴즈 데이터를 반환합니다
func GetTestGeneralQuizzes() []GeneralQuiz {
	now := time.Now()
	return []GeneralQuiz{
		{
			ID:          primitive.NewObjectID(),
			Category:    "지리",
			Difficulty:  "easy",
			Question:    "대한민국의 수도는?",
			Options:     []string{"서울", "부산", "대구", "인천"},
			Answer:      0,
			Explanation: "대한민국의 수도는 서울입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "역사",
			Difficulty:  "easy",
			Question:    "한글을 만든 사람은?",
			Options:     []string{"이순신", "세종대왕", "김구", "유관순"},
			Answer:      1,
			Explanation: "세종대왕이 훈민정음(한글)을 창제했습니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "과학",
			Difficulty:  "easy",
			Question:    "물의 화학식은?",
			Options:     []string{"H2O", "CO2", "O2", "NaCl"},
			Answer:      0,
			Explanation: "물의 화학식은 H2O입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "상식",
			Difficulty:  "easy",
			Question:    "일주일은 며칠인가?",
			Options:     []string{"5일", "6일", "7일", "8일"},
			Answer:      2,
			Explanation: "일주일은 7일입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "동물",
			Difficulty:  "medium",
			Question:    "가장 빠른 육지 동물은?",
			Options:     []string{"사자", "표범", "치타", "말"},
			Answer:      2,
			Explanation: "치타는 시속 최대 120km로 달릴 수 있는 가장 빠른 육지 동물입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "지리",
			Difficulty:  "medium",
			Question:    "세계에서 가장 큰 대양은?",
			Options:     []string{"대서양", "태평양", "인도양", "북극해"},
			Answer:      1,
			Explanation: "태평양은 지구 표면의 약 46%를 차지하는 가장 큰 대양입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "과학",
			Difficulty:  "medium",
			Question:    "태양계에서 가장 큰 행성은?",
			Options:     []string{"지구", "화성", "목성", "토성"},
			Answer:      2,
			Explanation: "목성은 태양계에서 가장 큰 행성입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "스포츠",
			Difficulty:  "medium",
			Question:    "올림픽은 몇 년마다 개최되나?",
			Options:     []string{"2년", "3년", "4년", "5년"},
			Answer:      2,
			Explanation: "하계 올림픽은 4년마다 개최됩니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "역사",
			Difficulty:  "medium",
			Question:    "제2차 세계대전이 끝난 해는?",
			Options:     []string{"1940년", "1942년", "1945년", "1950년"},
			Answer:      2,
			Explanation: "제2차 세계대전은 1945년에 끝났습니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "음악",
			Difficulty:  "medium",
			Question:    "피아노 건반은 흑백 합쳐 몇 개인가?",
			Options:     []string{"76개", "88개", "96개", "108개"},
			Answer:      1,
			Explanation: "표준 피아노 건반은 흰 건반 52개, 검은 건반 36개로 총 88개입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "문학",
			Difficulty:  "hard",
			Question:    "노벨문학상을 받은 한국인은?",
			Options:     []string{"김소월", "윤동주", "이승만", "김대중"},
			Answer:      3,
			Explanation: "김대중 대통령이 2000년 노벨평화상을 받았습니다. (노벨문학상은 아직 한국인 수상자 없음)",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "과학",
			Difficulty:  "hard",
			Question:    "빛의 속도는 초속 약 몇 km인가?",
			Options:     []string{"30만 km", "100만 km", "300만 km", "1000만 km"},
			Answer:      0,
			Explanation: "빛의 속도는 초속 약 30만 km(정확히는 299,792km)입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "역사",
			Difficulty:  "hard",
			Question:    "프랑스 혁명이 일어난 해는?",
			Options:     []string{"1776년", "1789년", "1800년", "1812년"},
			Answer:      1,
			Explanation: "프랑스 혁명은 1789년에 시작되었습니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "지리",
			Difficulty:  "hard",
			Question:    "세계에서 가장 높은 산은?",
			Options:     []string{"K2", "에베레스트", "킬리만자로", "백두산"},
			Answer:      1,
			Explanation: "에베레스트산은 해발 8,849m로 세계에서 가장 높은 산입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Category:    "예술",
			Difficulty:  "hard",
			Question:    "모나리자를 그린 화가는?",
			Options:     []string{"피카소", "고흐", "레오나르도 다 빈치", "미켈란젤로"},
			Answer:      2,
			Explanation: "모나리자는 레오나르도 다 빈치의 작품입니다.",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}
