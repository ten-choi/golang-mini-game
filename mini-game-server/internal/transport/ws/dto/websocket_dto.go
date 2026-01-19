package dto

import (
	"draw-and-guess-server/internal/graph/model"
)

// WebSocketRequest represents the base request structure for WebSocket messages
// @Description WebSocket 요청 메시지의 기본 구조
type WebSocketRequest struct {
	Type    string      `json:"type" example:"subscribe" description:"메시지 타입: subscribe, unsubscribe, message"`
	Channel string      `json:"channel" example:"game/room-123" description:"채널 이름 (구독/발행 대상)"`
	Data    interface{} `json:"data,omitempty" description:"전송할 데이터 (message 타입인 경우)"`
}

// WebSocketResponse represents the base response structure for WebSocket messages
// @Description WebSocket 응답 메시지의 기본 구조
type WebSocketResponse struct {
	Type    string      `json:"type" example:"message" description:"메시지 타입"`
	Channel string      `json:"channel" example:"game/room-123" description:"메시지가 발행된 채널"`
	Data    interface{} `json:"data" description:"수신된 데이터"`
}

// SubscribeRequest represents a channel subscription request
// @Description 채널 구독 요청
type SubscribeRequest struct {
	Type    string `json:"type" example:"subscribe"`
	Channel string `json:"channel" example:"game/room-abc123" description:"구독할 채널 (예: game/room-{roomId})"`
}

// UnsubscribeRequest represents a channel unsubscription request
// @Description 채널 구독 해제 요청
type UnsubscribeRequest struct {
	Type    string `json:"type" example:"unsubscribe"`
	Channel string `json:"channel" example:"game/room-abc123" description:"구독 해제할 채널"`
}

// ChatMessageData represents chat message data
// @Description 채팅 메시지 데이터
type ChatMessageData struct {
	RoomID   string `json:"roomId" example:"abc123" description:"방 ID"`
	UserID   string `json:"userId" example:"user-001" description:"발신자 ID"`
	Username string `json:"username" example:"홍길동" description:"발신자 닉네임"`
	Message  string `json:"message" example:"안녕하세요!" description:"채팅 메시지 내용"`
}

// GameStateData represents game state update data
// @Description 게임 상태 업데이트 데이터
type GameStateData struct {
	RoomID       string        `json:"roomId" example:"abc123"`
	CurrentRound int           `json:"currentRound" example:"2" description:"현재 라운드"`
	Drawer       string        `json:"drawer" example:"user-001" description:"현재 그림 그리는 플레이어"`
	TimeLeft     int           `json:"timeLeft" example:"45" description:"남은 시간 (초)"`
	Users        []*model.User `json:"users" description:"플레이어 목록"`
}

// DrawingData represents drawing action data
// @Description 그림 그리기 데이터
type DrawingData struct {
	RoomID string      `json:"roomId" example:"abc123"`
	Action string      `json:"action" example:"draw" description:"액션: draw, clear, undo"`
	Points []DrawPoint `json:"points,omitempty" description:"그리기 좌표"`
	Color  string      `json:"color,omitempty" example:"#FF0000" description:"색상 (Hex 코드)"`
	Width  int         `json:"width,omitempty" example:"5" description:"선 두께"`
}

// DrawPoint represents a single point in a drawing
// @Description 그리기 좌표 포인트
type DrawPoint struct {
	X float64 `json:"x" example:"120.5" description:"X 좌표"`
	Y float64 `json:"y" example:"85.3" description:"Y 좌표"`
}

// AnswerSubmitData represents answer submission data
// @Description 정답 제출 데이터
type AnswerSubmitData struct {
	RoomID   string `json:"roomId" example:"abc123"`
	UserID   string `json:"userId" example:"user-002"`
	Username string `json:"username" example:"김철수"`
	Answer   string `json:"answer" example:"사과" description:"제출한 정답"`
}

// CorrectAnswerData represents correct answer notification
// @Description 정답 알림 데이터
type CorrectAnswerData struct {
	RoomID   string `json:"roomId" example:"abc123"`
	UserID   string `json:"userId" example:"user-002"`
	Username string `json:"username" example:"김철수"`
	Answer   string `json:"answer" example:"사과" description:"정답"`
	Score    int    `json:"score" example:"100" description:"획득 점수"`
}

// RoundStartData represents round start notification
// @Description 라운드 시작 알림 데이터
type RoundStartData struct {
	RoomID    string `json:"roomId" example:"abc123"`
	Round     int    `json:"round" example:"3" description:"라운드 번호"`
	Drawer    string `json:"drawer" example:"user-001" description:"그림 그맴 플레이어"`
	Topic     string `json:"topic,omitempty" example:"사과" description:"주제 (그림 그리는 사람만 받음)"`
	TopicHint string `json:"topicHint" example:"과일 (2자)" description:"주제 힌트"`
	TimeLimit int    `json:"timeLimit" example:"60" description:"제한 시간 (초)"`
}

// RoundEndData represents round end notification
// @Description 라운드 종료 알림 데이터
type RoundEndData struct {
	RoomID     string        `json:"roomId" example:"abc123"`
	Round      int           `json:"round" example:"3"`
	Topic      string        `json:"topic" example:"사과" description:"정답"`
	Winners    []string      `json:"winners" description:"정답 맞춘 플레이어 목록"`
	Scoreboard []*model.User `json:"scoreboard" description:"현재 점수판"`
}

// GameEndData represents game end notification
// @Description 게임 종료 알림 데이터
type GameEndData struct {
	RoomID     string        `json:"roomId" example:"abc123"`
	Winner     *model.User   `json:"winner" description:"우승자"`
	FinalScore []*model.User `json:"finalScore" description:"최종 점수판"`
}

// UserJoinedData represents user join notification
// @Description 플레이어 입장 알림 데이터
type UserJoinedData struct {
	RoomID   string `json:"roomId" example:"abc123"`
	UserID   string `json:"userId" example:"user-003"`
	Username string `json:"username" example:"이영희"`
}

// UserLeftData represents user leave notification
// @Description 플레이어 퇴장 알림 데이터
type UserLeftData struct {
	RoomID   string `json:"roomId" example:"abc123"`
	UserID   string `json:"userId" example:"user-003"`
	Username string `json:"username" example:"이영희"`
}

// ErrorResponse represents WebSocket error response
// @Description WebSocket 에러 응답
type ErrorResponse struct {
	Type    string       `json:"type" example:"error"`
	Channel string       `json:"channel" example:"game/room-abc123"`
	Data    ErrorMessage `json:"data"`
}

// ErrorMessage represents error details
// @Description 에러 상세 정보
type ErrorMessage struct {
	Code    string `json:"code" example:"INVALID_CHANNEL" description:"에러 코드"`
	Message string `json:"message" example:"Invalid channel format" description:"에러 메시지"`
}
