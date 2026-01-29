package graph

import (
	"context"
	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/valkey"
	"draw-and-guess-server/pkg/dictionary"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// ========================================
// GameRoom Mutations
// ========================================

// CreateGameRoom is the resolver for the createGameRoom field.
func (r *mutationResolver) CreateGameRoom(ctx context.Context, input model.CreateGameRoomInput) (*model.GameRoom, error) {
	roomID := uuid.New().String()
	now := time.Now()

	// Set isPrivate based on input, default to false
	isPrivate := false
	if input.IsPrivate != nil {
		isPrivate = *input.IsPrivate
	}

	// Set roundTimeLimit based on input, default to 30 seconds
	roundTimeLimit := int32(30)
	if input.RoundTimeLimit != nil {
		// Validate roundTimeLimit (5-300 seconds)
		if *input.RoundTimeLimit < 5 || *input.RoundTimeLimit > 300 {
			return nil, fmt.Errorf("round time limit must be between 5 and 300 seconds")
		}
		roundTimeLimit = *input.RoundTimeLimit
	}

	room := &model.GameRoom{
		ID:             roomID,
		Name:           input.Name,
		GameType:       input.GameType,
		Status:         model.GameRoomStatus(common.RoomStatusWaiting),
		CurrentRound:   0,
		TotalRounds:    int32(input.TotalRounds),
		RoundTimeLimit: roundTimeLimit,
		Users:          []*model.GameUser{},
		MaxUsers:       int32(input.MaxUsers),
		HostUserID:     input.HostUserID,
		UsedQuizIds:    []string{}, // Initialize empty quiz history
		IsPrivate:      isPrivate,
		Password:       input.Password, // Set password if provided
		CreatedAt:      now,
	}

	// Fetch host user to get the name
	hostUser, err := r.UserService.GetByID(ctx, input.HostUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host user: %w", err)
	}

	// Add host as first user
	room.Users = append(room.Users, &model.GameUser{
		UserID:  input.HostUserID,
		Name:    hostUser.Name,
		Score:   0,
		IsReady: true,
	})

	roomMutex.Lock()
	gameRooms[roomID] = room
	roomMutex.Unlock()

	// Publish lobby update
	go publishLobbyUpdate()

	return room, nil
}

// JoinGameRoom is the resolver for the joinGameRoom field.
func (r *mutationResolver) JoinGameRoom(ctx context.Context, roomID string, userID string, password *string) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	// Check password if room is private
	if room.IsPrivate && room.Password != nil {
		if password == nil || *password != *room.Password {
			return nil, fmt.Errorf("incorrect password")
		}
	}

	// Check if user already in room
	for _, p := range room.Users {
		if p.UserID == userID {
			return room, nil // Already in room
		}
	}

	// Check if room is full
	if int32(len(room.Users)) >= room.MaxUsers {
		return nil, fmt.Errorf("room is full (max %d users)", room.MaxUsers)
	}

	// Fetch user to get the name
	user, err := r.UserService.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Add user
	newUser := &model.GameUser{
		UserID:  userID,
		Name:    user.Name,
		Score:   0,
		IsReady: false,
	}
	room.Users = append(room.Users, newUser)

	// Update the room in the map (important!)
	gameRooms[roomID] = room

	// Publish events
	go func() {
		publishLobbyUpdate()
		publishRoomUpdateToWebSocket(roomID, room)
	}()

	return room, nil
}

