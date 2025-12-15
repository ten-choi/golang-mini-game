package handlers

import (
	"context"
	"draw-and-guess-server/database"
	"draw-and-guess-server/models"
	"draw-and-guess-server/valkey"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var topics []models.GameTopic

const (
	roomKeyPrefix = "game_room:"
	roomTTL       = 3600 // 1 hour in seconds
)

// manageGameTimer updates the room timer every second and broadcasts via Valkey Pub/Sub
func manageGameTimer(roomID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var room models.GameRoom
		if err := valkey.GetJSON(roomKeyPrefix+roomID, &room); err != nil {
			log.Printf("Timer: Room %s not found, stopping timer", roomID)
			return
		}

		if room.GameStatus != "playing" {
			log.Printf("Timer: Game %s not playing, stopping timer", roomID)
			return
		}

		room.TimeLeft--

		if room.TimeLeft <= 0 {
			// 시간 초과 시: 정답자가 없으므로 현재 drawer 유지 (LastRoundWinner를 설정하지 않음)
			if err := advanceToNextRound(&room); err != nil {
				log.Printf("Timer: %v", err)
				room.GameStatus = "finished"
				room.TimeLeft = 0
			}
		}

		if err := persistRoom(roomID, &room); err != nil {
			log.Printf("Timer: Failed to update room %s: %v", roomID, err)
			return
		}

		broadcastGame(roomID, map[string]interface{}{
			"type": "timer_update",
			"data": gin.H{
				"room_id":      roomID,
				"time_left":    room.TimeLeft,
				"round_number": room.RoundNumber,
				"game_status":  room.GameStatus,
			},
		})

		if room.GameStatus == "finished" {
			log.Printf("Timer: Game %s finished", roomID)
			return
		}
	}
}

// GetGameRooms godoc
// @Summary Get game rooms
// @Description Get all active game rooms or a specific room by ID
// @Tags game-rooms
// @Accept json
// @Produce json
// @Param id query string false "Room ID (UUID)"
// @Success 200 {object} models.ApiResult{result=[]models.GameRoom}
// @Failure 404 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/rooms [get]
func GetGameRooms(c *gin.Context) {
	id := c.Query("id")

	if id != "" {
		// Get specific room
		var room models.GameRoom
		err := valkey.GetJSON(roomKeyPrefix+id, &room)
		if err != nil {
			respondError(c, http.StatusNotFound, "Room not found")
			return
		}

		respondSuccess(c, "Get Game Room", []models.GameRoom{room})
		return
	}

	// Get all active rooms
	keys, err := valkey.GetKeys(roomKeyPrefix + "*")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch game rooms")
		return
	}

	var rooms []models.GameRoom
	for _, key := range keys {
		var room models.GameRoom
		err := valkey.GetJSON(key, &room)
		if err == nil && room.IsActive {
			rooms = append(rooms, room)
		}
	}

	respondSuccess(c, "Get Game Rooms", rooms)
}

// CreateGameRoom godoc
// @Summary Create a new game room
// @Description Create a new game room with the specified drawer
// @Tags game-rooms
// @Accept x-www-form-urlencoded
// @Produce json
// @Param ldap_user formData string true "Username of the room creator"
// @Success 200 {object} models.ApiResult{result=models.GameRoom}
// @Failure 400 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/room [post]
func CreateGameRoom(c *gin.Context) {
	ldapUser := c.PostForm("ldap_user")
	if ldapUser == "" {
		respondError(c, http.StatusBadRequest, "ldap_user is missing")
		return
	}

	roomUUID := uuid.New().String()

	topic, err := selectNextWord(nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "No drawing topics are available. Please check MongoDB data.")
		return
	}

	room := models.GameRoom{
		UUID:         roomUUID,
		IsActive:     true,
		DrawerUser:   ldapUser,
		RoomCreator:  ldapUser,
		Players:      []models.Player{},
		RoundNumber:  1,
		TimeLeft:     60,
		GameStatus:   "waiting",
		UsedWords:    []string{},
		MaxRounds:    3,
		WinningScore: 3,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	applyTopicToRoom(&room, topic)

	if err := persistRoom(roomUUID, &room); err != nil {
		log.Printf("Failed to create game room: %v", err)
		respondError(c, http.StatusInternalServerError, "Failed to create game room")
		return
	}

	log.Printf("Created game room with ID: %s", roomUUID)

	respondSuccess(c, "Insert Game Room", map[string]interface{}{
		"room_id":                   roomUUID,
		"current_word":              room.CurrentWord,
		"current_word_translations": room.CurrentWordTranslations,
	})
}

