package valkey

import (
	"context"
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
		Password:     "", // No password by default
		DB:           0,  // Use default DB
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
