package websocket

import (
	"context"
	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/graph"
	"draw-and-guess-server/internal/valkey"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Global variables for managing WebSocket clients and Valkey subscriptions
var (
	clients             = make(map[*websocket.Conn]*Client)   // All connected WebSocket clients
	mutex               sync.Mutex                            // Mutex for thread-safe clients map access
	valkeySubscriptions = make(map[string]context.CancelFunc) // Valkey channel subscription cancel functions
	subMutex            sync.Mutex                            // Mutex for thread-safe valkeySubscriptions map access
)

// Client represents a WebSocket client with connection and subscription information
type Client struct {
	conn          *websocket.Conn // WebSocket connection
	roomID        string          // Room ID that the client belongs to
	UserID        string          // Client's user ID
	subscriptions map[string]bool // List of subscribed channels
	mu            sync.Mutex      // Mutex for thread-safe struct field access
}

// WSMessage represents the WebSocket message format
type WSMessage struct {
	Type    string      `json:"type"`    // Message type (subscribe, unsubscribe, message)
	Channel string      `json:"channel"` // Target channel name
	Data    interface{} `json:"data"`    // Message data
}

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (for development, restrict in production)
	},
}

// HandleWebSocket upgrades HTTP connection to WebSocket and registers client
func HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	// Create new client
	client := &Client{
		conn:          conn,
		subscriptions: make(map[string]bool),
	}

	// Add to global clients map
	mutex.Lock()
	clients[conn] = client
	mutex.Unlock()

	log.Println("New client connected.")

	// Start message handling goroutine
	go handleMessages(client)
}

// handleDrawingMessage processes drawing messages
func handleDrawingMessage(roomID string, data interface{}) {
	channel := common.ChannelGamePrefix + roomID
	drawingPayload := map[string]interface{}{
		"type":    "DRAW_EVENT",
		"payload": data,
	}
	handlePublish(channel, drawingPayload)
}

// handleQuizAnswer processes quiz answer submissions and awards points
func handleQuizAnswer(roomID string, data interface{}) {
	answerData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid quiz answer data format")
		return
	}

	userID, _ := answerData["userId"].(string)
	quizID, _ := answerData["quizId"].(string)
	userAnswer := answerData["answer"] // bool for OX, float64/int for QA

	if userID == "" {
		log.Printf("Missing userId in quiz answer")
		return
	}

	if quizID == "" {
		log.Printf("Missing quizId in quiz answer")
		return
	}

	// Get user's name from room
	room, exists := graph.GetGameRoom(roomID)
	if !exists {
		log.Printf("Room not found: %s", roomID)
		return
	}

	var username string
	for _, user := range room.Users {
		if user.UserID == userID {
			username = user.Name
			break
		}
	}

	if username == "" {
		log.Printf("User not found in room: userId=%s, roomId=%s", userID, roomID)
		return
	}

	// Get stored quiz info (answer + difficulty)
	quizInfo, exists := graph.GetQuizInfo(roomID, quizID)
	if !exists {
		log.Printf("Quiz answer not found: room=%s, quiz=%s", roomID, quizID)
		return
	}

	// Check if answer is correct
	isCorrect := false
	switch correctAnswer := quizInfo.Answer.(type) {
	case bool: // OX Quiz
		if userBool, ok := userAnswer.(bool); ok {
			isCorrect = (userBool == correctAnswer)
			log.Printf("[Quiz] OX - User: %v, Correct: %v, Match: %v", userBool, correctAnswer, isCorrect)
		}
	case int: // QA Quiz
		if userFloat, ok := userAnswer.(float64); ok {
			isCorrect = (int(userFloat) == correctAnswer)
			log.Printf("[Quiz] QA - User: %d, Correct: %d, Match: %v", int(userFloat), correctAnswer, isCorrect)
		}
	}

	log.Printf("[Quiz] Answer from %s (userId: %s) in room %s: quizId=%s, correct=%v", username, userID, roomID, quizID, isCorrect)

	// Add score silently (will be revealed at ROUND_ENDED)
	if isCorrect {
		score := graph.CalculateQuizScore(quizInfo.Difficulty)
		graph.AddScoreToUserSilent(roomID, username, score)
		log.Printf("[Quiz] ✅ %s will get %d points (difficulty: %d) - score added to pending, waiting for round end", username, score, quizInfo.Difficulty)
	} else {
		log.Printf("[Quiz] ❌ %s got wrong answer - no points, waiting for round end", username)
	}

	// Send answer submission confirmation only (no correct/incorrect info)
	channel := common.ChannelGamePrefix + roomID
	payload := map[string]interface{}{
		"type": "ANSWER_SUBMITTED",
		"payload": map[string]interface{}{
			"userId": userID,
			"quizId": quizID,
		},
	}
	handlePublish(channel, payload)
	log.Printf("[Quiz] 📝 Answer submitted by %s - waiting for timer to reveal results", username)
}

