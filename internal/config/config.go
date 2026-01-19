package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	MongoURI       string
	MongoDB        string
	ValkeyAddr     string
	ValkeyPassword string
	ServerPort     string
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
	MongoDB = getEnv("MONGO_DB", "draw_and_guess_db")
	ValkeyAddr = getEnv("VALKEY_ADDR", "localhost:6379")
	ValkeyPassword = getEnv("VALKEY_PASSWORD", "")
	ServerPort = getEnv("SERVER_PORT", "8080")

	// Validate essential configurations
	if ServerPort == "" {
		log.Fatal("SERVER_PORT must be set")
	}

	log.Printf("Config loaded - ServerPort: %s, ValkeyAddr: %s, MongoDB: %s/%s",
		ServerPort, ValkeyAddr, MongoURI, MongoDB)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
