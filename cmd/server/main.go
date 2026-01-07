package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/equitywala/backend/internal/config"
	"github.com/equitywala/backend/internal/infrastructure/database/mongodb"
	postgres "github.com/equitywala/backend/internal/infrastructure/database/postgres"
	"github.com/equitywala/backend/internal/interfaces/http/middleware"
	"github.com/equitywala/backend/internal/interfaces/http/routes"
	"github.com/equitywala/backend/internal/shared/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Starting Equitywala Stock Advisory Platform Backend")

	// Initialize PostgreSQL database
	postgresDB, err := postgres.NewConnection(cfg.Database.DSN())
	if err != nil {
		appLogger.Fatal("Failed to connect to PostgreSQL", "error", err)
	}
	appLogger.Info("Connected to PostgreSQL database")

	// Initialize MongoDB
	mongoClient, err := mongodb.NewConnection(cfg.MongoDB.URI)
	if err != nil {
		appLogger.Fatal("Failed to connect to MongoDB", "error", err)
	}
	appLogger.Info("Connected to MongoDB")

	// Set up Gin router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Apply middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Apply CORS middleware
	router.Use(middleware.CORSMiddleware(cfg))

	// Set up routes
	deps := &routes.Dependencies{
		PostgresDB: postgresDB,
		MongoDB:    mongoClient,
		Config:     cfg,
		Logger:     appLogger,
	}
	routes.SetupRoutes(router, deps)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info("Server starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", "error", err)
	}

	// Close database connections
	if err := postgres.CloseDB(postgresDB); err != nil {
		appLogger.Error("Error closing PostgreSQL connection", "error", err)
	}

	if err := mongoClient.Disconnect(ctx); err != nil {
		appLogger.Error("Error closing MongoDB connection", "error", err)
	}

	appLogger.Info("Server exited")
}