// handleWordchainSubmit processes wordchain word submissions
func handleWordchainSubmit(roomID string, data interface{}) {
	wordData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid wordchain submit data format")
		return
	}

	username, _ := wordData["username"].(string)
	word, _ := wordData["word"].(string)
	lastWord, _ := wordData["lastWord"].(string)

	if username == "" || word == "" {
		log.Printf("Missing required fields for wordchain submit")
		return
	}

	// Validate wordchain rules
	correct := true
	reason := ""
	isDuplicate := false

	// Check for duplicate words first
	if graph.IsWordchainDuplicate(roomID, word) {
		correct = false
		reason = "This word was already used. Please enter a different word."
		isDuplicate = true
		log.Printf("[Wordchain] Duplicate word: %s by %s", word, username)
	} else if lastWord == "" {
		// First word case (no lastWord) - accept any valid word
		correct = true
	} else {
		// Check if first character of new word matches last character of lastWord
		if len(lastWord) > 0 && len(word) > 0 {
			lastChar := []rune(lastWord)[len([]rune(lastWord))-1]
			firstChar := []rune(word)[0]

			if lastChar != firstChar {
				correct = false
				reason = fmt.Sprintf("First character does not match last character '%c' of '%s'", lastChar, lastWord)
			}
		}

		// Check if word ends with ん (invalid in wordchain)
		if correct && len(word) > 0 {
			lastChar := []rune(word)[len([]rune(word))-1]
			if lastChar == 'ん' {
				correct = false
				reason = "Words ending with 'ん' cannot be used"
			}
		}
	}

	log.Printf("Wordchain validation: username=%s, word=%s, correct=%v, duplicate=%v", username, word, correct, isDuplicate)

	// Check if it's this user's turn
	isCorrectTurn := graph.CheckWordchainTurn(roomID, username)
	if !isCorrectTurn {
		correct = false
		reason = "It's not your turn"
		log.Printf("[Wordchain] Not %s's turn in room %s", username, roomID)
	}

	// Publish result to all users via graph package
	graph.PublishWordchainResult(roomID, username, word, correct, reason)

	// If correct, update last word, give points, and move to next user
	if correct && isCorrectTurn {
		graph.UpdateWordchainState(roomID, word, username)
		graph.MoveToNextTurn(roomID)
	} else if isCorrectTurn && !isDuplicate {
		// If wrong answer (but not duplicate) and correct turn, end round
		log.Printf("[Wordchain] Wrong answer from %s: %s. Ending round.", username, reason)
		graph.EndWordchainRound(roomID, fmt.Sprintf("%s gave wrong answer: %s", username, reason))
	}
	// If duplicate, don't move turn - let same user try again
}

