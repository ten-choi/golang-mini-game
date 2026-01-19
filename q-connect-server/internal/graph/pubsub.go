package graph

import (
	"sync"

	"draw-and-guess-server/internal/graph/model"
)

// PubSub manages subscriptions for real-time updates
type PubSub struct {
	mu sync.RWMutex

	// Room-specific subscribers
	roomSubscribers      map[string]map[string]chan *model.GameRoom
	userJoinedSubs       map[string]map[string]chan *model.GameUser
	userLeftSubs         map[string]map[string]chan *model.GameUser
	userReadyUpdatedSubs map[string]map[string]chan *model.GameUser
	hostChangedSubs      map[string]map[string]chan *model.GameUser
	gameStartedSubs      map[string]map[string]chan *model.GameRoom
	roundStartedSubs     map[string]map[string]chan *model.GameRoom
	roundEndedSubs       map[string]map[string]chan *model.GameRoom
	gameEndedSubs        map[string]map[string]chan *model.GameRoom
	chatMessageSubs      map[string]map[string]chan *model.ChatMessage
	gameEventSubs        map[string]map[string]chan *model.GameEvent
	errorSubs            map[string]map[string]chan *model.ErrorEvent

	// Lobby subscribers (all rooms)
	lobbySubscribers map[string]chan []*model.GameRoom
}

// Global PubSub instance
var pubsub = NewPubSub()

// NewPubSub creates a new PubSub instance
func NewPubSub() *PubSub {
	return &PubSub{
		roomSubscribers:      make(map[string]map[string]chan *model.GameRoom),
		userJoinedSubs:       make(map[string]map[string]chan *model.GameUser),
		userLeftSubs:         make(map[string]map[string]chan *model.GameUser),
		userReadyUpdatedSubs: make(map[string]map[string]chan *model.GameUser),
		hostChangedSubs:      make(map[string]map[string]chan *model.GameUser),
		gameStartedSubs:      make(map[string]map[string]chan *model.GameRoom),
		roundStartedSubs:     make(map[string]map[string]chan *model.GameRoom),
		roundEndedSubs:       make(map[string]map[string]chan *model.GameRoom),
		gameEndedSubs:        make(map[string]map[string]chan *model.GameRoom),
		chatMessageSubs:      make(map[string]map[string]chan *model.ChatMessage),
		gameEventSubs:        make(map[string]map[string]chan *model.GameEvent),
		errorSubs:            make(map[string]map[string]chan *model.ErrorEvent),
		lobbySubscribers:     make(map[string]chan []*model.GameRoom),
	}
}

// GetPubSub returns the global PubSub instance
func GetPubSub() *PubSub {
	return pubsub
}

// SubscribeToRoom subscribes to a specific room's updates
func (ps *PubSub) SubscribeToRoom(roomID, subscriberID string) <-chan *model.GameRoom {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.roomSubscribers[roomID] == nil {
		ps.roomSubscribers[roomID] = make(map[string]chan *model.GameRoom)
	}

	ch := make(chan *model.GameRoom, 1)
	ps.roomSubscribers[roomID][subscriberID] = ch
	return ch
}

// UnsubscribeFromRoom unsubscribes from a room
func (ps *PubSub) UnsubscribeFromRoom(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.roomSubscribers[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

// PublishRoomUpdate publishes an update to all room subscribers
func (ps *PubSub) PublishRoomUpdate(room *model.GameRoom) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.roomSubscribers[room.ID]; ok {
		for _, ch := range subs {
			select {
			case ch <- room:
			default:
				// Channel full, skip
			}
		}
	}
}

// SubscribeToLobby subscribes to all room updates
func (ps *PubSub) SubscribeToLobby(subscriberID string) <-chan []*model.GameRoom {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan []*model.GameRoom, 1)
	ps.lobbySubscribers[subscriberID] = ch
	return ch
}

// UnsubscribeFromLobby unsubscribes from lobby updates
func (ps *PubSub) UnsubscribeFromLobby(subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ch, ok := ps.lobbySubscribers[subscriberID]; ok {
		close(ch)
		delete(ps.lobbySubscribers, subscriberID)
	}
}

