package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IPOAdvisoryRepo defines the interface for IPO advisory data persistence
type IPOAdvisoryRepo interface {
	// Create creates a new IPO advisory
	Create(ctx context.Context, advisory *models.IPOAdvisory) error

	// FindByID finds an IPO advisory by ID
	FindByID(ctx context.Context, id uuid.UUID) (*models.IPOAdvisory, error)

	// FindByStatus finds IPO advisories by status
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.IPOAdvisory, error)

	// FindPublished finds published IPO advisories
	FindPublished(ctx context.Context, limit, offset int) ([]*models.IPOAdvisory, error)

	// Update updates an existing IPO advisory
	Update(ctx context.Context, advisory *models.IPOAdvisory) error

	// Delete deletes an IPO advisory
	Delete(ctx context.Context, id uuid.UUID) error
}

// ipoAdvisoryRepo implements the IPO advisory repository interface
type ipoAdvisoryRepo struct {
	*RepoContext
}

// NewIPOAdvisoryRepo creates a new IPO advisory repository
func NewIPOAdvisoryRepo(ctx *RepoContext) IPOAdvisoryRepo {
	return &ipoAdvisoryRepo{RepoContext: ctx}
}

// Create creates a new IPO advisory
func (r *ipoAdvisoryRepo) Create(ctx context.Context, advisory *models.IPOAdvisory) error {
	if err := r.db.WithContext(ctx).Create(advisory).Error; err != nil {
		return fmt.Errorf("failed to create IPO advisory: %w", err)
	}
	return nil
}

// FindByID finds an IPO advisory by ID
func (r *ipoAdvisoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.IPOAdvisory, error) {
	var advisory models.IPOAdvisory
	if err := r.db.WithContext(ctx).
		Preload("AdvisoryType").
		Where("id = ?", id).
		First(&advisory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("IPO_ADVISORY_NOT_FOUND", "IPO advisory not found")
		}
		return nil, fmt.Errorf("failed to find IPO advisory: %w", err)
	}
	return &advisory, nil
}

// FindByStatus finds IPO advisories by status
func (r *ipoAdvisoryRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.IPOAdvisory, error) {
	var advisories []*models.IPOAdvisory
	query := r.db.WithContext(ctx).Preload("AdvisoryType")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&advisories).Error; err != nil {
		return nil, fmt.Errorf("failed to find IPO advisories: %w", err)
	}
	return advisories, nil
}

// FindPublished finds published IPO advisories
func (r *ipoAdvisoryRepo) FindPublished(ctx context.Context, limit, offset int) ([]*models.IPOAdvisory, error) {
	return r.FindByStatus(ctx, "listed", limit, offset)
}

// Update updates an existing IPO advisory
func (r *ipoAdvisoryRepo) Update(ctx context.Context, advisory *models.IPOAdvisory) error {
	if err := r.db.WithContext(ctx).Save(advisory).Error; err != nil {
		return fmt.Errorf("failed to update IPO advisory: %w", err)
	}
	return nil
}

// Delete deletes an IPO advisory
func (r *ipoAdvisoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.IPOAdvisory{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete IPO advisory: %w", err)
	}
	return nil
}
