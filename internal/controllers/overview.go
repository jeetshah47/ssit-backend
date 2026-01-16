package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/services"
)

// OverviewController handles overview page endpoints
type OverviewController struct {
	getOverviewService *services.GetOverviewService
}

// NewOverviewController creates a new overview controller
func NewOverviewController(getOverviewService *services.GetOverviewService) *OverviewController {
	return &OverviewController{
		getOverviewService: getOverviewService,
	}
}

// GetOverview retrieves all overview data
func (c *OverviewController) GetOverview(ctx *utils.Context) (interface{}, error) {
	query := services.GetOverviewQuery{
		StockWatchlistLimit:   10,
		MutualFundBasketLimit: 2,
		SectorSnapshotLimit:   10,
	}

	// Parse query parameters if provided
	if limit := ctx.GetQuery("stock_limit"); limit != "" {
		if parsedLimit := parseInt(limit); parsedLimit > 0 {
			query.StockWatchlistLimit = parsedLimit
		}
	}
	if limit := ctx.GetQuery("mf_limit"); limit != "" {
		if parsedLimit := parseInt(limit); parsedLimit > 0 {
			query.MutualFundBasketLimit = parsedLimit
		}
	}
	if limit := ctx.GetQuery("sector_limit"); limit != "" {
		if parsedLimit := parseInt(limit); parsedLimit > 0 {
			query.SectorSnapshotLimit = parsedLimit
		}
	}

	result, err := c.getOverviewService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sector_snapshots": result.SectorSnapshots,
		"stock_watchlist":  result.StockWatchlist,
		"mutual_funds":     result.MutualFunds,
	}, nil
}

// GetSectorSnapshots retrieves sector snapshots
func (c *OverviewController) GetSectorSnapshots(ctx *utils.Context) (interface{}, error) {
	query := services.GetOverviewQuery{
		SectorSnapshotLimit: 10,
	}

	if limit := ctx.GetQuery("limit"); limit != "" {
		if parsedLimit := parseInt(limit); parsedLimit > 0 {
			query.SectorSnapshotLimit = parsedLimit
		}
	}

	result, err := c.getOverviewService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.SectorSnapshots, nil
}

// GetStockWatchlist retrieves stock watchlist
func (c *OverviewController) GetStockWatchlist(ctx *utils.Context) (interface{}, error) {
	query := services.GetOverviewQuery{
		StockWatchlistLimit: 10,
	}

	if limit := ctx.GetQuery("limit"); limit != "" {
		if parsedLimit := parseInt(limit); parsedLimit > 0 {
			query.StockWatchlistLimit = parsedLimit
		}
	}

	result, err := c.getOverviewService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.StockWatchlist, nil
}

// GetMutualFunds retrieves mutual funds
func (c *OverviewController) GetMutualFunds(ctx *utils.Context) (interface{}, error) {
	query := services.GetOverviewQuery{
		MutualFundBasketLimit: 2,
	}

	if limit := ctx.GetQuery("limit"); limit != "" {
		if parsedLimit := parseInt(limit); parsedLimit > 0 {
			query.MutualFundBasketLimit = parsedLimit
		}
	}

	result, err := c.getOverviewService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.MutualFunds, nil
}