// PublishLobbyUpdate publishes all rooms to lobby subscribers
func (ps *PubSub) PublishLobbyUpdate(rooms []*model.GameRoom) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, ch := range ps.lobbySubscribers {
		select {
		case ch <- rooms:
		default:
			// Channel full, skip
		}
	}
}

// SubscribeToUserJoined subscribes to user join events
func (ps *PubSub) SubscribeToUserJoined(roomID, subscriberID string) <-chan *model.GameUser {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.userJoinedSubs[roomID] == nil {
		ps.userJoinedSubs[roomID] = make(map[string]chan *model.GameUser)
	}

	ch := make(chan *model.GameUser, 1)
	ps.userJoinedSubs[roomID][subscriberID] = ch
	return ch
}

// UnsubscribeFromUserJoined unsubscribes from user join events
func (ps *PubSub) UnsubscribeFromUserJoined(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.userJoinedSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

// PublishUserJoined publishes a user join event
func (ps *PubSub) PublishUserJoined(roomID string, user *model.GameUser) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.userJoinedSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- user:
			default:
			}
		}
	}
}

// SubscribeToUserLeft subscribes to user leave events
func (ps *PubSub) SubscribeToUserLeft(roomID, subscriberID string) <-chan *model.GameUser {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.userLeftSubs[roomID] == nil {
		ps.userLeftSubs[roomID] = make(map[string]chan *model.GameUser)
	}

	ch := make(chan *model.GameUser, 1)
	ps.userLeftSubs[roomID][subscriberID] = ch
	return ch
}

// UnsubscribeFromUserLeft unsubscribes from user leave events
func (ps *PubSub) UnsubscribeFromUserLeft(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.userLeftSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

// PublishUserLeft publishes a user leave event
func (ps *PubSub) PublishUserLeft(roomID string, user *model.GameUser) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.userLeftSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- user:
			default:
			}
		}
	}
}

// SubscribeToGameStarted subscribes to game start events
func (ps *PubSub) SubscribeToGameStarted(roomID, subscriberID string) <-chan *model.GameRoom {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.gameStartedSubs[roomID] == nil {
		ps.gameStartedSubs[roomID] = make(map[string]chan *model.GameRoom)
	}

	ch := make(chan *model.GameRoom, 1)
	ps.gameStartedSubs[roomID][subscriberID] = ch
	return ch
}

// UnsubscribeFromGameStarted unsubscribes from game start events
func (ps *PubSub) UnsubscribeFromGameStarted(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.gameStartedSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

// PublishGameStarted publishes a game start event
func (ps *PubSub) PublishGameStarted(room *model.GameRoom) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.gameStartedSubs[room.ID]; ok {
		for _, ch := range subs {
			select {
			case ch <- room:
			default:
			}
		}
	}
}

// UserReadyUpdated subscription methods
func (ps *PubSub) SubscribeToUserReadyUpdated(roomID, subscriberID string) <-chan *model.GameUser {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.userReadyUpdatedSubs[roomID] == nil {
		ps.userReadyUpdatedSubs[roomID] = make(map[string]chan *model.GameUser)
	}

	ch := make(chan *model.GameUser, 1)
	ps.userReadyUpdatedSubs[roomID][subscriberID] = ch
	return ch
}

func (ps *PubSub) UnsubscribeFromUserReadyUpdated(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.userReadyUpdatedSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

func (ps *PubSub) PublishUserReadyUpdated(roomID string, user *model.GameUser) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.userReadyUpdatedSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- user:
			default:
			}
		}
	}
}

// HostChanged subscription methods
func (ps *PubSub) SubscribeToHostChanged(roomID, subscriberID string) <-chan *model.GameUser {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.hostChangedSubs[roomID] == nil {
		ps.hostChangedSubs[roomID] = make(map[string]chan *model.GameUser)
	}

	ch := make(chan *model.GameUser, 1)
	ps.hostChangedSubs[roomID][subscriberID] = ch
	return ch
}

func (ps *PubSub) UnsubscribeFromHostChanged(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.hostChangedSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

func (ps *PubSub) PublishHostChanged(roomID string, user *model.GameUser) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.hostChangedSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- user:
			default:
			}
		}
	}
}

