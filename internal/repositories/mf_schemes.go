package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MFSchemeRepo defines the interface for MF scheme data persistence
type MFSchemeRepo interface {
	Create(ctx context.Context, scheme *models.MFScheme) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.MFScheme, error)
	FindAll(ctx context.Context, limit, offset int) ([]*models.MFScheme, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.MFScheme, error)
	Update(ctx context.Context, scheme *models.MFScheme) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// mfSchemeRepo implements the MF scheme repository interface
type mfSchemeRepo struct {
	*RepoContext
}

// NewMFSchemeRepo creates a new MF scheme repository
func NewMFSchemeRepo(ctx *RepoContext) MFSchemeRepo {
	return &mfSchemeRepo{RepoContext: ctx}
}

// Create creates a new MF scheme
func (r *mfSchemeRepo) Create(ctx context.Context, scheme *models.MFScheme) error {
	if err := r.db.WithContext(ctx).Create(scheme).Error; err != nil {
		return fmt.Errorf("failed to create MF scheme: %w", err)
	}
	return nil
}

// FindByID finds an MF scheme by ID
func (r *mfSchemeRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.MFScheme, error) {
	var scheme models.MFScheme
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&scheme).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("MF_SCHEME_NOT_FOUND", "MF scheme not found")
		}
		return nil, fmt.Errorf("failed to find MF scheme: %w", err)
	}
	return &scheme, nil
}

// FindAll finds all MF schemes
func (r *mfSchemeRepo) FindAll(ctx context.Context, limit, offset int) ([]*models.MFScheme, error) {
	var schemes []*models.MFScheme
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&schemes).Error; err != nil {
		return nil, fmt.Errorf("failed to find MF schemes: %w", err)
	}
	return schemes, nil
}

// FindByStatus finds MF schemes by status
func (r *mfSchemeRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.MFScheme, error) {
	var schemes []*models.MFScheme
	query := r.db.WithContext(ctx)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&schemes).Error; err != nil {
		return nil, fmt.Errorf("failed to find MF schemes: %w", err)
	}
	return schemes, nil
}

// Update updates an existing MF scheme
func (r *mfSchemeRepo) Update(ctx context.Context, scheme *models.MFScheme) error {
	if err := r.db.WithContext(ctx).Save(scheme).Error; err != nil {
		return fmt.Errorf("failed to update MF scheme: %w", err)
	}
	return nil
}

// Delete deletes an MF scheme
func (r *mfSchemeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.MFScheme{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete MF scheme: %w", err)
	}
	return nil
}
