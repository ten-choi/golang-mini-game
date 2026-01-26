package graph

import (
	"context"
	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/valkey"
	"draw-and-guess-server/pkg/dictionary"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// In-memory storage for game rooms
// Note: For production deployment with multiple instances,
// migrate to distributed storage (Redis/Valkey)
var (
	gameRooms = make(map[string]*model.GameRoom)
	roomMutex sync.RWMutex
)

// publishLobbyUpdate publishes current game rooms to all lobby subscribers
// Only includes active rooms (not empty, not finished)
func publishLobbyUpdate() {
	roomMutex.RLock()
	rooms := make([]*model.GameRoom, 0, len(gameRooms))
	for _, room := range gameRooms {
		// Skip empty rooms and finished games
		if len(room.Users) > 0 && room.Status != model.GameStatusFinished {
			rooms = append(rooms, room)
		}
	}
	roomMutex.RUnlock()

	log.Printf("[publishLobbyUpdate] Publishing %d active rooms (total in memory: %d)", len(rooms), len(gameRooms))

	// Publish to GraphQL subscribers
	GetPubSub().PublishLobbyUpdate(rooms)

	// Also publish to WebSocket clients via Valkey
	publishLobbyUpdateToWebSocket(rooms)
}

// GetGameRoom returns a game room by ID (thread-safe read)
func GetGameRoom(roomID string) (*model.GameRoom, bool) {
	roomMutex.RLock()
	defer roomMutex.RUnlock()
	room, exists := gameRooms[roomID]
	return room, exists
}

// RemoveUserFromRoom removes a user from a game room
// This is called when a WebSocket connection is closed
func RemoveUserFromRoom(roomID, userID string) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		log.Printf("[RemoveUserFromRoom] Room %s not found for removing user %s", roomID, userID)
		return
	}

	// Remove user
	newUsers := []*model.GameUser{}
	var leavingUser *model.GameUser
	for _, p := range room.Users {
		if p.UserID != userID {
			newUsers = append(newUsers, p)
		} else {
			leavingUser = p
		}
	}

	// No change if user wasn't in room
	if leavingUser == nil {
		return
	}

	room.Users = newUsers

	// Don't delete WAITING rooms even if empty - allow host to refresh
	// Only delete if room was in PLAYING or FINISHED state
	if len(room.Users) == 0 {
		if room.Status != model.GameStatusWaiting {
			delete(gameRooms, roomID)
			log.Printf("[RemoveUserFromRoom] Room %s deleted (empty, status: %s)", roomID, room.Status)
			go publishLobbyUpdate()
			return
		}
		log.Printf("[RemoveUserFromRoom] Room %s is empty but WAITING, keeping alive", roomID)
		// Room remains but with no users - they can rejoin
	}

	// Reassign host if needed
	if room.HostUserID == userID && len(room.Users) > 0 {
		room.HostUserID = room.Users[0].UserID
		log.Printf("[RemoveUserFromRoom] New host assigned in room %s: %s", roomID, room.HostUserID)
	}

	gameRooms[roomID] = room

	// Publish events - all in goroutine to avoid deadlock
	log.Printf("[RemoveUserFromRoom] User %s removed from room %s (WebSocket disconnect)", userID, roomID)

	go func() {
		publishLobbyUpdate()
		GetPubSub().PublishRoomUpdate(room)
		GetPubSub().PublishUserLeft(roomID, leavingUser)
		publishRoomUpdateToWebSocket(roomID, room)
	}()
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
		// Skip empty rooms and finished games
		if len(room.Users) == 0 || room.Status == model.GameStatusFinished {
			continue
		}
		if gameType == nil || room.GameType == *gameType {
			result = append(result, room)
		}
	}
	return result
}

// publishLobbyUpdateToWebSocket publishes lobby update to WebSocket clients via Valkey
func publishLobbyUpdateToWebSocket(rooms []*model.GameRoom) {
	channel := common.ChannelLobby

	// Create the message payload
	payload := map[string]interface{}{
		"type": "lobby_update",
		"data": rooms,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal lobby update: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish lobby update to WebSocket: %v", err)
	} else {
		log.Printf("Published lobby update to WebSocket (channel: lobby, %d rooms)", len(rooms))
	}
}

// CleanupStaleRooms removes empty rooms and finished games periodically
// This is a safety measure in case some rooms weren't properly cleaned up
func CleanupStaleRooms() {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	deletedCount := 0
	for roomID, room := range gameRooms {
		shouldDelete := false
		reason := ""

		// Delete empty rooms
		if len(room.Users) == 0 {
			shouldDelete = true
			reason = "empty"
		}

		// Delete finished games older than 5 minutes
		if room.Status == model.GameStatusFinished {
			shouldDelete = true
			reason = "finished"
		}

		if shouldDelete {
			delete(gameRooms, roomID)
			deletedCount++
			log.Printf("[CleanupStaleRooms] Deleted room %s (%s, status: %s, users: %d)",
				roomID, reason, room.Status, len(room.Users))
		}
	}

	if deletedCount > 0 {
		log.Printf("[CleanupStaleRooms] Cleaned up %d stale rooms", deletedCount)
		// Publish lobby update after cleanup
		go publishLobbyUpdate()
	}
}

