// Package domain contains pure business models for WebSocket-based game logic
// NOTE: For GraphQL, use internal/transport/graphql/model instead
package domain

import "time"

// GameRoom is used for WebSocket-based game state management
// This is separate from the GraphQL model.GameRoom
type GameRoom struct {
	ID                      string            `json:"id"`
	UUID                    string            `json:"uuid"`
	IsActive                bool              `json:"is_active"`
	DrawerUser              string            `json:"drawer_user"`
	RoomCreator             string            `json:"room_creator"`
	LastRoundWinner         string            `json:"last_round_winner,omitempty"`
	Players                 []Player          `json:"players"`
	CurrentWord             string            `json:"current_word"`
	CurrentWordTranslations map[string]string `json:"current_word_translations,omitempty"`
	RoundNumber             int               `json:"round_number"`
	TimeLeft                int               `json:"time_left"`
	GameStatus              string            `json:"game_status"` // waiting, playing, finished
	UsedWords               []string          `json:"used_words"`
	MaxRounds               int               `json:"max_rounds"`
	WinningScore            int               `json:"winning_score"`
	MaxPlayers              int               `json:"max_players"`
	GameType                string            `json:"game_type"` // ox, qa, wordchain
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
}

// Player represents a player in WebSocket game context
type Player struct {
	Username string `json:"username"`
	Score    int    `json:"score"`
	Attempts int    `json:"attempts"`
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
