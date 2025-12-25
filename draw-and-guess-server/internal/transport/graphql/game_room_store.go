package graph

import (
	"sync"

	"draw-and-guess-server/internal/graph/model"
)

// In-memory storage for game rooms
// TODO: Move to Redis/Valkey for production use
var (
	gameRooms = make(map[string]*model.GameRoom)
	roomMutex sync.RWMutex
)

// publishLobbyUpdate publishes current game rooms to all lobby subscribers
func publishLobbyUpdate() {
	roomMutex.RLock()
	rooms := make([]*model.GameRoom, 0, len(gameRooms))
	for _, room := range gameRooms {
		rooms = append(rooms, room)
	}
	roomMutex.RUnlock()

	GetPubSub().PublishLobbyUpdate(rooms)
}

// GetGameRoom returns a game room by ID (thread-safe read)
func GetGameRoom(roomID string) (*model.GameRoom, bool) {
	roomMutex.RLock()
	defer roomMutex.RUnlock()
	room, exists := gameRooms[roomID]
	return room, exists
}

// SetGameRoom stores a game room (thread-safe write)
func SetGameRoom(roomID string, room *model.GameRoom) {
	roomMutex.Lock()
	defer roomMutex.Unlock()
	gameRooms[roomID] = room
}

// DeleteGameRoomFromStore removes a game room (thread-safe)
func DeleteGameRoomFromStore(roomID string) {
	roomMutex.Lock()
	defer roomMutex.Unlock()
	delete(gameRooms, roomID)
}

// GetAllGameRooms returns all game rooms (thread-safe)
func GetAllGameRooms() []*model.GameRoom {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	rooms := make([]*model.GameRoom, 0, len(gameRooms))
	for _, room := range gameRooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// GetGameRoomsByType returns game rooms filtered by type (thread-safe)
func GetGameRoomsByType(gameType *model.GameType) []*model.GameRoom {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	result := []*model.GameRoom{}
	for _, room := range gameRooms {
		if gameType == nil || room.GameType == *gameType {
			result = append(result, room)
		}
	}
	return result
}
