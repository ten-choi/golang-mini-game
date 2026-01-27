package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("Starting difficulty migration script...")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get MongoDB configuration from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "draw_and_guess_db"
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}

	log.Println("MongoDB connected successfully")

	db := client.Database(dbName)

	// Migrate OX Quizzes
	log.Println("Migrating OX Quizzes...")
	oxCollection := db.Collection("ox_quizzes")
	oxMigrated, oxFailed := migrateDifficulty(ctx, oxCollection)
	log.Printf("OX Quizzes - Migrated: %d, Failed: %d", oxMigrated, oxFailed)

	// Migrate QA Quizzes
	log.Println("Migrating QA Quizzes...")
	qaCollection := db.Collection("qa_quizzes")
	qaMigrated, qaFailed := migrateDifficulty(ctx, qaCollection)
	log.Printf("QA Quizzes - Migrated: %d, Failed: %d", qaMigrated, qaFailed)

	log.Printf("Migration completed! Total migrated: %d, Total failed: %d",
		oxMigrated+qaMigrated, oxFailed+qaFailed)
}

func migrateDifficulty(ctx context.Context, collection *mongo.Collection) (int, int) {
	migrated := 0
	failed := 0

	// Find all documents with string difficulty
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to query collection: %v", err)
		return 0, 0
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Failed to decode document: %v", err)
			failed++
			continue
		}

		id := doc["_id"]
		difficulty := doc["difficulty"]

		// Check if difficulty is a string
		diffStr, isString := difficulty.(string)
		if !isString {
			// Already migrated or difficulty is already int
			continue
		}

		// Convert string to int
		diffInt := convertDifficultyToInt(diffStr)
		if diffInt == 0 {
			log.Printf("Skipping document %v with difficulty '%s'", id, diffStr)
			failed++
			continue
		}

		// Update the document
		filter := bson.M{"_id": id}
		update := bson.M{"$set": bson.M{"difficulty": diffInt}}

		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Failed to update document %v: %v", id, err)
			failed++
			continue
		}

		if result.ModifiedCount > 0 {
			log.Printf("Migrated document %v: '%s' → %d", id, diffStr, diffInt)
			migrated++
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
	}

	return migrated, failed
}

func convertDifficultyToInt(difficulty string) int {
	// Normalize the string (lowercase, trim spaces)
	normalized := strings.ToLower(strings.TrimSpace(difficulty))

	switch normalized {
	case "easy", "쉬움", "초급":
		return 1
	case "medium", "normal", "보통", "중급":
		return 2
	case "hard", "어려움", "고급":
		return 3
	case "very_hard", "veryhard", "very hard", "매우 어려움", "최고급":
		return 4
	case "extreme", "익스트림", "극한":
		return 5
	default:
		// Try to match unknown patterns - default to 2 (medium)
		log.Printf("Unknown difficulty value: '%s', defaulting to 2 (medium)", difficulty)
		return 2
	}
}
