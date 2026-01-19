package routes

import (
	"draw-and-guess-server/internal/database"
	"draw-and-guess-server/internal/graph"
	"draw-and-guess-server/internal/handlers"
	"draw-and-guess-server/internal/middleware"
	"draw-and-guess-server/internal/repository"
	"draw-and-guess-server/internal/service"
	"draw-and-guess-server/internal/websocket"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures all routes
func SetupRouter() *gin.Engine {
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.TraceID()) // Add trace ID for request tracking
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(middleware.ErrorHandler())

	// Health check (non-versioned)
	r.GET("/health", handlers.HealthCheck)

	// API routes
	api := r.Group("")
	{
		setupWebSocketRoutes(api)
		setupGraphQLRoutes(api)
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
func setupGraphQLRoutes(api *gin.RouterGroup) {
	// Initialize repositories
	usersCollection := database.GetCollection("users")
	wordsCollection := database.GetCollection("korean_words")
	oxQuizzesCollection := database.GetCollection("ox_quizzes")
	qaQuizzesCollection := database.GetCollection("qa_quizzes")
	userStatsCollection := database.GetCollection("user_stats")

	userRepo := repository.NewUserRepository(usersCollection)
	quizRepo := repository.NewQuizRepository(wordsCollection, oxQuizzesCollection, qaQuizzesCollection)
	userStatsRepo := repository.NewUserStatsRepository(userStatsCollection)

	// Initialize services
	userService := service.NewUserService(userRepo)
	quizService := service.NewQuizService(quizRepo)
	userStatsService := service.NewUserStatsService(userStatsRepo)

	// Initialize resolver with services
	resolver := &graph.Resolver{
		UserService:      userService,
		QuizService:      quizService,
		UserStatsService: userStatsService,
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// Add transports for GraphQL
	srv.AddTransport(transport.POST{})    // POST requests for mutations/queries
	srv.AddTransport(transport.GET{})     // GET requests for queries
	srv.AddTransport(transport.Websocket{ // WebSocket for subscriptions
		KeepAlivePingInterval: 10 * time.Second,
	})

	// GraphQL endpoint - POST for queries/mutations, WebSocket handled separately
	api.POST("/graphql", gin.WrapH(srv))

	// GET /graphql shows Apollo Sandbox UI (best interface)
	api.GET("/graphql", func(c *gin.Context) {
		// Check if it's a WebSocket upgrade request
		if c.GetHeader("Upgrade") == "websocket" {
			gin.WrapH(srv)(c)
			return
		}
		// Otherwise show Apollo Sandbox UI
		playground.ApolloSandboxHandler("Apollo Sandbox", "/graphql")(c.Writer, c.Request)
	})

	// Alternative Playground UIs
	// api.GET("/playground", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql")))
	// api.GET("/sandbox", gin.WrapH(playground.ApolloSandboxHandler("Apollo Sandbox", "/graphql")))
}
