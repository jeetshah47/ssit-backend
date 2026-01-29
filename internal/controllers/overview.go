package controllers

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/equitywala/backend/internal/services"
)

// OverviewController handles overview page endpoints
type OverviewController struct {
	getOverviewService        *services.GetOverviewService
	sectorSnapshotRepo        repositories.SectorSnapshotRepo
	stockBasketRepo           repositories.StockBasketRepo
	ipoAdvisoryRepo           repositories.IPOAdvisoryRepo
	weeklyMarketMoodRepo      repositories.WeeklyMarketMoodRepo
	getWeeklyMarketMoodService *services.GetWeeklyMarketMoodService
}

// NewOverviewController creates a new overview controller
func NewOverviewController(
	getOverviewService *services.GetOverviewService,
	sectorSnapshotRepo repositories.SectorSnapshotRepo,
	stockBasketRepo repositories.StockBasketRepo,
	ipoAdvisoryRepo repositories.IPOAdvisoryRepo,
	weeklyMarketMoodRepo repositories.WeeklyMarketMoodRepo,
	getWeeklyMarketMoodService *services.GetWeeklyMarketMoodService,
) *OverviewController {
	return &OverviewController{
		getOverviewService:         getOverviewService,
		sectorSnapshotRepo:         sectorSnapshotRepo,
		stockBasketRepo:            stockBasketRepo,
		ipoAdvisoryRepo:            ipoAdvisoryRepo,
		weeklyMarketMoodRepo:       weeklyMarketMoodRepo,
		getWeeklyMarketMoodService: getWeeklyMarketMoodService,
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

// GetMarketPulse retrieves market pulse data (from sector snapshots)
func (c *OverviewController) GetMarketPulse(ctx *utils.Context) (interface{}, error) {
	limit := 10
	if limitStr := ctx.GetQuery("limit"); limitStr != "" {
		if parsedLimit := parseInt(limitStr); parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	snapshots, err := c.sectorSnapshotRepo.FindAll(ctx.Request.Context(), limit)
	if err != nil {
		return nil, err
	}

	// Transform to market pulse format
	pulse := make([]map[string]interface{}, 0, len(snapshots))
	for _, snapshot := range snapshots {
		content := fmt.Sprintf("%s: %s", snapshot.SectorName, getSectorDescription(snapshot))
		pulse = append(pulse, map[string]interface{}{
			"id":      snapshot.ID.String(),
			"content": content,
			"updated": snapshot.UpdatedAt.Format("2006-01-02"),
		})
	}

	return pulse, nil
}

// GetTopCall retrieves today's top call (from latest stock basket or IPO)
func (c *OverviewController) GetTopCall(ctx *utils.Context) (interface{}, error) {
	// Try to get latest published stock basket with bullet idea
	baskets, err := c.stockBasketRepo.FindBulletIdeas(ctx.Request.Context(), 1)
	if err == nil && len(baskets) > 0 {
		basket := baskets[0]
		if len(basket.Items) > 0 {
			item := basket.Items[0]
			name := "N/A"
			if item.Stock.ID != uuid.Nil {
				name = item.Stock.Name
			}
			return map[string]interface{}{
				"type":      "Stock",
				"name":      name,
				"verdict":   getVerdictFromAction(item.Action),
				"rationale": item.Rationale,
			}, nil
		}
	}

	// Fallback to latest IPO
	ipos, err := c.ipoAdvisoryRepo.FindPublished(ctx.Request.Context(), 1, 0)
	if err == nil && len(ipos) > 0 {
		ipo := ipos[0]
		verdict := "Hold"
		if ipo.Suggestion != nil {
			verdict = *ipo.Suggestion
		}
		return map[string]interface{}{
			"type":     "IPO",
			"name":     ipo.IPOName,
			"verdict":  verdict,
			"rationale": fmt.Sprintf("IPO with GMP: %.2f", getGMPValue(ipo.GMP)),
		}, nil
	}

	// Default fallback
	return map[string]interface{}{
		"type":     "Stock",
		"name":     "N/A",
		"verdict":  "Hold",
		"rationale": "No active recommendations",
	}, nil
}

// GetWeeklyMarketMood retrieves weekly market mood
func (c *OverviewController) GetWeeklyMarketMood(ctx *utils.Context) (interface{}, error) {
	query := services.GetWeeklyMarketMoodQuery{
		MoodID: uuid.Nil, // Get latest
	}

	result, err := c.getWeeklyMarketMoodService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Mood, nil
}

// Helper functions
func getSectorDescription(snapshot *models.SectorSnapshot) string {
	if snapshot.ChangePercentage != nil {
		if *snapshot.ChangePercentage > 0 {
			return fmt.Sprintf("Up %.2f%%", *snapshot.ChangePercentage)
		} else if *snapshot.ChangePercentage < 0 {
			return fmt.Sprintf("Down %.2f%%", *snapshot.ChangePercentage)
		}
	}
	return "Stable"
}

func getGMPValue(gmp *float64) float64 {
	if gmp == nil {
		return 0.0
	}
	return *gmp
}
