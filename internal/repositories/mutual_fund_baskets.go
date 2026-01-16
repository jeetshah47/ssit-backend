package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MutualFundBasketRepo defines the interface for mutual fund basket data persistence
type MutualFundBasketRepo interface {
	// Create creates a new mutual fund basket
	Create(ctx context.Context, basket *models.MutualFundBasket) error

	// FindByID finds a mutual fund basket by ID with items
	FindByID(ctx context.Context, id uuid.UUID) (*models.MutualFundBasket, error)

	// FindByBasketType finds mutual fund baskets by basket type
	FindByBasketType(ctx context.Context, basketType string, limit, offset int) ([]*models.MutualFundBasket, error)

	// FindByStatus finds mutual fund baskets by status
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.MutualFundBasket, error)

	// FindPublished finds published mutual fund baskets
	FindPublished(ctx context.Context, limit, offset int) ([]*models.MutualFundBasket, error)

	// Update updates an existing mutual fund basket
	Update(ctx context.Context, basket *models.MutualFundBasket) error

	// Delete deletes a mutual fund basket
	Delete(ctx context.Context, id uuid.UUID) error
}

// mutualFundBasketRepo implements the mutual fund basket repository interface
type mutualFundBasketRepo struct {
	*RepoContext
}

// NewMutualFundBasketRepo creates a new mutual fund basket repository
func NewMutualFundBasketRepo(ctx *RepoContext) MutualFundBasketRepo {
	return &mutualFundBasketRepo{RepoContext: ctx}
}

// Create creates a new mutual fund basket
func (r *mutualFundBasketRepo) Create(ctx context.Context, basket *models.MutualFundBasket) error {
	if err := r.db.WithContext(ctx).Create(basket).Error; err != nil {
		return fmt.Errorf("failed to create mutual fund basket: %w", err)
	}
	return nil
}

// FindByID finds a mutual fund basket by ID with items
func (r *mutualFundBasketRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.MutualFundBasket, error) {
	var basket models.MutualFundBasket
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("AdvisoryType").
		Where("id = ?", id).
		First(&basket).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("MUTUAL_FUND_BASKET_NOT_FOUND", "mutual fund basket not found")
		}
		return nil, fmt.Errorf("failed to find mutual fund basket: %w", err)
	}
	return &basket, nil
}

// FindByBasketType finds mutual fund baskets by basket type
func (r *mutualFundBasketRepo) FindByBasketType(ctx context.Context, basketType string, limit, offset int) ([]*models.MutualFundBasket, error) {
	var baskets []*models.MutualFundBasket
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("AdvisoryType").
		Where("basket_type = ? AND status = ?", basketType, "published").
		Order("published_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&baskets).Error; err != nil {
		return nil, fmt.Errorf("failed to find mutual fund baskets: %w", err)
	}
	return baskets, nil
}

// FindByStatus finds mutual fund baskets by status
func (r *mutualFundBasketRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.MutualFundBasket, error) {
	var baskets []*models.MutualFundBasket
	query := r.db.WithContext(ctx).Preload("Items").Preload("AdvisoryType")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&baskets).Error; err != nil {
		return nil, fmt.Errorf("failed to find mutual fund baskets: %w", err)
	}
	return baskets, nil
}

// FindPublished finds published mutual fund baskets
func (r *mutualFundBasketRepo) FindPublished(ctx context.Context, limit, offset int) ([]*models.MutualFundBasket, error) {
	return r.FindByStatus(ctx, "published", limit, offset)
}

// Update updates an existing mutual fund basket
func (r *mutualFundBasketRepo) Update(ctx context.Context, basket *models.MutualFundBasket) error {
	if err := r.db.WithContext(ctx).Save(basket).Error; err != nil {
		return fmt.Errorf("failed to update mutual fund basket: %w", err)
	}
	return nil
}

// Delete deletes a mutual fund basket
func (r *mutualFundBasketRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.MutualFundBasket{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete mutual fund basket: %w", err)
	}
	return nil
}
