// Package models contains legacy models for WebSocket-based game logic
// NOTE: For GraphQL, use internal/graph/model instead
package models

import "time"

// LegacyGameRoom is used for WebSocket-based game state management
// This is separate from the GraphQL model.GameRoom
type LegacyGameRoom struct {
	ID                      string            `json:"id"`
	UUID                    string            `json:"uuid"`
	IsActive                bool              `json:"is_active"`
	DrawerUser              string            `json:"drawer_user"`                 // 현재 그림 그리는 사람
	RoomCreator             string            `json:"room_creator"`                // 방 생성자
	LastRoundWinner         string            `json:"last_round_winner,omitempty"` // 이전 라운드 우승자
	Players                 []LegacyPlayer    `json:"players"`                     // 참가자 목록
	CurrentWord             string            `json:"current_word"`
	CurrentWordTranslations map[string]string `json:"current_word_translations,omitempty"`
	RoundNumber             int               `json:"round_number"`  // 현재 라운드 (1-3)
	TimeLeft                int               `json:"time_left"`     // 남은 시간 (초)
	GameStatus              string            `json:"game_status"`   // waiting, playing, finished
	UsedWords               []string          `json:"used_words"`    // 사용된 단어들
	MaxRounds               int               `json:"max_rounds"`    // 최대 라운드 (3)
	WinningScore            int               `json:"winning_score"` // 우승 점수 (3)
	MaxPlayers              int               `json:"max_players"`   // 최대 플레이어 수 (4)
	GameType                string            `json:"game_type"`     // 게임 타입 (ox, qa, wordchain)
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
}

// LegacyPlayer represents a player in WebSocket game context
type LegacyPlayer struct {
	Username string `json:"username"`
	Score    int    `json:"score"`
	Attempts int    `json:"attempts"` // 이번 라운드 제출 횟수
}

// GameTopic represents a game topic with translations
type GameTopic struct {
	Canonical    string            `json:"canonical"`
	Translations map[string]string `json:"translations"`
}

// GameUser represents a user in game context
type GameUser struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Role     string `json:"role"` // drawer, player
}
