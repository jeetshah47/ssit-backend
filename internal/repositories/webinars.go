package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebinarRepo defines the interface for webinar data persistence
type WebinarRepo interface {
	Create(ctx context.Context, webinar *models.Webinar) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Webinar, error)
	FindAll(ctx context.Context, limit, offset int) ([]*models.Webinar, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Webinar, error)
	Update(ctx context.Context, webinar *models.Webinar) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// webinarRepo implements the webinar repository interface
type webinarRepo struct {
	*RepoContext
}

// NewWebinarRepo creates a new webinar repository
func NewWebinarRepo(ctx *RepoContext) WebinarRepo {
	return &webinarRepo{RepoContext: ctx}
}

// Create creates a new webinar
func (r *webinarRepo) Create(ctx context.Context, webinar *models.Webinar) error {
	if err := r.db.WithContext(ctx).Create(webinar).Error; err != nil {
		return fmt.Errorf("failed to create webinar: %w", err)
	}
	return nil
}

// FindByID finds a webinar by ID
func (r *webinarRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Webinar, error) {
	var webinar models.Webinar
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&webinar).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("WEBINAR_NOT_FOUND", "webinar not found")
		}
		return nil, fmt.Errorf("failed to find webinar: %w", err)
	}
	return &webinar, nil
}

// FindAll finds all webinars
func (r *webinarRepo) FindAll(ctx context.Context, limit, offset int) ([]*models.Webinar, error) {
	var webinars []*models.Webinar
	if err := r.db.WithContext(ctx).
		Order("date DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&webinars).Error; err != nil {
		return nil, fmt.Errorf("failed to find webinars: %w", err)
	}
	return webinars, nil
}

// FindByStatus finds webinars by status
func (r *webinarRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Webinar, error) {
	var webinars []*models.Webinar
	query := r.db.WithContext(ctx)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("date DESC, created_at DESC").Limit(limit).Offset(offset).Find(&webinars).Error; err != nil {
		return nil, fmt.Errorf("failed to find webinars: %w", err)
	}
	return webinars, nil
}

// Update updates an existing webinar
func (r *webinarRepo) Update(ctx context.Context, webinar *models.Webinar) error {
	if err := r.db.WithContext(ctx).Save(webinar).Error; err != nil {
		return fmt.Errorf("failed to update webinar: %w", err)
	}
	return nil
}

// Delete deletes a webinar
func (r *webinarRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Webinar{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete webinar: %w", err)
	}
	return nil
}
