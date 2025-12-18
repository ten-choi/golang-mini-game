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
	if err := valkey.PublishMessage(channel, string(jsonData)); err != nil {
		log.Printf("Failed to publish to Valkey: %v", err)
	}
}

// startValkeySubscription은 Valkey 채널을 구독하고 메시지를 WebSocket 클라이언트에 브로드캐스트
func startValkeySubscription(ctx context.Context, channel string) {
	log.Printf("Starting Valkey subscription for channel: %s", channel)

	// Valkey 채널 구독 시작
	err := valkey.SubscribeChannel(ctx, channel, func(message string) {
		log.Printf("Received message from Valkey channel %s: %s", channel, message)

		// 해당 채널을 구독 중인 모든 WebSocket 클라이언트에게 브로드캐스트
		mutex.Lock()
		defer mutex.Unlock()

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
	})

	// splitChannel은 채널 문자열을 '/' 구분자로 분리 (예: "game/room-123" -> ["game", "room-123"])
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
