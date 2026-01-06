package routes

import (
	"database/sql"

	"draw-and-guess-server/internal/graph"
	"draw-and-guess-server/internal/handlers"
	"draw-and-guess-server/internal/middleware"
	"draw-and-guess-server/internal/repository"
	"draw-and-guess-server/internal/service"
	"draw-and-guess-server/internal/websocket"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures all routes
func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.TraceID()) // Add trace ID for request tracking
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(middleware.ErrorHandler())

	// Health check (non-versioned)
	r.GET("/health", handlers.HealthCheck)

	// API v1 routes
	api := r.Group("/api/v1")
	{
		setupWebSocketRoutes(api)
		setupGraphQLRoutes(api, db)
	}

	return r
}

// setupWebSocketRoutes sets up WebSocket routes for real-time communication
func setupWebSocketRoutes(api *gin.RouterGroup) {
	ws := api.Group("/ws")
	{
		ws.GET("/lobby", websocket.HandleLobbyWebSocket)
		ws.GET("/rooms/:id", websocket.HandleRoomWebSocket)
	}
}

// setupGraphQLRoutes sets up GraphQL routes with gqlgen
func setupGraphQLRoutes(api *gin.RouterGroup, db *sql.DB) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	quizRepo := repository.NewQuizRepository(db)
	playerStatsRepo := repository.NewPlayerStatsRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	quizService := service.NewQuizService(quizRepo)
	playerStatsService := service.NewPlayerStatsService(playerStatsRepo)

	// Initialize resolver with services
	resolver := &graph.Resolver{
		UserService:        userService,
		QuizService:        quizService,
		PlayerStatsService: playerStatsService,
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// GraphQL endpoint
	api.POST("/graphql", gin.WrapH(srv))

	// Apollo Sandbox (GET) - Development environment
	api.GET("/graphql", gin.WrapH(playground.ApolloSandboxHandler("Apollo Sandbox", "/graphql")))
}
