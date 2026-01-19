package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"draw-and-guess-server/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	DB     *mongo.Database
)

// Connect establishes a connection to MongoDB
func Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.MongoURI)

	// Set connection pool settings
	clientOptions.SetMaxPoolSize(25)
	clientOptions.SetMinPoolSize(5)
	clientOptions.SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	Client = client
	DB = client.Database(config.MongoDB)
	log.Println("Successfully connected to MongoDB!")
	return nil
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if Client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Client.Disconnect(ctx)
}

// InitSchema creates indexes for MongoDB collections
func InitSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Users collection indexes
	usersCollection := DB.Collection("users")
	_, err := usersCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"UserName": 1},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    map[string]interface{}{"hange_id": 1},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: map[string]interface{}{"created_at": -1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Korean words collection indexes
	wordsCollection := DB.Collection("korean_words")
	_, err = wordsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"word": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"first_char": 1},
		},
		{
			Keys: map[string]interface{}{"last_char": 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create korean_words indexes: %w", err)
	}

	// OX quizzes collection indexes
	oxQuizzesCollection := DB.Collection("ox_quizzes")
	_, err = oxQuizzesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"category": 1},
		},
		{
			Keys: map[string]interface{}{"is_active": 1},
		},
		{
			Keys: map[string]interface{}{"difficulty": 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create ox_quizzes indexes: %w", err)
	}

	// QA quizzes collection indexes
	qaQuizzesCollection := DB.Collection("qa_quizzes")
	_, err = qaQuizzesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"category": 1},
		},
		{
			Keys: map[string]interface{}{"is_active": 1},
		},
		{
			Keys: map[string]interface{}{"difficulty": 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create qa_quizzes indexes: %w", err)
	}

	// User stats collection indexes
	userStatsCollection := DB.Collection("user_stats")
	_, err = userStatsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"user_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"total_score": -1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create user_stats indexes: %w", err)
	}

	// Shop items collection indexes
	shopItemsCollection := DB.Collection("shop_items")
	_, err = shopItemsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"item_type": 1},
		},
		{
			Keys: map[string]interface{}{"is_available": 1},
		},
		{
			Keys: map[string]interface{}{"is_featured": 1},
		},
		{
			Keys: map[string]interface{}{"price": 1},
		},
		{
			Keys: map[string]interface{}{"tags": 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create shop_items indexes: %w", err)
	}

	// User inventory collection indexes
	userInventoryCollection := DB.Collection("user_inventory")
	_, err = userInventoryCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{"user_id", 1}, {"item_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"user_id", 1}, {"item_type", 1}},
		},
		{
			Keys: bson.D{{"user_id", 1}, {"is_equipped", 1}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create user_inventory indexes: %w", err)
	}

	// Transactions collection indexes
	transactionsCollection := DB.Collection("transactions")
	_, err = transactionsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"user_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"user_id", 1}, {"type", 1}},
		},
		{
			Keys: bson.D{{"user_id", 1}, {"status", 1}},
		},
		{
			Keys: bson.D{{"created_at", -1}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create transactions indexes: %w", err)
	}

	log.Println("MongoDB indexes created successfully")
	return nil
}

// GetCollection returns a MongoDB collection by name
func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}
