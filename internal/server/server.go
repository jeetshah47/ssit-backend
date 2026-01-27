package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/equitywala/backend/internal/core/config"
	"github.com/equitywala/backend/internal/common/logger"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	logger     logger.Logger
}

// NewServer creates a new server instance
func NewServer(router *gin.Engine, cfg *config.Config, appLogger logger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		logger: appLogger,
	}
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Server starting", "port", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	return nil
}

// StartAsync starts the server in a goroutine
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			s.logger.Fatal("Failed to start server", "error", err)
		}
	}()
}