// LeaveGameRoom is the resolver for the leaveGameRoom field.
func (r *mutationResolver) LeaveGameRoom(ctx context.Context, roomID string, userID string) (*model.GameRoom, error) {
	roomMutex.Lock()

	room, exists := gameRooms[roomID]
	if !exists {
		roomMutex.Unlock()
		return nil, fmt.Errorf("room not found")
	}

	log.Printf("[LeaveGameRoom] User %s leaving room %s (current users: %d)", userID, roomID, len(room.Users))

	// Remove user and capture the leaving user
	var leavingUser *model.GameUser
	newUsers := []*model.GameUser{}
	for _, p := range room.Users {
		if p.UserID != userID {
			newUsers = append(newUsers, p)
		} else {
			leavingUser = p
		}
	}
	room.Users = newUsers

	// User not found in room
	if leavingUser == nil {
		log.Printf("[LeaveGameRoom] Warning: User %s not found in room %s", userID, roomID)
		roomMutex.Unlock()
		return room, nil
	}

	// Delete room if empty (consistent with RemoveUserFromRoom logic)
	if len(room.Users) == 0 {
		// Always delete empty rooms when user explicitly leaves
		log.Printf("[LeaveGameRoom] Room %s is now empty, deleting room (status: %s)", roomID, room.Status)
		delete(gameRooms, roomID)
		roomMutex.Unlock() // Unlock before goroutines

		// Cleanup pre-loaded quizzes if it's a quiz game
		if room.GameType == model.GameTypeOx || room.GameType == model.GameTypeQa {
			go cleanupPreloadedQuizzes(roomID)
		}

		go func() {
			publishLobbyUpdate()
			// Notify that room is deleted via WebSocket
			publishRoomDeletedToWebSocket(roomID)
		}()

		// Return nil to indicate room was deleted
		return nil, nil
	}

	// Reassign host if needed
	if room.HostUserID == userID && len(room.Users) > 0 {
		room.HostUserID = room.Users[0].UserID
		log.Printf("[LeaveGameRoom] New host assigned in room %s: %s", roomID, room.HostUserID)
	}

	// Update the room in the map (important!)
	gameRooms[roomID] = room
	roomMutex.Unlock() // Unlock before goroutines

	// Publish events
	log.Printf("[LeaveGameRoom] User %s left room %s, remaining users: %d", leavingUser.Name, roomID, len(room.Users))

	// All publish operations in goroutine to avoid deadlock
	go func() {
		publishLobbyUpdate()
		publishRoomUpdateToWebSocket(roomID, room)
	}()

	return room, nil
}

// StartGame is the resolver for the startGame field.
func (r *mutationResolver) StartGame(ctx context.Context, roomID string) (*model.GameRoom, error) {
	roomMutex.Lock()
	room, exists := gameRooms[roomID]
	if !exists {
		roomMutex.Unlock()
		return nil, fmt.Errorf("room not found")
	}

	// Check if all users are ready
	if len(room.Users) == 0 {
		roomMutex.Unlock()
		return nil, fmt.Errorf("cannot start game: no users in room")
	}

	for _, user := range room.Users {
		if !user.IsReady {
			roomMutex.Unlock()
			return nil, fmt.Errorf("cannot start game: not all users are ready")
		}
	}

	room.Status = model.GameRoomStatus(common.RoomStatusPlaying)
	room.CurrentRound = 1

	// Update the room in the map (important!)
	gameRooms[roomID] = room
	roomMutex.Unlock()

	// Update all users' status to ingame
	for _, user := range room.Users {
		statusKey := common.RedisKeyUserStatus + user.UserID
		err := valkey.Client.Set(ctx, statusKey, common.UserStatusInGame, time.Duration(common.RedisUserStatusTTL)*time.Second).Err()
		if err != nil {
			log.Printf("[StartGame] Failed to set user %s status to ingame: %v", user.UserID, err)
		} else {
			log.Printf("[StartGame] User %s status set to ingame", user.UserID)
		}
	}

	// Publish events
	go func() {
		publishLobbyUpdate()
		publishRoomUpdateToWebSocket(roomID, room)
	}()

	// Start game management based on game type
	log.Printf("[StartGame] ========================================")
	log.Printf("[StartGame] Room %s game type: %s", roomID, room.GameType)
	log.Printf("[StartGame] GameType value as string: '%s'", string(room.GameType))
	log.Printf("[StartGame] GameType length: %d", len(room.GameType))
	log.Printf("[StartGame] GameTypeOx constant: '%s'", model.GameTypeOx)
	log.Printf("[StartGame] GameTypeQa constant: '%s'", model.GameTypeQa)
	log.Printf("[StartGame] GameTypeWordchain constant: '%s'", model.GameTypeWordchain)
	log.Printf("[StartGame] Is OX? %v", room.GameType == model.GameTypeOx)
	log.Printf("[StartGame] Is QA? %v", room.GameType == model.GameTypeQa)
	log.Printf("[StartGame] Is Wordchain? %v", room.GameType == model.GameTypeWordchain)
	log.Printf("[StartGame] ========================================")

	if room.GameType == model.GameTypeWordchain {
		log.Printf("[StartGame] Starting wordchain game for room %s", roomID)
		go startWordchainGame(roomID)
	} else if room.GameType == model.GameTypeOx || room.GameType == model.GameTypeQa {
		log.Printf("[StartGame] Starting quiz game for room %s (type: %s)", roomID, room.GameType)
		// Use background context for goroutine
		go startQuizGame(context.Background(), roomID, room.GameType, r.QuizService)
	} else {
		log.Printf("[StartGame] ERROR: Unknown/unmatched game type: '%s' (len=%d)", room.GameType, len(room.GameType))
	}

	return room, nil
}

