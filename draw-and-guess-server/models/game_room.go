package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GameRoom struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UUID                    string             `bson:"uuid" json:"uuid"`
	IsActive                bool               `bson:"is_active" json:"is_active"`
	DrawerUser              string             `bson:"drawer_user" json:"drawer_user"`                       // 현재 그림 그리는 사람
	RoomCreator             string             `bson:"room_creator" json:"room_creator"`                     // 방 생성자
	LastRoundWinner         string             `bson:"last_round_winner,omitempty" json:"last_round_winner"` // 이전 라운드 우승자
	Players                 []Player           `bson:"players" json:"players"`                               // 참가자 목록
	CurrentWord             string             `bson:"current_word" json:"current_word"`
	CurrentWordTranslations map[string]string  `bson:"current_word_translations,omitempty" json:"current_word_translations,omitempty"`
	RoundNumber             int                `bson:"round_number" json:"round_number"`   // 현재 라운드 (1-3)
	TimeLeft                int                `bson:"time_left" json:"time_left"`         // 남은 시간 (초)
	GameStatus              string             `bson:"game_status" json:"game_status"`     // waiting, playing, finished
	UsedWords               []string           `bson:"used_words" json:"used_words"`       // 사용된 단어들
	MaxRounds               int                `bson:"max_rounds" json:"max_rounds"`       // 최대 라운드 (3)
	WinningScore            int                `bson:"winning_score" json:"winning_score"` // 우승 점수 (3)
	GameType                GameType           `bson:"game_type" json:"game_type"`
	CreatedAt               time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt               time.Time          `bson:"updated_at" json:"updated_at"`
}

type Player struct {
	Username string `bson:"username" json:"username"`
	Score    int    `bson:"score" json:"score"`
	Attempts int    `bson:"attempts" json:"attempts"` // 이번 라운드 제출 횟수
}

// GameType은 게임 타입을 나타냅니다 (enum 패턴)
type GameType string

const (
	GameTypeOX    GameType = "ox"    // OX 퀴즈
	GameTypeGuess GameType = "guess" // 그림 맞추기
)

// IsValid는 GameType이 유효한 값인지 검증합니다
func (g GameType) IsValid() bool {
	switch g {
	case GameTypeOX, GameTypeGuess:
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
