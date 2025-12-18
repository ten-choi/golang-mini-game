package handlers

import (
	"context"
	"draw-and-guess-server/src/models"
	"draw-and-guess-server/src/valkey"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 전역 변수: 게임 주제 목록
var topics []models.GameTopic

const (
	roomKeyPrefix = "game_room:" // Valkey에 저장될 방 키 접두사
	roomTTL       = 3600         // 방 데이터 유효 시간 (1시간)
)

// manageGameTimer는 게임 타이머를 관리하고 매초 남은 시간을 브로드캐스트
// 고루틴으로 실행되며 게임이 진행 중일 때 1초마다 시간을 감소시킴
func manageGameTimer(roomID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Valkey에서 방 정보 조회
		var room models.GameRoom
		if err := valkey.GetJSON(roomKeyPrefix+roomID, &room); err != nil {
			log.Printf("Timer: Room %s not found, stopping timer", roomID)
			return
		}

		// 게임이 진행 중이 아니면 타이머 중지
		if room.GameStatus != "playing" {
			log.Printf("Timer: Game %s not playing, stopping timer", roomID)
			return
		}

		// 남은 시간 감소
		room.TimeLeft--

		// 시간이 0이 되면 다음 라운드로 진행
		if room.TimeLeft <= 0 {
			// 시간 초과 시: 정답자가 없으므로 현재 drawer 유지 (LastRoundWinner를 설정하지 않음)
			if err := advanceToNextRound(&room); err != nil {
				log.Printf("Timer: %v", err)
				room.GameStatus = "finished"
				room.TimeLeft = 0
			}
		}

		// 방 정보 저장
		if err := persistRoom(roomID, &room); err != nil {
			log.Printf("Timer: Failed to update room %s: %v", roomID, err)
			return
		}

		// 모든 클라이언트에게 타이머 업데이트 브로드캐스트
		broadcastGame(roomID, map[string]interface{}{
			"type": "timer_update",
			"data": gin.H{
				"room_id":      roomID,
				"time_left":    room.TimeLeft,
				"round_number": room.RoundNumber,
				"game_status":  room.GameStatus,
			},
		})

		// 게임이 종료되면 타이머 중지
		if room.GameStatus == "finished" {
			log.Printf("Timer: Game %s finished", roomID)
			return
		}
	}
}

// GetGameRooms는 게임 방 목록을 조회하는 API 핸들러
// GET /game/rooms?id={roomID}
// Query 파라미터: id (선택) - 특정 방 ID를 지정하면 해당 방만 반환
// 용도: 모든 활성 게임 방 목록을 조회하거나 특정 방의 정보를 조회
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
		// 특정 방 조회
		var room models.GameRoom
		err := valkey.GetJSON(roomKeyPrefix+id, &room)
		if err != nil {
			JSONNotFound(c, "Room not found")
			return
		}

		JSONSuccess(c, []models.GameRoom{room})
		return
	}

	// 모든 활성 방 조회
	keys, err := valkey.GetKeys(roomKeyPrefix + "*")
	if err != nil {
		JSONInternalError(c, "Failed to fetch game rooms")
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

	JSONSuccess(c, rooms)
}