// DeleteGameRoom is the resolver for the deleteGameRoom field.
func (r *mutationResolver) DeleteGameRoom(ctx context.Context, roomID string) (bool, error) {
	roomMutex.Lock()
	room, exists := gameRooms[roomID]
	delete(gameRooms, roomID)
	roomMutex.Unlock()

	// Cleanup pre-loaded quizzes if it's a quiz game
	if exists && (room.GameType == model.GameTypeOx || room.GameType == model.GameTypeQa) {
		cleanupPreloadedQuizzes(roomID)
	}

	go publishLobbyUpdate()

	return true, nil
}

// ========================================
// GameRoom Queries
// ========================================

// GameRoom is the resolver for the gameRoom field.
func (r *queryResolver) GameRoom(ctx context.Context, id string) (*model.GameRoom, error) {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	room, exists := gameRooms[id]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	return room, nil
}

// GameRooms is the resolver for the gameRooms field.
func (r *queryResolver) GameRooms(ctx context.Context, gameType *model.GameType, status *model.GameRoomStatus, includePrivate *bool, hasSpace *bool, limit *int32, sortBy *string) ([]*model.GameRoom, error) {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	result := []*model.GameRoom{}
	for _, room := range gameRooms {
		// Skip empty rooms (should have been deleted)
		if len(room.Users) == 0 {
			continue
		}

		// Skip finished games (should be cleaned up)
		if room.Status == model.GameRoomStatus(common.RoomStatusFinished) {
			continue
		}

		// Filter by game type
		if gameType != nil && room.GameType != *gameType {
			continue
		}

		// Filter by status
		if status != nil && room.Status != *status {
			continue
		}

		// Filter by private rooms
		if includePrivate == nil || !*includePrivate {
			if room.IsPrivate {
				continue
			}
		}

		// Filter by available space
		if hasSpace != nil && *hasSpace {
			if len(room.Users) >= int(room.MaxUsers) {
				continue
			}
		}

		result = append(result, room)

		// Apply limit
		if limit != nil && len(result) >= int(*limit) {
			break
		}
	}

	return result, nil
}

// MyCurrentRoom is the resolver for the myCurrentRoom field.
func (r *queryResolver) MyCurrentRoom(ctx context.Context, Name string) (*model.GameRoom, error) {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	// Find room where Name is a user
	for _, room := range gameRooms {
		for _, user := range room.Users {
			if user.Name == Name {
				return room, nil
			}
		}
	}

	return nil, nil // Not in any room
}

// ========================================
// GameRoom Subscriptions
// ========================================