// handleLobbyChatMessage processes lobby chat messages and saves/broadcasts them
func handleLobbyChatMessage(client *Client, data interface{}) {
	// Parse data
	chatData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid lobby chat message data format")
		return
	}

	userID, _ := chatData["userId"].(string)
	username, _ := chatData["username"].(string)
	message, _ := chatData["message"].(string)

	if userID == "" || message == "" {
		log.Printf("Missing required lobby chat message fields")
		return
	}

	log.Printf("Lobby chat from %s: %s", username, message)

	// Save chat message to Valkey
	ctx := context.Background()
	if err := valkey.SaveLobbyChatMessage(ctx, userID, username, message); err != nil {
		log.Printf("Failed to save lobby chat message: %v", err)
	}

	// Publish message to lobby channel
	chatPayload := map[string]interface{}{
		"type": "LOBBY_CHAT",
		"payload": map[string]interface{}{
			"userId":   userID,
			"username": username,
			"message":  message,
		},
	}

	handlePublish(common.ChannelLobby, chatPayload)
}

// handleMessages continuously receives and processes messages from client
func handleMessages(client *Client) {
	// Clean up client on function exit
	defer func() {
		// Auto-remove player from room on disconnect
		if client.roomID != "" && client.roomID != common.ChannelLobby && client.UserID != "" {
			log.Printf("Auto-removing user %s from room %s due to WebSocket disconnect", client.UserID, client.roomID)
			removeUserFromRoom(client.roomID, client.UserID)
		}

		mutex.Lock()
		delete(clients, client.conn)
		mutex.Unlock()
		client.conn.Close()
		log.Println("Client disconnected.")
	}()

	log.Println("Waiting for messages from client...")

	// Message receiving loop
	for {
		// Read message from client
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			} else {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		log.Printf("Received message: %s", string(msg))

		// Parse JSON
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// Process message based on type
		switch wsMsg.Type {
		case "subscribe":
			if wsMsg.Channel != "" {
				handleSubscribe(client, wsMsg.Channel)
			}
		case "unsubscribe":
			if wsMsg.Channel != "" {
				handleUnsubscribe(client, wsMsg.Channel)
			}
		case "lobby_chat":
			handleLobbyChatMessage(client, wsMsg.Data)
		case "chat":
			handleChatMessage(client, wsMsg.Data)
		case "drawing":
			if client.roomID != "" {
				handleDrawingMessage(client.roomID, wsMsg.Data)
			}
		case "wordchain_submit":
			if client.roomID != "" {
				handleWordchainSubmit(client.roomID, wsMsg.Data)
			}
		case "quiz_answer":
			if client.roomID != "" {
				handleQuizAnswer(client.roomID, wsMsg.Data)
			}
		default:
			log.Printf("Unknown message type: %s", wsMsg.Type)
		}
	}
}

// handleSubscribe processes client's channel subscription request
func handleSubscribe(client *Client, channel string) {
	// Add to client subscriptions
	client.mu.Lock()
	client.subscriptions[channel] = true
	userID := client.UserID
	if client.roomID == "" && len(channel) > 0 {
		// Extract room ID from channel (e.g. "game/room-uuid" -> "room-uuid")
		parts := splitChannel(channel)
		if len(parts) > 1 {
			client.roomID = parts[1]
		}
	}
	client.mu.Unlock()

	// Start Valkey subscription if not already created
	subMutex.Lock()
	if _, exists := valkeySubscriptions[channel]; !exists {
		ctx, cancel := context.WithCancel(context.Background())
		valkeySubscriptions[channel] = cancel
		go startValkeySubscription(ctx, channel) // Start Valkey subscription as goroutine
		log.Printf("[Subscribe] 🆕 Started Valkey subscription for channel: %s", channel)
	}
	subMutex.Unlock()

	// Log subscription based on channel type (lobby or game room)
	if channel == common.ChannelLobby {
		logger := common.GetLogger()
		logger.Info("[Subscribe] 🏠 LOBBY - User '%s' subscribed to lobby", userID)
	} else if len(channel) > len(common.ChannelGamePrefix) && channel[:len(common.ChannelGamePrefix)] == common.ChannelGamePrefix {
		roomID := channel[len(common.ChannelGamePrefix):]
		log.Printf("[Subscribe] 🎮 GAME ROOM - User '%s' subscribed to room: %s", userID, roomID)
	} else {
		log.Printf("[Subscribe] ✅ User '%s' subscribed to channel: %s", userID, channel)
	}
}

