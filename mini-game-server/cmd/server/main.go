package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/config"
	"draw-and-guess-server/internal/database"
	"draw-and-guess-server/internal/graph"
	"draw-and-guess-server/internal/routes"
	"draw-and-guess-server/internal/valkey"
	"draw-and-guess-server/pkg/dictionary"
	"draw-and-guess-server/pkg/utils"
)

func main() {
	logger := common.GetLogger()
	logger.Info("Starting Draw & Guess Game Server...")

	// Initialize config
	config.Init()

	// Initialize Snowflake ID generator (node ID 1, can be configured per instance)
	if err := utils.InitSnowflake(1); err != nil {
		logger.Error("Failed to initialize Snowflake ID generator: %v", err)
		os.Exit(1)
	}
	logger.Info("✓ Snowflake ID generator initialized")

	// Load Japanese word dictionary (required for wordchain game)
	dictPath := filepath.Join("data", "japanese_words_jmdict.txt")
	dict := dictionary.GetInstance()
	if err := dict.LoadFromFile(dictPath); err != nil {
		logger.Error("Failed to load word dictionary: %v", err)
		logger.Error("Dictionary is required for wordchain game - server cannot start")
		os.Exit(1)
	}
	logger.Info("✓ Word dictionary loaded (%d words)", dict.GetWordCount())

	// Connect to Valkey (optional)
	if err := valkey.Connect(); err != nil {
		logger.Warn("Valkey connection failed: %v (continuing without cache)", err)
	} else {
		logger.Info("✓ Connected to Valkey")
		defer func() {
			if err := valkey.Close(); err != nil {
				logger.Error("Failed to close Valkey connection: %v", err)
			}
		}()
	}

	// Connect to MongoDB (required)
	if err := database.Connect(); err != nil {
		logger.Error("Failed to connect to MongoDB: %v", err)
		os.Exit(1)
	}
	logger.Info("✓ Connected to MongoDB")
	defer func() {
		if err := database.Disconnect(); err != nil {
			logger.Error("Failed to disconnect MongoDB: %v", err)
		}
	}()

	// Initialize database schema (create indexes)
	if err := database.InitSchema(); err != nil {
		logger.Error("Failed to initialize database schema: %v", err)
		os.Exit(1)
	}
	logger.Info("✓ Database schema initialized")

	// Start room cleanup scheduler
	graph.StartRoomCleanupScheduler()
	logger.Info("✓ Room cleanup scheduler started")

	// Setup router
	r := routes.SetupRouter()

	// Configure server
	srv := &http.Server{
		Addr:           "0.0.0.0:" + config.ServerPort,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in goroutine
	go func() {
		logger.Info("🚀 Server started on %s", srv.Addr)
		logger.Info("GraphQL Playground: http://localhost:%s/graphql", config.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start: %v", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	gracefulShutdown(srv, logger)
}

func gracefulShutdown(srv *http.Server, logger *common.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}

	logger.Info("✓ Server exited gracefully")
	logger.Sync() // Flush any buffered log entries
}