// GameRoomUpdated is the resolver for the gameRoomUpdated field.
func (r *subscriptionResolver) GameRoomUpdated(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	ch := make(chan *model.GameRoom)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// LobbyUpdated is the resolver for the lobbyUpdated field.
func (r *subscriptionResolver) LobbyUpdated(ctx context.Context) (<-chan []*model.GameRoom, error) {
	ch := make(chan []*model.GameRoom)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// UserJoined is the resolver for the userJoined field.
func (r *subscriptionResolver) UserJoined(ctx context.Context, roomID string) (<-chan *model.GameUser, error) {
	ch := make(chan *model.GameUser)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// startWordchainGame manages wordchain game flow
func startWordchainGame(roomID string) {
	roomMutex.Lock()
	room, exists := gameRooms[roomID]
	if !exists {
		roomMutex.Unlock()
		return
	}

	// Generate initial word and set first turn to host (first user)
	dict := dictionary.GetInstance()
	initialWord := dict.GetRandomWord()

	room.WordchainLastWord = &initialWord
	if len(room.Users) > 0 {
		firstUserID := room.Users[0].UserID
		room.CurrentTurnUserID = &firstUserID
	}

	// Initialize used words array with initial word
	room.WordchainUsedWords = []string{initialWord}
	currentRound := room.CurrentRound
	roomMutex.Unlock()

	// Broadcast initial word prompt (nil for prompt since we're just showing lastWord)
	publishWordchainPromptToWebSocket(roomID, nil, initialWord, currentRound)

	// Publish room update to show changes
	publishRoomUpdateToWebSocket(roomID, room)

	// Start turn timer for first user
	StartWordchainTurnTimer(roomID)
}

// UserLeft is the resolver for the userLeft field.
func (r *subscriptionResolver) UserLeft(ctx context.Context, roomID string) (<-chan *model.GameUser, error) {
	ch := make(chan *model.GameUser)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// GameStarted is the resolver for the gameStarted field.
func (r *subscriptionResolver) GameStarted(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	ch := make(chan *model.GameRoom)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// UpdateGameRoom is the resolver for the updateGameRoom field.
func (r *mutationResolver) UpdateGameRoom(ctx context.Context, roomID string, input model.UpdateGameRoomInput) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	// Update fields if provided
	if input.Name != nil {
		room.Name = *input.Name
	}
	if input.MaxUsers != nil {
		room.MaxUsers = *input.MaxUsers
	}
	if input.TotalRounds != nil {
		room.TotalRounds = *input.TotalRounds
	}
	if input.RoundTimeLimit != nil {
		// Validate roundTimeLimit (5-300 seconds)
		if *input.RoundTimeLimit < 5 || *input.RoundTimeLimit > 300 {
			return nil, fmt.Errorf("round time limit must be between 5 and 300 seconds")
		}
		room.RoundTimeLimit = *input.RoundTimeLimit
	}
	if input.IsPrivate != nil {
		room.IsPrivate = *input.IsPrivate
	}
	if input.Password != nil {
		room.Password = input.Password
	}

	// Update the room in the map (important!)
	gameRooms[roomID] = room

	// Publish update
	go func() {
		publishLobbyUpdate()
		publishRoomUpdateToWebSocket(roomID, room)
	}()

	return room, nil
}

// TransferHost is the resolver for the transferHost field.
func (r *mutationResolver) TransferHost(ctx context.Context, roomID string, newHostUserID string) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	// Check if new host is in the room
	found := false
	for _, p := range room.Users {
		if p.Name == newHostUserID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("new host is not in the room")
	}

	room.HostUserID = newHostUserID

	// Update the room in the map (important!)
	gameRooms[roomID] = room

	// Publish update
	go func() {
		publishRoomUpdateToWebSocket(roomID, room)
	}()

	return room, nil
}

// SetReady is the resolver for the setReady field.
func (r *mutationResolver) SetReady(ctx context.Context, roomID string, userID string, ready bool) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	// Find user and update ready status
	for i, p := range room.Users {
		if p.UserID == userID {
			room.Users[i].IsReady = ready

			// Update the room in the map (important!)
			gameRooms[roomID] = room

			// Publish update
			go func() {
				publishRoomUpdateToWebSocket(roomID, room)
			}()

			return room, nil
		}
	}

	return nil, fmt.Errorf("user not found in room")
}

// StartRound is the resolver for the startRound field.
func (r *mutationResolver) StartRound(ctx context.Context, roomID string) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	if room.Status != model.GameRoomStatus(common.RoomStatusPlaying) {
		return nil, fmt.Errorf("game is not in playing status")
	}

	room.CurrentRound++

	// Publish update
	go func() {
		publishRoomUpdateToWebSocket(roomID, room)
	}()

	return room, nil
}

// SubmitAnswer is the resolver for the submitAnswer field.
func (r *mutationResolver) SubmitAnswer(ctx context.Context, roomID string, Name string, answer string) (*model.AnswerResult, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	// Find user
	var user *model.GameUser
	for i := range room.Users {
		if room.Users[i].Name == Name {
			user = room.Users[i]
			break
		}
	}
	if user == nil {
		return nil, fmt.Errorf("user not found in room")
	}

	// Answer validation is handled by game-specific WebSocket logic
	// This mutation is primarily for tracking and compatibility
	emptyStr := ""
	result := &model.AnswerResult{
		IsCorrect:     false,
		EarnedScore:   0,
		TotalScore:    user.Score,
		CorrectAnswer: &emptyStr,
	}

	return result, nil
}

// EndGame is the resolver for the endGame field.
func (r *mutationResolver) EndGame(ctx context.Context, roomID string) (*model.GameResult, error) {
	roomMutex.Lock()
	room, exists := gameRooms[roomID]
	if !exists {
		roomMutex.Unlock()
		return nil, fmt.Errorf("room not found")
	}

	// Calculate rankings before changing status
	rankings := calculateRankings(room.Users)

	room.Status = model.GameRoomStatus(common.RoomStatusFinished)

	// Update all users' status back to lobby
	for _, user := range room.Users {
		statusKey := common.RedisKeyUserStatus + user.UserID
		err := valkey.Client.Set(ctx, statusKey, common.UserStatusLobby, time.Duration(common.RedisUserStatusTTL)*time.Second).Err()
		if err != nil {
			log.Printf("[EndGame] Failed to set user %s status to lobby: %v", user.UserID, err)
		} else {
			log.Printf("[EndGame] User %s status set to lobby", user.UserID)
		}
	}

	// Keep room in FINISHED state temporarily
	gameRooms[roomID] = room
	roomMutex.Unlock()

	// Reset room to WAITING after 5 seconds
	go func() {
		time.Sleep(common.RoomResetDelaySeconds * time.Second)
		roomMutex.Lock()
		defer roomMutex.Unlock()

		room, exists := gameRooms[roomID]
		if !exists {
			return
		}

		log.Printf("[EndGame] Resetting room %s from FINISHED to WAITING", roomID)
		room.Status = model.GameRoomStatus(common.RoomStatusWaiting)
		room.CurrentRound = 0

		// Reset all users' ready status
		for i := range room.Users {
			room.Users[i].IsReady = false
			room.Users[i].Score = 0
		}

		gameRooms[roomID] = room
		go publishRoomUpdateToWebSocket(roomID, room)
		go publishLobbyUpdate()
	}()

	// Cleanup pre-loaded quizzes if it's a quiz game
	if room.GameType == model.GameTypeOx || room.GameType == model.GameTypeQa {
		go cleanupPreloadedQuizzes(roomID)
	}

	// Determine winner
	winner := ""
	if len(rankings) > 0 {
		winner = rankings[0].Name
	}

	// Publish update and deletion events
	go func() {
		publishLobbyUpdate()
		// Broadcast game ended with rankings via WebSocket
		publishGameEndedWithRankings(roomID, room, rankings)
		// Notify that room is deleted via WebSocket
		publishRoomDeletedToWebSocket(roomID)
	}()

	if len(rankings) > 0 {
		log.Printf("[EndGame] Game ended for room %s. Winner: %s (%d points)", roomID, winner, rankings[0].Score)
	} else {
		log.Printf("[EndGame] Game ended for room %s. No rankings (empty room)", roomID)
	}

	return &model.GameResult{
		Room:     room,
		Rankings: rankings,
		Winner:   winner,
	}, nil
}

// SendChat is the resolver for the sendChat field.
func (r *mutationResolver) SendChat(ctx context.Context, roomID string, userID string, message string) (*model.ChatMessage, error) {
	chatMsg := &model.ChatMessage{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		Username:  userID,
		Message:   message,
		Timestamp: time.Now(),
	}

	// Chat messages handled by WebSocket layer
	return chatMsg, nil
}

// GameConfig is the resolver for the gameConfig field.
func (r *queryResolver) GameConfig(ctx context.Context) (*model.GameConfig, error) {
	return &model.GameConfig{
		MaxUsers:      8,
		RoundDuration: 60,
		DrawingTime:   30,
		GuessingTime:  30,
		RoundsPerGame: 5,
	}, nil
}

// RandomWordchainPrompt is the resolver for the randomWordchainPrompt field.
func (r *queryResolver) RandomWordchainPrompt(ctx context.Context) (*model.WordchainPrompt, error) {
	// TODO: Implement actual word database
	words := []string{"gorira", "neko", "inu", "sakana", "usagi"}
	word := words[time.Now().Unix()%int64(len(words))]

	return &model.WordchainPrompt{
		Word: word,
		Hint: nil,
	}, nil
}

// GameRoomsUpdated is the resolver for the gameRoomsUpdated field.
func (r *subscriptionResolver) GameRoomsUpdated(ctx context.Context) (<-chan []*model.GameRoom, error) {
	ch := make(chan []*model.GameRoom)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// UserReadyUpdated is the resolver for the userReadyUpdated field.
func (r *subscriptionResolver) UserReadyUpdated(ctx context.Context, roomID string) (<-chan *model.GameUser, error) {
	ch := make(chan *model.GameUser)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// HostChanged is the resolver for the hostChanged field.
func (r *subscriptionResolver) HostChanged(ctx context.Context, roomID string) (<-chan *model.GameUser, error) {
	ch := make(chan *model.GameUser)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// RoundStarted is the resolver for the roundStarted field.
func (r *subscriptionResolver) RoundStarted(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	ch := make(chan *model.GameRoom)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// RoundEnded is the resolver for the roundEnded field.
func (r *subscriptionResolver) RoundEnded(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	ch := make(chan *model.GameRoom)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// GameEnded is the resolver for the gameEnded field.
func (r *subscriptionResolver) GameEnded(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	ch := make(chan *model.GameRoom)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// ChatMessage is the resolver for the chatMessage field.
func (r *subscriptionResolver) ChatMessage(ctx context.Context, roomID string) (<-chan *model.ChatMessage, error) {
	ch := make(chan *model.ChatMessage)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// GameEvent is the resolver for the gameEvent field.
func (r *subscriptionResolver) GameEvent(ctx context.Context, roomID string) (<-chan *model.GameEvent, error) {
	ch := make(chan *model.GameEvent)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// Error is the resolver for the error field.
func (r *subscriptionResolver) Error(ctx context.Context, roomID string) (<-chan *model.ErrorEvent, error) {
	ch := make(chan *model.ErrorEvent)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// MyEvents is the resolver for the myEvents field.
func (r *subscriptionResolver) MyEvents(ctx context.Context, Name string) (<-chan *model.GameEvent, error) {
	ch := make(chan *model.GameEvent)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// UserConnectionStatus is the resolver for the userConnectionStatus field.
func (r *subscriptionResolver) UserConnectionStatus(ctx context.Context, roomID string) (<-chan *model.UserConnection, error) {
	ch := make(chan *model.UserConnection, 1)

	// Connection status is managed by WebSocket layer
	// This subscription is a placeholder for future implementation
	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

// ========================================
// WebSocket Helper Functions
// ========================================

// StartWordchainTurnTimer starts a timer for the current turn in wordchain game
func StartWordchainTurnTimer(roomID string) {
	go func() {
		roomMutex.Lock()
		room, exists := gameRooms[roomID]
		if !exists || room.Status != model.GameRoomStatus(common.RoomStatusPlaying) {
			roomMutex.Unlock()
			return
		}

		timeLimit := int(room.RoundTimeLimit)
		now := time.Now()
		room.WordchainTurnStartTime = &now
		gameRooms[roomID] = room
		roomMutex.Unlock()

		log.Printf("[Wordchain] Turn timer started for %s: %d seconds", *room.CurrentTurnUserID, timeLimit)

		// Countdown timer
		for i := timeLimit; i >= 0; i-- {
			roomMutex.Lock()
			room, exists := gameRooms[roomID]
			if !exists || room.Status != model.GameRoomStatus(common.RoomStatusPlaying) {
				roomMutex.Unlock()
				return
			}

			// Check if turn has changed (user answered correctly)
			if room.WordchainTurnStartTime == nil || !room.WordchainTurnStartTime.Equal(now) {
				roomMutex.Unlock()
				log.Printf("[Wordchain] Turn changed, stopping timer")
				return
			}
			roomMutex.Unlock()

			// Broadcast time remaining
			publishTimerToWebSocket(roomID, i)

			if i > 0 {
				time.Sleep(1 * time.Second)
			}
		}

		// Time's up - end round
		log.Printf("[Wordchain] Time's up for %s, ending round", *room.CurrentTurnUserID)
		EndWordchainRound(roomID, fmt.Sprintf("%s ?? ??", *room.CurrentTurnUserID))
	}()
}

// calculateRankings calculates player rankings based on scores
func calculateRankings(users []*model.GameUser) []*model.GameRanking {
	if len(users) == 0 {
		return []*model.GameRanking{}
	}

	// Create a copy and sort by score (descending)
	sorted := make([]*model.GameUser, len(users))
	copy(sorted, users)

	// Sort by score descending
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Score < sorted[j].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	rankings := make([]*model.GameRanking, len(sorted))
	currentRank := 1

	for i, user := range sorted {
		// Handle ties (same score = same rank)
		if i > 0 && sorted[i].Score == sorted[i-1].Score {
			// Same rank as previous
			rankings[i] = &model.GameRanking{
				Rank:   int32(currentRank),
				UserID: user.UserID,
				Name:   user.Name,
				Score:  user.Score,
			}
		} else {
			// New rank
			currentRank = i + 1
			rankings[i] = &model.GameRanking{
				Rank:   int32(currentRank),
				UserID: user.UserID,
				Name:   user.Name,
				Score:  user.Score,
			}
		}
	}

	log.Printf("[Rankings] Calculated rankings for %d users", len(rankings))
	for i, r := range rankings {
		log.Printf("[Rankings] #%d: Rank %d - %s (%d points)", i+1, r.Rank, r.Name, r.Score)
	}

	return rankings
}

// publishGameEndedWithRankings broadcasts game end event with rankings to WebSocket
func publishGameEndedWithRankings(roomID string, room *model.GameRoom, rankings []*model.GameRanking) {
	channel := common.ChannelGamePrefix + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type": "GAME_ENDED",
		"payload": map[string]interface{}{
			"room":     room,
			"rankings": rankings,
			"winner":   "",
		},
	}

	// Set winner if there are rankings
	if len(rankings) > 0 {
		payload["payload"].(map[string]interface{})["winner"] = rankings[0].Name
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal game ended message: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish game ended to WebSocket: %v", err)
	} else {
		log.Printf("[GAME_ENDED] Published to channel %s with %d rankings", channel, len(rankings))
	}
}

// publishRoomUpdateToWebSocket publishes room update to WebSocket clients via Valkey
func publishRoomUpdateToWebSocket(roomID string, room *model.GameRoom) {
	channel := common.ChannelGamePrefix + roomID

	// Create the message payload
	payload := map[string]interface{}{
		"type": "room_update",
		"data": room,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal room update: %v", err)
		return
	}

	// Publish to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish room update to WebSocket: %v", err)
	} else {
		log.Printf("Published room update to WebSocket channel: %s", channel)
	}
}
