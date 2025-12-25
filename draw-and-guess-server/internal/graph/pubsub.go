package graph

import (
	"sync"

	"draw-and-guess-server/internal/graph/model"
)

// PubSub manages subscriptions for real-time updates
type PubSub struct {
	mu sync.RWMutex

	// Room-specific subscribers
	roomSubscribers  map[string]map[string]chan *model.GameRoom
	playerJoinedSubs map[string]map[string]chan *model.Player
	playerLeftSubs   map[string]map[string]chan *model.Player
	gameStartedSubs  map[string]map[string]chan *model.GameRoom

	// Lobby subscribers (all rooms)
	lobbySubscribers map[string]chan []*model.GameRoom
}

// Global PubSub instance
var pubsub = NewPubSub()

// NewPubSub creates a new PubSub instance
func NewPubSub() *PubSub {
	return &PubSub{
		roomSubscribers:  make(map[string]map[string]chan *model.GameRoom),
		playerJoinedSubs: make(map[string]map[string]chan *model.Player),
		playerLeftSubs:   make(map[string]map[string]chan *model.Player),
		gameStartedSubs:  make(map[string]map[string]chan *model.GameRoom),
		lobbySubscribers: make(map[string]chan []*model.GameRoom),
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

// SubscribeToPlayerJoined subscribes to player join events
func (ps *PubSub) SubscribeToPlayerJoined(roomID, subscriberID string) <-chan *model.Player {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.playerJoinedSubs[roomID] == nil {
		ps.playerJoinedSubs[roomID] = make(map[string]chan *model.Player)
	}

	ch := make(chan *model.Player, 1)
	ps.playerJoinedSubs[roomID][subscriberID] = ch
	return ch
}

// UnsubscribeFromPlayerJoined unsubscribes from player join events
func (ps *PubSub) UnsubscribeFromPlayerJoined(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.playerJoinedSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

// PublishPlayerJoined publishes a player join event
func (ps *PubSub) PublishPlayerJoined(roomID string, player *model.Player) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.playerJoinedSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- player:
			default:
			}
		}
	}
}

// SubscribeToPlayerLeft subscribes to player leave events
func (ps *PubSub) SubscribeToPlayerLeft(roomID, subscriberID string) <-chan *model.Player {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.playerLeftSubs[roomID] == nil {
		ps.playerLeftSubs[roomID] = make(map[string]chan *model.Player)
	}

	ch := make(chan *model.Player, 1)
	ps.playerLeftSubs[roomID][subscriberID] = ch
	return ch
}

// UnsubscribeFromPlayerLeft unsubscribes from player leave events
func (ps *PubSub) UnsubscribeFromPlayerLeft(roomID, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.playerLeftSubs[roomID]; ok {
		if ch, ok := subs[subscriberID]; ok {
			close(ch)
			delete(subs, subscriberID)
		}
	}
}

// PublishPlayerLeft publishes a player leave event
func (ps *PubSub) PublishPlayerLeft(roomID string, player *model.Player) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.playerLeftSubs[roomID]; ok {
		for _, ch := range subs {
			select {
			case ch <- player:
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