// CreateGameRoom은 새로운 게임 방을 생성하는 API 핸들러
// POST /game/room
// Form 데이터: ldap_user (방장의 사용자명)
// 용도: 새 게임 방을 생성하고 첫 번째 라운드의 그림 주제를 설정
// @Summary Create a new game room
// @Description Create a new game room with the specified drawer
// @Tags game-rooms
// @Accept x-www-form-urlencoded
// @Produce json
// @Param ldap_user formData string true "Username of the room creator"
// @Param game_type formData string true "Game type: ox, general, or guess"
// @Success 200 {object} models.ApiResult{result=models.GameRoom}
// @Failure 400 {object} models.ApiResult
// @Failure 500 {object} models.ApiResult
// @Router /game/room [post]
func CreateGameRoom(c *gin.Context) {
	ldapUser := c.PostForm("ldap_user")
	if ldapUser == "" {
		JSONBadRequest(c, "ldap_user is missing")
		return
	}

	// 게임 타입 가져오기 (기본값: guess)
	gameTypeStr := c.DefaultPostForm("game_type", "guess")
	gameType := models.GameType(gameTypeStr)

	// 게임 타입 검증
	if !gameType.IsValid() {
		JSONBadRequest(c, "Invalid game_type. Must be: ox, general, or guess")
		return
	}

	// UUID 생성 (방 고유 ID)
	roomUUID := uuid.New().String()

	// 방 초기 설정
	room := models.GameRoom{
		UUID:         roomUUID,
		IsActive:     true,
		DrawerUser:   ldapUser, // 방장 (guess 타입에서는 그림 그리는 사람)
		RoomCreator:  ldapUser,
		Players:      []models.Player{},
		RoundNumber:  1,
		TimeLeft:     60,
		GameStatus:   "waiting", // 대기 중 상태로 시작
		UsedWords:    []string{},
		MaxRounds:    3,        // 최대 3라운드
		WinningScore: 3,        // 승리 점수 3점
		MaxPlayers:   4,        // 최대 플레이어 수
		GameType:     gameType, // 게임 타입
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 게임 타입에 따라 초기 문제/주제 설정
	if gameType == models.GameTypeGuess {
		// 그림 맞추기: 첫 번째 그림 주제 선택
		topic, err := selectNextWord(nil)
		if err != nil {
			JSONInternalError(c, "No drawing topics are available. Please check MongoDB data.")
			return
		}
		applyTopicToRoom(&room, topic)
	} else {
		// OX 또는 일반 퀴즈: 게임 시작 시 문제 설정
		room.CurrentWord = "" // 문제는 게임 시작 시 설정
	}

	// Valkey에 방 정보 저장
	if err := persistRoom(roomUUID, &room); err != nil {
		log.Printf("Failed to create game room: %v", err)
		JSONInternalError(c, "Failed to create game room")
		return
	}

	log.Printf("Created game room with ID: %s, Type: %s", roomUUID, gameType)

	// WebSocket으로 방 목록 업데이트 브로드캐스트
	broadcastRoomListUpdate()

	response := map[string]interface{}{
		"room_id":   roomUUID,
		"game_type": gameType,
	}

	if gameType == models.GameTypeGuess {
		response["current_word"] = room.CurrentWord
		response["current_word_translations"] = room.CurrentWordTranslations
	}

	JSONSuccess(c, response)
}

// JoinGameRoom은 게임 방에 참가하는 API 핸들러
// POST /game/room/:id/join
// Form 데이터: username (참가할 플레이어의 사용자명)
// 용도: 기존 게임 방에 새 플레이어를 추가
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
		JSONBadRequest(c, "id or username is missing")
		return
	}

	// Valkey에서 방 정보 조회
	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		JSONNotFound(c, "Game room not found")
		return
	}

	// 이미 참가한 플레이어인지 확인
	if findPlayerIndex(room.Players, username) != -1 {
		JSONBadRequest(c, "Already joined")
		return
	}

	// 최대 인원 체크 (방장 제외 최대 3명, 총 4명)
	maxPlayers := 4
	if room.MaxPlayers > 0 {
		maxPlayers = room.MaxPlayers
	}

	// 방장을 제외한 플레이어 수 확인
	if len(room.Players) >= maxPlayers-1 {
		JSONBadRequest(c, "Room is full (max 4 players)")
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

	// 변경사항 저장
	if err := persistRoom(id, &room); err != nil {
		JSONInternalError(c, "Failed to join room")
		return
	}

	log.Printf("Player %s joined room %s (total players: %d)", username, id, len(room.Players))

	// WebSocket으로 방 상태 업데이트 브로드캐스트
	broadcastRoomUpdate(id)
	broadcastRoomListUpdate()

	JSONSuccess(c, map[string]interface{}{
		"room_id":      id,
		"player_count": len(room.Players),
	})
}

