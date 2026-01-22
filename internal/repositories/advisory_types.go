package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"gorm.io/gorm"
)

// AdvisoryTypeRepo defines the interface for advisory type data persistence
type AdvisoryTypeRepo interface {
	// FindByName finds an advisory type by name
	FindByName(ctx context.Context, name string) (*models.AdvisoryType, error)
	// Create creates a new advisory type
	Create(ctx context.Context, advisoryType *models.AdvisoryType) error
}

// advisoryTypeRepo implements the advisory type repository interface
type advisoryTypeRepo struct {
	*RepoContext
}

// NewAdvisoryTypeRepo creates a new advisory type repository
func NewAdvisoryTypeRepo(ctx *RepoContext) AdvisoryTypeRepo {
	return &advisoryTypeRepo{RepoContext: ctx}
}

// FindByName finds an advisory type by name
func (r *advisoryTypeRepo) FindByName(ctx context.Context, name string) (*models.AdvisoryType, error) {
	var advisoryType models.AdvisoryType
	if err := r.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&advisoryType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("ADVISORY_TYPE_NOT_FOUND", "advisory type not found")
		}
		return nil, fmt.Errorf("failed to find advisory type: %w", err)
	}
	return &advisoryType, nil
}

// Create creates a new advisory type
func (r *advisoryTypeRepo) Create(ctx context.Context, advisoryType *models.AdvisoryType) error {
	if err := r.db.WithContext(ctx).Create(advisoryType).Error; err != nil {
		return fmt.Errorf("failed to create advisory type: %w", err)
	}
	return nil
}
