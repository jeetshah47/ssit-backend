package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ETFBasketRepo defines the interface for ETF basket data persistence
type ETFBasketRepo interface {
	// Create creates a new ETF basket
	Create(ctx context.Context, basket *models.ETFBasket) error

	// FindByID finds an ETF basket by ID with items
	FindByID(ctx context.Context, id uuid.UUID) (*models.ETFBasket, error)

	// FindByStatus finds ETF baskets by status
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ETFBasket, error)

	// FindPublished finds published ETF baskets
	FindPublished(ctx context.Context, limit, offset int) ([]*models.ETFBasket, error)

	// Update updates an existing ETF basket
	Update(ctx context.Context, basket *models.ETFBasket) error

	// Delete deletes an ETF basket
	Delete(ctx context.Context, id uuid.UUID) error

	// CreateItem creates a new ETF basket item
	CreateItem(ctx context.Context, item *models.ETFBasketItem) error
}

// etfBasketRepo implements the ETF basket repository interface
type etfBasketRepo struct {
	*RepoContext
}

// NewETFBasketRepo creates a new ETF basket repository
func NewETFBasketRepo(ctx *RepoContext) ETFBasketRepo {
	return &etfBasketRepo{RepoContext: ctx}
}

// Create creates a new ETF basket
func (r *etfBasketRepo) Create(ctx context.Context, basket *models.ETFBasket) error {
	if err := r.db.WithContext(ctx).Create(basket).Error; err != nil {
		return fmt.Errorf("failed to create ETF basket: %w", err)
	}
	return nil
}

// FindByID finds an ETF basket by ID with items
func (r *etfBasketRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.ETFBasket, error) {
	var basket models.ETFBasket
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("AdvisoryType").
		Where("id = ?", id).
		First(&basket).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("ETF_BASKET_NOT_FOUND", "ETF basket not found")
		}
		return nil, fmt.Errorf("failed to find ETF basket: %w", err)
	}
	return &basket, nil
}

// FindByStatus finds ETF baskets by status
func (r *etfBasketRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ETFBasket, error) {
	var baskets []*models.ETFBasket
	query := r.db.WithContext(ctx).Preload("Items").Preload("AdvisoryType")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&baskets).Error; err != nil {
		return nil, fmt.Errorf("failed to find ETF baskets: %w", err)
	}
	return baskets, nil
}

// FindPublished finds published ETF baskets
func (r *etfBasketRepo) FindPublished(ctx context.Context, limit, offset int) ([]*models.ETFBasket, error) {
	return r.FindByStatus(ctx, "published", limit, offset)
}

// Update updates an existing ETF basket
func (r *etfBasketRepo) Update(ctx context.Context, basket *models.ETFBasket) error {
	if err := r.db.WithContext(ctx).Save(basket).Error; err != nil {
		return fmt.Errorf("failed to update ETF basket: %w", err)
	}
	return nil
}

// Delete deletes an ETF basket
func (r *etfBasketRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.ETFBasket{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete ETF basket: %w", err)
	}
	return nil
}

// CreateItem creates a new ETF basket item
func (r *etfBasketRepo) CreateItem(ctx context.Context, item *models.ETFBasketItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("failed to create ETF basket item: %w", err)
	}
	return nil
}