// StartGame은 게임을 시작하는 API 핸들러
// POST /game/room/:id/start
// 용도: 대기 중인 게임을 시작하고 타이머를 활성화
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
		JSONBadRequest(c, "id is missing")
		return
	}

	// Valkey에서 방 정보 조회
	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		JSONNotFound(c, "Game room not found")
		return
	}

	// 최소 1명 이상의 플레이어 필요 확인
	if len(room.Players) < 1 {
		JSONBadRequest(c, "Need at least 1 player to start")
		return
	}

	// 게임 상태를 "playing"으로 변경
	room.GameStatus = "playing"
	room.TimeLeft = 60
	room.UpdatedAt = time.Now()

	// 변경사항 저장
	if err := persistRoom(id, &room); err != nil {
		JSONInternalError(c, "Failed to start game")
		return
	}

	log.Printf("Game started in room %s by %s", id, room.DrawerUser)

	// 모든 클라이언트에게 게임 시작 알림
	broadcastGame(id, map[string]interface{}{
		"type":        "update",
		"game_status": "playing",
	})

	// 방 목록 업데이트 (게임 중 방은 참가 불가)
	broadcastRoomListUpdate()

	// 타이머 관리 고루틴 시작
	go manageGameTimer(id)

	JSONSuccess(c, map[string]interface{}{
		"game_status":  "playing",
		"round_number": room.RoundNumber,
	})
}

// CheckAnswer는 플레이어의 정답을 확인하는 API 핸들러
// POST /game/room/:id/answer
// Form 데이터: username, answer (플레이어가 제출한 답)
// 용도: 정답 여부를 확인하고 점수를 부여하며, 정답이면 다음 라운드로 진행
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
		JSONBadRequest(c, "id, answer or username is missing")
		return
	}

	// Valkey에서 방 정보 조회
	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		JSONNotFound(c, "Game room not found")
		return
	}

	// 플레이어 찾기
	playerIndex := findPlayerIndex(room.Players, username)
	if playerIndex == -1 {
		JSONBadRequest(c, "Player not found")
		return
	}

	// 3회 제출 제한 확인
	if room.Players[playerIndex].Attempts >= 3 {
		JSONBadRequest(c, "Maximum attempts reached (3/3)")
		return
	}

	// 시도 횟수 증가
	room.Players[playerIndex].Attempts++

	// 정답 여부 확인
	isCorrect := isAnswerCorrect(answer, &room)

	if isCorrect {
		// 정답인 경우: 점수 부여 및 다음 라운드 진행
		room.Players[playerIndex].Score += 2
		room.LastRoundWinner = username // 정답자를 다음 라운드 drawer로 설정
		gameFinished := room.Players[playerIndex].Score >= room.WinningScore

		// 게임이 끝나지 않았고 라운드가 남았으면 다음 라운드로
		if !gameFinished && room.RoundNumber < room.MaxRounds {
			if err := advanceToNextRound(&room); err != nil {
				JSONInternalError(c, "No drawing topics are available. Please check MongoDB data.")
				return
			}
		}

		// 게임 종료 조건 확인
		if gameFinished || room.RoundNumber > room.MaxRounds {
			room.GameStatus = "finished"
		}

		if err := persistRoom(id, &room); err != nil {
			JSONInternalError(c, "Failed to update game state")
			return
		}

		JSONSuccess(c, map[string]interface{}{
			"is_correct":                true,
			"current_word":              room.CurrentWord,
			"current_word_translations": room.CurrentWordTranslations,
			"player_score":              room.Players[playerIndex].Score,
			"game_finished":             gameFinished || room.RoundNumber > room.MaxRounds,
			"round_number":              room.RoundNumber,
		})
		return
	}

	// 오답인 경우: 남은 시도 횟수 반환
	if err := persistRoom(id, &room); err != nil {
		JSONInternalError(c, "Failed to update attempt")
		return
	}

	JSONSuccess(c, map[string]interface{}{
		"is_correct":      false,
		"attempts_left":   3 - room.Players[playerIndex].Attempts,
		"current_attempt": room.Players[playerIndex].Attempts,
	})
}