// handleUnsubscribe processes client's channel unsubscription request
func handleUnsubscribe(client *Client, channel string) {
	client.mu.Lock()
	userID := client.UserID
	delete(client.subscriptions, channel) // Remove from subscription list
	client.mu.Unlock()

	// Log unsubscription based on channel type (lobby or game room)
	if channel == common.ChannelLobby {
		logger := common.GetLogger()
		logger.Info("[Unsubscribe] 🏠 LOBBY - User '%s' unsubscribed from lobby", userID)
	} else if len(channel) > len(common.ChannelGamePrefix) && channel[:len(common.ChannelGamePrefix)] == common.ChannelGamePrefix {
		roomID := channel[len(common.ChannelGamePrefix):]
		log.Printf("[Unsubscribe] 🎮 GAME ROOM - User '%s' unsubscribed from room: %s", userID, roomID)
	} else {
		log.Printf("[Unsubscribe] ❌ User '%s' unsubscribed from channel: %s", userID, channel)
	}

	// Send unsubscription confirmation message to client
	confirmMsg := map[string]interface{}{
		"type":    "unsubscribed",
		"channel": channel,
		"message": "Successfully unsubscribed from " + channel,
	}
	if msgBytes, err := json.Marshal(confirmMsg); err == nil {
		client.conn.WriteMessage(websocket.TextMessage, msgBytes)
	}
}

// handlePublish publishes message to Valkey channel
func handlePublish(channel string, data interface{}) {
	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal data: %v", err)
		return
	}

	log.Printf("Publishing to Valkey channel: %s", channel)

	// Publish message to Valkey
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish to Valkey: %v", err)
	}
}

// handleChatMessage processes chat messages and broadcasts to all clients in the room
func handleChatMessage(client *Client, data interface{}) {
	// Parse data
	chatData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid chat message data format")
		return
	}

	roomID, _ := chatData["roomId"].(string)
	userID, _ := chatData["userId"].(string)
	username, _ := chatData["username"].(string)
	message, _ := chatData["message"].(string)

	if roomID == "" || userID == "" || message == "" {
		log.Printf("Missing required chat message fields")
		return
	}

	log.Printf("Chat message from %s in room %s: %s", username, roomID, message)

	// Publish chat message to Valkey channel (real-time only, no storage)
	channel := common.ChannelGamePrefix + roomID
	chatPayload := ChatMessagePayload{
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		Message:  message,
	}

	handlePublish(channel, map[string]interface{}{
		"type":    "CHAT_MESSAGE",
		"payload": chatPayload,
	})
}

// startValkeySubscription subscribes to Valkey channel and broadcasts messages to WebSocket clients
func startValkeySubscription(ctx context.Context, channel string) {
	log.Printf("Starting Valkey subscription for channel: %s", channel)

	// Start Valkey channel subscription
	pubsub := valkey.SubscribeChannel(ctx, channel)
	if pubsub == nil {
		log.Printf("Failed to subscribe to channel: %s", channel)
		return
	}
	defer pubsub.Close()

	// Message receiving loop
	ch := pubsub.Channel()
	for msg := range ch {
		message := msg.Payload
		log.Printf("Received message from Valkey channel %s: %s", channel, message)

		// Broadcast to all WebSocket clients subscribed to this channel
		mutex.Lock()
		subscribedCount := 0
		totalClients := len(clients)
		for _, client := range clients {
			// Check if client is subscribed to this channel
			client.mu.Lock()
			subscribed := client.subscriptions[channel]
			client.mu.Unlock()

			if subscribed {
				subscribedCount++
				// Create WebSocket message
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

				// Send message to client
				if err := client.conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					log.Printf("Error sending message to client: %v", err)
				}
			}
		}
		log.Printf("[Broadcast] Channel '%s': sent to %d/%d clients", channel, subscribedCount, totalClients)
		mutex.Unlock()
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

// removeUserFromRoom removes player from room by userID
func removeUserFromRoom(roomID, userID string) {
	graph.RemoveUserFromRoom(roomID, userID)
}