// RoundStarted subscription methods
func (ps *PubSub) SubscribeToRoundStarted(roomID, subscriberID string) <-chan *model.GameRoom {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.roundStartedSubs[roomID] == nil {
		ps.roundStartedSubs[roomID] = make(map[string]chan *model.GameRoom)
	}

	ch := make(chan *model.GameRoom, 1)
	ps.roundStartedSubs[roomID][subscriberID] = ch
	return ch
}

func (ps *PubSub) UnsubscribeFromRoundStarted(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.roundStartedSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

func (ps *PubSub) PublishRoundStarted(roomID string, room *model.GameRoom) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.roundStartedSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- room:
			default:
			}
		}
	}
}

// RoundEnded subscription methods
func (ps *PubSub) SubscribeToRoundEnded(roomID, subscriberID string) <-chan *model.GameRoom {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.roundEndedSubs[roomID] == nil {
		ps.roundEndedSubs[roomID] = make(map[string]chan *model.GameRoom)
	}

	ch := make(chan *model.GameRoom, 1)
	ps.roundEndedSubs[roomID][subscriberID] = ch
	return ch
}

func (ps *PubSub) UnsubscribeFromRoundEnded(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.roundEndedSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

func (ps *PubSub) PublishRoundEnded(roomID string, room *model.GameRoom) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.roundEndedSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- room:
			default:
			}
		}
	}
}

// GameEnded subscription methods
func (ps *PubSub) SubscribeToGameEnded(roomID, subscriberID string) <-chan *model.GameRoom {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.gameEndedSubs[roomID] == nil {
		ps.gameEndedSubs[roomID] = make(map[string]chan *model.GameRoom)
	}

	ch := make(chan *model.GameRoom, 1)
	ps.gameEndedSubs[roomID][subscriberID] = ch
	return ch
}

func (ps *PubSub) UnsubscribeFromGameEnded(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.gameEndedSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

func (ps *PubSub) PublishGameEnded(roomID string, room *model.GameRoom) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.gameEndedSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- room:
			default:
			}
		}
	}
}

// ChatMessage subscription methods
func (ps *PubSub) SubscribeToChatMessage(roomID, subscriberID string) <-chan *model.ChatMessage {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.chatMessageSubs[roomID] == nil {
		ps.chatMessageSubs[roomID] = make(map[string]chan *model.ChatMessage)
	}

	ch := make(chan *model.ChatMessage, 1)
	ps.chatMessageSubs[roomID][subscriberID] = ch
	return ch
}

func (ps *PubSub) UnsubscribeFromChatMessage(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.chatMessageSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

func (ps *PubSub) PublishChatMessage(roomID string, message *model.ChatMessage) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.chatMessageSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- message:
			default:
			}
		}
	}
}

// GameEvent subscription methods
func (ps *PubSub) SubscribeToGameEvent(roomID, subscriberID string) <-chan *model.GameEvent {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.gameEventSubs[roomID] == nil {
		ps.gameEventSubs[roomID] = make(map[string]chan *model.GameEvent)
	}

	ch := make(chan *model.GameEvent, 1)
	ps.gameEventSubs[roomID][subscriberID] = ch
	return ch
}

func (ps *PubSub) UnsubscribeFromGameEvent(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.gameEventSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

func (ps *PubSub) PublishGameEvent(roomID string, event *model.GameEvent) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.gameEventSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

// ErrorEvent subscription methods
func (ps *PubSub) SubscribeToError(roomID, subscriberID string) <-chan *model.ErrorEvent {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.errorSubs[roomID] == nil {
		ps.errorSubs[roomID] = make(map[string]chan *model.ErrorEvent)
	}

	ch := make(chan *model.ErrorEvent, 1)
	ps.errorSubs[roomID][subscriberID] = ch
	return ch
}

func (ps *PubSub) UnsubscribeFromError(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.errorSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

func (ps *PubSub) PublishError(roomID string, errorEvent *model.ErrorEvent) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.errorSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- errorEvent:
			default:
			}
		}
	}
}
