package config

import (
	"os"
)

var (
	MongoURI     string
	ValkeyAddr   string
	DatabaseName string
	ServerPort   string
)

func Init() {
	MongoURI = getEnv("MONGO_URI", "mongodb://localhost:27017")
	ValkeyAddr = getEnv("VALKEY_ADDR", "localhost:6379")
	DatabaseName = getEnv("DATABASE_NAME", "draw_and_guess_db")
	ServerPort = getEnv("SERVER_PORT", "8080")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
