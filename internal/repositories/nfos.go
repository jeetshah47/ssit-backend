package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NFORepo defines the interface for NFO data persistence
type NFORepo interface {
	Create(ctx context.Context, nfo *models.NFO) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.NFO, error)
	FindAll(ctx context.Context, limit, offset int) ([]*models.NFO, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.NFO, error)
	FindOpen(ctx context.Context) ([]*models.NFO, error)
	Update(ctx context.Context, nfo *models.NFO) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// nfoRepo implements the NFO repository interface
type nfoRepo struct {
	*RepoContext
}

// NewNFORepo creates a new NFO repository
func NewNFORepo(ctx *RepoContext) NFORepo {
	return &nfoRepo{RepoContext: ctx}
}

// Create creates a new NFO
func (r *nfoRepo) Create(ctx context.Context, nfo *models.NFO) error {
	if err := r.db.WithContext(ctx).Create(nfo).Error; err != nil {
		return fmt.Errorf("failed to create NFO: %w", err)
	}
	return nil
}

// FindByID finds an NFO by ID
func (r *nfoRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.NFO, error) {
	var nfo models.NFO
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&nfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("NFO_NOT_FOUND", "NFO not found")
		}
		return nil, fmt.Errorf("failed to find NFO: %w", err)
	}
	return &nfo, nil
}

// FindAll finds all NFOs
func (r *nfoRepo) FindAll(ctx context.Context, limit, offset int) ([]*models.NFO, error) {
	var nfos []*models.NFO
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&nfos).Error; err != nil {
		return nil, fmt.Errorf("failed to find NFOs: %w", err)
	}
	return nfos, nil
}

// FindByStatus finds NFOs by status
func (r *nfoRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.NFO, error) {
	var nfos []*models.NFO
	query := r.db.WithContext(ctx)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&nfos).Error; err != nil {
		return nil, fmt.Errorf("failed to find NFOs: %w", err)
	}
	return nfos, nil
}

// FindOpen finds currently open NFOs
func (r *nfoRepo) FindOpen(ctx context.Context) ([]*models.NFO, error) {
	var nfos []*models.NFO
	if err := r.db.WithContext(ctx).
		Where("status = ?", "open").
		Order("open_date ASC").
		Find(&nfos).Error; err != nil {
		return nil, fmt.Errorf("failed to find open NFOs: %w", err)
	}
	return nfos, nil
}

// Update updates an existing NFO
func (r *nfoRepo) Update(ctx context.Context, nfo *models.NFO) error {
	if err := r.db.WithContext(ctx).Save(nfo).Error; err != nil {
		return fmt.Errorf("failed to update NFO: %w", err)
	}
	return nil
}

// Delete deletes an NFO
func (r *nfoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.NFO{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete NFO: %w", err)
	}
	return nil
}
