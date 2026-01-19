package main

import (
	"context"
	"log"
	"time"

	"draw-and-guess-server/internal/config"
	"draw-and-guess-server/internal/database"
	"draw-and-guess-server/internal/models"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// Initialize configuration
	config.Init()

	// Connect to MongoDB
	err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("✓ Connected to MongoDB")

	// Initialize Snowflake ID generator
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("Failed to create Snowflake node: %v", err)
	}
	log.Println("✓ Snowflake ID generator initialized")

	ctx := context.Background()
	db := database.Client.Database("draw_and_guess_db")
	oxCollection := db.Collection("ox_quizzes")
	qaCollection := db.Collection("qa_quizzes")

	// Check current count
	oxCount, _ := oxCollection.CountDocuments(ctx, bson.M{})
	qaCount, _ := qaCollection.CountDocuments(ctx, bson.M{})
	log.Printf("Current counts - OX: %d, QA: %d", oxCount, qaCount)

	// Add more OX quizzes
	oxQuizzes := []models.OXQuiz{
		{
			ID:          node.Generate().Int64(),
			Category:    "역사",
			Difficulty:  "easy",
			Question:    "세종대왕이 한글을 창제하였다.",
			Answer:      true,
			Explanation: "세종대왕이 1443년에 훈민정음(한글)을 창제하였습니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "지리",
			Difficulty:  "easy",
			Question:    "에베레스트는 세계에서 가장 높은 산이다.",
			Answer:      true,
			Explanation: "에베레스트 산은 해발 8,849m로 세계에서 가장 높은 산입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "상식",
			Difficulty:  "easy",
			Question:    "일주일은 8일이다.",
			Answer:      false,
			Explanation: "일주일은 7일입니다 (월, 화, 수, 목, 금, 토, 일).",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  "easy",
			Question:    "다이아몬드는 탄소로 이루어져 있다.",
			Answer:      true,
			Explanation: "다이아몬드는 탄소 원자가 결정 구조를 이루고 있습니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "문화",
			Difficulty:  "easy",
			Question:    "피카소는 프랑스 화가이다.",
			Answer:      false,
			Explanation: "파블로 피카소는 스페인 출신의 화가입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "상식",
			Difficulty:  "easy",
			Question:    "한국의 수도는 서울이다.",
			Answer:      true,
			Explanation: "대한민국의 수도는 서울특별시입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  "easy",
			Question:    "태양은 동쪽에서 떠서 서쪽으로 진다.",
			Answer:      true,
			Explanation: "지구의 자전 방향 때문에 태양은 동쪽에서 떠서 서쪽으로 집니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  "medium",
			Question:    "물은 섭씨 100도에서 끓는다.",
			Answer:      true,
			Explanation: "표준 기압(1기압)에서 물은 100°C에서 끓습니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  "easy",
			Question:    "사람의 심장은 오른쪽 가슴에 있다.",
			Answer:      false,
			Explanation: "심장은 가슴의 중앙에서 약간 왼쪽에 위치합니다.",
		},
	}

	// Insert OX quizzes
	for _, quiz := range oxQuizzes {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure unique IDs
		_, err := oxCollection.InsertOne(ctx, quiz)
		if err != nil {
			log.Printf("Failed to insert OX quiz: %v", err)
		} else {
			question := quiz.Question
			if len(question) > 30 {
				question = question[:30]
			}
			log.Printf("✓ Inserted OX quiz: %s (ID: %d)", question, quiz.ID)
		}
	}

	// Add more QA quizzes
	qaQuizzes := []models.GeneralQuiz{
		{
			ID:          node.Generate().Int64(),
			Category:    "상식",
			Difficulty:  "easy",
			Question:    "다음 중 김치의 주재료는?",
			Options:     []string{"배추", "양파", "감자", "당근"},
			Answer:      0,
			Explanation: "김치는 주로 배추를 절여서 만듭니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "경제",
			Difficulty:  "easy",
			Question:    "한국의 화폐 단위는?",
			Options:     []string{"원", "엔", "달러", "유로"},
			Answer:      0,
			Explanation: "대한민국의 화폐 단위는 원(₩)입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  "easy",
			Question:    "지구의 위성은?",
			Options:     []string{"달", "화성", "금성", "목성"},
			Answer:      0,
			Explanation: "지구의 유일한 자연 위성은 달입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "역사",
			Difficulty:  "medium",
			Question:    "조선을 건국한 인물은?",
			Options:     []string{"이성계", "세종대왕", "정조", "광개토대왕"},
			Answer:      0,
			Explanation: "이성계(태조)가 1392년에 조선을 건국했습니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "지리",
			Difficulty:  "easy",
			Question:    "프랑스의 수도는?",
			Options:     []string{"파리", "런던", "베를린", "로마"},
			Answer:      0,
			Explanation: "프랑스의 수도는 파리입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "음악",
			Difficulty:  "medium",
			Question:    "피아노의 건반은 몇 개?",
			Options:     []string{"88개", "76개", "61개", "100개"},
			Answer:      0,
			Explanation: "표준 피아노는 88개의 건반을 가지고 있습니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "스포츠",
			Difficulty:  "easy",
			Question:    "올림픽은 몇 년마다 개최되나?",
			Options:     []string{"4년", "2년", "5년", "3년"},
			Answer:      0,
			Explanation: "하계/동계 올림픽 모두 4년마다 개최됩니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  "medium",
			Question:    "빛의 속도는 대략?",
			Options:     []string{"30만 km/s", "10만 km/s", "50만 km/s", "100만 km/s"},
			Answer:      0,
			Explanation: "빛의 속도는 약 299,792km/s입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "미술",
			Difficulty:  "easy",
			Question:    "모나리자를 그린 화가는?",
			Options:     []string{"레오나르도 다 빈치", "피카소", "고흐", "모네"},
			Answer:      0,
			Explanation: "모나리자는 레오나르도 다 빈치의 대표작입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "수학",
			Difficulty:  "easy",
			Question:    "원주율(π)은 대략?",
			Options:     []string{"3.14", "2.71", "1.41", "4.20"},
			Answer:      0,
			Explanation: "원주율(π)은 약 3.14159...입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "상식",
			Difficulty:  "easy",
			Question:    "1년은 몇 개월?",
			Options:     []string{"12개월", "10개월", "14개월", "11개월"},
			Answer:      0,
			Explanation: "1년은 12개월(1월~12월)입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "기술",
			Difficulty:  "medium",
			Question:    "인터넷 프로토콜은?",
			Options:     []string{"TCP/IP", "HTTP/2", "FTP/S", "SMTP"},
			Answer:      0,
			Explanation: "TCP/IP는 인터넷의 기본 통신 프로토콜입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "동물",
			Difficulty:  "easy",
			Question:    "한국의 국가 동물은?",
			Options:     []string{"호랑이", "곰", "사자", "독수리"},
			Answer:      0,
			Explanation: "호랑이는 대한민국의 상징 동물입니다.",
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "영화",
			Difficulty:  "medium",
			Question:    "2020년 아카데미 작품상 수상작은?",
			Options:     []string{"기생충", "어벤져스", "겨울왕국", "타이타닉"},
			Answer:      0,
			Explanation: "기생충은 2020년 아카데미 작품상을 수상한 한국 영화입니다.",
		},
	}

	// Insert QA quizzes
	for _, quiz := range qaQuizzes {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure unique IDs
		_, err := qaCollection.InsertOne(ctx, quiz)
		if err != nil {
			log.Printf("Failed to insert QA quiz: %v", err)
		} else {
			question := quiz.Question
			if len(question) > 30 {
				question = question[:30]
			}
			log.Printf("✓ Inserted QA quiz: %s (ID: %d)", question, quiz.ID)
		}
	}

	// Final count
	oxCount, _ = oxCollection.CountDocuments(ctx, bson.M{})
	qaCount, _ = qaCollection.CountDocuments(ctx, bson.M{})
	log.Printf("\n========================================")
	log.Printf("✓ Quiz addition completed!")
	log.Printf("Total OX Quizzes in DB: %d", oxCount)
	log.Printf("Total QA Quizzes in DB: %d", qaCount)
	log.Printf("========================================")
}
