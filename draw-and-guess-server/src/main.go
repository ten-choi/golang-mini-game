package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
			"http://10.33.255.58:8080",
			"http://localhost:8080",
			"http://localhost:3000",
			"http://localhost:5174", // Vite dev server
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

	// Logging middleware (after CORS)
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

	// Quiz routes
	r.GET("/app/quiz/ox", handlers.GetOXQuiz)
	r.GET("/app/quiz/general", handlers.GetGeneralQuiz)
	r.GET("/app/quiz/:type", handlers.GetQuizByType)

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

	// GraphQL UI options
	r.GET("/graphql/playground", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql")))
	r.GET("/graphql/sandbox", gin.WrapH(playground.ApolloSandboxHandler("Apollo Sandbox", "/graphql")))

	// GraphQL endpoint
	r.POST("/graphql", gin.WrapH(h))
	r.GET("/graphql", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/graphql/sandbox")
	})

	// Start server with graceful shutdown
	port := config.ServerPort
	addr := "0.0.0.0:" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server started on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
