package websocket

import (
	"context"
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

// 전역 변수: WebSocket 클라이언트와 Valkey 구독 관리
var (
	clients             = make(map[*websocket.Conn]*Client)   // 연결된 모든 WebSocket 클라이언트
	mutex               sync.Mutex                            // clients map 동시성 제어
	valkeySubscriptions = make(map[string]context.CancelFunc) // Valkey 채널별 구독 취소 함수
	subMutex            sync.Mutex                            // valkeySubscriptions map 동시성 제어
)

// Client는 WebSocket 클라이언트 정보를 담는 구조체
type Client struct {
	conn          *websocket.Conn // WebSocket 연결
	roomID        string          // 클라이언트가 속한 방 ID
	username      string          // 클라이언트의 사용자명
	subscriptions map[string]bool // 구독 중인 채널 목록
	mu            sync.Mutex      // 구조체 필드 동시성 제어
}

// WSMessage는 WebSocket 메시지 형식
type WSMessage struct {
	Type    string      `json:"type"`    // 메시지 타입 (subscribe, unsubscribe, message)
	Channel string      `json:"channel"` // 대상 채널명
	Data    interface{} `json:"data"`    // 메시지 데이터
}

// WebSocket 업그레이더 설정
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 origin 허용 (개발용, 운영에서는 제한 필요)
	},
}

// HandleWebSocket은 HTTP 연결을 WebSocket으로 업그레이드하고 클라이언트를 등록
func HandleWebSocket(c *gin.Context) {
	// HTTP 연결을 WebSocket으로 업그레이드
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	// 새 클라이언트 생성
	client := &Client{
		conn:          conn,
		subscriptions: make(map[string]bool),
	}

	// 전역 클라이언트 맵에 추가
	mutex.Lock()
	clients[conn] = client
	mutex.Unlock()

	log.Println("New client connected.")

	// 고루틴으로 메시지 처리 시작
	go handleMessages(client)
}

// HandleLobbyWebSocket은 로비 전용 WebSocket 핸들러
func HandleLobbyWebSocket(c *gin.Context) {
	// HTTP 연결을 WebSocket으로 업그레이드
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	// 새 클라이언트 생성
	client := &Client{
		conn:          conn,
		roomID:        "lobby",
		subscriptions: make(map[string]bool),
	}

	// 전역 클라이언트 맵에 추가
	mutex.Lock()
	clients[conn] = client
	mutex.Unlock()

	log.Println("New lobby client connected.")

	// 자동으로 로비 채널 구독
	handleSubscribe(client, "lobby")

	// 로비 채팅 히스토리 전송
	ctx := context.Background()
	chatHistory, err := valkey.GetLobbyChatHistory(ctx, 1000)
	if err != nil {
		log.Printf("Failed to get lobby chat history: %v", err)
	} else if len(chatHistory) > 0 {
		// 채팅 히스토리를 클라이언트에게 전송
		historyPayload := map[string]interface{}{
			"type":     "LOBBY_CHAT_HISTORY",
			"messages": chatHistory,
		}
		if data, err := json.Marshal(historyPayload); err == nil {
			client.conn.WriteMessage(websocket.TextMessage, data)
			log.Printf("Sent %d chat history messages to client", len(chatHistory))
		}
	}

	// 고루틴으로 로비 메시지 처리 시작
	go handleLobbyMessages(client)
}

// HandleRoomWebSocket은 게임 방 전용 WebSocket 핸들러
func HandleRoomWebSocket(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		log.Printf("Room ID is required for room WebSocket")
		c.JSON(400, gin.H{"error": "room ID is required"})
		return
	}

	// HTTP 연결을 WebSocket으로 업그레이드
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	// 새 클라이언트 생성
	client := &Client{
		conn:          conn,
		roomID:        roomID,
		subscriptions: make(map[string]bool),
	}

	// 전역 클라이언트 맵에 추가
	mutex.Lock()
	clients[conn] = client
	mutex.Unlock()

	log.Printf("New room client connected to room: %s", roomID)

	// 자동으로 게임 방 채널 구독
	roomChannel := "game/" + roomID
	handleSubscribe(client, roomChannel)

	// 고루틴으로 방 메시지 처리 시작
	go handleRoomMessages(client, roomID)
}

