package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	MongoURI     string
	ValkeyAddr   string
	DatabaseName string
	ServerPort   string
)

func Init() {
	// Load .env file if it exists (for local development)
	// In production/k8s, environment variables are set directly
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables or defaults")
	} else {
		log.Println("Loaded configuration from .env file")
	}

	MongoURI = getEnv("MONGO_URI", "mongodb://localhost:27017")
	ValkeyAddr = getEnv("VALKEY_ADDR", "localhost:6379")
	DatabaseName = getEnv("DATABASE_NAME", "draw_and_guess_db")
	ServerPort = getEnv("SERVER_PORT", "8080")

	log.Printf("Config loaded - ServerPort: %s, ValkeyAddr: %s, MongoURI: %s", ServerPort, ValkeyAddr, MongoURI)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
