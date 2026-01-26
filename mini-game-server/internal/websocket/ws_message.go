package websocket

import (
	"encoding/json"
)

// WSSuccessMessage represents a successful WebSocket message format
// Format: { "type": "DRAW_EVENT", "payload": { ... } }
type WSSuccessMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSErrorMessage represents an error WebSocket message format
// Format: { "type": "ERROR", "code": "UNAUTHORIZED", "message": "token expired" }
type WSErrorMessage struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GameEventPayload represents game state change data
type GameEventPayload struct {
	EventType string      `json:"eventType"` // "user_joined", "user_left", "game_started", "round_started"
	RoomID    string      `json:"roomId"`
	Data      interface{} `json:"data"`
}

// ChatMessagePayload는 채팅 메시지 데이터
type ChatMessagePayload struct {
	RoomID   string `json:"roomId"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// DrawingEventPayload는 그림 그리기 데이터
type DrawingEventPayload struct {
	RoomID    string      `json:"roomId"`
	UserID    string      `json:"userId"`
	StrokeID  string      `json:"strokeId,omitempty"`
	Points    []DrawPoint `json:"points,omitempty"`
	Color     string      `json:"color,omitempty"`
	LineWidth float64     `json:"lineWidth,omitempty"`
	Action    string      `json:"action"` // "start", "draw", "end", "clear"
}

// DrawPoint는 그리기에서 하나의 점을 나타냄
type DrawPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BuildWSSuccess는 성공 WebSocket 메시지를 생성
func BuildWSSuccess(msgType string, payload interface{}) []byte {
	msg := WSSuccessMessage{
		Type:    msgType,
		Payload: payload,
	}
	bytes, _ := json.Marshal(msg)
	return bytes
}

// BuildWSError는 에러 WebSocket 메시지를 생성
func BuildWSError(code string, message string) []byte {
	msg := WSErrorMessage{
		Type:    "ERROR",
		Code:    code,
		Message: message,
	}
	bytes, _ := json.Marshal(msg)
	return bytes
}

// BuildGameEvent는 게임 이벤트 WebSocket 메시지를 생성
func BuildGameEvent(roomID string, eventType string, data interface{}) []byte {
	payload := GameEventPayload{
		EventType: eventType,
		RoomID:    roomID,
		Data:      data,
	}
	return BuildWSSuccess("GAME_EVENT", payload)
}

// BuildChatMessage는 채팅 WebSocket 메시지를 생성
func BuildChatMessage(roomID, userID, userName, message string) []byte {
	payload := ChatMessagePayload{
		RoomID:   roomID,
		UserID:   userID,
		Username: userName,
		Message:  message,
	}
	return BuildWSSuccess("CHAT_MESSAGE", payload)
}

// BuildDrawingEvent는 그리기 WebSocket 메시지를 생성
func BuildDrawingEvent(roomID, userID, action string, data interface{}) []byte {
	payload := DrawingEventPayload{
		RoomID: roomID,
		UserID: userID,
		Action: action,
	}

	// Add additional fields based on action and data type
	if drawData, ok := data.(map[string]interface{}); ok {
		if strokeID, ok := drawData["strokeId"].(string); ok {
			payload.StrokeID = strokeID
		}
		if color, ok := drawData["color"].(string); ok {
			payload.Color = color
		}
		if lineWidth, ok := drawData["lineWidth"].(float64); ok {
			payload.LineWidth = lineWidth
		}
		if points, ok := drawData["points"].([]DrawPoint); ok {
			payload.Points = points
		}
	}

	return BuildWSSuccess("DRAW_EVENT", payload)
}
