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

// CreateMutualFundBasketCmd represents the command for creating a mutual fund basket
type CreateMutualFundBasketCmd struct {
	BasketType  string // general, index_fund, sectoral, nfo_review
	Name        string
	Description *string
	MaxSchemes  int
	Items       []*CreateMutualFundBasketItemCmd
	PublishedBy uuid.UUID
}

// CreateMutualFundBasketItemCmd represents a mutual fund basket item in create command
type CreateMutualFundBasketItemCmd struct {
	SchemeName   string
	SchemeCode   *string
	EntryPrice   float64
	ExitPrice    *float64
	CurrentNAV   *float64
	Trend        *string // positive, negative, neutral
	DisplayOrder int
}

// CreateMutualFundBasketCmdOutputData represents the output
type CreateMutualFundBasketCmdOutputData struct {
	Basket *models.MutualFundBasket
}

// CreateMutualFundBasketService handles mutual fund basket creation
type CreateMutualFundBasketService struct {
	mutualFundBasketRepo repositories.MutualFundBasketRepo
	advisoryTypeRepo     repositories.AdvisoryTypeRepo
}

// NewCreateMutualFundBasketService creates a new create mutual fund basket service
func NewCreateMutualFundBasketService(
	mutualFundBasketRepo repositories.MutualFundBasketRepo,
	advisoryTypeRepo repositories.AdvisoryTypeRepo,
) *CreateMutualFundBasketService {
	return &CreateMutualFundBasketService{
		mutualFundBasketRepo: mutualFundBasketRepo,
		advisoryTypeRepo:     advisoryTypeRepo,
	}
}

// Execute executes the CreateMutualFundBasket command
func (s *CreateMutualFundBasketService) Execute(ctx context.Context, cmd CreateMutualFundBasketCmd) (*CreateMutualFundBasketCmdOutputData, error) {
	// Validate command
	if err := s.validateCommand(cmd); err != nil {
		return nil, err
	}

	// Get mutual_fund_basket advisory type
	advisoryType, err := s.advisoryTypeRepo.FindByName(ctx, "mutual_fund_basket")
	if err != nil {
		return nil, fmt.Errorf("failed to find advisory type: %w", err)
	}
	advisoryTypeID := advisoryType.ID

	maxSchemes := cmd.MaxSchemes
	if maxSchemes == 0 {
		maxSchemes = 10
	}

	now := time.Now()
	basket := &models.MutualFundBasket{
		ID:             uuid.New(),
		AdvisoryTypeID: advisoryTypeID,
		BasketType:     cmd.BasketType,
		Name:           cmd.Name,
		Description:    cmd.Description,
		MaxSchemes:     maxSchemes,
		Status:         "draft",
		PublishedBy:    &cmd.PublishedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Create items
	items := make([]*models.MutualFundBasketItem, 0, len(cmd.Items))
	for _, itemCmd := range cmd.Items {
		item := &models.MutualFundBasketItem{
			ID:                 uuid.New(),
			MutualFundBasketID: basket.ID,
			SchemeName:         itemCmd.SchemeName,
			SchemeCode:         itemCmd.SchemeCode,
			EntryPrice:         itemCmd.EntryPrice,
			ExitPrice:          itemCmd.ExitPrice,
			CurrentNAV:         itemCmd.CurrentNAV,
			Trend:              itemCmd.Trend,
			Status:             "active",
			DisplayOrder:       itemCmd.DisplayOrder,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		items = append(items, item)
	}
	// Convert to slice of models (not pointers) for GORM
	basketItems := make([]models.MutualFundBasketItem, 0, len(items))
	for _, item := range items {
		basketItems = append(basketItems, *item)
	}
	basket.Items = basketItems

	// Persist via repository
	if err := s.mutualFundBasketRepo.Create(ctx, basket); err != nil {
		return nil, fmt.Errorf("failed to create mutual fund basket: %w", err)
	}

	return &CreateMutualFundBasketCmdOutputData{Basket: basket}, nil
}

// validateCommand validates the command input
func (s *CreateMutualFundBasketService) validateCommand(cmd CreateMutualFundBasketCmd) error {
	if cmd.Name == "" {
		return errors.NewDomainError("INVALID_INPUT", "name is required")
	}
	validBasketTypes := map[string]bool{
		"general":    true,
		"index_fund": true,
		"sectoral":   true,
		"nfo_review": true,
	}
	if !validBasketTypes[cmd.BasketType] {
		return errors.NewDomainError("INVALID_INPUT", "invalid basket type")
	}
	if len(cmd.Items) > cmd.MaxSchemes {
		return errors.NewDomainError("INVALID_INPUT", "items exceed max schemes limit")
	}
	return nil
}

// GetMutualFundBasketQuery represents the query for getting a mutual fund basket
type GetMutualFundBasketQuery struct {
	BasketID uuid.UUID
}

// GetMutualFundBasketResult represents the output
type GetMutualFundBasketResult struct {
	Basket *models.MutualFundBasket
}

// GetMutualFundBasketService handles mutual fund basket retrieval
type GetMutualFundBasketService struct {
	mutualFundBasketRepo repositories.MutualFundBasketRepo
}

// NewGetMutualFundBasketService creates a new get mutual fund basket service
func NewGetMutualFundBasketService(mutualFundBasketRepo repositories.MutualFundBasketRepo) *GetMutualFundBasketService {
	return &GetMutualFundBasketService{
		mutualFundBasketRepo: mutualFundBasketRepo,
	}
}

// Execute executes the GetMutualFundBasket query
func (s *GetMutualFundBasketService) Execute(ctx context.Context, query GetMutualFundBasketQuery) (*GetMutualFundBasketResult, error) {
	basket, err := s.mutualFundBasketRepo.FindByID(ctx, query.BasketID)
	if err != nil {
		return nil, err
	}

	return &GetMutualFundBasketResult{Basket: basket}, nil
}

// ListMutualFundBasketsQuery represents the query for listing mutual fund baskets
type ListMutualFundBasketsQuery struct {
	BasketType string
	Status     string
	Limit      int
	Offset     int
}

// ListMutualFundBasketsResult represents the output
type ListMutualFundBasketsResult struct {
	Baskets []*models.MutualFundBasket
}

// ListMutualFundBasketsService handles mutual fund basket listing
type ListMutualFundBasketsService struct {
	mutualFundBasketRepo repositories.MutualFundBasketRepo
}

// NewListMutualFundBasketsService creates a new list mutual fund baskets service
func NewListMutualFundBasketsService(mutualFundBasketRepo repositories.MutualFundBasketRepo) *ListMutualFundBasketsService {
	return &ListMutualFundBasketsService{
		mutualFundBasketRepo: mutualFundBasketRepo,
	}
}

// Execute executes the ListMutualFundBaskets query
func (s *ListMutualFundBasketsService) Execute(ctx context.Context, query ListMutualFundBasketsQuery) (*ListMutualFundBasketsResult, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	var baskets []*models.MutualFundBasket
	var err error

	if query.BasketType != "" {
		baskets, err = s.mutualFundBasketRepo.FindByBasketType(ctx, query.BasketType, query.Limit, query.Offset)
	} else {
		baskets, err = s.mutualFundBasketRepo.FindByStatus(ctx, query.Status, query.Limit, query.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list mutual fund baskets: %w", err)
	}

	return &ListMutualFundBasketsResult{Baskets: baskets}, nil
}
