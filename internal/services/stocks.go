package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// CreateStockCmd represents the command for creating a stock
type CreateStockCmd struct {
	Name     string
	Symbol   string
	Exchange string
	Sector   *string
}

// UpdateStockCmd represents the command for updating a stock
type UpdateStockCmd struct {
	ID       uuid.UUID
	Name     string
	Symbol   string
	Exchange string
	Sector   *string
	IsActive bool
}

// StockService handles stock operations
type StockService struct {
	stockRepo repositories.StockRepo
}

// NewStockService creates a new stock service
func NewStockService(stockRepo repositories.StockRepo) *StockService {
	return &StockService{
		stockRepo: stockRepo,
	}
}

// Create executes the CreateStock command
func (s *StockService) Create(ctx context.Context, cmd CreateStockCmd) (*models.Stock, error) {
	// Validate
	if cmd.Name == "" || cmd.Symbol == "" || cmd.Exchange == "" {
		return nil, errors.NewDomainError("INVALID_INPUT", "Name, Symbol and Exchange are required")
	}

	// Normalize
	cmd.Symbol = strings.ToUpper(cmd.Symbol)
	cmd.Exchange = strings.ToUpper(cmd.Exchange)

	// Check existence
	exists, err := s.stockRepo.FindBySymbolAndExchange(ctx, cmd.Symbol, cmd.Exchange)
	if err == nil && exists != nil {
		return nil, errors.NewDomainError("STOCK_ALREADY_EXISTS", "Stock with this symbol and exchange already exists")
	}

	stock := &models.Stock{
		Name:     cmd.Name,
		Symbol:   cmd.Symbol,
		Exchange: cmd.Exchange,
		Sector:   cmd.Sector,
		IsActive: true,
	}

	if err := s.stockRepo.Create(ctx, stock); err != nil {
		return nil, fmt.Errorf("failed to create stock: %w", err)
	}

	return stock, nil
}

// GetAll returns all stocks
func (s *StockService) GetAll(ctx context.Context) ([]models.Stock, error) {
	return s.stockRepo.FindAll(ctx)
}

// GetByID returns a stock by ID
func (s *StockService) GetByID(ctx context.Context, id uuid.UUID) (*models.Stock, error) {
	return s.stockRepo.FindByID(ctx, id)
}

// Update updates a stock
func (s *StockService) Update(ctx context.Context, cmd UpdateStockCmd) (*models.Stock, error) {
	stock, err := s.stockRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	// Normalize
	cmd.Symbol = strings.ToUpper(cmd.Symbol)
	cmd.Exchange = strings.ToUpper(cmd.Exchange)

	// Check if updating symbol/exchange conflicts with another stock
	if stock.Symbol != cmd.Symbol || stock.Exchange != cmd.Exchange {
		existing, err := s.stockRepo.FindBySymbolAndExchange(ctx, cmd.Symbol, cmd.Exchange)
		if err == nil && existing != nil && existing.ID != stock.ID {
			return nil, errors.NewDomainError("STOCK_ALREADY_EXISTS", "Stock with this symbol and exchange already exists")
		}
	}

	stock.Name = cmd.Name
	stock.Symbol = cmd.Symbol
	stock.Exchange = cmd.Exchange
	stock.Sector = cmd.Sector
	stock.IsActive = cmd.IsActive

	if err := s.stockRepo.Update(ctx, stock); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	return stock, nil
}

// Delete deletes a stock
func (s *StockService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.stockRepo.Delete(ctx, id)
}