// UpdateGameRoom은 게임 방 정보를 업데이트하는 API 핸들러
// PATCH /game/room/:id
// Form 데이터: action (선택, "advance_round"로 다음 라운드 진행), user_count (선택), is_active (선택)
// 용도: 방의 상태를 변경하거나 다음 라운드로 강제 진행
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
		JSONBadRequest(c, "id is missing in URL")
		return
	}

	// "advance_round" 액션 처리: 다음 라운드로 강제 진행
	if action := c.PostForm("action"); action == "advance_round" {
		var room models.GameRoom
		err := valkey.GetJSON(roomKeyPrefix+id, &room)
		if err != nil {
			JSONNotFound(c, "Room not found")
			return
		}

		// 다음 라운드로 진행
		if err := advanceToNextRound(&room); err != nil {
			JSONInternalError(c, "No drawing topics are available. Please check MongoDB data.")
			return
		}

		if err := persistRoom(id, &room); err != nil {
			JSONInternalError(c, "Failed to advance round")
			return
		}

		JSONSuccess(c, map[string]interface{}{
			"room_id":      id,
			"round_number": room.RoundNumber,
		})
		return
	}

	// Get existing room
	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		JSONNotFound(c, "Game room not found")
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
		JSONInternalError(c, "Failed to update game room")
		return
	}

	JSONSuccess(c, map[string]interface{}{
		"room_id":   id,
		"is_active": room.IsActive,
	})
}

// DeleteGameRoom은 게임 방을 삭제하는 API 핸들러
// DELETE /game/room/:id
// 용도: 게임 방을 완전히 제거 (Valkey에서 삭제)
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
		JSONBadRequest(c, "id is missing in URL")
		return
	}

	// Valkey에서 방 데이터 완전 삭제
	err := valkey.DeleteKey(roomKeyPrefix + id)
	if err != nil {
		log.Printf("Failed to delete game room: %v", err)
		JSONInternalError(c, "Failed to delete game room")
		return
	}

	JSONSuccess(c, map[string]interface{}{
		"room_id": id,
		"deleted": true,
	})
}

// LeaveGameRoom은 게임 방에서 나가는 API 핸들러
// POST /game/room/:id/leave?username={username}
// Query 파라미터: username (나갈 플레이어의 사용자명)
// 용도: 플레이어를 방에서 제거하고, 모든 플레이어가 나가면 방을 삭제
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
		JSONBadRequest(c, "id or username is missing")
		return
	}

	// Valkey에서 방 정보 조회
	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		JSONNotFound(c, "Room not found")
		return
	}

	// 플레이어를 방에서 제거
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
		JSONNotFound(c, "Player not in room")
		return
	}

	// 방이 비어있으면 (플레이어도 없고 방장도 나간 경우) 방 삭제
	if len(newPlayers) == 0 && room.DrawerUser == username {
		if err := valkey.DeleteKey(roomKeyPrefix + id); err != nil {
			log.Printf("Failed to delete empty room: %v", err)
		}

		log.Printf("Room %s deleted (empty)", id)
		broadcastRoomListUpdate()

		JSONSuccess(c, map[string]interface{}{
			"room_id": id,
			"deleted": true,
		})
		return
	}

	// 남은 플레이어로 방 업데이트

	// Update room with remaining players
	room.Players = newPlayers
	room.UpdatedAt = time.Now()

	if err := persistRoom(id, &room); err != nil {
		JSONInternalError(c, "Failed to update room")
		return
	}

	log.Printf("Player %s left room %s (remaining players: %d)", username, id, len(room.Players))

	// WebSocket으로 방 상태 업데이트 브로드캐스트
	broadcastRoomUpdate(id)
	broadcastRoomListUpdate()

	JSONSuccess(c, map[string]interface{}{
		"room_id":      id,
		"player_count": len(room.Players),
	})
}

