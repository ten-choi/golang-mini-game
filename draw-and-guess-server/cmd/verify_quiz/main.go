package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Quiz struct {
	ID          int64     `bson:"_id"`
	Type        string    `bson:"type"`
	Category    string    `bson:"category"`
	Difficulty  string    `bson:"difficulty"`
	Question    string    `bson:"question"`
	Options     []string  `bson:"options,omitempty"`
	Answer      int       `bson:"answer,omitempty"`
	Explanation string    `bson:"explanation"`
	IsActive    bool      `bson:"is_active"`
	CreatedAt   time.Time `bson:"created_at"`
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

	// QA 퀴즈 개수 확인
	count, err := collection.CountDocuments(ctx, bson.M{"is_active": true, "type": "QA"})
	if err != nil {
		log.Fatalf("Failed to count: %v", err)
	}
	fmt.Printf("\n========================================\n")
	fmt.Printf("✅ Total QA quizzes in qa_quizzes collection: %d\n", count)
	fmt.Printf("========================================\n\n")

	// 첫 3개 퀴즈 가져오기
	findOptions := options.Find().SetLimit(3)
	cursor, err := collection.Find(ctx, bson.M{"is_active": true, "type": "QA"}, findOptions)
	if err != nil {
		log.Fatalf("Failed to find: %v", err)
	}
	defer cursor.Close(ctx)

	fmt.Println("📋 Sample Quizzes:\n")
	i := 1
	for cursor.Next(ctx) {
		var quiz Quiz
		if err := cursor.Decode(&quiz); err != nil {
			log.Printf("Failed to decode: %v", err)
			continue
		}
		
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Quiz #%d (ID: %d)\n", i, quiz.ID)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Type: %s\n", quiz.Type)
		fmt.Printf("Category: %s\n", quiz.Category)
		fmt.Printf("Difficulty: %s\n", quiz.Difficulty)
		fmt.Printf("Question: %s\n", quiz.Question)
		fmt.Printf("Options:\n")
		for idx, opt := range quiz.Options {
			marker := "  "
			if idx == quiz.Answer {
				marker = "✓ "
			}
			fmt.Printf("  %s[%d] %s\n", marker, idx, opt)
		}
		fmt.Printf("Correct Answer: %d\n", quiz.Answer)
		fmt.Printf("Explanation: %s\n", quiz.Explanation)
		fmt.Printf("Is Active: %v\n\n", quiz.IsActive)
		i++
	}

	// JSON으로도 출력
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📦 JSON Format (for debugging):")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	cursor2, _ := collection.Find(ctx, bson.M{"is_active": true, "type": "QA"}, options.Find().SetLimit(1))
	defer cursor2.Close(ctx)
	if cursor2.Next(ctx) {
		var quiz Quiz
		cursor2.Decode(&quiz)
		jsonData, _ := json.MarshalIndent(quiz, "", "  ")
		fmt.Println(string(jsonData))
	}
}
