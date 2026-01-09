package services

import (
	"gorm.io/gorm"
	"github.com/equitywala/backend/internal/common/logger"
)

// ServiceContext provides context for services
type ServiceContext struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewServiceContext creates a new service context
func NewServiceContext(db *gorm.DB, logger logger.Logger) *ServiceContext {
	return &ServiceContext{
		db:     db,
		logger: logger,
	}
}

// GetDb returns the database connection
func (s *ServiceContext) GetDb() *gorm.DB {
	return s.db
}

// GetLogger returns the logger
func (s *ServiceContext) GetLogger() logger.Logger {
	return s.logger
}

