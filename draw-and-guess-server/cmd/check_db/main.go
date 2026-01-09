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

type QAQuiz struct {
	ID         int64    `bson:"id"`
	Question   string   `bson:"question"`
	Options    []string `bson:"options"`
	Answer     int      `bson:"answer"`
	Category   string   `bson:"category"`
	Difficulty string   `bson:"difficulty"`
	Type       string   `bson:"type"`
	IsActive   bool     `bson:"is_active"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB 연결
	mongoURI := "mongodb://admin:password@10.33.255.58:30017/draw_and_guess_db"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("draw_and_guess_db")
	collection := db.Collection("qa_quizzes")

	// 활성화된 QA 퀴즈 개수 확인
	count, err := collection.CountDocuments(ctx, bson.M{"is_active": true})
	if err != nil {
		log.Fatalf("Failed to count documents: %v", err)
	}
	fmt.Printf("✅ Total active QA quizzes: %d\n\n", count)

	// type 필드로 필터링한 개수
	countWithType, err := collection.CountDocuments(ctx, bson.M{
		"is_active": true,
		"type":      "QA",
	})
	if err != nil {
		log.Fatalf("Failed to count documents with type: %v", err)
	}
	fmt.Printf("✅ Active QA quizzes with type='QA': %d\n\n", countWithType)

	// 샘플 퀴즈 3개 가져오기
	findOptions := options.Find().SetLimit(3)
	cursor, err := collection.Find(ctx, bson.M{"is_active": true}, findOptions)
	if err != nil {
		log.Fatalf("Failed to find documents: %v", err)
	}
	defer cursor.Close(ctx)

	fmt.Println("========== Sample Quizzes ==========")
	for cursor.Next(ctx) {
		var quiz QAQuiz
		if err := cursor.Decode(&quiz); err != nil {
			log.Printf("Failed to decode: %v", err)
			continue
		}
		fmt.Printf("ID: %d\n", quiz.ID)
		fmt.Printf("Question: %s\n", quiz.Question)
		fmt.Printf("Type: %s\n", quiz.Type)
		fmt.Printf("Category: %s\n", quiz.Category)
		fmt.Printf("Options: %v\n", quiz.Options)
		fmt.Printf("Answer: %d\n", quiz.Answer)
		fmt.Printf("IsActive: %v\n\n", quiz.IsActive)
	}
}
