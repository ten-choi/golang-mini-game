package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"draw-and-guess-server/config"
	"draw-and-guess-server/database"
	_ "draw-and-guess-server/docs"
	"draw-and-guess-server/handlers"
	"draw-and-guess-server/valkey"
	"draw-and-guess-server/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Draw and Guess Game API
// @version 1.0
// @description API for Draw and Guess multiplayer game
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /app

// @schemes http https
func main() {
	// Initialize config
	config.Init()

	// Connect to Valkey
	if err := valkey.Connect(); err != nil {
		log.Fatalf("Failed to connect to Valkey: %v", err)
	}

	// Connect to MongoDB (required for topics)
	if err := database.Connect(); err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	if err := handlers.LoadTopicsFromMongo(context.Background()); err != nil {
		log.Fatalf("Failed to load topics from MongoDB: %v", err)
	}
	defer func() {
		if err := database.Disconnect(); err != nil {
			log.Printf("Failed to disconnect MongoDB: %v", err)
		}
	}()

	// Create Gin router
	r := gin.Default()

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true // Allow all origins for ngrok compatibility
	corsConfig.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	// Logging middleware
	r.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latency := endTime.Sub(startTime)
		latencyInMilliseconds := float64(latency) / float64(time.Millisecond)

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		endpoint := c.Request.URL.Path

		var logMessage string
		if method == "GET" {
			params := c.Request.URL.RawQuery
			logMessage = fmt.Sprintf("| %d | %.3fms | %s | %s %s?%s",
				statusCode, latencyInMilliseconds, clientIP, method, endpoint, params)
		} else {
			logMessage = fmt.Sprintf("| %d | %.3fms | %s | %s %s",
				statusCode, latencyInMilliseconds, clientIP, method, endpoint)
		}

		log.Println(logMessage)
	})

	// Game Room routes
	r.GET("/app/game/rooms", handlers.GetGameRooms)
	r.POST("/app/game/room", handlers.CreateGameRoom)
	r.POST("/app/game/room/:id/join", handlers.JoinGameRoom)
	r.POST("/app/game/room/:id/leave", handlers.LeaveGameRoom)
	r.POST("/app/game/room/:id/start", handlers.StartGame)
	r.POST("/app/game/room/:id/answer", handlers.CheckAnswer)
	r.POST("/app/game/room/:id/chat", handlers.HandleChatMessage)
	r.PATCH("/app/game/room/:id", handlers.UpdateGameRoom)
	r.DELETE("/app/game/room/:id", handlers.DeleteGameRoom)

	// User routes
	r.GET("/app/user/:id", handlers.GetUser)
	r.POST("/app/user", handlers.CreateUser)
	r.PATCH("/app/user/:id", handlers.UpdateUser)

	// WebSocket route
	r.GET("/app/ws", websocket.HandleWebSocket)

	// AsyncAPI documentation routes
	r.GET("/app/asyncapi", handlers.AsyncAPIDocumentation)
	r.GET("/app/asyncapi.yaml", handlers.GetAsyncAPISpec)
	r.GET("/app/asyncapi.json", handlers.GetAsyncAPISpec)

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := config.ServerPort
	log.Printf("Starting server on port %s...", port)
	log.Printf("Swagger UI (REST API) available at: http://localhost:%s/swagger/index.html", port)
	log.Printf("AsyncAPI (WebSocket) available at: http://localhost:%s/app/asyncapi", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