// JoinGameRoom godoc
// @Summary Join a game room
// @Description Join an existing game room
// @Tags game-rooms
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path string true "Room ID (UUID)"
// @Param ldap_user formData string true "Username of the player joining"
// @Success 200 {object} models.ApiResult{result=models.GameRoom}
// @Failure 400 {object} models.ApiResult
// @Failure 404 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/room/{id}/join [post]
func JoinGameRoom(c *gin.Context) {
	id := c.Param("id")
	username := c.PostForm("username")

	if id == "" || username == "" {
		respondError(c, http.StatusBadRequest, "id or username is missing")
		return
	}

	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		respondError(c, http.StatusNotFound, "Game room not found")
		return
	}

	if findPlayerIndex(room.Players, username) != -1 {
		respondError(c, http.StatusBadRequest, "Already joined")
		return
	}

	// 새 플레이어 추가
	newPlayer := models.Player{
		Username: username,
		Score:    0,
		Attempts: 0,
	}
	room.Players = append(room.Players, newPlayer)
	room.UpdatedAt = time.Now()

	if err := persistRoom(id, &room); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to join room")
		return
	}

	respondSuccess(c, "Joined Game Room", nil)
}

// StartGame godoc
// @Summary Start a game
// @Description Start the game in a room (host only)
// @Tags game-rooms
// @Accept json
// @Produce json
// @Param id path string true "Room ID (UUID)"
// @Success 200 {object} models.ApiResult{result=models.GameRoom}
// @Failure 400 {object} models.ApiResult
// @Failure 403 {object} models.ApiResult
// @Failure 404 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/room/{id}/start [post]
func StartGame(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		respondError(c, http.StatusBadRequest, "id is missing")
		return
	}

	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		respondError(c, http.StatusNotFound, "Game room not found")
		return
	}

	// 최소 2명 이상 필요
	if len(room.Players) < 1 {
		respondError(c, http.StatusBadRequest, "Need at least 1 player to start")
		return
	}

	// Update room directly
	room.GameStatus = "playing"
	room.TimeLeft = 60
	room.UpdatedAt = time.Now()

	if err := persistRoom(id, &room); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to start game")
		return
	}

	broadcastGame(id, map[string]interface{}{
		"type":        "update",
		"game_status": "playing",
	})

	// Start timer management goroutine
	go manageGameTimer(id)

	respondSuccess(c, "Game started", map[string]interface{}{
		"game_status":  "playing",
		"round_number": room.RoundNumber,
	})
}

