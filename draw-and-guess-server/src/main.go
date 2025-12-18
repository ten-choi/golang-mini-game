package main

import (
	"context"
	"log"
	"net/http"
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

	// -----------------------------
	// CORS (server-safe settings)
	// -----------------------------
	// IMPORTANT:
	// - If AllowCredentials = true, you MUST NOT use AllowAllOrigins = true.
	// - Add Apollo Studio/Sandbox related headers.
	corsConfig := cors.Config{
		AllowOrigins: []string{
			"https://studio.apollographql.com",
			// Add your real domain(s) here if you have them, e.g.:
			// "https://your-service.example.com",
			// Local dev (optional):
			"http://localhost:8080",
			"http://localhost:3000",
		},
		AllowMethods: []string{"GET", "POST", "OPTIONS", "PATCH", "PUT", "DELETE"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",

			// Apollo Sandbox / Studio often sends these:
			"apollo-require-preflight",
			"x-apollo-operation-name",
			"apollo-client-name",
			"apollo-client-version",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// (Optional but helpful) Explicit OPTIONS handler for /graphql
	// Some proxies/ingress setups can be picky about OPTIONS.
	r.OPTIONS("/graphql", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// Logging middleware
	r.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		latency := time.Since(startTime)

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		endpoint := c.Request.URL.Path

		log.Printf("| %d | %.3fms | %s | %s %s",
			statusCode,
			float64(latency)/float64(time.Millisecond),
			clientIP,
			method,
			endpoint,
		)
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

	// GraphQL handler
	h := gqlhandler.New(&gqlhandler.Config{
		Schema: &graphql.Schema,
		Pretty: true,
	})

	// GraphQL UI + endpoint
	// Using "/graphql" makes the UI use the current host automatically.
	r.GET("/graphql", gin.WrapH(playground.ApolloSandboxHandler("GraphQL", "/graphql")))
	r.POST("/graphql", gin.WrapH(h))

	// Start server
	port := config.ServerPort
	addr := "0.0.0.0:" + port
	log.Printf("Server started on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
