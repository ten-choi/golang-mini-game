package valkey

import (
	"context"
	"draw-and-guess-server/config"
	"encoding/json"
	"log"

	"github.com/valkey-io/valkey-go"
)

var Client valkey.Client

func Connect() error {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{config.ValkeyAddr},
	})
	if err != nil {
		return err
	}

	Client = client

	ctx := context.Background()
	err = Client.Do(ctx, Client.B().Ping().Build()).Error()
	if err != nil {
		return err
	}

	log.Println("Connected to Valkey!")
	return nil
}

// PublishMessage publishes a message to a channel
func PublishMessage(channel string, message interface{}) error {
	ctx := context.Background()
	var data string

	// If message is already a string, use it directly
	if str, ok := message.(string); ok {
		data = str
	} else {
		// Otherwise, marshal to JSON
		jsonData, err := json.Marshal(message)
		if err != nil {
			return err
		}
		data = string(jsonData)
	}

	log.Printf("Publishing to Valkey channel '%s': %s", channel, data)
	return Client.Do(ctx, Client.B().Publish().Channel(channel).Message(data).Build()).Error()
}

// SubscribeChannel subscribes to a channel and returns messages
func SubscribeChannel(ctx context.Context, channel string, handler func(string)) error {
	err := Client.Receive(ctx, Client.B().Subscribe().Channel(channel).Build(), func(msg valkey.PubSubMessage) {
		handler(msg.Message)
	})
	return err
}

// SetJSON sets a JSON value with expiration
func SetJSON(key string, value interface{}, expirationSeconds int) error {
	ctx := context.Background()
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return Client.Do(ctx, Client.B().Set().Key(key).Value(string(jsonData)).ExSeconds(int64(expirationSeconds)).Build()).Error()
}

// GetJSON gets a JSON value
func GetJSON(key string, result interface{}) error {
	ctx := context.Background()
	cmd := Client.B().Get().Key(key).Build()
	resp := Client.Do(ctx, cmd)

	if resp.Error() != nil {
		return resp.Error()
	}

	data, err := resp.ToString()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), result)
}

// DeleteKey deletes a key
func DeleteKey(key string) error {
	ctx := context.Background()
	return Client.Do(ctx, Client.B().Del().Key(key).Build()).Error()
}

// GetKeys gets all keys matching a pattern
func GetKeys(pattern string) ([]string, error) {
	ctx := context.Background()
	cmd := Client.B().Keys().Pattern(pattern).Build()
	resp := Client.Do(ctx, cmd)

	if resp.Error() != nil {
		return nil, resp.Error()
	}

	keys, err := resp.AsStrSlice()
	if err != nil {
		return nil, err
	}

	return keys, nil
}
