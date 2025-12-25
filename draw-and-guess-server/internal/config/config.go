package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string
	ValkeyAddr       string
	ServerPort       string
)

func Init() {
	// Load .env file if it exists (for local development)
	// In production/k8s, environment variables are set directly
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables or defaults")
	} else {
		log.Println("Loaded configuration from .env file")
	}

	PostgresHost = getEnv("POSTGRES_HOST", "localhost")
	PostgresPort = getEnv("POSTGRES_PORT", "5432")
	PostgresUser = getEnv("POSTGRES_USER", "postgres")
	PostgresPassword = getEnv("POSTGRES_PASSWORD", "password")
	PostgresDBName = getEnv("POSTGRES_DB", "draw_and_guess_db")
	ValkeyAddr = getEnv("VALKEY_ADDR", "localhost:6379")
	ServerPort = getEnv("SERVER_PORT", "8080")

	// Validate essential configurations
	if ServerPort == "" {
		log.Fatal("SERVER_PORT must be set")
	}

	log.Printf("Config loaded - ServerPort: %s, ValkeyAddr: %s, PostgreSQL: %s:%s/%s",
		ServerPort, ValkeyAddr, PostgresHost, PostgresPort, PostgresDBName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