// CheckAnswer godoc
// @Summary Check answer
// @Description Check if the submitted answer is correct
// @Tags game-rooms
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path string true "Room ID (UUID)"
// @Param ldap_user formData string true "Username"
// @Param answer formData string true "Answer to check"
// @Success 200 {object} models.ApiResult{result=models.GameRoom}
// @Failure 400 {object} models.ApiResult
// @Failure 404 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/room/{id}/answer [post]
func CheckAnswer(c *gin.Context) {
	id := c.Param("id")
	answer := c.PostForm("answer")
	username := c.PostForm("username")

	if id == "" || answer == "" || username == "" {
		respondError(c, http.StatusBadRequest, "id, answer or username is missing")
		return
	}

	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		respondError(c, http.StatusNotFound, "Game room not found")
		return
	}

	playerIndex := findPlayerIndex(room.Players, username)
	if playerIndex == -1 {
		respondError(c, http.StatusBadRequest, "Player not found")
		return
	}

	// 3회 제출 제한 확인
	if room.Players[playerIndex].Attempts >= 3 {
		respondError(c, http.StatusBadRequest, "Maximum attempts reached (3/3)")
		return
	}

	// 시도 횟수 증가
	room.Players[playerIndex].Attempts++

	isCorrect := isAnswerCorrect(answer, &room)

	if isCorrect {
		room.Players[playerIndex].Score += 2
		room.LastRoundWinner = username // 정답자를 다음 라운드 drawer로 설정
		gameFinished := room.Players[playerIndex].Score >= room.WinningScore

		if !gameFinished && room.RoundNumber < room.MaxRounds {
			if err := advanceToNextRound(&room); err != nil {
				respondError(c, http.StatusInternalServerError, "No drawing topics are available. Please check MongoDB data.")
				return
			}
		}

		if gameFinished || room.RoundNumber > room.MaxRounds {
			room.GameStatus = "finished"
		}

		if err := persistRoom(id, &room); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to update game state")
			return
		}

		respondSuccess(c, "Correct answer!", map[string]interface{}{
			"is_correct":                true,
			"current_word":              room.CurrentWord,
			"current_word_translations": room.CurrentWordTranslations,
			"player_score":              room.Players[playerIndex].Score,
			"game_finished":             gameFinished || room.RoundNumber > room.MaxRounds,
			"round_number":              room.RoundNumber,
		})
		return
	}

	if err := persistRoom(id, &room); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to update attempt")
		return
	}

	respondSuccess(c, "Wrong answer", map[string]interface{}{
		"is_correct":      false,
		"attempts_left":   3 - room.Players[playerIndex].Attempts,
		"current_attempt": room.Players[playerIndex].Attempts,
	})
}

// UpdateGameRoom godoc
// @Summary Update game room
// @Description Update game room properties or advance to the next round
// @Tags game-rooms
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path string true "Room ID (UUID)"
// @Param action formData string false "Action to perform (e.g., 'advance_round')"
// @Param user_count formData int false "Number of users"
// @Param is_active formData bool false "Room active status"
// @Success 200 {object} models.ApiResult{result=models.GameRoom}
// @Failure 400 {object} models.ApiResult
// @Failure 404 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/room/{id} [patch]
func UpdateGameRoom(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondError(c, http.StatusBadRequest, "id is missing in URL")
		return
	}

	// Check for advance_round action
	if action := c.PostForm("action"); action == "advance_round" {
		var room models.GameRoom
		err := valkey.GetJSON(roomKeyPrefix+id, &room)
		if err != nil {
			respondError(c, http.StatusNotFound, "Room not found")
			return
		}

		// Advance to next round
		if err := advanceToNextRound(&room); err != nil {
			respondError(c, http.StatusInternalServerError, "No drawing topics are available. Please check MongoDB data.")
			return
		}

		if err := persistRoom(id, &room); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to advance round")
			return
		}

		respondSuccess(c, "Round advanced", nil)
		return
	}

	// Get existing room
	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		respondError(c, http.StatusNotFound, "Game room not found")
		return
	}

	// Update fields
	room.UpdatedAt = time.Now()

	if userCountStr := c.PostForm("user_count"); userCountStr != "" {
		userCount, err := strconv.Atoi(userCountStr)
		if err == nil {
			// Note: GameRoom model doesn't have UserCount field in the current schema
			// If needed, add it to the model first
			_ = userCount
		}
	}

	if isActiveStr := c.PostForm("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			room.IsActive = isActive
		}
	}

	// Save updated room to Valkey
	if err := persistRoom(id, &room); err != nil {
		log.Printf("Failed to update game room: %v", err)
		respondError(c, http.StatusInternalServerError, "Failed to update game room")
		return
	}

	respondSuccess(c, "Update Game Room", nil)
}

// DeleteGameRoom godoc
// @Summary Delete game room
// @Description Delete a game room (host only)
// @Tags game-rooms
// @Accept json
// @Produce json
// @Param id path string true "Room ID (UUID)"
// @Success 200 {object} models.ApiResult
// @Failure 404 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/room/{id} [delete]
func DeleteGameRoom(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondError(c, http.StatusBadRequest, "id is missing in URL")
		return
	}

	// Hard delete the room
	err := valkey.DeleteKey(roomKeyPrefix + id)
	if err != nil {
		log.Printf("Failed to delete game room: %v", err)
		respondError(c, http.StatusInternalServerError, "Failed to delete game room")
		return
	}

	respondSuccess(c, "Delete Game Room", nil)
}

