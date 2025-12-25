package models

import "time"

type GameRoom struct {
	ID                      string            `json:"id"`
	UUID                    string            `json:"uuid"`
	IsActive                bool              `json:"is_active"`
	DrawerUser              string            `json:"drawer_user"`                 // 현재 그림 그리는 사람
	RoomCreator             string            `json:"room_creator"`                // 방 생성자
	LastRoundWinner         string            `json:"last_round_winner,omitempty"` // 이전 라운드 우승자
	Players                 []Player          `json:"players"`                     // 참가자 목록
	CurrentWord             string            `json:"current_word"`
	CurrentWordTranslations map[string]string `json:"current_word_translations,omitempty"`
	RoundNumber             int               `json:"round_number"`  // 현재 라운드 (1-3)
	TimeLeft                int               `json:"time_left"`     // 남은 시간 (초)
	GameStatus              string            `json:"game_status"`   // waiting, playing, finished
	UsedWords               []string          `json:"used_words"`    // 사용된 단어들
	MaxRounds               int               `json:"max_rounds"`    // 최대 라운드 (3)
	WinningScore            int               `json:"winning_score"` // 우승 점수 (3)
	MaxPlayers              int               `json:"max_players"`   // 최대 플레이어 수 (4)
	GameType                GameType          `json:"game_type"`     // 게임 타입 (ox, general, guess)
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
}

type Player struct {
	Username string `json:"username"`
	Score    int    `json:"score"`
	Attempts int    `json:"attempts"` // 이번 라운드 제출 횟수
}

// GameType은 게임 타입을 나타냅니다 (enum 패턴)
type GameType string

const (
	GameTypeOX        GameType = "ox"        // OX 퀴즈
	GameTypeQA        GameType = "qa"        // 일반 상식 퀴즈 (객관식)
	GameTypeWordChain GameType = "wordchain" // 끝말잇기
)

// IsValid는 GameType이 유효한 값인지 검증합니다
func (g GameType) IsValid() bool {
	switch g {
	case GameTypeOX, GameTypeQA, GameTypeWordChain:
		return true
	}
	return false
}

// String은 GameType을 문자열로 반환합니다
func (g GameType) String() string {
	return string(g)
}

type GameTopic struct {
	Canonical    string            `bson:"canonical" json:"canonical"`
	Translations map[string]string `bson:"translations" json:"translations"`
}

// GetTestTopics는 테스트용 하드코딩된 퀴즈 데이터를 반환합니다
// GuessQuiz 데이터를 GameTopic 형식으로 변환하여 반환합니다
func GetTestTopics() []GameTopic {
	guessQuizzes := GetTestGuessQuizzes()
	topics := make([]GameTopic, len(guessQuizzes))

	for i, quiz := range guessQuizzes {
		topics[i] = GameTopic{
			Canonical:    quiz.Topic,
			Translations: quiz.Translations,
		}
	}

	return topics
}

type ApiResult struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

type GameUser struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Role     string `json:"role"` // drawer, player
}
