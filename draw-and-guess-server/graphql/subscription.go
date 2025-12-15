package graphql

import (
	"context"
	"draw-and-guess-server/valkey"
	"encoding/json"
	"log"
	"time"

	"github.com/graphql-go/graphql"
)

// SubscribeGameEvents는 게임 이벤트를 실시간으로 전송합니다
func SubscribeGameEvents(ctx context.Context, roomID string, sendFunc func(interface{})) error {
	channel := "game/room-" + roomID

	return valkey.SubscribeChannel(ctx, channel, func(message string) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(message), &data); err != nil {
			log.Printf("Failed to parse game event: %v", err)
			return
		}

		event := map[string]interface{}{
			"type":      data["type"],
			"roomId":    roomID,
			"data":      message,
			"timestamp": time.Now().Format(time.RFC3339),
		}

		sendFunc(event)
	})
}

// SubscribeChatMessages는 채팅 메시지를 실시간으로 전송합니다
func SubscribeChatMessages(ctx context.Context, roomID string, sendFunc func(interface{})) error {
	channel := "chat/room-" + roomID

	return valkey.SubscribeChannel(ctx, channel, func(message string) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(message), &data); err != nil {
			log.Printf("Failed to parse chat message: %v", err)
			return
		}

		chatMsg := map[string]interface{}{
			"roomId":    roomID,
			"userId":    data["user_id"],
			"username":  data["username"],
			"message":   data["message"],
			"timestamp": time.Now().Format(time.RFC3339),
		}

		sendFunc(chatMsg)
	})
}

// SubscribeDrawingUpdates는 그림 그리기 데이터를 실시간으로 전송합니다
func SubscribeDrawingUpdates(ctx context.Context, roomID string, sendFunc func(interface{})) error {
	channel := "draw/room-" + roomID

	return valkey.SubscribeChannel(ctx, channel, func(message string) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(message), &data); err != nil {
			log.Printf("Failed to parse drawing data: %v", err)
			return
		}

		drawingData := map[string]interface{}{
			"roomId":    roomID,
			"action":    data["action"],
			"data":      message,
			"timestamp": time.Now().Format(time.RFC3339),
		}

		sendFunc(drawingData)
	})
}

// CreateSubscriptionResolver는 Subscription resolver를 생성합니다
func CreateSubscriptionResolver(subscriptionType string, roomID string) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		sourceChan := make(chan interface{})

		go func() {
			defer close(sourceChan)

			switch subscriptionType {
			case "gameEvents":
				err := SubscribeGameEvents(ctx, roomID, func(data interface{}) {
					select {
					case sourceChan <- data:
					case <-ctx.Done():
						return
					}
				})
				if err != nil {
					log.Printf("Game events subscription error: %v", err)
				}

			case "chatMessages":
				err := SubscribeChatMessages(ctx, roomID, func(data interface{}) {
					select {
					case sourceChan <- data:
					case <-ctx.Done():
						return
					}
				})
				if err != nil {
					log.Printf("Chat messages subscription error: %v", err)
				}

			case "drawingUpdates":
				err := SubscribeDrawingUpdates(ctx, roomID, func(data interface{}) {
					select {
					case sourceChan <- data:
					case <-ctx.Done():
						return
					}
				})
				if err != nil {
					log.Printf("Drawing updates subscription error: %v", err)
				}
			}
		}()

		return sourceChan, nil
	}
}
