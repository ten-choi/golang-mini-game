package websocket

import (
	"context"
	"draw-and-guess-server/src/valkey"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	clients             = make(map[*websocket.Conn]*Client)
	mutex               sync.Mutex
	valkeySubscriptions = make(map[string]context.CancelFunc)
	subMutex            sync.Mutex
)

type Client struct {
	conn          *websocket.Conn
	roomID        string
	subscriptions map[string]bool
	mu            sync.Mutex
}

type WSMessage struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	client := &Client{
		conn:          conn,
		subscriptions: make(map[string]bool),
	}

	mutex.Lock()
	clients[conn] = client
	mutex.Unlock()

	log.Println("New client connected.")

	go handleMessages(client)
}

func handleMessages(client *Client) {
	defer func() {
		mutex.Lock()
		delete(clients, client.conn)
		mutex.Unlock()
		client.conn.Close()
		log.Println("Client disconnected.")
	}()

	log.Println("Waiting for messages from client...")

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			} else {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		log.Printf("Received raw message: %s", string(msg))

		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		log.Printf("Parsed message: type=%s, channel=%s", wsMsg.Type, wsMsg.Channel)

		switch wsMsg.Type {
		case "subscribe":
			handleSubscribe(client, wsMsg.Channel)
		case "unsubscribe":
			handleUnsubscribe(client, wsMsg.Channel)
		case "message":
			handlePublish(wsMsg.Channel, wsMsg.Data)
		default:
			log.Printf("Unknown message type: %s", wsMsg.Type)
		}
	}
}

func handleSubscribe(client *Client, channel string) {
	client.mu.Lock()
	client.subscriptions[channel] = true
	if client.roomID == "" && len(channel) > 0 {
		// Extract room ID from channel (e.g., "game/room-uuid" -> "room-uuid")
		parts := splitChannel(channel)
		if len(parts) > 1 {
			client.roomID = parts[1]
		}
	}
	client.mu.Unlock()

	// Start Valkey subscription if not already active
	subMutex.Lock()
	if _, exists := valkeySubscriptions[channel]; !exists {
		ctx, cancel := context.WithCancel(context.Background())
		valkeySubscriptions[channel] = cancel
		go startValkeySubscription(ctx, channel)
	}
	subMutex.Unlock()

	log.Printf("Client subscribed to channel: %s", channel)
}

func handleUnsubscribe(client *Client, channel string) {
	client.mu.Lock()
	delete(client.subscriptions, channel)
	client.mu.Unlock()
	log.Printf("Client unsubscribed from channel: %s", channel)
}

func handlePublish(channel string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal data: %v", err)
		return
	}

	log.Printf("Publishing to Valkey channel: %s", channel)

	if err := valkey.PublishMessage(channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish to Valkey: %v", err)
	}
}

func startValkeySubscription(ctx context.Context, channel string) {
	log.Printf("Starting Valkey subscription for channel: %s", channel)

	err := valkey.SubscribeChannel(ctx, channel, func(message string) {
		log.Printf("Received message from Valkey channel %s: %s", channel, message)

		// Broadcast to all WebSocket clients subscribed to this channel
		mutex.Lock()
		defer mutex.Unlock()

		for _, client := range clients {
			client.mu.Lock()
			subscribed := client.subscriptions[channel]
			client.mu.Unlock()

			if subscribed {
				wsMsg := WSMessage{
					Type:    "message",
					Channel: channel,
					Data:    json.RawMessage(message),
				}
				msgBytes, err := json.Marshal(wsMsg)
				if err != nil {
					log.Printf("Failed to marshal message: %v", err)
					continue
				}

				if err := client.conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					log.Printf("Error sending message to client: %v", err)
				}
			}
		}
	})

	if err != nil {
		log.Printf("Valkey subscription error for channel %s: %v", channel, err)
	}
}

func splitChannel(channel string) []string {
	parts := []string{}
	current := ""
	for _, ch := range channel {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
