package repositories

import (
	"context"

	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockRepo defines the interface for stock repository
type StockRepo interface {
	Create(ctx context.Context, stock *models.Stock) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Stock, error)
	FindAll(ctx context.Context) ([]models.Stock, error)
	Update(ctx context.Context, stock *models.Stock) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindBySymbolAndExchange(ctx context.Context, symbol, exchange string) (*models.Stock, error)
}

// stockRepo implements StockRepo
type stockRepo struct {
	db *gorm.DB
}

// NewStockRepo creates a new stock repository
func NewStockRepo(db *gorm.DB) StockRepo {
	return &stockRepo{
		db: db,
	}
}

// Create creates a new stock
func (r *stockRepo) Create(ctx context.Context, stock *models.Stock) error {
	return r.db.WithContext(ctx).Create(stock).Error
}

// FindByID finds a stock by ID
func (r *stockRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Stock, error) {
	var stock models.Stock
	if err := r.db.WithContext(ctx).First(&stock, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &stock, nil
}

// FindAll finds all stocks
func (r *stockRepo) FindAll(ctx context.Context) ([]models.Stock, error) {
	var stocks []models.Stock
	if err := r.db.WithContext(ctx).Order("name asc").Find(&stocks).Error; err != nil {
		return nil, err
	}
	return stocks, nil
}

// Update updates a stock
func (r *stockRepo) Update(ctx context.Context, stock *models.Stock) error {
	return r.db.WithContext(ctx).Save(stock).Error
}

// Delete deletes a stock
func (r *stockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Stock{}, "id = ?", id).Error
}

// FindBySymbolAndExchange finds a stock by symbol and exchange
func (r *stockRepo) FindBySymbolAndExchange(ctx context.Context, symbol, exchange string) (*models.Stock, error) {
	var stock models.Stock
	if err := r.db.WithContext(ctx).Where("symbol = ? AND exchange = ?", symbol, exchange).First(&stock).Error; err != nil {
		return nil, err
	}
	return &stock, nil
}

