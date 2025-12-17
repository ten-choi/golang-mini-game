package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"draw-and-guess-server/config"
	"draw-and-guess-server/database"
	"draw-and-guess-server/graphql"
	"draw-and-guess-server/handlers"
	"draw-and-guess-server/valkey"
	"draw-and-guess-server/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	gqlhandler "github.com/graphql-go/handler"
)

func main() {
	// Initialize config
	config.Init()

	// Connect to Valkey
	if err := valkey.Connect(); err != nil {
		log.Fatalf("Failed to connect to Valkey: %v", err)
	}

	// Connect to MongoDB (optional for now - using hardcoded topics)
	if err := database.Connect(); err != nil {
		log.Printf("MongoDB connection warning: %v (continuing with hardcoded topics)", err)
	} else {
		defer func() {
			if err := database.Disconnect(); err != nil {
				log.Printf("Failed to disconnect MongoDB: %v", err)
			}
		}()
	}

	// Load topics (currently using hardcoded data for testing)
	if err := handlers.LoadTopicsFromMongo(context.Background()); err != nil {
		log.Fatalf("Failed to load topics: %v", err)
	}

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

	// GraphQL endpoint
	h := gqlhandler.New(&gqlhandler.Config{
		Schema:   &graphql.Schema,
		Pretty:   true,
		GraphiQL: true,
	})
	r.POST("/graphql", gin.WrapH(h))
	r.GET("/graphql", gin.WrapH(h))

	// Start server
	port := config.ServerPort
	log.Printf("Starting server on port %s...", port)
	log.Printf("GraphQL Playground available at: http://localhost:%s/graphql", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
