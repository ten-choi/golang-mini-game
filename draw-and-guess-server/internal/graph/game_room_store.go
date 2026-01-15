package graph

import (
	"context"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/valkey"
	"draw-and-guess-server/pkg/dictionary"
	"encoding/json"
	"log"
	"sync"
	"time"
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

// RemovePlayerFromRoom removes a player from a game room
// This is called when a WebSocket connection is closed
func RemovePlayerFromRoom(roomID, username string) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		log.Printf("[RemovePlayerFromRoom] Room %s not found for removing player %s", roomID, username)
		return
	}

	// Remove player
	newPlayers := []*model.Player{}
	var leavingPlayer *model.Player
	for _, p := range room.Players {
		if p.Username != username {
			newPlayers = append(newPlayers, p)
		} else {
			leavingPlayer = p
		}
	}

	// No change if player wasn't in room
	if leavingPlayer == nil {
		return
	}

	room.Players = newPlayers

	// Don't delete WAITING rooms even if empty - allow host to refresh
	// Only delete if room was in PLAYING or FINISHED state
	if len(room.Players) == 0 {
		if room.Status != model.GameStatusWaiting {
			delete(gameRooms, roomID)
			log.Printf("[RemovePlayerFromRoom] Room %s deleted (empty, status: %s)", roomID, room.Status)
			go publishLobbyUpdate()
			return
		}
		log.Printf("[RemovePlayerFromRoom] Room %s is empty but WAITING, keeping alive", roomID)
		// Room remains but with no players - they can rejoin
	}

	// Reassign host if needed
	if room.HostUsername == username && len(room.Players) > 0 {
		room.HostUsername = room.Players[0].Username
		log.Printf("[RemovePlayerFromRoom] New host assigned in room %s: %s", roomID, room.HostUsername)
	}

	gameRooms[roomID] = room

	// Publish events
	log.Printf("[RemovePlayerFromRoom] Player %s removed from room %s (WebSocket disconnect)", username, roomID)
	go func() {
		GetPubSub().PublishRoomUpdate(room)
		GetPubSub().PublishPlayerLeft(roomID, leavingPlayer)
		publishLobbyUpdate()
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
		if gameType == nil || room.GameType == *gameType {
			result = append(result, room)
		}
	}
	return result
}

// publishLobbyUpdateToWebSocket publishes lobby update to WebSocket clients via Valkey
func publishLobbyUpdateToWebSocket(rooms []*model.GameRoom) {
	channel := "lobby"

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

// publishQuizToWebSocket publishes a quiz to players in a game room via WebSocket
func publishQuizToWebSocket(roomID string, quiz interface{}) {
	channel := "game/" + roomID

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

// publishTimerToWebSocket publishes the remaining time to players in a game room via WebSocket
func publishTimerToWebSocket(roomID string, timeLeft int) {
	channel := "game/" + roomID

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
	channel := "game/" + roomID

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

// publishGameEndToWebSocket publishes game end message with final scores to WebSocket clients
func publishGameEndToWebSocket(roomID string, room *model.GameRoom) {
	channel := "game/" + roomID

	// Prepare player scores sorted by score
	type PlayerScore struct {
		Username string `json:"username"`
		Score    int32  `json:"score"`
	}

	scores := make([]PlayerScore, len(room.Players))
	for i, player := range room.Players {
		scores[i] = PlayerScore{
			Username: player.Username,
			Score:    player.Score,
		}
	}

	// Create the message payload
	payload := map[string]interface{}{
		"type": "game_end",
		"data": map[string]interface{}{
			"totalRounds": room.TotalRounds,
			"players":     scores,
			"message":     "게임이 종료되었습니다! 최종 점수를 확인하세요.",
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

// publishWordchainPromptToWebSocket publishes a wordchain prompt to players via WebSocket
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
func publishWordResultToWebSocket(roomID string, playerName string, word string, correct bool, reason string) {
	channel := "game/" + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type":       "word_result",
		"playerName": playerName,
		"word":       word,
		"correct":    correct,
		"reason":     reason,
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
func PublishWordchainResult(roomID string, playerName string, word string, correct bool, reason string) {
	publishWordResultToWebSocket(roomID, playerName, word, correct, reason)
}

// UpdateWordchainState updates the last word and player score after correct answer
func UpdateWordchainState(roomID string, word string, username string) {
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

	// Update player score
	for i, player := range room.Players {
		if player.Username == username {
			room.Players[i].Score += 100
			log.Printf("Player %s scored 100 points (new score: %d)", username, room.Players[i].Score)
			break
		}
	}

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
func CheckWordchainTurn(roomID string, username string) bool {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return false
	}

	return room.CurrentTurnUsername != nil && *room.CurrentTurnUsername == username
}

// MoveToNextTurn moves to the next player's turn
func MoveToNextTurn(roomID string) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists || len(room.Players) == 0 {
		return
	}

	// Find current player index
	currentIndex := -1
	for i, player := range room.Players {
		if room.CurrentTurnUsername != nil && player.Username == *room.CurrentTurnUsername {
			currentIndex = i
			break
		}
	}

	// Move to next player (circular)
	nextIndex := (currentIndex + 1) % len(room.Players)
	nextUsername := room.Players[nextIndex].Username
	room.CurrentTurnUsername = &nextUsername

	log.Printf("[Wordchain] Turn moved to: %s (index: %d)", room.CurrentTurnUsername, nextIndex)

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
		room.CurrentTurnUsername = &emptyStr
		room.WordchainTurnStartTime = nil

		// Keep players but reset their ready status
		for i := range room.Players {
			room.Players[i].IsReady = false
		}

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
	room.WordchainLastWord = &emptyStr2   // Will be set by new round start
	room.CurrentTurnUsername = &emptyStr2 // Will be set by new round start
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
	if len(room.Players) > 0 {
		firstUsername := room.Players[0].Username
		room.CurrentTurnUsername = &firstUsername
	}

	// Initialize with initial word
	room.WordchainUsedWords = []string{initialWord}
	roomMutex.Unlock()

	// Broadcast round start and initial word
	go GetPubSub().PublishRoundStarted(roomID, room)
	publishWordchainPromptToWebSocket(roomID, nil, initialWord)
	publishRoomUpdateToWebSocket(roomID, room)

	// Start turn timer
	StartWordchainTurnTimer(roomID)
}
