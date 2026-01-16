package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockBasketRepo defines the interface for stock basket data persistence
type StockBasketRepo interface {
	// Create creates a new stock basket
	Create(ctx context.Context, basket *models.StockBasket) error

	// FindByID finds a stock basket by ID with items
	FindByID(ctx context.Context, id uuid.UUID) (*models.StockBasket, error)

	// FindByStatus finds stock baskets by status
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.StockBasket, error)

	// FindPublished finds published stock baskets
	FindPublished(ctx context.Context, limit, offset int) ([]*models.StockBasket, error)

	// Update updates an existing stock basket
	Update(ctx context.Context, basket *models.StockBasket) error

	// Delete deletes a stock basket
	Delete(ctx context.Context, id uuid.UUID) error

	// FindLatestPublished finds the latest published basket
	FindLatestPublished(ctx context.Context) (*models.StockBasket, error)
}

// stockBasketRepo implements the stock basket repository interface
type stockBasketRepo struct {
	*RepoContext
}

// NewStockBasketRepo creates a new stock basket repository
func NewStockBasketRepo(ctx *RepoContext) StockBasketRepo {
	return &stockBasketRepo{RepoContext: ctx}
}

// Create creates a new stock basket
func (r *stockBasketRepo) Create(ctx context.Context, basket *models.StockBasket) error {
	if err := r.db.WithContext(ctx).Create(basket).Error; err != nil {
		return fmt.Errorf("failed to create stock basket: %w", err)
	}
	return nil
}

// FindByID finds a stock basket by ID with items
func (r *stockBasketRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.StockBasket, error) {
	var basket models.StockBasket
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("AdvisoryType").
		Where("id = ?", id).
		First(&basket).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("STOCK_BASKET_NOT_FOUND", "stock basket not found")
		}
		return nil, fmt.Errorf("failed to find stock basket: %w", err)
	}
	return &basket, nil
}

// FindByStatus finds stock baskets by status
func (r *stockBasketRepo) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*models.StockBasket, error) {
	var baskets []*models.StockBasket
	query := r.db.WithContext(ctx).Preload("Items").Preload("AdvisoryType")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&baskets).Error; err != nil {
		return nil, fmt.Errorf("failed to find stock baskets: %w", err)
	}
	return baskets, nil
}

// FindPublished finds published stock baskets
func (r *stockBasketRepo) FindPublished(ctx context.Context, limit, offset int) ([]*models.StockBasket, error) {
	return r.FindByStatus(ctx, "published", limit, offset)
}

// Update updates an existing stock basket
func (r *stockBasketRepo) Update(ctx context.Context, basket *models.StockBasket) error {
	if err := r.db.WithContext(ctx).Save(basket).Error; err != nil {
		return fmt.Errorf("failed to update stock basket: %w", err)
	}
	return nil
}

// Delete deletes a stock basket
func (r *stockBasketRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.StockBasket{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete stock basket: %w", err)
	}
	return nil
}

// FindLatestPublished finds the latest published basket
func (r *stockBasketRepo) FindLatestPublished(ctx context.Context) (*models.StockBasket, error) {
	var basket models.StockBasket
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Where("status = ?", "published").
		Order("published_at DESC").
		First(&basket).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("STOCK_BASKET_NOT_FOUND", "no published stock basket found")
		}
		return nil, fmt.Errorf("failed to find latest published stock basket: %w", err)
	}
	return &basket, nil
}