// HandleChatMessage는 채팅 메시지를 처리하는 API 핸들러
// POST /game/room/:id/chat
// JSON 데이터: {username, message}
// 용도: 채팅 메시지를 받아서 정답인지 확인하고, 정답이면 점수를 부여하고 다음 라운드로 진행
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
		JSONBadRequest(c, "Room ID is required")
		return
	}

	var request struct {
		Username string `json:"username"`
		Message  string `json:"message"`
	}

	// JSON 요청 바디 파싱
	if err := c.ShouldBindJSON(&request); err != nil {
		JSONBadRequest(c, "Invalid request body")
		return
	}

	// Valkey에서 방 정보 조회
	var room models.GameRoom
	err := valkey.GetJSON(roomKeyPrefix+id, &room)
	if err != nil {
		JSONNotFound(c, "Room not found")
		return
	}

	// 게임이 진행 중이 아니면 일반 채팅으로 처리
	if room.GameStatus != "playing" {
		JSONSuccess(c, map[string]interface{}{
			"is_correct": false,
			"is_chat":    true,
		})
		return
	}

	// 그림 그리는 사람은 정답을 맞출 수 없음
	if request.Username == room.DrawerUser {
		JSONSuccess(c, map[string]interface{}{
			"is_correct": false,
			"is_chat":    true,
		})
		return
	}

	// 메시지가 정답인지 확인
	isCorrect := isAnswerCorrect(request.Message, &room)

	if isCorrect {
		// 정답인 경우: 정답자에게 2점, 그림 그린 사람에게 1점 부여
		if idx := findPlayerIndex(room.Players, request.Username); idx != -1 {
			room.Players[idx].Score += 2
		}
		if idx := findPlayerIndex(room.Players, room.DrawerUser); idx != -1 {
			room.Players[idx].Score++
		}

		room.LastRoundWinner = request.Username // 정답자를 다음 라운드 drawer로 설정

		// 다음 라운드로 진행
		if err := advanceToNextRound(&room); err != nil {
			JSONInternalError(c, "No drawing topics are available. Please check MongoDB data.")
			return
		}

		// 변경사항 저장 및 브로드캐스트
		if err := persistRoom(id, &room); err != nil {
			log.Printf("Failed to update room: %v", err)
		} else {
			broadcastGame(id, map[string]interface{}{
				"type": "update",
			})
		}
	}

	JSONSuccess(c, map[string]interface{}{
		"is_correct": isCorrect,
		"is_chat":    !isCorrect,
	})
}

// LoadTopicsFromMongo는 MongoDB에서 게임 주제를 로드하는 함수
// 현재는 테스트를 위해 하드코딩된 데이터를 사용합니다.
func LoadTopicsFromMongo(ctx context.Context) error {
	// models.GetTestTopics()에서 테스트용 하드코딩 데이터 로드
	topics = models.GetTestTopics()
	log.Printf("Loaded %d topics from hardcoded data (test mode)", len(topics))
	return nil
}

// persistRoom은 방 정보를 Valkey에 저장하는 함수 (TTL 설정 포함)
func persistRoom(roomID string, room *models.GameRoom) error {
	room.UpdatedAt = time.Now()
	if room.UUID == "" {
		room.UUID = roomID
	}
	return valkey.SetJSON(roomKeyPrefix+roomID, room, roomTTL)
}

