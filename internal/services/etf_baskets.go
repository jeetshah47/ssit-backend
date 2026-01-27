package services

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// GetETFBasketQuery represents the query for getting an ETF basket
type GetETFBasketQuery struct {
	BasketID uuid.UUID
}

// GetETFBasketResult represents the output
type GetETFBasketResult struct {
	Basket *models.ETFBasket
}

// GetETFBasketService handles ETF basket retrieval
type GetETFBasketService struct {
	etfBasketRepo repositories.ETFBasketRepo
}

// NewGetETFBasketService creates a new get ETF basket service
func NewGetETFBasketService(etfBasketRepo repositories.ETFBasketRepo) *GetETFBasketService {
	return &GetETFBasketService{
		etfBasketRepo: etfBasketRepo,
	}
}

// Execute executes the GetETFBasket query
func (s *GetETFBasketService) Execute(ctx context.Context, query GetETFBasketQuery) (*GetETFBasketResult, error) {
	basket, err := s.etfBasketRepo.FindByID(ctx, query.BasketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ETF basket: %w", err)
	}

	return &GetETFBasketResult{
		Basket: basket,
	}, nil
}

// ListETFBasketsQuery represents the query for listing ETF baskets
type ListETFBasketsQuery struct {
	Status string
	Limit  int
	Offset int
}

// ListETFBasketsResult represents the output
type ListETFBasketsResult struct {
	Baskets []*models.ETFBasket
}

// ListETFBasketsService handles ETF basket listing
type ListETFBasketsService struct {
	etfBasketRepo repositories.ETFBasketRepo
}

// NewListETFBasketsService creates a new list ETF baskets service
func NewListETFBasketsService(etfBasketRepo repositories.ETFBasketRepo) *ListETFBasketsService {
	return &ListETFBasketsService{
		etfBasketRepo: etfBasketRepo,
	}
}

// Execute executes the ListETFBaskets query
func (s *ListETFBasketsService) Execute(ctx context.Context, query ListETFBasketsQuery) (*ListETFBasketsResult, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 10
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	var baskets []*models.ETFBasket
	var err error

	if query.Status == "published" {
		baskets, err = s.etfBasketRepo.FindPublished(ctx, limit, offset)
	} else {
		baskets, err = s.etfBasketRepo.FindByStatus(ctx, query.Status, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list ETF baskets: %w", err)
	}

	return &ListETFBasketsResult{
		Baskets: baskets,
	}, nil
}