// StartRoomCleanupScheduler starts a background goroutine that periodically cleans up stale rooms
func StartRoomCleanupScheduler() {
	ticker := time.NewTicker(30 * time.Second) // Run every 30 seconds
	go func() {
		log.Printf("[StartRoomCleanupScheduler] Room cleanup scheduler started")
		for range ticker.C {
			CleanupStaleRooms()
		}
	}()
}

// publishQuizToWebSocket publishes a quiz to users in a game room via WebSocket
func publishQuizToWebSocket(roomID string, quiz interface{}) {
	channel := common.ChannelGamePrefix + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type": "quiz",
		"quiz": quiz,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal quiz: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish quiz to WebSocket: %v", err)
	} else {
		log.Printf("Published quiz to game channel: %s", channel)
	}
}

// publishTimerToWebSocket publishes the remaining time to users in a game room via WebSocket
func publishTimerToWebSocket(roomID string, timeLeft int) {
	channel := common.ChannelGamePrefix + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type":     "timer",
		"timeLeft": timeLeft,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal timer: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish timer to WebSocket: %v", err)
	}
}

// publishRoundEndToWebSocket publishes round end message to WebSocket clients
func publishRoundEndToWebSocket(roomID string, roundNumber int, reason string) {
	channel := common.ChannelGamePrefix + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type": "round_end",
		"data": map[string]interface{}{
			"round":  roundNumber,
			"reason": reason,
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal round end: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish round end to WebSocket: %v", err)
	} else {
		log.Printf("Published round %d end to WebSocket: %s", roundNumber, reason)
	}
}

// publishRoomDeletedToWebSocket publishes room deletion notification to WebSocket clients
func publishRoomDeletedToWebSocket(roomID string) {
	channel := common.ChannelGamePrefix + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type": "room_deleted",
		"data": map[string]interface{}{
			"roomId": roomID,
			"reason": "Room has been deleted",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal room deleted message: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish room deletion to WebSocket: %v", err)
	} else {
		log.Printf("Published room deletion notification for room: %s", roomID)
	}
}

// publishGameEndToWebSocket publishes game end message with final scores to WebSocket clients
func publishGameEndToWebSocket(roomID string, room *model.GameRoom) {
	channel := common.ChannelGamePrefix + roomID

	// Prepare user scores sorted by score
	type UserScore struct {
		UserID string `json:"userID"`
		Score  int32  `json:"score"`
	}

	scores := make([]UserScore, len(room.Users))
	for i, user := range room.Users {
		scores[i] = UserScore{
			UserID: user.UserID,
			Score:  user.Score,
		}
	}

	// Create the message payload
	payload := map[string]interface{}{
		"type": "game_end",
		"data": map[string]interface{}{
			"totalRounds": room.TotalRounds,
			"users":       scores,
			"message":     "??? ???????! ?? ??? ?????",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal game end: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish game end to WebSocket: %v", err)
	} else {
		log.Printf("Published game end to WebSocket for room: %s", roomID)
	}
}

// publishWordchainPromptToWebSocket publishes a wordchain prompt to users via WebSocket
func publishWordchainPromptToWebSocket(roomID string, prompt *model.WordchainPrompt, lastWord string) {
	channel := "game/" + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type":     "wordchain_prompt",
		"prompt":   prompt,
		"lastWord": lastWord,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal wordchain prompt: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish wordchain prompt to WebSocket: %v", err)
	} else {
		log.Printf("Published wordchain prompt to game channel: %s", channel)
	}
}

// publishWordResultToWebSocket publishes wordchain validation result via WebSocket
func publishWordResultToWebSocket(roomID string, userName string, word string, correct bool, reason string) {
	channel := "game/" + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type":     "word_result",
		"userName": userName,
		"word":     word,
		"correct":  correct,
		"reason":   reason,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal word result: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish word result to WebSocket: %v", err)
	} else {
		log.Printf("Published word result to game channel: %s (correct: %v)", channel, correct)
	}
}

// PublishWordchainResult publishes wordchain validation result (called from WebSocket handler)
func PublishWordchainResult(roomID string, userName string, word string, correct bool, reason string) {
	publishWordResultToWebSocket(roomID, userName, word, correct, reason)
}

// UpdateWordchainState updates the last word and user score after correct answer
func UpdateWordchainState(roomID string, word string, userID string) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		log.Printf("Room not found for wordchain update: %s", roomID)
		return
	}

	// Update last word
	room.WordchainLastWord = &word

	// Add word to used words list
	if room.WordchainUsedWords == nil {
		room.WordchainUsedWords = []string{}
	}
	room.WordchainUsedWords = append(room.WordchainUsedWords, word)

	// Update user score
	for i, user := range room.Users {
		if user.UserID == userID {
			room.Users[i].Score += 100
			log.Printf("User %s scored 100 points (new score: %d)", userID, room.Users[i].Score)
			break
		}
	}

	// Update the room in the map (important!)
	gameRooms[roomID] = room

	// Publish room update
	go publishRoomUpdateToWebSocket(roomID, room)
}

