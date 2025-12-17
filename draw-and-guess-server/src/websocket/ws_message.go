package websocket

import (
	"encoding/json"
)

// WSSuccessMessage represents a successful WebSocket message
// Format: { "type": "DRAW_EVENT", "payload": { ... } }
type WSSuccessMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSErrorMessage represents an error WebSocket message
// Format: { "type": "ERROR", "code": "UNAUTHORIZED", "message": "token expired" }
type WSErrorMessage struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GameEventPayload represents game state change data
type GameEventPayload struct {
	EventType string      `json:"eventType"` // "player_joined", "player_left", "game_started", "round_started"
	RoomID    string      `json:"roomId"`
	Data      interface{} `json:"data"`
}

// ChatMessagePayload represents a chat message
type ChatMessagePayload struct {
	RoomID     string `json:"roomId"`
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
	Message    string `json:"message"`
}

// DrawingEventPayload represents drawing stroke data
type DrawingEventPayload struct {
	RoomID    string      `json:"roomId"`
	PlayerID  string      `json:"playerId"`
	StrokeID  string      `json:"strokeId,omitempty"`
	Points    []DrawPoint `json:"points,omitempty"`
	Color     string      `json:"color,omitempty"`
	LineWidth float64     `json:"lineWidth,omitempty"`
	Action    string      `json:"action"` // "start", "draw", "end", "clear"
}

// DrawPoint represents a single point in a drawing stroke
type DrawPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BuildWSSuccess creates a success WebSocket message
func BuildWSSuccess(msgType string, payload interface{}) []byte {
	msg := WSSuccessMessage{
		Type:    msgType,
		Payload: payload,
	}
	bytes, _ := json.Marshal(msg)
	return bytes
}

// BuildWSError creates an error WebSocket message
func BuildWSError(code string, message string) []byte {
	msg := WSErrorMessage{
		Type:    "ERROR",
		Code:    code,
		Message: message,
	}
	bytes, _ := json.Marshal(msg)
	return bytes
}

// BuildGameEvent creates a game event WebSocket message
func BuildGameEvent(roomID string, eventType string, data interface{}) []byte {
	payload := GameEventPayload{
		EventType: eventType,
		RoomID:    roomID,
		Data:      data,
	}
	return BuildWSSuccess("GAME_EVENT", payload)
}

// BuildChatMessage creates a chat WebSocket message
func BuildChatMessage(roomID, playerID, playerName, message string) []byte {
	payload := ChatMessagePayload{
		RoomID:     roomID,
		PlayerID:   playerID,
		PlayerName: playerName,
		Message:    message,
	}
	return BuildWSSuccess("CHAT_MESSAGE", payload)
}

// BuildDrawingEvent creates a drawing WebSocket message
func BuildDrawingEvent(roomID, playerID, action string, data interface{}) []byte {
	payload := DrawingEventPayload{
		RoomID:   roomID,
		PlayerID: playerID,
		Action:   action,
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
