package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"draw-and-guess-server/src/config"
	"draw-and-guess-server/src/database"
	"draw-and-guess-server/src/graphql"
	"draw-and-guess-server/src/handlers"
	"draw-and-guess-server/src/valkey"
	"draw-and-guess-server/src/websocket"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	gqlhandler "github.com/graphql-go/handler"
)

func main() {
	// Initialize config
	config.Init()

	// Connect to Valkey
	if err := valkey.Connect(); err != nil {
		log.Printf("Valkey connection warning: %v (continuing without cache)", err)
	} else {
		log.Println("Successfully connected to Valkey")
	}

	// Connect to MongoDB
	if err := database.Connect(); err != nil {
		log.Printf("MongoDB connection warning: %v (continuing with hardcoded topics)", err)
	} else {
		log.Println("Successfully connected to MongoDB")
		defer func() {
			if err := database.Disconnect(); err != nil {
				log.Printf("Failed to disconnect MongoDB: %v", err)
			}
		}()
	}

	// Load topics
	if err := handlers.LoadTopicsFromMongo(context.Background()); err != nil {
		log.Printf("Failed to load topics from MongoDB, using hardcoded data: %v", err)
	}

	// Create Gin router
	r := gin.Default()

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
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

		logMessage := fmt.Sprintf("| %d | %.3fms | %s | %s %s",
			statusCode, latencyInMilliseconds, clientIP, method, endpoint)
		log.Println(logMessage)
	})

	// Health check
	r.GET("/health", handlers.HealthCheck)

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
		Schema: &graphql.Schema,
		Pretty: true,
	})
	r.GET("/graphql", gin.WrapH(playground.ApolloSandboxHandler("GraphQL", "/graphql")))
	r.POST("/graphql", gin.WrapH(h))

	// Start server
	port := config.ServerPort
	log.Printf("Server started on port %s", port)
	log.Printf("Apollo Sandbox: http://localhost:%s/graphql", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
