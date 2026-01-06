package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"draw-and-guess-server/internal/config"
	"draw-and-guess-server/pkg/utils"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("Loading configuration...")
	config.Init()

	log.Println("Connecting to database...")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword, config.PostgresDBName)

	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize Snowflake ID generator
	if err := utils.InitSnowflake(1); err != nil {
		log.Fatalf("Failed to initialize snowflake: %v", err)
	}

	ctx := context.Background()

	// Insert OX Quizzes
	log.Println("Inserting OX quiz data...")
	oxQuizzes := []struct {
		category    string
		difficulty  string
		question    string
		answer      bool
		explanation string
	}{
		{"Science", "easy", "The Earth is flat.", false, "The Earth is approximately spherical in shape."},
		{"History", "easy", "The Great Wall of China was built to keep out invaders.", true, "The Great Wall was primarily built as a defense system against invasions."},
		{"Geography", "easy", "Mount Everest is the tallest mountain in the world.", true, "Mount Everest stands at 8,849 meters above sea level."},
		{"Science", "medium", "Water boils at 100 degrees Celsius at sea level.", true, "At standard atmospheric pressure, water boils at 100°C."},
		{"Biology", "easy", "Humans have five senses.", true, "The traditional five senses are sight, hearing, touch, taste, and smell."},
		{"History", "medium", "The United States declared independence in 1776.", true, "The Declaration of Independence was signed on July 4, 1776."},
		{"Science", "easy", "The Sun revolves around the Earth.", false, "The Earth revolves around the Sun. This is called heliocentric model."},
		{"Math", "easy", "A triangle has four sides.", false, "A triangle has three sides. A quadrilateral has four sides."},
		{"Technology", "easy", "HTML stands for HyperText Markup Language.", true, "HTML is the standard markup language for creating web pages."},
		{"Geography", "easy", "Africa is the largest continent by area.", false, "Asia is the largest continent. Africa is the second largest."},
		{"Biology", "medium", "Sharks are mammals.", false, "Sharks are fish, not mammals. They breathe through gills."},
		{"Physics", "medium", "Light travels faster than sound.", true, "Light travels at approximately 299,792 km/s, sound at about 343 m/s."},
		{"History", "easy", "World War II ended in 1945.", true, "World War II officially ended on September 2, 1945."},
		{"Science", "easy", "Diamonds are made of carbon.", true, "Diamonds are composed entirely of carbon atoms."},
		{"Geography", "medium", "The Amazon River is longer than the Nile River.", false, "The Nile River is generally considered the longest river."},
	}

	for _, quiz := range oxQuizzes {
		id := utils.GenerateID()
		_, err := database.ExecContext(ctx,
			`INSERT INTO ox_quizzes (id, category, difficulty, question, answer, explanation, usage_count, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, 0, true, $7, $7)
			ON CONFLICT (id) DO NOTHING`,
			id, quiz.category, quiz.difficulty, quiz.question, quiz.answer, quiz.explanation, time.Now())
		if err != nil {
			log.Printf("Failed to insert OX quiz: %v", err)
		}
	}

	// Insert QA Quizzes
	log.Println("Inserting QA quiz data...")
	qaQuizzes := []struct {
		category   string
		difficulty string
		question   string
		options    []string
		answer     int
	}{
		{"General Knowledge", "easy", "What is the capital of France?", []string{"London", "Paris", "Berlin", "Madrid"}, 1},
		{"Science", "easy", "What is H2O commonly known as?", []string{"Water", "Oxygen", "Hydrogen", "Carbon"}, 0},
		{"History", "medium", "Who painted the Mona Lisa?", []string{"Michelangelo", "Leonardo da Vinci", "Raphael", "Donatello"}, 1},
		{"Geography", "easy", "Which continent is known as the Dark Continent?", []string{"Asia", "Africa", "South America", "Australia"}, 1},
		{"Science", "medium", "What is the chemical symbol for gold?", []string{"Go", "Au", "Gd", "Ag"}, 1},
		{"Sports", "easy", "How many players are on a soccer team?", []string{"9", "10", "11", "12"}, 2},
		{"Technology", "easy", "What does CPU stand for?", []string{"Computer Processing Unit", "Central Processing Unit", "Central Program Utility", "Computer Program Unit"}, 1},
		{"Literature", "medium", "Who wrote Romeo and Juliet?", []string{"Charles Dickens", "William Shakespeare", "Jane Austen", "Mark Twain"}, 1},
		{"Math", "easy", "What is the result of 7 x 8?", []string{"54", "55", "56", "57"}, 2},
		{"Science", "easy", "What planet is known as the Red Planet?", []string{"Venus", "Jupiter", "Mars", "Saturn"}, 2},
		{"History", "hard", "In which year did Christopher Columbus discover America?", []string{"1490", "1491", "1492", "1493"}, 2},
		{"Geography", "medium", "What is the smallest country in the world?", []string{"Monaco", "Vatican City", "San Marino", "Liechtenstein"}, 1},
		{"Music", "easy", "How many keys does a standard piano have?", []string{"76", "85", "88", "92"}, 2},
		{"Science", "medium", "What is the speed of light in vacuum?", []string{"299792 km/s", "300000 km/s", "299000 km/s", "298000 km/s"}, 0},
		{"General Knowledge", "easy", "What is the largest ocean on Earth?", []string{"Atlantic", "Pacific", "Indian", "Arctic"}, 1},
	}

	for _, quiz := range qaQuizzes {
		id := utils.GenerateID()
		optionsJSON, _ := json.Marshal(quiz.options)
		_, err := database.ExecContext(ctx,
			`INSERT INTO qa_quizzes (id, category, difficulty, question, options, answer, usage_count, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, 0, true, $7, $7)
			ON CONFLICT (id) DO NOTHING`,
			id, quiz.category, quiz.difficulty, quiz.question, optionsJSON, quiz.answer, time.Now())
		if err != nil {
			log.Printf("Failed to insert QA quiz: %v", err)
		}
	}

	// Verify counts
	var oxCount, qaCount int
	database.QueryRowContext(ctx, "SELECT COUNT(*) FROM ox_quizzes WHERE is_active = true").Scan(&oxCount)
	database.QueryRowContext(ctx, "SELECT COUNT(*) FROM qa_quizzes WHERE is_active = true").Scan(&qaCount)

	fmt.Printf("\n✅ Quiz data inserted successfully!\n")
	fmt.Printf("   OX Quizzes: %d\n", oxCount)
	fmt.Printf("   QA Quizzes: %d\n", qaCount)
}