// selectNextWord는 아직 사용되지 않은 랜덤 주제를 선택하는 함수
func selectNextWord(used []string) (models.GameTopic, error) {
	if len(topics) == 0 {
		return models.GameTopic{}, errors.New("no topics loaded")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	usedSet := make(map[string]struct{}, len(used))
	for _, word := range used {
		usedSet[strings.ToLower(word)] = struct{}{}
	}

	// 사용하지 않은 주제를 찾기 위해 여러 번 시도
	for attempts := 0; attempts < len(topics)*2; attempts++ {
		candidate := ensureTopicTranslations(topics[r.Intn(len(topics))])
		if _, exists := usedSet[strings.ToLower(candidate.Canonical)]; !exists {
			return candidate, nil
		}
	}

	// 모든 주제를 사용했으면 랜덤으로 선택
	return ensureTopicTranslations(topics[r.Intn(len(topics))]), nil
}

// applyTopicToRoom은 선택된 주제를 방에 적용하는 함수
func applyTopicToRoom(room *models.GameRoom, topic models.GameTopic) {
	room.CurrentWord = topic.Canonical
	translations := make(map[string]string, len(topic.Translations))
	for key, value := range topic.Translations {
		translations[key] = value
	}
	room.CurrentWordTranslations = translations
	room.UsedWords = append(room.UsedWords, topic.Canonical)
}

// advanceToNextRound는 다음 라운드로 진행하는 함수
func advanceToNextRound(room *models.GameRoom) error {
	// 다음 라운드 drawer 결정: 이전 라운드 우승자가 있으면 그 사람, 없으면 현재 drawer 유지
	if room.LastRoundWinner != "" {
		room.DrawerUser = room.LastRoundWinner
	}
	// LastRoundWinner가 비어있으면 현재 DrawerUser를 그대로 유지

	// 라운드 번호 증가
	room.RoundNumber++
	if room.RoundNumber > room.MaxRounds {
		room.GameStatus = "finished"
		room.TimeLeft = 0
		return nil
	}

	// 새로운 주제 선택
	topic, err := selectNextWord(room.UsedWords)
	if err != nil {
		return err
	}

	// 새 주제 적용 및 타이머 리셋
	applyTopicToRoom(room, topic)
	room.TimeLeft = 60
	room.LastRoundWinner = "" // 새 라운드 시작 시 초기화

	// 모든 플레이어의 시도 횟수 초기화
	for i := range room.Players {
		room.Players[i].Attempts = 0
	}

	return nil
}

// isAnswerCorrect는 제출된 답이 정답인지 확인하는 함수
func isAnswerCorrect(answer string, room *models.GameRoom) bool {
	normalized := normalizeAnswer(answer)
	if normalized == "" {
		return false
	}

	// 현재 단어와 비교
	if normalizeAnswer(room.CurrentWord) == normalized {
		return true
	}

	// 번역된 단어들과 비교
	for _, value := range room.CurrentWordTranslations {
		if normalizeAnswer(value) == normalized {
			return true
		}
	}

	return false
}

// normalizeAnswer는 답변을 정규화하는 함수 (소문자 변환, 공백 제거)
func normalizeAnswer(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// ensureTopicTranslations는 주제의 번역 정보를 보장하는 함수
func ensureTopicTranslations(topic models.GameTopic) models.GameTopic {
	if topic.Translations == nil {
		topic.Translations = map[string]string{}
	}
	if topic.Canonical != "" {
		topic.Translations["en"] = topic.Canonical
	}
	return topic
}

// broadcastGame은 게임 업데이트를 Valkey Pub/Sub을 통해 브로드캐스트하는 함수
func broadcastGame(roomID string, payload map[string]interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal game payload: %v", err)
		return
	}

	// Valkey를 통해 해당 방의 모든 구독자에게 메시지 발행
	if err := valkey.PublishMessage("game/"+roomID, string(data)); err != nil {
		log.Printf("Failed to publish game payload: %v", err)
	}
}

// findPlayerIndex는 플레이어 목록에서 특정 사용자의 인덱스를 찾는 함수 (없으면 -1 반환)
func findPlayerIndex(players []models.Player, username string) int {
	for i, player := range players {
		if player.Username == username {
			return i
		}
	}
	return -1
}

// broadcastRoomUpdate는 특정 방의 상태가 변경되었을 때 해당 방에 있는 모든 클라이언트에게 업데이트를 전송
func broadcastRoomUpdate(roomID string) {
	var room models.GameRoom
	if err := valkey.GetJSON(roomKeyPrefix+roomID, &room); err != nil {
		log.Printf("Failed to get room for broadcast: %v", err)
		return
	}

	payload := map[string]interface{}{
		"type": "room_update",
		"data": room,
	}

	broadcastGame(roomID, payload)
}

// broadcastRoomListUpdate는 방 목록이 변경되었을 때 (방 생성/삭제/상태 변경) 모든 클라이언트에게 업데이트를 전송
func broadcastRoomListUpdate() {
	// 전역 채널로 방 목록 업데이트 브로드캐스트
	payload := map[string]interface{}{
		"type": "room_list_update",
		"data": map[string]interface{}{
			"message": "Room list has been updated",
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal room list payload: %v", err)
		return
	}

	// 전역 채널로 발행
	if err := valkey.PublishMessage("lobby", string(data)); err != nil {
		log.Printf("Failed to publish room list update: %v", err)
	}
}
