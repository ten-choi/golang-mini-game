package valkey

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"draw-and-guess-server/internal/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// Connect establishes a connection to Valkey (Redis)
func Connect() error {
	Client = redis.NewClient(&redis.Options{
		Addr:         config.ValkeyAddr,
		Password:     config.ValkeyPassword, // Use password from config
		DB:           0,                     // Use default DB
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("valkey connection failed: %w", err)
	}

	log.Println("✓ Connected to Valkey")
	return nil
}

// Close closes the Valkey connection
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// PublishMessage publishes a message to a Redis channel
func PublishMessage(ctx context.Context, channel string, message interface{}) error {
	if Client == nil {
		return fmt.Errorf("valkey client is not initialized")
	}
	return Client.Publish(ctx, channel, message).Err()
}

// SubscribeChannel subscribes to a Redis channel and returns a PubSub
func SubscribeChannel(ctx context.Context, channels ...string) *redis.PubSub {
	if Client == nil {
		log.Println("Warning: valkey client is not initialized")
		return nil
	}
	return Client.Subscribe(ctx, channels...)
}

// SaveLobbyChatMessage saves a lobby chat message to Valkey list (max 1000 messages)
func SaveLobbyChatMessage(ctx context.Context, playerID, playerName, message string) error {
	if Client == nil {
		return fmt.Errorf("valkey client is not initialized")
	}

	// Create chat message JSON
	chatMsg := map[string]interface{}{
		"playerId":   playerID,
		"playerName": playerName,
		"message":    message,
		"timestamp":  time.Now().Unix(),
	}

	// Convert to JSON string
	msgBytes, err := json.Marshal(chatMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal chat message: %w", err)
	}

	// Add to list (LPUSH adds to the head)
	if err := Client.LPush(ctx, "lobby:chat:messages", msgBytes).Err(); err != nil {
		return fmt.Errorf("failed to save chat message: %w", err)
	}

	// Keep only the latest 1000 messages
	if err := Client.LTrim(ctx, "lobby:chat:messages", 0, 999).Err(); err != nil {
		return fmt.Errorf("failed to trim chat messages: %w", err)
	}

	return nil
}

// GetLobbyChatHistory retrieves the last N lobby chat messages
func GetLobbyChatHistory(ctx context.Context, count int64) ([]map[string]interface{}, error) {
	if Client == nil {
		return nil, fmt.Errorf("valkey client is not initialized")
	}

	if count > 1000 {
		count = 1000
	}

	// Get messages from list (LRANGE 0 count-1)
	msgs, err := Client.LRange(ctx, "lobby:chat:messages", 0, count-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	// Parse JSON messages
	var chatHistory []map[string]interface{}
	for i := len(msgs) - 1; i >= 0; i-- { // Reverse to show oldest first
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(msgs[i]), &msg); err != nil {
			log.Printf("Failed to unmarshal chat message: %v", err)
			continue
		}
		chatHistory = append(chatHistory, msg)
	}

	return chatHistory, nil
}
