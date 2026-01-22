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

	// FindBulletIdeas finds published stock baskets with bullet ideas
	FindBulletIdeas(ctx context.Context, limit int) ([]*models.StockBasket, error)

	// FindRecommendations finds stock basket items that are recommendations (not bullet ideas)
	FindRecommendations(ctx context.Context, limit int) ([]*models.StockBasketItem, error)

	// Item operations
	CreateItem(ctx context.Context, item *models.StockBasketItem) error
	FindItemByID(ctx context.Context, id uuid.UUID) (*models.StockBasketItem, error)
	UpdateItem(ctx context.Context, item *models.StockBasketItem) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
	FindByAdvisoryType(ctx context.Context, advisoryTypeID uuid.UUID, limit, offset int) ([]*models.StockBasket, error)
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

// FindBulletIdeas finds published stock baskets with bullet ideas
func (r *stockBasketRepo) FindBulletIdeas(ctx context.Context, limit int) ([]*models.StockBasket, error) {
	var baskets []*models.StockBasket
	if err := r.db.WithContext(ctx).
		Preload("Items", "is_bullet_idea = ?", true).
		Where("status = ? AND is_bullet_idea = ?", "published", true).
		Order("published_at DESC").
		Limit(limit).
		Find(&baskets).Error; err != nil {
		return nil, fmt.Errorf("failed to find bullet ideas: %w", err)
	}
	return baskets, nil
}

// FindRecommendations finds stock basket items that are recommendations (not bullet ideas)
func (r *stockBasketRepo) FindRecommendations(ctx context.Context, limit int) ([]*models.StockBasketItem, error) {
	var items []*models.StockBasketItem
	if err := r.db.WithContext(ctx).
		Joins("JOIN stock_baskets ON stock_basket_items.stock_basket_id = stock_baskets.id").
		Where("stock_baskets.status = ? AND stock_basket_items.is_bullet_idea = ?", "published", false).
		Order("stock_baskets.published_at DESC, stock_basket_items.display_order ASC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to find recommendations: %w", err)
	}
	return items, nil
}

// CreateItem creates a new stock basket item
func (r *stockBasketRepo) CreateItem(ctx context.Context, item *models.StockBasketItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("failed to create stock basket item: %w", err)
	}
	return nil
}

// FindItemByID finds a stock basket item by ID
func (r *stockBasketRepo) FindItemByID(ctx context.Context, id uuid.UUID) (*models.StockBasketItem, error) {
	var item models.StockBasketItem
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("STOCK_BASKET_ITEM_NOT_FOUND", "stock basket item not found")
		}
		return nil, fmt.Errorf("failed to find stock basket item: %w", err)
	}
	return &item, nil
}

// UpdateItem updates a stock basket item
func (r *stockBasketRepo) UpdateItem(ctx context.Context, item *models.StockBasketItem) error {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return fmt.Errorf("failed to update stock basket item: %w", err)
	}
	return nil
}

// DeleteItem deletes a stock basket item
func (r *stockBasketRepo) DeleteItem(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.StockBasketItem{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete stock basket item: %w", err)
	}
	return nil
}

// FindByAdvisoryType finds stock baskets by advisory type
func (r *stockBasketRepo) FindByAdvisoryType(ctx context.Context, advisoryTypeID uuid.UUID, limit, offset int) ([]*models.StockBasket, error) {
	var baskets []*models.StockBasket
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Where("advisory_type_id = ?", advisoryTypeID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&baskets).Error; err != nil {
		return nil, fmt.Errorf("failed to find stock baskets by advisory type: %w", err)
	}
	return baskets, nil
}
