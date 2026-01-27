package services

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
)

// GetOverviewQuery represents the query for getting overview data
type GetOverviewQuery struct {
	StockWatchlistLimit    int
	MutualFundBasketLimit  int
	SectorSnapshotLimit    int
}

// GetOverviewResult represents the output of GetOverview query
type GetOverviewResult struct {
	SectorSnapshots []*models.SectorSnapshot
	StockWatchlist  []*models.StockBasketItem
	MutualFunds     []*MutualFundBasketWithItems
}

// MutualFundBasketWithItems represents a mutual fund basket with its items
type MutualFundBasketWithItems struct {
	BasketID   string
	BasketName string
	BasketType string
	Items      []*models.MutualFundBasketItem
}

// GetOverviewService handles overview data retrieval
type GetOverviewService struct {
	sectorSnapshotRepo    repositories.SectorSnapshotRepo
	stockBasketRepo       repositories.StockBasketRepo
	mutualFundBasketRepo  repositories.MutualFundBasketRepo
}

// NewGetOverviewService creates a new get overview service
func NewGetOverviewService(
	sectorSnapshotRepo repositories.SectorSnapshotRepo,
	stockBasketRepo repositories.StockBasketRepo,
	mutualFundBasketRepo repositories.MutualFundBasketRepo,
) *GetOverviewService {
	return &GetOverviewService{
		sectorSnapshotRepo:   sectorSnapshotRepo,
		stockBasketRepo:      stockBasketRepo,
		mutualFundBasketRepo: mutualFundBasketRepo,
	}
}

// Execute executes the GetOverview query
func (s *GetOverviewService) Execute(ctx context.Context, query GetOverviewQuery) (*GetOverviewResult, error) {
	// Set defaults
	if query.SectorSnapshotLimit == 0 {
		query.SectorSnapshotLimit = 10
	}
	if query.StockWatchlistLimit == 0 {
		query.StockWatchlistLimit = 10
	}
	if query.MutualFundBasketLimit == 0 {
		query.MutualFundBasketLimit = 2
	}

	// Get sector snapshots
	sectorSnapshots, err := s.sectorSnapshotRepo.FindAll(ctx, query.SectorSnapshotLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get sector snapshots: %w", err)
	}

	// Get latest published stock basket for watchlist
	stockBasket, err := s.stockBasketRepo.FindLatestPublished(ctx)
	var stockWatchlist []*models.StockBasketItem
	if err == nil && stockBasket != nil {
		// Filter active items and limit
		for i := range stockBasket.Items {
			item := &stockBasket.Items[i]
			if item.Status == "active" && len(stockWatchlist) < query.StockWatchlistLimit {
				stockWatchlist = append(stockWatchlist, item)
			}
		}
	}

	// Get published mutual fund baskets
	mutualFundBaskets, err := s.mutualFundBasketRepo.FindPublished(ctx, query.MutualFundBasketLimit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get mutual fund baskets: %w", err)
	}

	// Transform mutual fund baskets
	mutualFunds := make([]*MutualFundBasketWithItems, 0, len(mutualFundBaskets))
	for _, basket := range mutualFundBaskets {
		items := make([]*models.MutualFundBasketItem, 0, len(basket.Items))
		for i := range basket.Items {
			items = append(items, &basket.Items[i])
		}
		mutualFunds = append(mutualFunds, &MutualFundBasketWithItems{
			BasketID:   basket.ID.String(),
			BasketName: basket.Name,
			BasketType: basket.BasketType,
			Items:      items,
		})
	}

	return &GetOverviewResult{
		SectorSnapshots: sectorSnapshots,
		StockWatchlist: stockWatchlist,
		MutualFunds:    mutualFunds,
	}, nil
}
