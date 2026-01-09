package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB 연결
	clientOptions := options.Client().ApplyURI("mongodb://admin:password@10.33.255.58:30017/draw_and_guess_db?authSource=admin")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("draw_and_guess_db")
	usersCollection := db.Collection("users")

	// 기존 인덱스 확인
	fmt.Println("Current indexes:")
	cursor, err := usersCollection.Indexes().List(ctx)
	if err != nil {
		log.Fatalf("Failed to list indexes: %v", err)
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err = cursor.All(ctx, &indexes); err != nil {
		log.Fatalf("Failed to decode indexes: %v", err)
	}

	for _, index := range indexes {
		fmt.Printf("  - %v\n", index["name"])
		if keys, ok := index["key"].(bson.M); ok {
			for key := range keys {
				fmt.Printf("    Key: %s\n", key)
			}
		}
	}

	// username_1 인덱스 삭제
	fmt.Println("\nDropping username_1 index...")
	_, err = usersCollection.Indexes().DropOne(ctx, "username_1")
	if err != nil {
		if mongo.IsDuplicateKeyError(err) || err.Error() == "index not found" {
			fmt.Println("  Index not found or already dropped")
		} else {
			log.Printf("Warning: Failed to drop username_1 index: %v", err)
		}
	} else {
		fmt.Println("  ✓ username_1 index dropped successfully")
	}

	// 인덱스 다시 확인
	fmt.Println("\nIndexes after cleanup:")
	cursor, err = usersCollection.Indexes().List(ctx)
	if err != nil {
		log.Fatalf("Failed to list indexes: %v", err)
	}
	defer cursor.Close(ctx)

	indexes = []bson.M{}
	if err = cursor.All(ctx, &indexes); err != nil {
		log.Fatalf("Failed to decode indexes: %v", err)
	}

	for _, index := range indexes {
		fmt.Printf("  - %v\n", index["name"])
	}

	fmt.Println("\n✓ Index cleanup completed!")
}