// handleRoomMessages는 게임 방 클라이언트로부터 메시지를 계속 수신하고 처리
func handleRoomMessages(client *Client, roomID string) {
	// 함수 종료 시 클라이언트 정리
	defer func() {
		mutex.Lock()
		delete(clients, client.conn)
		mutex.Unlock()
		client.conn.Close()
		log.Printf("Room client disconnected from room: %s", roomID)
	}()

	log.Printf("Waiting for room messages from client in room: %s", roomID)

	// 메시지 수신 루프
	for {
		// 클라이언트로부터 메시지 읽기
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			} else {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		log.Printf("Received room message in %s: %s", roomID, string(msg))

		// JSON 파싱
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("Failed to parse room message: %v", err)
			continue
		}

		// 방 메시지 타입에 따라 처리
		switch wsMsg.Type {
		case "chat":
			handleChatMessage(client, wsMsg.Data)
		case "drawing":
			handleDrawingMessage(roomID, wsMsg.Data)
		case "game_action":
			handleGameAction(roomID, wsMsg.Data)
		case "wordchain_submit":
			handleWordchainSubmit(roomID, wsMsg.Data)
		default:
			log.Printf("Unknown room message type: %s", wsMsg.Type)
		}
	}
}

// handleDrawingMessage는 그림 그리기 메시지를 처리
func handleDrawingMessage(roomID string, data interface{}) {
	channel := "game/" + roomID
	drawingPayload := map[string]interface{}{
		"type":    "DRAW_EVENT",
		"payload": data,
	}
	handlePublish(channel, drawingPayload)
}

// handleGameAction은 게임 액션 메시지를 처리
func handleGameAction(roomID string, data interface{}) {
	channel := "game/" + roomID
	actionPayload := map[string]interface{}{
		"type":    "GAME_ACTION",
		"payload": data,
	}
	handlePublish(channel, actionPayload)
}

// handleWordchainSubmit은 끝말잇기 단어 제출을 처리
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
		reason = "이미 사용된 단어입니다. 다른 단어를 입력해주세요."
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
				reason = fmt.Sprintf("'%s'의 마지막 글자 '%c'와 시작이 일치하지 않습니다", lastWord, lastChar)
			}
		}

		// Check if word ends with ん (invalid in wordchain)
		if correct && len(word) > 0 {
			lastChar := []rune(word)[len([]rune(word))-1]
			if lastChar == 'ん' {
				correct = false
				reason = "ん으로 끝나는 단어는 사용할 수 없습니다"
			}
		}
	}

	log.Printf("Wordchain validation: username=%s, word=%s, correct=%v, duplicate=%v", username, word, correct, isDuplicate)

	// Check if it's this user's turn
	isCorrectTurn := graph.CheckWordchainTurn(roomID, username)
	if !isCorrectTurn {
		correct = false
		reason = "당신의 차례가 아닙니다"
		log.Printf("[Wordchain] Not %s's turn in room %s", username, roomID)
	}

	// Publish result to all players via graph package
	graph.PublishWordchainResult(roomID, username, word, correct, reason)

	// If correct, update last word, give points, and move to next player
	if correct && isCorrectTurn {
		graph.UpdateWordchainState(roomID, word, username)
		graph.MoveToNextTurn(roomID)
	} else if isCorrectTurn && !isDuplicate {
		// If wrong answer (but not duplicate) and correct turn, end round
		log.Printf("[Wordchain] Wrong answer from %s: %s. Ending round.", username, reason)
		graph.EndWordchainRound(roomID, fmt.Sprintf("%s님이 오답: %s", username, reason))
	}
	// If duplicate, don't move turn - let same player try again
}

// handleLobbyMessages는 로비 클라이언트로부터 메시지를 계속 수신하고 처리
func handleLobbyMessages(client *Client) {
	// 함수 종료 시 클라이언트 정리
	defer func() {
		// 연결 해제 시 자동으로 방에서 플레이어 제거
		if client.roomID != "" && client.roomID != "lobby" && client.username != "" {
			log.Printf("Auto-removing player %s from room %s due to WebSocket disconnect", client.username, client.roomID)
			removePlayerFromRoom(client.roomID, client.username)
		}

		mutex.Lock()
		delete(clients, client.conn)
		mutex.Unlock()
		client.conn.Close()
		log.Println("Lobby client disconnected.")
	}()

	log.Println("Waiting for lobby messages from client...")

	// 메시지 수신 루프
	for {
		// 클라이언트로부터 메시지 읽기
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			} else {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		log.Printf("Received lobby message: %s", string(msg))

		// JSON 파싱
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("Failed to parse lobby message: %v", err)
			continue
		}

		// 메시지 타입에 따라 처리
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
		case "identify":
			handleIdentify(client, wsMsg.Data)
		case "message":
			// Handle game messages sent through lobby WebSocket
			handleGameMessage(client, wsMsg)
		default:
			log.Printf("Unknown lobby message type: %s", wsMsg.Type)
		}
	}
}

