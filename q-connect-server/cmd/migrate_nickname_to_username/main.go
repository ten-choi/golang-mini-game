package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "draw_and_guess"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	log.Println("MongoDB connected")

	collection := client.Database(dbName).Collection("users")

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatal("Count failed:", err)
	}
	log.Printf("Total documents: %d\n", count)

	if count == 0 {
		log.Println("No data to migrate")
		return
	}

	userNameCount, err := collection.CountDocuments(ctx, bson.M{"userName": bson.M{"$exists": true}})
	if err != nil {
		log.Fatal("Name check failed:", err)
	}
	log.Printf("Documents with userName: %d\n", userNameCount)

	if userNameCount == 0 {
		log.Println("Already migrated")
		return
	}

	log.Println("Renaming userName to UserName...")
	result, err := collection.UpdateMany(
		ctx,
		bson.M{"userName": bson.M{"$exists": true}},
		bson.M{"$rename": bson.M{"userName": "UserName"}},
	)
	if err != nil {
		log.Fatal("Rename failed:", err)
	}

	log.Printf("Migration complete: %d documents updated\n", result.ModifiedCount)

	_, _ = collection.Indexes().DropOne(ctx, "userName_1")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "UserName", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	}
	indexName, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatal("Index creation failed:", err)
	}
	log.Printf("Index created: %s\n", indexName)

	fmt.Println("Done!")
}
