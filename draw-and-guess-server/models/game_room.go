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
	CreatedAt               time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt               time.Time          `bson:"updated_at" json:"updated_at"`
}

type Player struct {
	Username string `bson:"username" json:"username"`
	Score    int    `bson:"score" json:"score"`
	Attempts int    `bson:"attempts" json:"attempts"` // 이번 라운드 제출 횟수
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
