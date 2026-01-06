package graph

import (
	"context"
	"draw-and-guess-server/internal/graph/model"
	"fmt"
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

	room := &model.GameRoom{
		ID:           roomID,
		Name:         input.Name,
		GameType:     input.GameType,
		Status:       model.GameStatusWaiting,
		CurrentRound: 0,
		TotalRounds:  int32(input.TotalRounds),
		Players:      []*model.Player{},
		MaxPlayers:   int32(input.MaxPlayers),
		HostUsername: input.HostUsername,
		UsedQuizIds:  []string{}, // Initialize empty quiz history
		CreatedAt:    now,
	}

	// Add host as first player
	room.Players = append(room.Players, &model.Player{
		Username:    input.HostUsername,
		DisplayName: input.HostUsername,
		Score:       0,
		IsReady:     true,
	})

	roomMutex.Lock()
	gameRooms[roomID] = room
	roomMutex.Unlock()

	// Publish lobby update
	go publishLobbyUpdate()

	return room, nil
}

// JoinGameRoom is the resolver for the joinGameRoom field.
func (r *mutationResolver) JoinGameRoom(ctx context.Context, roomID string, username string, password *string) (*model.GameRoom, error) {
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

	// Check if player already in room
	for _, p := range room.Players {
		if p.Username == username {
			return room, nil // Already in room
		}
	}

	// Add player
	newPlayer := &model.Player{
		Username:    username,
		DisplayName: username,
		Score:       0,
		IsReady:     false,
	}
	room.Players = append(room.Players, newPlayer)

	// Publish events
	go func() {
		GetPubSub().PublishRoomUpdate(room)
		GetPubSub().PublishPlayerJoined(roomID, newPlayer)
		publishLobbyUpdate()
	}()

	return room, nil
}

// LeaveGameRoom is the resolver for the leaveGameRoom field.
func (r *mutationResolver) LeaveGameRoom(ctx context.Context, roomID string, username string) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	// Remove player and capture the leaving player
	var leavingPlayer *model.Player
	newPlayers := []*model.Player{}
	for _, p := range room.Players {
		if p.Username != username {
			newPlayers = append(newPlayers, p)
		} else {
			leavingPlayer = p
		}
	}
	room.Players = newPlayers

	// Delete room if empty
	if len(room.Players) == 0 {
		delete(gameRooms, roomID)
		go publishLobbyUpdate()
		return room, nil
	}

	// Publish events
	if leavingPlayer != nil {
		go func() {
			GetPubSub().PublishRoomUpdate(room)
			GetPubSub().PublishPlayerLeft(roomID, leavingPlayer)
			publishLobbyUpdate()
		}()
	}

	return room, nil
}

// StartGame is the resolver for the startGame field.
func (r *mutationResolver) StartGame(ctx context.Context, roomID string) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	room.Status = model.GameStatusPlaying
	room.CurrentRound = 1

	// Publish events
	go func() {
		GetPubSub().PublishRoomUpdate(room)
		GetPubSub().PublishGameStarted(room)
		publishLobbyUpdate()
	}()

	return room, nil
}

// DeleteGameRoom is the resolver for the deleteGameRoom field.
func (r *mutationResolver) DeleteGameRoom(ctx context.Context, roomID string) (bool, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	delete(gameRooms, roomID)
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
func (r *queryResolver) GameRooms(ctx context.Context, gameType *model.GameType) ([]*model.GameRoom, error) {
	roomMutex.RLock()
	defer roomMutex.RUnlock()

	result := []*model.GameRoom{}
	for _, room := range gameRooms {
		if gameType == nil || room.GameType == *gameType {
			result = append(result, room)
		}
	}

	return result, nil
}

// ========================================
// GameRoom Subscriptions
// ========================================

// GameRoomUpdated is the resolver for the gameRoomUpdated field.
func (r *subscriptionResolver) GameRoomUpdated(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToRoom(roomID, subscriberID)

	// Unsubscribe when context is done
	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromRoom(roomID, subscriberID)
	}()

	return ch, nil
}

