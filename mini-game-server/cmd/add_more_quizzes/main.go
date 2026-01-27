package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OXQuiz struct
type OXQuiz struct {
	ID          int64     `bson:"_id"`
	Category    string    `bson:"category"`
	Difficulty  int       `bson:"difficulty"`
	Question    string    `bson:"question"`
	Answer      bool      `bson:"answer"`
	Explanation string    `bson:"explanation"`
	UsageCount  int       `bson:"usage_count"`
	IsActive    bool      `bson:"is_active"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

// GeneralQuiz struct
type GeneralQuiz struct {
	ID          int64     `bson:"_id"`
	Category    string    `bson:"category"`
	Difficulty  int       `bson:"difficulty"`
	Question    string    `bson:"question"`
	Options     []string  `bson:"options"`
	Answer      int       `bson:"answer"`
	Explanation string    `bson:"explanation"`
	ImageURL    string    `bson:"image_url"`
	UsageCount  int       `bson:"usage_count"`
	IsActive    bool      `bson:"is_active"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

func main() {
	log.Println("Starting quiz addition script...")

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

	// Get collections
	oxCollection := db.Collection("ox_quizzes")
	qaCollection := db.Collection("general_quizzes")

	// Initialize Snowflake node for ID generation
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("Failed to create snowflake node: %v", err)
	}

	ctx = context.Background()

	// Check current counts
	oxCount, _ := oxCollection.CountDocuments(ctx, bson.M{})
	qaCount, _ := qaCollection.CountDocuments(ctx, bson.M{})
	log.Printf("Current counts - OX: %d, QA: %d", oxCount, qaCount)

	// Add more OX quizzes
	oxQuizzes := []OXQuiz{
		{
			ID:          node.Generate().Int64(),
			Category:    "역사",
			Difficulty:  1,
			Question:    "세종대왕이 한글을 창제하였다.",
			Answer:      true,
			Explanation: "세종대왕이 1443년에 훈민정음(한글)을 창제하였습니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "지리",
			Difficulty:  1,
			Question:    "에베레스트는 세계에서 가장 높은 산이다.",
			Answer:      true,
			Explanation: "에베레스트 산은 해발 8,849m로 세계에서 가장 높은 산입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  1,
			Question:    "태양계는 8개의 행성이다.",
			Answer:      false,
			Explanation: "태양계는 7개의 행성입니다 (명왕성이 제외됨, 2006년).",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "기술",
			Difficulty:  1,
			Question:    "인터넷은 군사용으로 처음 개발되었다.",
			Answer:      true,
			Explanation: "인터넷은 미국 국방부가 개발한 ARPANET에서 시작되었습니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "기술",
			Difficulty:  1,
			Question:    "블루투스는 무선 통신 기술이다.",
			Answer:      true,
			Explanation: "블루투스는 단거리 무선 통신 표준 기술입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "지리",
			Difficulty:  1,
			Question:    "한국의 수도는 서울이다.",
			Answer:      true,
			Explanation: "대한민국의 수도는 서울특별시입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "기술",
			Difficulty:  1,
			Question:    "태양은 동쪽에서 떠서 서쪽으로 진다.",
			Answer:      true,
			Explanation: "지구의 자전 방향 때문에 태양은 동쪽에서 떠서 서쪽으로 집니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "기술",
			Difficulty:  2,
			Question:    "물의 끓는 점은 100도이다.",
			Answer:      true,
			Explanation: "표준 대기압에서 물의 끓는 점은 100°C입니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "기술",
			Difficulty:  2,
			Question:    "지구의 중심은 액체로 이루어져있다.",
			Answer:      false,
			Explanation: "지구의 중심은 고체 상태의 내핵과 액체 상태의 외핵으로 이루어져 있습니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "스포츠",
			Difficulty:  2,
			Question:    "올림픽은 4년마다 개최된다.",
			Answer:      true,
			Explanation: "하계 올림픽과 동계 올림픽 모두 4년마다 개최됩니다.",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Insert OX quizzes
	for _, quiz := range oxQuizzes {
		_, err := oxCollection.InsertOne(ctx, quiz)
		if err != nil {
			log.Printf("Failed to insert OX quiz: %v", err)
			continue
		}
		log.Printf("Added OX quiz: %s", quiz.Question)
	}

	// Add QA quizzes
	qaQuizzes := []GeneralQuiz{
		{
			ID:          node.Generate().Int64(),
			Category:    "역사",
			Difficulty:  1,
			Question:    "대한민국의 수도는?",
			Options:     []string{"서울", "부산", "대전", "인천"},
			Answer:      0,
			Explanation: "대한민국의 수도는 서울특별시입니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  2,
			Question:    "물의 화학식은?",
			Options:     []string{"H2O", "CO2", "O2", "H2SO4"},
			Answer:      0,
			Explanation: "물의 화학식은 H2O(수소 2개, 산소 1개)입니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "수학",
			Difficulty:  1,
			Question:    "2 + 2 = ?",
			Options:     []string{"3", "4", "5", "6"},
			Answer:      1,
			Explanation: "2 더하기 2는 4입니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "지리",
			Difficulty:  2,
			Question:    "세계에서 가장 큰 대륙은?",
			Options:     []string{"아시아", "아프리카", "유럽", "남미"},
			Answer:      0,
			Explanation: "아시아는 면적이 약 4,400만 km²로 세계에서 가장 큰 대륙입니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "역사",
			Difficulty:  3,
			Question:    "제2차 세계대전이 끝난 해는?",
			Options:     []string{"1943", "1944", "1945", "1946"},
			Answer:      2,
			Explanation: "제2차 세계대전은 1945년에 종료되었습니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "과학",
			Difficulty:  3,
			Question:    "빛의 속도는 약 얼마인가?",
			Options:     []string{"30만 km/s", "10만 km/s", "50만 km/s", "100만 km/s"},
			Answer:      0,
			Explanation: "진공에서 빛의 속도는 약 30만 km/s (정확히는 299,792,458 m/s)입니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "기술",
			Difficulty:  4,
			Question:    "최초의 프로그래밍 언어는?",
			Options:     []string{"FORTRAN", "COBOL", "Assembly", "Plankalkül"},
			Answer:      3,
			Explanation: "Plankalkül은 1942-1945년 사이에 콘라드 추제가 설계한 최초의 고급 프로그래밍 언어입니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "스포츠",
			Difficulty:  2,
			Question:    "축구 경기에서 한 팀은 몇 명의 선수로 구성되는가?",
			Options:     []string{"9명", "10명", "11명", "12명"},
			Answer:      2,
			Explanation: "축구 경기는 각 팀당 골키퍼를 포함한 11명의 선수로 진행됩니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "예술",
			Difficulty:  3,
			Question:    "모나리자를 그린 화가는?",
			Options:     []string{"피카소", "고흐", "레오나르도 다빈치", "모네"},
			Answer:      2,
			Explanation: "모나리자는 레오나르도 다빈치가 1503-1519년에 그린 작품입니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          node.Generate().Int64(),
			Category:    "음악",
			Difficulty:  2,
			Question:    "피아노의 건반은 총 몇 개인가?",
			Options:     []string{"76개", "88개", "92개", "100개"},
			Answer:      1,
			Explanation: "표준 피아노는 88개의 건반(52개 흰 건반, 36개 검은 건반)으로 구성됩니다.",
			ImageURL:    "",
			UsageCount:  0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Insert QA quizzes
	for _, quiz := range qaQuizzes {
		_, err := qaCollection.InsertOne(ctx, quiz)
		if err != nil {
			log.Printf("Failed to insert QA quiz: %v", err)
			continue
		}
		log.Printf("Added QA quiz: %s", quiz.Question)
	}

	// Final count
	oxCount, _ = oxCollection.CountDocuments(ctx, bson.M{})
	qaCount, _ = qaCollection.CountDocuments(ctx, bson.M{})
	log.Printf("Final counts - OX: %d, QA: %d", oxCount, qaCount)
	log.Println("Quiz addition completed successfully!")
}
