package application

import (
	"gorm.io/gorm"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/equitywala/backend/internal/core/config"
	"github.com/equitywala/backend/internal/common/logger"
)

// Application represents the application context
type Application interface {
	GetDb() *gorm.DB
	GetMongoDB() *mongo.Client
	GetConfig() *config.Config
	GetLogger() logger.Logger
}

// application implements the Application interface
type application struct {
	db     *gorm.DB
	mongoDB *mongo.Client
	config  *config.Config
	logger  logger.Logger
}

// NewApplication creates a new application instance
func NewApplication(
	db *gorm.DB,
	mongoDB *mongo.Client,
	cfg *config.Config,
	logger logger.Logger,
) Application {
	return &application{
		db:     db,
		mongoDB: mongoDB,
		config:  cfg,
		logger:  logger,
	}
}

// GetDb returns the database connection
func (a *application) GetDb() *gorm.DB {
	return a.db
}

// GetMongoDB returns the MongoDB connection
func (a *application) GetMongoDB() *mongo.Client {
	return a.mongoDB
}

// GetConfig returns the configuration
func (a *application) GetConfig() *config.Config {
	return a.config
}

// GetLogger returns the logger
func (a *application) GetLogger() logger.Logger {
	return a.logger
}

