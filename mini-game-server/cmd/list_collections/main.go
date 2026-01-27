package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("Listing all collections in database...")

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
	log.Printf("Database: %s\n", dbName)

	db := client.Database(dbName)

	// List all collections
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}

	fmt.Println("\n=== All Collections ===")
	for _, collName := range collections {
		coll := db.Collection(collName)
		count, _ := coll.CountDocuments(ctx, bson.M{})
		fmt.Printf("- %s (%d documents)\n", collName, count)

		// Check for string difficulty
		if count > 0 {
			stringDiff, _ := coll.CountDocuments(ctx, bson.M{"difficulty": bson.M{"$type": "string"}})
			intDiff, _ := coll.CountDocuments(ctx, bson.M{"difficulty": bson.M{"$type": "int"}})
			if stringDiff > 0 || intDiff > 0 {
				fmt.Printf("  └─ Difficulty: String=%d, Int=%d\n", stringDiff, intDiff)
			}
		}
	}

	fmt.Println("\n=== Check completed ===")
}