// handleLobbyChatMessage는 로비 채팅 메시지를 처리하고 저장 및 브로드캐스트
func handleLobbyChatMessage(client *Client, data interface{}) {
	// 데이터 파싱
	chatData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid lobby chat message data format")
		return
	}

	playerID, _ := chatData["playerId"].(string)
	playerName, _ := chatData["playerName"].(string)
	message, _ := chatData["message"].(string)

	if playerID == "" || message == "" {
		log.Printf("Missing required lobby chat message fields")
		return
	}

	log.Printf("Lobby chat from %s: %s", playerName, message)

	// Valkey에 채팅 메시지 저장
	ctx := context.Background()
	if err := valkey.SaveLobbyChatMessage(ctx, playerID, playerName, message); err != nil {
		log.Printf("Failed to save lobby chat message: %v", err)
	}

	// 로비 채널에 메시지 발행
	chatPayload := map[string]interface{}{
		"type": "LOBBY_CHAT",
		"payload": map[string]interface{}{
			"playerId":   playerID,
			"playerName": playerName,
			"message":    message,
		},
	}

	handlePublish("lobby", chatPayload)
}

// handleGameMessage handles game-related messages sent through lobby WebSocket
func handleGameMessage(client *Client, wsMsg WSMessage) {
	// Extract the actual game message from wsMsg.Data
	dataMap, ok := wsMsg.Data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid game message data format")
		return
	}

	gameType, _ := dataMap["type"].(string)
	gameData := dataMap["data"]

	log.Printf("Handling game message: type=%s, roomID=%s", gameType, client.roomID)

	// Route to appropriate handler based on game message type
	switch gameType {
	case "wordchain_submit":
		if client.roomID != "" {
			handleWordchainSubmit(client.roomID, gameData)
		} else {
			log.Printf("Cannot handle wordchain_submit: client has no roomID")
		}
	case "chat":
		handleChatMessage(client, gameData)
	case "drawing":
		if client.roomID != "" {
			handleDrawingMessage(client.roomID, gameData)
		}
	case "game_action":
		if client.roomID != "" {
			handleGameAction(client.roomID, gameData)
		}
	default:
		log.Printf("Unknown game message type: %s", gameType)
	}
}

// handleMessages는 클라이언트로부터 메시지를 계속 수신하고 처리
func handleMessages(client *Client) {
	// 함수 종료 시 클라이언트 정리
	defer func() {
		mutex.Lock()
		delete(clients, client.conn) // 클라이언트 목록에서 제거
		mutex.Unlock()
		client.conn.Close() // WebSocket 연결 종료
		log.Println("Client disconnected.")
	}()

	log.Println("Waiting for messages from client...")

	// 메시지 수신 루프
	for {
		// 클라이언트로부터 메시지 읽기
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

		// JSON 파싱
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		log.Printf("Parsed message: type=%s, channel=%s", wsMsg.Type, wsMsg.Channel)

		// 메시지 타입에 따라 처리
		switch wsMsg.Type {
		case "subscribe": // 채널 구독 요청
			handleSubscribe(client, wsMsg.Channel)
		case "unsubscribe": // 채널 구독 취소 요청
			handleUnsubscribe(client, wsMsg.Channel)
		case "message": // 메시지 발행 요청
			handlePublish(wsMsg.Channel, wsMsg.Data)
		case "chat": // 채팅 메시지 요청
			handleChatMessage(client, wsMsg.Data)
		case "identify": // 클라이언트 식별 요청
			handleIdentify(client, wsMsg.Data)
		default:
			log.Printf("Unknown message type: %s", wsMsg.Type)
		}
	}
}

