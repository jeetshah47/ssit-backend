package repositories

import "gorm.io/gorm"

// RepoContext provides database context for repositories
type RepoContext struct {
	db *gorm.DB
}

// NewRepoContext creates a new repository context
func NewRepoContext(db *gorm.DB) *RepoContext {
	return &RepoContext{db: db}
}

// GetDb returns the database connection
func (r *RepoContext) GetDb() *gorm.DB {
	return r.db
}

