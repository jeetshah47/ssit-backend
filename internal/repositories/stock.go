package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockRepo defines the interface for the master stock list
type StockRepo interface {
	List(ctx context.Context, limit, offset int) ([]*models.Stock, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Stock, error)
	FindByName(ctx context.Context, name string) (*models.Stock, error)
	Create(ctx context.Context, stock *models.Stock) error
	Update(ctx context.Context, stock *models.Stock) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type stockRepo struct {
	*RepoContext
}

// NewStockRepo creates a new stock repository
func NewStockRepo(ctx *RepoContext) StockRepo {
	return &stockRepo{RepoContext: ctx}
}

func (r *stockRepo) List(ctx context.Context, limit, offset int) ([]*models.Stock, error) {
	var list []*models.Stock
	res := r.db.WithContext(ctx).Order("name ASC").Limit(limit).Offset(offset).Find(&list)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to list stocks: %w", res.Error)
	}
	return list, nil
}

func (r *stockRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Stock, error) {
	var s models.Stock
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("STOCK_NOT_FOUND", "stock not found")
		}
		return nil, fmt.Errorf("failed to find stock: %w", err)
	}
	return &s, nil
}

func (r *stockRepo) FindByName(ctx context.Context, name string) (*models.Stock, error) {
	var s models.Stock
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("STOCK_NOT_FOUND", "stock not found")
		}
		return nil, fmt.Errorf("failed to find stock: %w", err)
	}
	return &s, nil
}

func (r *stockRepo) Create(ctx context.Context, stock *models.Stock) error {
	if err := r.db.WithContext(ctx).Create(stock).Error; err != nil {
		return fmt.Errorf("failed to create stock: %w", err)
	}
	return nil
}

func (r *stockRepo) Update(ctx context.Context, stock *models.Stock) error {
	if err := r.db.WithContext(ctx).Save(stock).Error; err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}
	return nil
}

func (r *stockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Stock{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete stock: %w", err)
	}
	return nil
}
