package main

import (
	"context"
	"log"
	"time"

	"draw-and-guess-server/internal/config"
	"draw-and-guess-server/internal/database"
	"draw-and-guess-server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// Load configuration
	config.Init()

	// Connect to MongoDB
	err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	ctx := context.Background()
	oxCollection := database.DB.Collection("ox_quizzes")
	qaCollection := database.DB.Collection("qa_quizzes")

	log.Println("Starting to add quizzes...")

	// Add OX Quizzes
	oxQuizzes := []models.OXQuiz{
		{
			ID:          generateID(),
			Category:    "상식",
			Difficulty:  "EASY",
			Question:    "한국의 수도는 서울이다.",
			Answer:      true,
			Explanation: "서울은 대한민국의 수도입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "상식",
			Difficulty:  "EASY",
			Question:    "태양은 서쪽에서 뜬다.",
			Answer:      false,
			Explanation: "태양은 동쪽에서 떠서 서쪽으로 집니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "과학",
			Difficulty:  "EASY",
			Question:    "물은 섭씨 100도에서 끓는다.",
			Answer:      true,
			Explanation: "표준 기압에서 물의 끓는점은 섭씨 100도입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "과학",
			Difficulty:  "MEDIUM",
			Question:    "인간의 심장은 오른쪽 가슴에 위치한다.",
			Answer:      false,
			Explanation: "심장은 가슴의 중앙에서 약간 왼쪽에 위치합니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "역사",
			Difficulty:  "EASY",
			Question:    "한글을 만든 사람은 세종대왕이다.",
			Answer:      true,
			Explanation: "세종대왕이 1443년에 한글을 창제했습니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "지리",
			Difficulty:  "MEDIUM",
			Question:    "에베레스트 산은 한국에 있다.",
			Answer:      false,
			Explanation: "에베레스트 산은 네팔과 티베트 경계에 있습니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "상식",
			Difficulty:  "EASY",
			Question:    "일주일은 7일이다.",
			Answer:      true,
			Explanation: "일주일은 월요일부터 일요일까지 7일입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "과학",
			Difficulty:  "MEDIUM",
			Question:    "다이아몬드는 탄소로 이루어져 있다.",
			Answer:      true,
			Explanation: "다이아몬드는 순수한 탄소 결정체입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "스포츠",
			Difficulty:  "EASY",
			Question:    "축구 경기는 11명이 한 팀을 이룬다.",
			Answer:      true,
			Explanation: "축구는 골키퍼 포함 11명이 한 팀입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "문화",
			Difficulty:  "MEDIUM",
			Question:    "피카소는 한국 화가이다.",
			Answer:      false,
			Explanation: "파블로 피카소는 스페인 출신의 화가입니다.",
			CreatedAt:   time.Now(),
		},
	}

	// Insert OX Quizzes
	for _, quiz := range oxQuizzes {
		// Check if already exists
		count, err := oxCollection.CountDocuments(ctx, bson.M{"question": quiz.Question})
		if err != nil {
			log.Printf("Error checking OX quiz: %v", err)
			continue
		}
		if count > 0 {
			log.Printf("OX Quiz already exists: %s", quiz.Question)
			continue
		}

		_, err = oxCollection.InsertOne(ctx, quiz)
		if err != nil {
			log.Printf("Failed to insert OX quiz: %v", err)
		} else {
			log.Printf("✓ Added OX quiz: %s", quiz.Question)
		}
	}

	// Add QA Quizzes
	qaQuizzes := []models.GeneralQuiz{
		{
			ID:          generateID(),
			Category:    "상식",
			Difficulty:  "EASY",
			Question:    "한국의 화폐 단위는?",
			Options:     []string{"원", "엔", "달러", "유로"},
			Answer:      0,
			Explanation: "한국의 화폐 단위는 원(₩)입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "과학",
			Difficulty:  "EASY",
			Question:    "지구의 위성은?",
			Options:     []string{"태양", "화성", "달", "금성"},
			Answer:      2,
			Explanation: "달은 지구의 유일한 자연 위성입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "역사",
			Difficulty:  "MEDIUM",
			Question:    "조선을 건국한 인물은?",
			Options:     []string{"이성계", "세종대왕", "광개토대왕", "김유신"},
			Answer:      0,
			Explanation: "이성계(태조)가 1392년에 조선을 건국했습니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "지리",
			Difficulty:  "EASY",
			Question:    "프랑스의 수도는?",
			Options:     []string{"런던", "베를린", "파리", "로마"},
			Answer:      2,
			Explanation: "파리는 프랑스의 수도이자 최대 도시입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "음악",
			Difficulty:  "MEDIUM",
			Question:    "피아노의 건반은 흰색과 검은색을 합쳐 몇 개?",
			Options:     []string{"76개", "88개", "96개", "108개"},
			Answer:      1,
			Explanation: "표준 피아노는 88개의 건반을 가지고 있습니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "스포츠",
			Difficulty:  "EASY",
			Question:    "올림픽은 몇 년마다 개최되는가?",
			Options:     []string{"2년", "3년", "4년", "5년"},
			Answer:      2,
			Explanation: "하계 올림픽과 동계 올림픽 모두 4년마다 개최됩니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "과학",
			Difficulty:  "MEDIUM",
			Question:    "빛의 속도는 초당 약 얼마?",
			Options:     []string{"30만 km", "100만 km", "1억 km", "10억 km"},
			Answer:      0,
			Explanation: "빛의 속도는 초당 약 30만 킬로미터(299,792 km)입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "문화",
			Difficulty:  "EASY",
			Question:    "모나리자를 그린 화가는?",
			Options:     []string{"고흐", "모네", "레오나르도 다 빈치", "미켈란젤로"},
			Answer:      2,
			Explanation: "모나리자는 레오나르도 다 빈치의 대표작입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "수학",
			Difficulty:  "EASY",
			Question:    "원주율(π)의 근사값은?",
			Options:     []string{"2.14", "3.14", "4.14", "5.14"},
			Answer:      1,
			Explanation: "원주율 π는 약 3.14159...입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "상식",
			Difficulty:  "EASY",
			Question:    "1년은 몇 개월?",
			Options:     []string{"10개월", "11개월", "12개월", "13개월"},
			Answer:      2,
			Explanation: "1년은 1월부터 12월까지 12개월입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "기술",
			Difficulty:  "MEDIUM",
			Question:    "인터넷의 기본 프로토콜은?",
			Options:     []string{"FTP", "HTTP", "TCP/IP", "SMTP"},
			Answer:      2,
			Explanation: "TCP/IP는 인터넷의 기본 통신 프로토콜입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "동물",
			Difficulty:  "EASY",
			Question:    "고양이과 동물 중 가장 큰 동물은?",
			Options:     []string{"사자", "호랑이", "표범", "치타"},
			Answer:      1,
			Explanation: "호랑이는 고양이과 동물 중 가장 큰 종입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "영화",
			Difficulty:  "MEDIUM",
			Question:    "아카데미 작품상을 받은 한국 영화는?",
			Options:     []string{"올드보이", "기생충", "부산행", "타짜"},
			Answer:      1,
			Explanation: "기생충은 2020년 아카데미 작품상을 받았습니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "언어",
			Difficulty:  "EASY",
			Question:    "영어 알파벳은 몇 개?",
			Options:     []string{"24개", "26개", "28개", "30개"},
			Answer:      1,
			Explanation: "영어 알파벳은 A부터 Z까지 26개입니다.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          generateID(),
			Category:    "음식",
			Difficulty:  "EASY",
			Question:    "김치의 주재료는?",
			Options:     []string{"무", "배추", "오이", "양파"},
			Answer:      1,
			Explanation: "배추김치는 한국의 대표적인 김치입니다.",
			CreatedAt:   time.Now(),
		},
	}

	// Insert QA Quizzes
	for _, quiz := range qaQuizzes {
		// Check if already exists
		count, err := qaCollection.CountDocuments(ctx, bson.M{"question": quiz.Question})
		if err != nil {
			log.Printf("Error checking QA quiz: %v", err)
			continue
		}
		if count > 0 {
			log.Printf("QA Quiz already exists: %s", quiz.Question)
			continue
		}

		_, err = qaCollection.InsertOne(ctx, quiz)
		if err != nil {
			log.Printf("Failed to insert QA quiz: %v", err)
		} else {
			log.Printf("✓ Added QA quiz: %s", quiz.Question)
		}
	}

	// Print summary
	oxCount, _ := oxCollection.CountDocuments(ctx, bson.M{})
	qaCount, _ := qaCollection.CountDocuments(ctx, bson.M{})

	log.Println("\n========================================")
	log.Printf("✓ Quiz addition completed!")
	log.Printf("Total OX Quizzes in DB: %d", oxCount)
	log.Printf("Total QA Quizzes in DB: %d", qaCount)
	log.Println("========================================")
}

func generateID() int64 {
	return time.Now().UnixNano() / 1000000
}
