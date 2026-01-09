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

type Quiz struct {
	ID          int64     `bson:"_id"`
	Type        string    `bson:"type"`
	Category    string    `bson:"category"`
	Difficulty  string    `bson:"difficulty"`
	Question    string    `bson:"question"`
	Options     []string  `bson:"options,omitempty"`
	Answer      int       `bson:"answer,omitempty"`
	IsAnswer    bool      `bson:"is_answer,omitempty"`
	Explanation string    `bson:"explanation"`
	ImageURL    string    `bson:"image_url,omitempty"`
	UsageCount  int       `bson:"usage_count"`
	IsActive    bool      `bson:"is_active"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

func main() {
	// MongoDB 연결
	ctx := context.Background()
	mongoURI := "mongodb://admin:password@10.33.255.58:30017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("draw_and_guess_db")
	quizCollection := db.Collection("qa_quizzes")

	// QA 퀴즈 개수 확인
	qaCount, err := quizCollection.CountDocuments(ctx, bson.M{"type": "QA"})
	if err != nil {
		log.Fatalf("Failed to count QA quizzes: %v", err)
	}

	fmt.Printf("Current QA quizzes count: %d\n", qaCount)

	if qaCount > 0 {
		fmt.Println("QA quizzes already exist. Showing first 5:")
		cursor, err := quizCollection.Find(ctx, bson.M{"type": "QA"}, options.Find().SetLimit(5))
		if err != nil {
			log.Fatalf("Failed to find QA quizzes: %v", err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var quiz Quiz
			if err := cursor.Decode(&quiz); err != nil {
				log.Printf("Failed to decode quiz: %v", err)
				continue
			}
			fmt.Printf("ID: %d, Question: %s, Options: %v\n", quiz.ID, quiz.Question, quiz.Options)
		}
		return
	}

	fmt.Println("No QA quizzes found. Adding sample quizzes...")

	// 샘플 QA 퀴즈 데이터
	now := time.Now()
	qaQuizzes := []Quiz{
		{
			ID:          1001,
			Type:        "QA",
			Category:    "일반상식",
			Difficulty:  "쉬움",
			Question:    "대한민국의 수도는?",
			Options:     []string{"서울", "부산", "인천", "대구"},
			Answer:      0,
			Explanation: "대한민국의 수도는 서울입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1002,
			Type:        "QA",
			Category:    "과학",
			Difficulty:  "쉬움",
			Question:    "물의 화학식은?",
			Options:     []string{"CO2", "H2O", "O2", "H2"},
			Answer:      1,
			Explanation: "물의 화학식은 H2O입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1003,
			Type:        "QA",
			Category:    "역사",
			Difficulty:  "보통",
			Question:    "세종대왕이 창제한 문자는?",
			Options:     []string{"한자", "한글", "가나", "알파벳"},
			Answer:      1,
			Explanation: "세종대왕은 훈민정음(한글)을 창제했습니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1004,
			Type:        "QA",
			Category:    "일반상식",
			Difficulty:  "쉬움",
			Question:    "1년은 몇 개월?",
			Options:     []string{"10개월", "11개월", "12개월", "13개월"},
			Answer:      2,
			Explanation: "1년은 12개월입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1005,
			Type:        "QA",
			Category:    "과학",
			Difficulty:  "보통",
			Question:    "태양계에서 가장 큰 행성은?",
			Options:     []string{"지구", "화성", "목성", "토성"},
			Answer:      2,
			Explanation: "태양계에서 가장 큰 행성은 목성입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1006,
			Type:        "QA",
			Category:    "일반상식",
			Difficulty:  "쉬움",
			Question:    "무지개는 몇 가지 색?",
			Options:     []string{"5가지", "6가지", "7가지", "8가지"},
			Answer:      2,
			Explanation: "무지개는 빨주노초파남보 7가지 색입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1007,
			Type:        "QA",
			Category:    "지리",
			Difficulty:  "보통",
			Question:    "일본의 수도는?",
			Options:     []string{"오사카", "교토", "도쿄", "후쿠오카"},
			Answer:      2,
			Explanation: "일본의 수도는 도쿄입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1008,
			Type:        "QA",
			Category:    "수학",
			Difficulty:  "쉬움",
			Question:    "10 + 5 = ?",
			Options:     []string{"13", "14", "15", "16"},
			Answer:      2,
			Explanation: "10 + 5 = 15입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1009,
			Type:        "QA",
			Category:    "동물",
			Difficulty:  "쉬움",
			Question:    "가장 빠른 육상 동물은?",
			Options:     []string{"사자", "치타", "표범", "호랑이"},
			Answer:      1,
			Explanation: "치타가 가장 빠른 육상 동물입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1010,
			Type:        "QA",
			Category:    "일반상식",
			Difficulty:  "쉬움",
			Question:    "사계절 순서는?",
			Options:     []string{"봄-여름-가을-겨울", "여름-가을-겨울-봄", "가을-겨울-봄-여름", "겨울-봄-여름-가을"},
			Answer:      0,
			Explanation: "사계절은 봄, 여름, 가을, 겨울 순서입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// 데이터 삽입
	var documents []interface{}
	for _, quiz := range qaQuizzes {
		documents = append(documents, quiz)
	}

	result, err := quizCollection.InsertMany(ctx, documents)
	if err != nil {
		log.Fatalf("Failed to insert quizzes: %v", err)
	}

	fmt.Printf("Successfully inserted %d QA quizzes\n", len(result.InsertedIDs))
	fmt.Println("QA quizzes added successfully!")
}
