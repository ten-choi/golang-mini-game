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
	log.Println("Checking quiz database status...")

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

	// Check OX Quizzes
	oxCollection := db.Collection("ox_quizzes")
	oxCount, _ := oxCollection.CountDocuments(ctx, bson.M{})
	log.Printf("\n=== OX Quizzes Collection ===")
	log.Printf("Total documents: %d", oxCount)

	if oxCount > 0 {
		// Get first 3 documents
		cursor, err := oxCollection.Find(ctx, bson.M{}, options.Find().SetLimit(3))
		if err == nil {
			defer cursor.Close(ctx)
			log.Println("\nSample OX Quizzes:")
			i := 1
			for cursor.Next(ctx) {
				var doc bson.M
				if err := cursor.Decode(&doc); err == nil {
					log.Printf("\n[%d] ID: %v", i, doc["_id"])
					log.Printf("    Question: %v", doc["question"])
					log.Printf("    Difficulty: %v (type: %T)", doc["difficulty"], doc["difficulty"])
					log.Printf("    Category: %v", doc["category"])
					i++
				}
			}
		}

		// Count by difficulty type
		stringDiff, _ := oxCollection.CountDocuments(ctx, bson.M{"difficulty": bson.M{"$type": "string"}})
		intDiff, _ := oxCollection.CountDocuments(ctx, bson.M{"difficulty": bson.M{"$type": "int"}})
		log.Printf("\nDifficulty types - String: %d, Int: %d", stringDiff, intDiff)
	}

	// Check QA Quizzes
	qaCollection := db.Collection("general_quizzes")
	qaCount, _ := qaCollection.CountDocuments(ctx, bson.M{})
	log.Printf("\n=== General Quizzes Collection ===")
	log.Printf("Total documents: %d", qaCount)

	if qaCount > 0 {
		// Get first 3 documents
		cursor, err := qaCollection.Find(ctx, bson.M{}, options.Find().SetLimit(3))
		if err == nil {
			defer cursor.Close(ctx)
			log.Println("\nSample QA Quizzes:")
			i := 1
			for cursor.Next(ctx) {
				var doc bson.M
				if err := cursor.Decode(&doc); err == nil {
					log.Printf("\n[%d] ID: %v", i, doc["_id"])
					log.Printf("    Question: %v", doc["question"])
					log.Printf("    Difficulty: %v (type: %T)", doc["difficulty"], doc["difficulty"])
					log.Printf("    Category: %v", doc["category"])
					i++
				}
			}
		}

		// Count by difficulty type
		stringDiff, _ := qaCollection.CountDocuments(ctx, bson.M{"difficulty": bson.M{"$type": "string"}})
		intDiff, _ := qaCollection.CountDocuments(ctx, bson.M{"difficulty": bson.M{"$type": "int"}})
		log.Printf("\nDifficulty types - String: %d, Int: %d", stringDiff, intDiff)
	}

	fmt.Println("\n=== Check completed ===")
}
