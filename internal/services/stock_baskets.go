package services

import (
	"context"
	"fmt"
	"time"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// CreateStockBasketCmd represents the command for creating a stock basket
type CreateStockBasketCmd struct {
	Name         string
	Description  *string
	IsBulletIdea bool
	Items        []*CreateStockBasketItemCmd
	PublishedBy  uuid.UUID
}

// CreateStockBasketItemCmd represents a stock basket item in create command
type CreateStockBasketItemCmd struct {
	StockName       string
	StockSymbol     *string
	CMP             float64
	Target          float64
	StopLoss        *float64
	EntryRangeMin   *float64
	EntryRangeMax   *float64
	Action          *string
	RiskLevel       *string
	TimeHorizon     *string
	Rationale       *string
	ReportURL       *string
	FundamentalsURL *string
	DisplayOrder    int
	IsBulletIdea    bool
}

// CreateStockBasketCmdOutputData represents the output
type CreateStockBasketCmdOutputData struct {
	Basket *models.StockBasket
}

// CreateStockBasketService handles stock basket creation
type CreateStockBasketService struct {
	stockBasketRepo  repositories.StockBasketRepo
	advisoryTypeRepo repositories.AdvisoryTypeRepo
}

// NewCreateStockBasketService creates a new create stock basket service
func NewCreateStockBasketService(
	stockBasketRepo repositories.StockBasketRepo,
	advisoryTypeRepo repositories.AdvisoryTypeRepo,
) *CreateStockBasketService {
	return &CreateStockBasketService{
		stockBasketRepo:  stockBasketRepo,
		advisoryTypeRepo: advisoryTypeRepo,
	}
}

// Execute executes the CreateStockBasket command
func (s *CreateStockBasketService) Execute(ctx context.Context, cmd CreateStockBasketCmd) (*CreateStockBasketCmdOutputData, error) {
	// Validate command
	if err := s.validateCommand(cmd); err != nil {
		return nil, err
	}

	// Get stock_basket advisory type
	advisoryType, err := s.advisoryTypeRepo.FindByName(ctx, "stock_basket")
	if err != nil {
		return nil, fmt.Errorf("failed to find advisory type: %w", err)
	}
	advisoryTypeID := advisoryType.ID

	now := time.Now()
	basket := &models.StockBasket{
		ID:             uuid.New(),
		AdvisoryTypeID: advisoryTypeID,
		Name:           cmd.Name,
		Description:    cmd.Description,
		IsBulletIdea:   cmd.IsBulletIdea,
		Status:         "draft",
		PublishedBy:    &cmd.PublishedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Create items
	items := make([]*models.StockBasketItem, 0, len(cmd.Items))
	for _, itemCmd := range cmd.Items {
		item := &models.StockBasketItem{
			ID:              uuid.New(),
			StockBasketID:   basket.ID,
			StockName:       itemCmd.StockName,
			StockSymbol:     itemCmd.StockSymbol,
			CMP:             itemCmd.CMP,
			Target:          itemCmd.Target,
			StopLoss:        itemCmd.StopLoss,
			EntryRangeMin:   itemCmd.EntryRangeMin,
			EntryRangeMax:   itemCmd.EntryRangeMax,
			Action:          itemCmd.Action,
			RiskLevel:       itemCmd.RiskLevel,
			TimeHorizon:     itemCmd.TimeHorizon,
			Rationale:       itemCmd.Rationale,
			ReportURL:       itemCmd.ReportURL,
			FundamentalsURL: itemCmd.FundamentalsURL,
			DisplayOrder:    itemCmd.DisplayOrder,
			IsBulletIdea:    itemCmd.IsBulletIdea,
			Status:          "active",
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		items = append(items, item)
	}
	// Convert to slice of models (not pointers) for GORM
	basketItems := make([]models.StockBasketItem, 0, len(items))
	for _, item := range items {
		basketItems = append(basketItems, *item)
	}
	basket.Items = basketItems

	// Persist via repository
	if err := s.stockBasketRepo.Create(ctx, basket); err != nil {
		return nil, fmt.Errorf("failed to create stock basket: %w", err)
	}

	return &CreateStockBasketCmdOutputData{Basket: basket}, nil
}

// validateCommand validates the command input
func (s *CreateStockBasketService) validateCommand(cmd CreateStockBasketCmd) error {
	if cmd.Name == "" {
		return errors.NewDomainError("INVALID_INPUT", "name is required")
	}
	if len(cmd.Items) == 0 {
		return errors.NewDomainError("INVALID_INPUT", "at least one item is required")
	}
	return nil
}

// GetStockBasketQuery represents the query for getting a stock basket
type GetStockBasketQuery struct {
	BasketID uuid.UUID
}

// GetStockBasketResult represents the output
type GetStockBasketResult struct {
	Basket *models.StockBasket
}

// GetStockBasketService handles stock basket retrieval
type GetStockBasketService struct {
	stockBasketRepo repositories.StockBasketRepo
}

// NewGetStockBasketService creates a new get stock basket service
func NewGetStockBasketService(stockBasketRepo repositories.StockBasketRepo) *GetStockBasketService {
	return &GetStockBasketService{
		stockBasketRepo: stockBasketRepo,
	}
}

// Execute executes the GetStockBasket query
func (s *GetStockBasketService) Execute(ctx context.Context, query GetStockBasketQuery) (*GetStockBasketResult, error) {
	basket, err := s.stockBasketRepo.FindByID(ctx, query.BasketID)
	if err != nil {
		return nil, err
	}

	return &GetStockBasketResult{Basket: basket}, nil
}

// ListStockBasketsQuery represents the query for listing stock baskets
type ListStockBasketsQuery struct {
	Status string
	Limit  int
	Offset int
}

// ListStockBasketsResult represents the output
type ListStockBasketsResult struct {
	Baskets []*models.StockBasket
}

// ListStockBasketsService handles stock basket listing
type ListStockBasketsService struct {
	stockBasketRepo repositories.StockBasketRepo
}

// NewListStockBasketsService creates a new list stock baskets service
func NewListStockBasketsService(stockBasketRepo repositories.StockBasketRepo) *ListStockBasketsService {
	return &ListStockBasketsService{
		stockBasketRepo: stockBasketRepo,
	}
}

// Execute executes the ListStockBaskets query
func (s *ListStockBasketsService) Execute(ctx context.Context, query ListStockBasketsQuery) (*ListStockBasketsResult, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	baskets, err := s.stockBasketRepo.FindByStatus(ctx, query.Status, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list stock baskets: %w", err)
	}

	return &ListStockBasketsResult{Baskets: baskets}, nil
}

// UpdateStockBasketCmd represents the command for updating a stock basket
type UpdateStockBasketCmd struct {
	BasketID    uuid.UUID
	Name        *string
	Description *string
	Status      *string
}

// UpdateStockBasketCmdOutputData represents the output
type UpdateStockBasketCmdOutputData struct {
	Basket *models.StockBasket
}

// UpdateStockBasketService handles stock basket updates
type UpdateStockBasketService struct {
	stockBasketRepo repositories.StockBasketRepo
}

// NewUpdateStockBasketService creates a new update stock basket service
func NewUpdateStockBasketService(stockBasketRepo repositories.StockBasketRepo) *UpdateStockBasketService {
	return &UpdateStockBasketService{
		stockBasketRepo: stockBasketRepo,
	}
}

// Execute executes the UpdateStockBasket command
func (s *UpdateStockBasketService) Execute(ctx context.Context, cmd UpdateStockBasketCmd) (*UpdateStockBasketCmdOutputData, error) {
	// Get existing basket
	basket, err := s.stockBasketRepo.FindByID(ctx, cmd.BasketID)
	if err != nil {
		return nil, err
	}

	// Check if published - core fields immutable
	if basket.IsPublished() {
		// Only allow status updates for published baskets
		if cmd.Status != nil {
			basket.Status = *cmd.Status
		}
	} else {
		// Allow all updates for draft baskets
		if cmd.Name != nil {
			basket.Name = *cmd.Name
		}
		if cmd.Description != nil {
			basket.Description = cmd.Description
		}
		if cmd.Status != nil {
			basket.Status = *cmd.Status
			if *cmd.Status == "published" {
				now := time.Now()
				basket.PublishedAt = &now
			}
		}
	}

	// Update
	if err := s.stockBasketRepo.Update(ctx, basket); err != nil {
		return nil, fmt.Errorf("failed to update stock basket: %w", err)
	}

	return &UpdateStockBasketCmdOutputData{Basket: basket}, nil
}