// LobbyUpdated is the resolver for the lobbyUpdated field.
func (r *subscriptionResolver) LobbyUpdated(ctx context.Context) (<-chan []*model.GameRoom, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToLobby(subscriberID)

	// Unsubscribe when context is done
	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromLobby(subscriberID)
	}()

	return ch, nil
}

// PlayerJoined is the resolver for the playerJoined field.
func (r *subscriptionResolver) PlayerJoined(ctx context.Context, roomID string) (<-chan *model.Player, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToPlayerJoined(roomID, subscriberID)

	// Unsubscribe when context is done
	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromPlayerJoined(roomID, subscriberID)
	}()

	return ch, nil
}

// PlayerLeft is the resolver for the playerLeft field.
func (r *subscriptionResolver) PlayerLeft(ctx context.Context, roomID string) (<-chan *model.Player, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToPlayerLeft(roomID, subscriberID)

	// Unsubscribe when context is done
	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromPlayerLeft(roomID, subscriberID)
	}()

	return ch, nil
}

// GameStarted is the resolver for the gameStarted field.
func (r *subscriptionResolver) GameStarted(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToGameStarted(roomID, subscriberID)

	// Unsubscribe when context is done
	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromGameStarted(roomID, subscriberID)
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
	if input.MaxPlayers != nil {
		room.MaxPlayers = *input.MaxPlayers
	}
	if input.TotalRounds != nil {
		room.TotalRounds = *input.TotalRounds
	}
	if input.IsPrivate != nil {
		room.IsPrivate = *input.IsPrivate
	}
	if input.Password != nil {
		room.Password = input.Password
	}

	// Publish update
	go func() {
		GetPubSub().PublishRoomUpdate(room)
		publishLobbyUpdate()
	}()

	return room, nil
}

// TransferHost is the resolver for the transferHost field.
func (r *mutationResolver) TransferHost(ctx context.Context, roomID string, newHostUsername string) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	// Check if new host is in the room
	found := false
	for _, p := range room.Players {
		if p.Username == newHostUsername {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("new host is not in the room")
	}

	room.HostUsername = newHostUsername

	// Publish update
	go func() {
		GetPubSub().PublishRoomUpdate(room)
		GetPubSub().PublishHostChanged(roomID, &model.Player{Username: newHostUsername, DisplayName: newHostUsername})
	}()

	return room, nil
}

// SetReady is the resolver for the setReady field.
func (r *mutationResolver) SetReady(ctx context.Context, roomID string, username string, ready bool) (bool, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return false, fmt.Errorf("room not found")
	}

	// Find player and update ready status
	for i, p := range room.Players {
		if p.Username == username {
			room.Players[i].IsReady = ready

			// Publish update
			go func() {
				GetPubSub().PublishRoomUpdate(room)
				GetPubSub().PublishPlayerReadyUpdated(roomID, room.Players[i])
			}()

			return true, nil
		}
	}

	return false, fmt.Errorf("player not found in room")
}

// StartRound is the resolver for the startRound field.
func (r *mutationResolver) StartRound(ctx context.Context, roomID string) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	if room.Status != model.GameStatusPlaying {
		return nil, fmt.Errorf("game is not in playing status")
	}

	room.CurrentRound++

	// Publish update
	go func() {
		GetPubSub().PublishRoomUpdate(room)
		GetPubSub().PublishRoundStarted(roomID, room)
	}()

	return room, nil
}

// SubmitAnswer is the resolver for the submitAnswer field.
func (r *mutationResolver) SubmitAnswer(ctx context.Context, roomID string, username string, answer string) (bool, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return false, fmt.Errorf("room not found")
	}

	// Find player
	var player *model.Player
	for i := range room.Players {
		if room.Players[i].Username == username {
			player = room.Players[i]
			break
		}
	}
	if player == nil {
		return false, fmt.Errorf("player not found in room")
	}

	// TODO: Validate answer and update score
	// This is a placeholder - actual logic depends on game type

	return true, nil
}

// EndGame is the resolver for the endGame field.
func (r *mutationResolver) EndGame(ctx context.Context, roomID string) (*model.GameRoom, error) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	room.Status = model.GameStatusFinished

	// Publish update
	go func() {
		GetPubSub().PublishRoomUpdate(room)
		GetPubSub().PublishGameEnded(roomID, room)
	}()

	return room, nil
}