// LeaveGameRoom godoc
// @Summary Leave game room
// @Description Remove a player from the game room
// @Tags game-rooms
// @Accept json
// @Produce json
// @Param id path string true "Room ID (UUID)"
// @Param username query string true "Username of the player leaving"
// @Success 200 {object} models.ApiResult{result=models.GameRoom}
// @Failure 400 {object} models.ApiResult
// @Failure 404 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/room/{id}/leave [post]
func LeaveGameRoom(c *gin.Context) {
	id := c.Param("id")
	username := c.Query("username")

	if id == "" || username == "" {
		respondError(c, http.StatusBadRequest, "id or username is missing")
		return
	}

	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		respondError(c, http.StatusNotFound, "Room not found")
		return
	}

	// Remove player from room
	newPlayers := []models.Player{}
	found := false
	for _, player := range room.Players {
		if player.Username != username {
			newPlayers = append(newPlayers, player)
		} else {
			found = true
		}
	}

	if !found {
		respondError(c, http.StatusNotFound, "Player not in room")
		return
	}

	// If room is empty (no players and drawer left), delete the room
	if len(newPlayers) == 0 && room.DrawerUser == username {
		if err := valkey.DeleteKey(roomKeyPrefix + id); err != nil {
			log.Printf("Failed to delete empty room: %v", err)
		}

		respondSuccess(c, "Left room and room deleted", nil)
		return
	}

	// Update room with remaining players
	room.Players = newPlayers
	room.UpdatedAt = time.Now()

	if err := persistRoom(id, &room); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to update room")
		return
	}

	respondSuccess(c, "Left room", nil)
}