// IsWordchainDuplicate checks if a word has already been used in the current game
func IsWordchainDuplicate(roomID string, word string) bool {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return false
	}

	for _, usedWord := range room.WordchainUsedWords {
		if usedWord == word {
			return true
		}
	}

	return false
}

// CheckWordchainTurn checks if it's the given user's turn
func CheckWordchainTurn(roomID string, userID string) bool {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return false
	}

	return room.CurrentTurnUserID != nil && *room.CurrentTurnUserID == userID
}

// MoveToNextTurn moves to the next user's turn
func MoveToNextTurn(roomID string) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists || len(room.Users) == 0 {
		return
	}

	// Find current user index
	currentIndex := -1
	for i, user := range room.Users {
		if room.CurrentTurnUserID != nil && user.UserID == *room.CurrentTurnUserID {
			currentIndex = i
			break
		}
	}

	// Move to next user (circular)
	nextIndex := (currentIndex + 1) % len(room.Users)
	nextUserName := room.Users[nextIndex].UserID
	room.CurrentTurnUserID = &nextUserName

	log.Printf("[Wordchain] Turn moved to: %s (index: %d)", *room.CurrentTurnUserID, nextIndex)

	// Update the room in the map (important!)
	gameRooms[roomID] = room

	// Broadcast updated room state with current turn info
	go publishRoomUpdateToWebSocket(roomID, room)

	// Restart timer for new turn
	go StartWordchainTurnTimer(roomID)
}

// EndWordchainRound ends the current round and starts the next one
func EndWordchainRound(roomID string, reason string) {
	roomMutex.Lock()
	room, exists := gameRooms[roomID]
	if !exists {
		roomMutex.Unlock()
		return
	}

	currentRound := int(room.CurrentRound)
	totalRounds := int(room.TotalRounds)

	log.Printf("[Wordchain] Round %d ended: %s", currentRound, reason)

	// Broadcast round end message
	publishRoundEndToWebSocket(roomID, currentRound, reason)

	// Check if game should end
	if currentRound >= totalRounds {
		// Reset room to WAITING state instead of FINISHED
		room.Status = model.GameStatusWaiting
		room.CurrentRound = 0
		room.WordchainUsedWords = []string{}
		emptyStr := ""
		room.WordchainLastWord = &emptyStr
		room.CurrentTurnUserID = &emptyStr
		room.WordchainTurnStartTime = nil

		// Keep users but reset their ready status
		for i := range room.Users {
			room.Users[i].IsReady = false
		}

		// Update the room in the map (important!)
		gameRooms[roomID] = room
		roomMutex.Unlock()

		log.Printf("[Wordchain] Game finished after %d rounds. Room reset to WAITING", totalRounds)

		// Broadcast game end message with final scores
		go publishGameEndToWebSocket(roomID, room)
		go publishRoomUpdateToWebSocket(roomID, room)
		go GetPubSub().PublishRoomUpdate(room)
		go publishLobbyUpdate()
		return
	}

	// Start next round
	room.CurrentRound++
	room.WordchainUsedWords = []string{} // Reset used words
	emptyStr2 := ""
	room.WordchainLastWord = &emptyStr2 // Will be set by new round start
	room.CurrentTurnUserID = &emptyStr2 // Will be set by new round start

	// Update the room in the map (important!)
	gameRooms[roomID] = room
	roomMutex.Unlock()

	log.Printf("[Wordchain] Starting round %d", room.CurrentRound)

	// Small delay before starting next round
	time.Sleep(3 * time.Second)

	// Start new round (same logic as startWordchainGame)
	startWordchainRound(roomID)
}

// startWordchainRound starts a new wordchain round
func startWordchainRound(roomID string) {
	roomMutex.Lock()
	room, exists := gameRooms[roomID]
	if !exists {
		roomMutex.Unlock()
		return
	}

	// Generate initial word for new round
	dict := dictionary.GetInstance()
	initialWord := dict.GetRandomWord()

	room.WordchainLastWord = &initialWord
	if len(room.Users) > 0 {
		firstUserName := room.Users[0].UserID
		room.CurrentTurnUserID = &firstUserName
	}

	// Initialize with initial word
	room.WordchainUsedWords = []string{initialWord}

	// Update the room in the map (important!)
	gameRooms[roomID] = room
	roomMutex.Unlock()

	// Broadcast round start and initial word
	go GetPubSub().PublishRoundStarted(roomID, room)
	publishWordchainPromptToWebSocket(roomID, nil, initialWord)
	publishRoomUpdateToWebSocket(roomID, room)

	// Start turn timer
	StartWordchainTurnTimer(roomID)
}
