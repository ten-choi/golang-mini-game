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

type OXQuiz struct {
	ID          int64     `bson:"_id"`
	Type        string    `bson:"type"`
	Category    string    `bson:"category"`
	Difficulty  string    `bson:"difficulty"`
	Question    string    `bson:"question"`
	IsAnswer    bool      `bson:"is_answer"`
	Explanation string    `bson:"explanation"`
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
	oxCollection := db.Collection("ox_quizzes")

	// OX 퀴즈 개수 확인
	count, err := oxCollection.CountDocuments(ctx, bson.M{"type": "OX"})
	if err != nil {
		log.Fatalf("Failed to count OX quizzes: %v", err)
	}

	fmt.Printf("Current OX quizzes count: %d\n", count)

	if count > 0 {
		fmt.Println("OX quizzes already exist. Showing first 5:")
		findOptions := options.Find().SetLimit(5)
		cursor, err := oxCollection.Find(ctx, bson.M{"type": "OX"}, findOptions)
		if err != nil {
			log.Fatalf("Failed to find OX quizzes: %v", err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var quiz OXQuiz
			if err := cursor.Decode(&quiz); err != nil {
				log.Printf("Failed to decode quiz: %v", err)
				continue
			}
			fmt.Printf("ID: %d, Question: %s, Answer: %v\n", quiz.ID, quiz.Question, quiz.IsAnswer)
		}
		return
	}

	fmt.Println("No OX quizzes found. Adding sample quizzes...")

	// 샘플 OX 퀴즈 데이터
	now := time.Now()
	oxQuizzes := []OXQuiz{
		{
			ID:          2001,
			Type:        "OX",
			Category:    "일반상식",
			Difficulty:  "쉬움",
			Question:    "태양은 동쪽에서 뜬다.",
			IsAnswer:    true,
			Explanation: "태양은 동쪽에서 떠서 서쪽으로 집니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2002,
			Type:        "OX",
			Category:    "과학",
			Difficulty:  "쉬움",
			Question:    "물은 섭씨 0도에서 얼어서 얼음이 된다.",
			IsAnswer:    true,
			Explanation: "물은 0도(섭씨)에서 얼어 얼음이 됩니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2003,
			Type:        "OX",
			Category:    "지리",
			Difficulty:  "쉬움",
			Question:    "한국은 섬나라이다.",
			IsAnswer:    false,
			Explanation: "한국은 반도 국가로, 북쪽으로 대륙과 연결되어 있습니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2004,
			Type:        "OX",
			Category:    "역사",
			Difficulty:  "보통",
			Question:    "세종대왕은 한글을 창제했다.",
			IsAnswer:    true,
			Explanation: "세종대왕은 1443년 훈민정음(한글)을 창제했습니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2005,
			Type:        "OX",
			Category:    "과학",
			Difficulty:  "보통",
			Question:    "지구는 평평하다.",
			IsAnswer:    false,
			Explanation: "지구는 구형(공 모양)입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2006,
			Type:        "OX",
			Category:    "수학",
			Difficulty:  "쉬움",
			Question:    "1 + 1 = 2이다.",
			IsAnswer:    true,
			Explanation: "기본적인 산술 연산입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2007,
			Type:        "OX",
			Category:    "과학",
			Difficulty:  "보통",
			Question:    "사람은 물 없이 한 달을 살 수 있다.",
			IsAnswer:    false,
			Explanation: "사람은 물 없이 약 3일 정도만 생존할 수 있습니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2008,
			Type:        "OX",
			Category:    "동물",
			Difficulty:  "쉬움",
			Question:    "고양이는 포유류이다.",
			IsAnswer:    true,
			Explanation: "고양이는 포유류 동물입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2009,
			Type:        "OX",
			Category:    "지리",
			Difficulty:  "보통",
			Question:    "에베레스트는 세계에서 가장 높은 산이다.",
			IsAnswer:    true,
			Explanation: "에베레스트산은 해발 8,849m로 세계에서 가장 높은 산입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2010,
			Type:        "OX",
			Category:    "역사",
			Difficulty:  "보통",
			Question:    "제2차 세계대전은 1945년에 끝났다.",
			IsAnswer:    true,
			Explanation: "제2차 세계대전은 1945년 9월 2일 일본의 항복으로 종결되었습니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// 데이터 삽입
	var docs []interface{}
	for _, quiz := range oxQuizzes {
		docs = append(docs, quiz)
	}

	result, err := oxCollection.InsertMany(ctx, docs)
	if err != nil {
		log.Fatalf("Failed to insert OX quizzes: %v", err)
	}

	fmt.Printf("Successfully inserted %d OX quizzes\n", len(result.InsertedIDs))
	fmt.Println("OX quizzes added successfully!")
}