// HandleChatMessage godoc
// @Summary Handle chat message
// @Description Process chat messages and check if they are correct answers
// @Tags game-rooms
// @Accept json
// @Produce json
// @Param id path string true "Room ID (UUID)"
// @Param message body object{username=string,message=string} true "Chat message data"
// @Success 200 {object} models.ApiResult{result=map[string]interface{}}
// @Failure 400 {object} models.ApiResult
// @Failure 404 {object} models.ApiResult
// @Router /game/room/{id}/chat [post]
func HandleChatMessage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondError(c, http.StatusBadRequest, "Room ID is required")
		return
	}

	var request struct {
		Username string `json:"username"`
		Message  string `json:"message"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		respondError(c, http.StatusNotFound, "Room not found")
		return
	}

	// If game is not playing, just return the message as regular chat
	if room.GameStatus != "playing" {
		respondSuccess(c, "Chat message", map[string]interface{}{
			"is_correct": false,
			"is_chat":    true,
		})
		return
	}

	// Drawer cannot guess - only players can
	if request.Username == room.DrawerUser {
		respondSuccess(c, "Chat message", map[string]interface{}{
			"is_correct": false,
			"is_chat":    true,
		})
		return
	}

	// Check if message is correct answer
	isCorrect := isAnswerCorrect(request.Message, &room)

	if isCorrect {
		if idx := findPlayerIndex(room.Players, request.Username); idx != -1 {
			room.Players[idx].Score += 2
		}
		if idx := findPlayerIndex(room.Players, room.DrawerUser); idx != -1 {
			room.Players[idx].Score++
		}

		room.LastRoundWinner = request.Username // 정답자를 다음 라운드 drawer로 설정

		if err := advanceToNextRound(&room); err != nil {
			respondError(c, http.StatusInternalServerError, "No drawing topics are available. Please check MongoDB data.")
			return
		}

		if err := persistRoom(id, &room); err != nil {
			log.Printf("Failed to update room: %v", err)
		} else {
			broadcastGame(id, map[string]interface{}{
				"type": "update",
			})
		}
	}

	respondSuccess(c, "Message processed", map[string]interface{}{
		"is_correct": isCorrect,
		"is_chat":    !isCorrect,
	})
}

// LoadTopicsFromMongo refreshes in-memory topics from MongoDB.
func LoadTopicsFromMongo(ctx context.Context) error {
	collection := database.GetCollection("game_topics")
	if collection == nil {
		return errors.New("mongodb is not connected")
	}

	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(queryCtx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(queryCtx)

	var loaded []models.GameTopic
	for cursor.Next(queryCtx) {
		var topic models.GameTopic
		if err := cursor.Decode(&topic); err != nil {
			return err
		}
		if topic.Canonical == "" {
			continue
		}
		if topic.Translations == nil {
			topic.Translations = map[string]string{}
		}
		if _, exists := topic.Translations["en"]; !exists {
			topic.Translations["en"] = topic.Canonical
		}
		loaded = append(loaded, topic)
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(loaded) == 0 {
		return errors.New("no topics found in MongoDB")
	}

	topics = loaded
	log.Printf("Loaded %d topics from MongoDB", len(loaded))
	return nil
}

// persistRoom sets the update timestamp and stores the room in Valkey with TTL.
func persistRoom(roomID string, room *models.GameRoom) error {
	room.UpdatedAt = time.Now()
	if room.UUID == "" {
		room.UUID = roomID
	}
	return valkey.SetJSON(roomKeyPrefix+roomID, room, roomTTL)
}

// selectNextWord returns a random topic that has not been used yet.
func selectNextWord(used []string) (models.GameTopic, error) {
	if len(topics) == 0 {
		return models.GameTopic{}, errors.New("no topics loaded")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	usedSet := make(map[string]struct{}, len(used))
	for _, word := range used {
		usedSet[strings.ToLower(word)] = struct{}{}
	}

	for attempts := 0; attempts < len(topics)*2; attempts++ {
		candidate := ensureTopicTranslations(topics[r.Intn(len(topics))])
		if _, exists := usedSet[strings.ToLower(candidate.Canonical)]; !exists {
			return candidate, nil
		}
	}

	return ensureTopicTranslations(topics[r.Intn(len(topics))]), nil
}

func applyTopicToRoom(room *models.GameRoom, topic models.GameTopic) {
	room.CurrentWord = topic.Canonical
	translations := make(map[string]string, len(topic.Translations))
	for key, value := range topic.Translations {
		translations[key] = value
	}
	room.CurrentWordTranslations = translations
	room.UsedWords = append(room.UsedWords, topic.Canonical)
}

func advanceToNextRound(room *models.GameRoom) error {
	// 다음 라운드 drawer 결정: 이전 라운드 우승자가 있으면 그 사람, 없으면 현재 drawer 유지
	if room.LastRoundWinner != "" {
		room.DrawerUser = room.LastRoundWinner
	}
	// LastRoundWinner가 비어있으면 현재 DrawerUser를 그대로 유지

	// 다음 라운드 준비
	room.RoundNumber++
	if room.RoundNumber > room.MaxRounds {
		room.GameStatus = "finished"
		room.TimeLeft = 0
		return nil
	}

	topic, err := selectNextWord(room.UsedWords)
	if err != nil {
		return err
	}

	applyTopicToRoom(room, topic)
	room.TimeLeft = 60
	room.LastRoundWinner = "" // 새 라운드 시작 시 초기화

	// 모든 플레이어의 시도 횟수 초기화
	for i := range room.Players {
		room.Players[i].Attempts = 0
	}

	return nil
}

func isAnswerCorrect(answer string, room *models.GameRoom) bool {
	normalized := normalizeAnswer(answer)
	if normalized == "" {
		return false
	}

	if normalizeAnswer(room.CurrentWord) == normalized {
		return true
	}

	for _, value := range room.CurrentWordTranslations {
		if normalizeAnswer(value) == normalized {
			return true
		}
	}

	return false
}

func normalizeAnswer(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func ensureTopicTranslations(topic models.GameTopic) models.GameTopic {
	if topic.Translations == nil {
		topic.Translations = map[string]string{}
	}
	if topic.Canonical != "" {
		topic.Translations["en"] = topic.Canonical
	}
	return topic
}

// broadcastGame sends a JSON payload to the room-specific Pub/Sub channel.
func broadcastGame(roomID string, payload map[string]interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal game payload: %v", err)
		return
	}

	if err := valkey.PublishMessage("game/"+roomID, string(data)); err != nil {
		log.Printf("Failed to publish game payload: %v", err)
	}
}

// findPlayerIndex returns the player's index or -1 if missing.
func findPlayerIndex(players []models.Player, username string) int {
	for i, player := range players {
		if player.Username == username {
			return i
		}
	}
	return -1
}