// handleSubscribe는 클라이언트의 채널 구독 요청 처리
func handleSubscribe(client *Client, channel string) {
	// 클라이언트의 구독 목록에 추가
	client.mu.Lock()
	client.subscriptions[channel] = true
	if client.roomID == "" && len(channel) > 0 {
		// 채널명에서 방 ID 추출 (예: "game/room-uuid" -> "room-uuid")
		parts := splitChannel(channel)
		if len(parts) > 1 {
			client.roomID = parts[1]
		}
	}
	client.mu.Unlock()

	// Valkey 구독이 아직 활성화되지 않았으면 시작
	subMutex.Lock()
	if _, exists := valkeySubscriptions[channel]; !exists {
		ctx, cancel := context.WithCancel(context.Background())
		valkeySubscriptions[channel] = cancel
		go startValkeySubscription(ctx, channel) // 고루틴으로 Valkey 구독 시작
	}
	subMutex.Unlock()

	log.Printf("Client subscribed to channel: %s", channel)
}

// handleUnsubscribe는 클라이언트의 채널 구독 취소 처리
func handleUnsubscribe(client *Client, channel string) {
	client.mu.Lock()
	delete(client.subscriptions, channel) // 구독 목록에서 제거
	client.mu.Unlock()
	log.Printf("Client unsubscribed from channel: %s", channel)
}

// handlePublish는 Valkey 채널에 메시지 발행
func handlePublish(channel string, data interface{}) {
	// 데이터를 JSON으로 직렬화
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal data: %v", err)
		return
	}

	log.Printf("Publishing to Valkey channel: %s", channel)

	// Valkey에 메시지 발행
	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish to Valkey: %v", err)
	}
}

// handleChatMessage는 채팅 메시지를 처리하고 같은 방의 모든 클라이언트에게 브로드캐스트
func handleChatMessage(client *Client, data interface{}) {
	// 데이터 파싱
	chatData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid chat message data format")
		return
	}

	roomID, _ := chatData["roomId"].(string)
	playerID, _ := chatData["playerId"].(string)
	playerName, _ := chatData["playerName"].(string)
	message, _ := chatData["message"].(string)

	if roomID == "" || playerID == "" || message == "" {
		log.Printf("Missing required chat message fields")
		return
	}

	log.Printf("Chat message from %s in room %s: %s", playerName, roomID, message)

	// 채팅 메시지를 Valkey 채널에 발행
	channel := "game/" + roomID
	chatPayload := ChatMessagePayload{
		RoomID:     roomID,
		PlayerID:   playerID,
		PlayerName: playerName,
		Message:    message,
	}

	handlePublish(channel, map[string]interface{}{
		"type":    "CHAT_MESSAGE",
		"payload": chatPayload,
	})
}

// startValkeySubscription은 Valkey 채널을 구독하고 메시지를 WebSocket 클라이언트에 브로드캐스트
func startValkeySubscription(ctx context.Context, channel string) {
	log.Printf("Starting Valkey subscription for channel: %s", channel)

	// Valkey 채널 구독 시작
	pubsub := valkey.SubscribeChannel(ctx, channel)
	if pubsub == nil {
		log.Printf("Failed to subscribe to channel: %s", channel)
		return
	}
	defer pubsub.Close()

	// 메시지 수신 루프
	ch := pubsub.Channel()
	for msg := range ch {
		message := msg.Payload
		log.Printf("Received message from Valkey channel %s: %s", channel, message)

		// 해당 채널을 구독 중인 모든 WebSocket 클라이언트에게 브로드캐스트
		mutex.Lock()
		for _, client := range clients {
			// 클라이언트가 이 채널을 구독 중인지 확인
			client.mu.Lock()
			subscribed := client.subscriptions[channel]
			client.mu.Unlock()

			if subscribed {
				// WebSocket 메시지 생성
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

				// 클라이언트에게 메시지 전송
				if err := client.conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					log.Printf("Error sending message to client: %v", err)
				}
			}
		}
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

// handleIdentify는 클라이언트의 식별 정보를 처리
func handleIdentify(client *Client, data interface{}) {
	identifyData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid identify data format")
		return
	}

	username, _ := identifyData["username"].(string)
	roomID, _ := identifyData["roomId"].(string)

	if username != "" {
		client.mu.Lock()
		client.username = username
		if roomID != "" && roomID != "lobby" {
			client.roomID = roomID
		}
		client.mu.Unlock()
		log.Printf("Client identified: username=%s, roomId=%s", username, roomID)
	}
}

// removePlayerFromRoom는 방에서 플레이어를 제거
func removePlayerFromRoom(roomID, username string) {
	graph.RemovePlayerFromRoom(roomID, username)
}