// SendChat is the resolver for the sendChat field.
func (r *mutationResolver) SendChat(ctx context.Context, roomID string, username string, message string) (*model.ChatMessage, error) {
	chatMsg := &model.ChatMessage{
		ID:          uuid.New().String(),
		RoomID:      roomID,
		Username:    username,
		DisplayName: username,
		Message:     message,
		Timestamp:   time.Now(),
	}

	// Publish chat message
	go GetPubSub().PublishChatMessage(roomID, chatMsg)

	return chatMsg, nil
}

// GameConfig is the resolver for the gameConfig field.
func (r *queryResolver) GameConfig(ctx context.Context) (*model.GameConfig, error) {
	return &model.GameConfig{
		MaxPlayers:    8,
		RoundDuration: 60,
		DrawingTime:   30,
		GuessingTime:  30,
		RoundsPerGame: 5,
	}, nil
}

// RandomWordchainPrompt is the resolver for the randomWordchainPrompt field.
func (r *queryResolver) RandomWordchainPrompt(ctx context.Context) (*model.WordchainPrompt, error) {
	// TODO: Implement actual word database
	words := []string{"사과", "포도", "딸기", "바나나", "수박"}
	word := words[time.Now().Unix()%int64(len(words))]

	return &model.WordchainPrompt{
		Word: word,
		Hint: nil,
	}, nil
}

// GameRoomsUpdated is the resolver for the gameRoomsUpdated field.
func (r *subscriptionResolver) GameRoomsUpdated(ctx context.Context) (<-chan []*model.GameRoom, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToLobby(subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromLobby(subscriberID)
	}()

	return ch, nil
}

// PlayerReadyUpdated is the resolver for the playerReadyUpdated field.
func (r *subscriptionResolver) PlayerReadyUpdated(ctx context.Context, roomID string) (<-chan *model.Player, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToPlayerReadyUpdated(roomID, subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromPlayerReadyUpdated(roomID, subscriberID)
	}()

	return ch, nil
}

// HostChanged is the resolver for the hostChanged field.
func (r *subscriptionResolver) HostChanged(ctx context.Context, roomID string) (<-chan *model.Player, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToHostChanged(roomID, subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromHostChanged(roomID, subscriberID)
	}()

	return ch, nil
}

// RoundStarted is the resolver for the roundStarted field.
func (r *subscriptionResolver) RoundStarted(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToRoundStarted(roomID, subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromRoundStarted(roomID, subscriberID)
	}()

	return ch, nil
}

// RoundEnded is the resolver for the roundEnded field.
func (r *subscriptionResolver) RoundEnded(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToRoundEnded(roomID, subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromRoundEnded(roomID, subscriberID)
	}()

	return ch, nil
}

// GameEnded is the resolver for the gameEnded field.
func (r *subscriptionResolver) GameEnded(ctx context.Context, roomID string) (<-chan *model.GameRoom, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToGameEnded(roomID, subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromGameEnded(roomID, subscriberID)
	}()

	return ch, nil
}

// ChatMessage is the resolver for the chatMessage field.
func (r *subscriptionResolver) ChatMessage(ctx context.Context, roomID string) (<-chan *model.ChatMessage, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToChatMessage(roomID, subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromChatMessage(roomID, subscriberID)
	}()

	return ch, nil
}

// GameEvent is the resolver for the gameEvent field.
func (r *subscriptionResolver) GameEvent(ctx context.Context, roomID string) (<-chan *model.GameEvent, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToGameEvent(roomID, subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromGameEvent(roomID, subscriberID)
	}()

	return ch, nil
}

// Error is the resolver for the error field.
func (r *subscriptionResolver) Error(ctx context.Context, roomID string) (<-chan *model.ErrorEvent, error) {
	subscriberID := uuid.New().String()
	ch := GetPubSub().SubscribeToError(roomID, subscriberID)

	go func() {
		<-ctx.Done()
		GetPubSub().UnsubscribeFromError(roomID, subscriberID)
	}()

	return ch, nil
}
