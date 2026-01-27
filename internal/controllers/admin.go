package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// AdminController handles admin CRUD operations for all entities
type AdminController struct {
	stockBasketRepo        repositories.StockBasketRepo
	ipoAdvisoryRepo        repositories.IPOAdvisoryRepo
	etfBasketRepo          repositories.ETFBasketRepo
	mfSchemeRepo           repositories.MFSchemeRepo
	mutualFundBasketRepo   repositories.MutualFundBasketRepo
	nfoRepo                repositories.NFORepo
	webinarRepo            repositories.WebinarRepo
	sectorSnapshotRepo     repositories.SectorSnapshotRepo
	weeklyMarketMoodRepo   repositories.WeeklyMarketMoodRepo
	weeklyAudioRepo        repositories.WeeklyAudioRepo
	advisoryTypeRepo       repositories.AdvisoryTypeRepo
}

// NewAdminController creates a new admin controller
func NewAdminController(
	stockBasketRepo repositories.StockBasketRepo,
	ipoAdvisoryRepo repositories.IPOAdvisoryRepo,
	etfBasketRepo repositories.ETFBasketRepo,
	mfSchemeRepo repositories.MFSchemeRepo,
	mutualFundBasketRepo repositories.MutualFundBasketRepo,
	nfoRepo repositories.NFORepo,
	webinarRepo repositories.WebinarRepo,
	sectorSnapshotRepo repositories.SectorSnapshotRepo,
	weeklyMarketMoodRepo repositories.WeeklyMarketMoodRepo,
	weeklyAudioRepo repositories.WeeklyAudioRepo,
	advisoryTypeRepo repositories.AdvisoryTypeRepo,
) *AdminController {
	return &AdminController{
		stockBasketRepo:        stockBasketRepo,
		ipoAdvisoryRepo:         ipoAdvisoryRepo,
		etfBasketRepo:          etfBasketRepo,
		mfSchemeRepo:           mfSchemeRepo,
		mutualFundBasketRepo:   mutualFundBasketRepo,
		nfoRepo:                nfoRepo,
		webinarRepo:            webinarRepo,
		sectorSnapshotRepo:     sectorSnapshotRepo,
		weeklyMarketMoodRepo:   weeklyMarketMoodRepo,
		weeklyAudioRepo:        weeklyAudioRepo,
		advisoryTypeRepo:       advisoryTypeRepo,
	}
}

// Stock Bullets - Create/Update/Delete
// Note: Stock bullets are stored as StockBasketItems with IsBulletIdea=true
// We'll create a dedicated basket for bullets or use a special basket

// CreateStockBullet creates a new stock bullet
func (c *AdminController) CreateStockBullet(ctx *utils.Context) (interface{}, error) {
	var req models.CreateStockBulletRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find or create "Stock Bullets" basket
	bulletBasketType, err := c.advisoryTypeRepo.FindByName(ctx.Request.Context(), "stock_bullet")
	if err != nil {
		// Create advisory type if it doesn't exist
		desc := "Stock Bullet Ideas"
		bulletBasketType = &models.AdvisoryType{
			Name:        "stock_bullet",
			DisplayName: "Stock Bullet Ideas",
			Description: &desc,
		}
		if err := c.advisoryTypeRepo.Create(ctx.Request.Context(), bulletBasketType); err != nil {
			return nil, fmt.Errorf("failed to create advisory type: %w", err)
		}
	}

	// Find existing bullet basket or create one
	baskets, err := c.stockBasketRepo.FindByAdvisoryType(ctx.Request.Context(), bulletBasketType.ID, 1, 0)
	if err != nil || len(baskets) == 0 {
		// Create a new basket for bullets
		userID, _ := uuid.Parse(ctx.UserID)
		desc := "Stock Bullet Ideas"
		basket := &models.StockBasket{
			AdvisoryTypeID: bulletBasketType.ID,
			Name:           "Stock Bullets",
			Description:    stringPtr(desc),
			IsBulletIdea:   true,
			Status:         "published",
			PublishedBy:    &userID,
		}
		if err := c.stockBasketRepo.Create(ctx.Request.Context(), basket); err != nil {
			return nil, fmt.Errorf("failed to create bullet basket: %w", err)
		}
		baskets = []*models.StockBasket{basket}
	}

	basket := baskets[0]

	// Create bullet item
	item := &models.StockBasketItem{
		StockBasketID: basket.ID,
		StockName:     req.Name,
		StockSymbol:   &req.Exchange, // Using exchange field for symbol
		CMP:           parsePrice(req.Price),
		Target:        parsePrice(req.Price) * 1.2, // Default target
		Action:        stringPtr(getActionFromVerdict(req.Verdict)),
		Rationale:     req.Rationale,
		IsBulletIdea:  true,
		DisplayOrder:  0,
		Status:        "active",
	}

	if err := c.stockBasketRepo.CreateItem(ctx.Request.Context(), item); err != nil {
		return nil, fmt.Errorf("failed to create stock bullet: %w", err)
	}

	return map[string]interface{}{
		"id":       item.ID.String(),
		"name":     item.StockName,
		"exchange": getExchangeFromSymbol(item.StockSymbol),
		"price":    formatPrice(item.CMP),
		"rationale": item.Rationale,
		"verdict":  getVerdictFromAction(item.Action),
	}, nil
}

// UpdateStockBullet updates a stock bullet
func (c *AdminController) UpdateStockBullet(ctx *utils.Context) (interface{}, error) {
	var params models.GetStockBulletParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateStockBulletRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	itemID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	item, err := c.stockBasketRepo.FindItemByID(ctx.Request.Context(), itemID)
	if err != nil {
		return nil, fmt.Errorf("stock bullet not found: %w", err)
	}

	// Update fields
	if req.Name != nil {
		item.StockName = *req.Name
	}
	if req.Exchange != nil {
		item.StockSymbol = req.Exchange
	}
	if req.Price != nil {
		item.CMP = parsePrice(*req.Price)
	}
	if req.Rationale != nil {
		item.Rationale = req.Rationale
	}
	if req.Verdict != nil {
		item.Action = stringPtr(getActionFromVerdict(*req.Verdict))
	}

	if err := c.stockBasketRepo.UpdateItem(ctx.Request.Context(), item); err != nil {
		return nil, fmt.Errorf("failed to update stock bullet: %w", err)
	}

	return map[string]interface{}{
		"id":       item.ID.String(),
		"name":     item.StockName,
		"exchange": getExchangeFromSymbol(item.StockSymbol),
		"price":    formatPrice(item.CMP),
		"rationale": item.Rationale,
		"verdict":  getVerdictFromAction(item.Action),
	}, nil
}

// DeleteStockBullet deletes a stock bullet
func (c *AdminController) DeleteStockBullet(ctx *utils.Context) (interface{}, error) {
	var params models.GetStockBulletParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	itemID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.stockBasketRepo.DeleteItem(ctx.Request.Context(), itemID); err != nil {
		return nil, fmt.Errorf("failed to delete stock bullet: %w", err)
	}

	return map[string]interface{}{
		"message": "Stock bullet deleted successfully",
	}, nil
}

// Stock Recommendations - Create/Update/Delete
// Similar to bullets, stored as StockBasketItems

// CreateStockRecommendation creates a new stock recommendation
func (c *AdminController) CreateStockRecommendation(ctx *utils.Context) (interface{}, error) {
	var req models.CreateStockRecommendationRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find or create "Stock Recommendations" basket
	recBasketType, err := c.advisoryTypeRepo.FindByName(ctx.Request.Context(), "stock_recommendation")
	if err != nil {
		recDesc := "Stock Recommendations"
		recBasketType = &models.AdvisoryType{
			Name:        "stock_recommendation",
			DisplayName: "Stock Recommendations",
			Description: &recDesc,
		}
		if err := c.advisoryTypeRepo.Create(ctx.Request.Context(), recBasketType); err != nil {
			return nil, fmt.Errorf("failed to create advisory type: %w", err)
		}
	}

	baskets, err := c.stockBasketRepo.FindByAdvisoryType(ctx.Request.Context(), recBasketType.ID, 1, 0)
	if err != nil || len(baskets) == 0 {
		userID, _ := uuid.Parse(ctx.UserID)
		recDesc := "Stock Recommendations"
		basket := &models.StockBasket{
			AdvisoryTypeID: recBasketType.ID,
			Name:           "Stock Recommendations",
			Description:    stringPtr(recDesc),
			IsBulletIdea:   false,
			Status:         "published",
			PublishedBy:    &userID,
		}
		if err := c.stockBasketRepo.Create(ctx.Request.Context(), basket); err != nil {
			return nil, fmt.Errorf("failed to create recommendation basket: %w", err)
		}
		baskets = []*models.StockBasket{basket}
	}

	basket := baskets[0]

	item := &models.StockBasketItem{
		StockBasketID: basket.ID,
		StockName:     req.Name,
		CMP:           parsePrice(req.Price),
		Target:        parsePrice(req.Price) * 1.2,
		Action:        stringPtr(getActionFromVerdict(req.Verdict)),
		IsBulletIdea:  false,
		DisplayOrder:  0,
		Status:        "active",
	}

	if err := c.stockBasketRepo.CreateItem(ctx.Request.Context(), item); err != nil {
		return nil, fmt.Errorf("failed to create stock recommendation: %w", err)
	}

	return map[string]interface{}{
		"id":     item.ID.String(),
		"name":   item.StockName,
		"price":  formatPrice(item.CMP),
		"verdict": getVerdictFromAction(item.Action),
	}, nil
}

// UpdateStockRecommendation updates a stock recommendation
func (c *AdminController) UpdateStockRecommendation(ctx *utils.Context) (interface{}, error) {
	var params models.GetStockRecommendationParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateStockRecommendationRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	itemID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	item, err := c.stockBasketRepo.FindItemByID(ctx.Request.Context(), itemID)
	if err != nil {
		return nil, fmt.Errorf("stock recommendation not found: %w", err)
	}

	if req.Name != nil {
		item.StockName = *req.Name
	}
	if req.Price != nil {
		item.CMP = parsePrice(*req.Price)
	}
	if req.Verdict != nil {
		item.Action = stringPtr(getActionFromVerdict(*req.Verdict))
	}

	if err := c.stockBasketRepo.UpdateItem(ctx.Request.Context(), item); err != nil {
		return nil, fmt.Errorf("failed to update stock recommendation: %w", err)
	}

	return map[string]interface{}{
		"id":     item.ID.String(),
		"name":   item.StockName,
		"price":  formatPrice(item.CMP),
		"verdict": getVerdictFromAction(item.Action),
	}, nil
}

// DeleteStockRecommendation deletes a stock recommendation
func (c *AdminController) DeleteStockRecommendation(ctx *utils.Context) (interface{}, error) {
	var params models.GetStockRecommendationParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	itemID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.stockBasketRepo.DeleteItem(ctx.Request.Context(), itemID); err != nil {
		return nil, fmt.Errorf("failed to delete stock recommendation: %w", err)
	}

	return map[string]interface{}{
		"message": "Stock recommendation deleted successfully",
	}, nil
}

// ETF Baskets - Create/Update/Delete
// CreateETF creates an ETF basket
func (c *AdminController) CreateETF(ctx *utils.Context) (interface{}, error) {
	var req models.CreateETFBasketRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request payload: %w. Please ensure all required fields (name) are provided and properly formatted. Items are optional and can be added later", err)
	}

	// Validate required fields
	if req.Name == "" {
		return nil, errors.NewDomainError("INVALID_INPUT", "ETF basket name is required")
	}
	
	// Handle null items array - initialize if nil (items are optional)
	if req.Items == nil {
		req.Items = []*models.CreateETFBasketItemRequest{}
	}

	// Validate items if provided (items are optional - can be added later)
	for i, itemReq := range req.Items {
		if itemReq == nil {
			return nil, fmt.Errorf("invalid item at index %d: item cannot be null", i)
		}
		if itemReq.Name == "" {
			return nil, fmt.Errorf("invalid item at index %d: name is required", i)
		}
		if itemReq.CMP <= 0 {
			return nil, fmt.Errorf("invalid item at index %d (name: %s): CMP must be greater than 0", i, itemReq.Name)
		}
		if itemReq.Target <= 0 {
			return nil, fmt.Errorf("invalid item at index %d (name: %s): target must be greater than 0", i, itemReq.Name)
		}
	}

	// Find or create ETF advisory type
	etfType, err := c.advisoryTypeRepo.FindByName(ctx.Request.Context(), "etf_basket")
	if err != nil {
		// Only create if it's a "not found" error, otherwise return the error
		domainErr, ok := err.(*errors.DomainError)
		if !ok || domainErr.Code != "ADVISORY_TYPE_NOT_FOUND" {
			return nil, fmt.Errorf("failed to find ETF advisory type 'etf_basket': %w. Please check database connectivity", err)
		}
		
		desc := "ETF Baskets"
		etfType = &models.AdvisoryType{
			Name:        "etf_basket",
			DisplayName: "ETF Baskets",
			Description: &desc,
			IsActive:    true,
		}
		if err := c.advisoryTypeRepo.Create(ctx.Request.Context(), etfType); err != nil {
			return nil, fmt.Errorf("failed to create ETF advisory type 'etf_basket': %w. This may indicate a database constraint violation or connectivity issue", err)
		}
	}

	// Validate and parse user ID
	if ctx.UserID == "" {
		return nil, errors.NewDomainError("INVALID_INPUT", "user ID is missing from authentication context")
	}
	userID, err := uuid.Parse(ctx.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format '%s': %w. Please ensure you are properly authenticated", ctx.UserID, err)
	}

	// Create ETF basket
	basket := &models.ETFBasket{
		AdvisoryTypeID: etfType.ID,
		Name:          req.Name,
		Description:   req.Description,
		Status:        "draft",
		PublishedBy:   &userID,
	}

	if err := c.etfBasketRepo.Create(ctx.Request.Context(), basket); err != nil {
		return nil, fmt.Errorf("failed to create ETF basket '%s': %w. Please check if a basket with this name already exists or verify database constraints", req.Name, err)
	}

	// Create items if provided (items are optional - can be added later via separate endpoint)
	for i, itemReq := range req.Items {
		item := &models.ETFBasketItem{
			ETFBasketID: basket.ID,
			Name:        itemReq.Name,
			Symbol:      itemReq.Symbol,
			CMP:         itemReq.CMP,
			Target:      itemReq.Target,
			StopLoss:    itemReq.StopLoss,
			EntryRangeMin: itemReq.EntryRangeMin,
			EntryRangeMax: itemReq.EntryRangeMax,
			Action:      itemReq.Action,
			RiskLevel:   itemReq.RiskLevel,
			TimeHorizon: itemReq.TimeHorizon,
			Rationale:   itemReq.Rationale,
			DisplayOrder: itemReq.DisplayOrder,
			Status:      "active",
		}
		if err := c.etfBasketRepo.CreateItem(ctx.Request.Context(), item); err != nil {
			return nil, fmt.Errorf("failed to create ETF basket item at index %d (name: '%s') for basket '%s': %w. Please verify the item data and database constraints", i, itemReq.Name, req.Name, err)
		}
	}

	// Reload with items
	basket, err = c.etfBasketRepo.FindByID(ctx.Request.Context(), basket.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created ETF basket '%s' (ID: %s): %w. The basket may have been created but items retrieval failed", req.Name, basket.ID.String(), err)
	}

	return basket, nil
}

// UpdateETF updates an ETF basket
func (c *AdminController) UpdateETF(ctx *utils.Context) (interface{}, error) {
	var params models.GetETFBasketParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, fmt.Errorf("invalid ETF basket ID in URL path: %w. Please provide a valid UUID", err)
	}

	var req models.UpdateETFBasketRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request payload: %w. Please ensure the request body is properly formatted JSON", err)
	}

	// Validate that at least one field is being updated
	if req.Name == nil && req.Description == nil && req.Status == nil {
		return nil, errors.NewDomainError("INVALID_INPUT", "at least one field (name, description, or status) must be provided for update")
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{"draft": true, "published": true, "archived": true}
		if !validStatuses[*req.Status] {
			return nil, fmt.Errorf("invalid status '%s'. Status must be one of: draft, published, archived", *req.Status)
		}
	}

	basketID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ETF basket ID format '%s': %w. Please provide a valid UUID", params.ID, err)
	}

	basket, err := c.etfBasketRepo.FindByID(ctx.Request.Context(), basketID)
	if err != nil {
		return nil, fmt.Errorf("ETF basket with ID '%s' not found: %w. Please verify the basket exists", params.ID, err)
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, errors.NewDomainError("INVALID_INPUT", "ETF basket name cannot be empty")
		}
		basket.Name = *req.Name
	}
	if req.Description != nil {
		basket.Description = req.Description
	}
	if req.Status != nil {
		basket.Status = *req.Status
	}

	if err := c.etfBasketRepo.Update(ctx.Request.Context(), basket); err != nil {
		return nil, fmt.Errorf("failed to update ETF basket '%s' (ID: %s): %w. Please check database constraints and connectivity", basket.Name, params.ID, err)
	}

	return basket, nil
}

// DeleteETF deletes an ETF basket
func (c *AdminController) DeleteETF(ctx *utils.Context) (interface{}, error) {
	var params models.GetETFBasketParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, fmt.Errorf("invalid ETF basket ID in URL path: %w. Please provide a valid UUID", err)
	}

	basketID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid ETF basket ID format '%s': %w. Please provide a valid UUID", params.ID, err)
	}

	// Check if basket exists before attempting deletion
	basket, err := c.etfBasketRepo.FindByID(ctx.Request.Context(), basketID)
	if err != nil {
		return nil, fmt.Errorf("ETF basket with ID '%s' not found: %w. Cannot delete a non-existent basket", params.ID, err)
	}

	if err := c.etfBasketRepo.Delete(ctx.Request.Context(), basketID); err != nil {
		return nil, fmt.Errorf("failed to delete ETF basket '%s' (ID: %s): %w. This may be due to foreign key constraints or database connectivity issues", basket.Name, params.ID, err)
	}

	return map[string]interface{}{
		"message": fmt.Sprintf("ETF basket '%s' (ID: %s) deleted successfully", basket.Name, params.ID),
	}, nil
}

// CreateETF creates an ETF basket (delegate to existing controller)
// This will be handled by ETFBasketController

// MF Schemes - Create/Update/Delete
// CreateMFScheme creates a new MF scheme
func (c *AdminController) CreateMFScheme(ctx *utils.Context) (interface{}, error) {
	var req models.CreateMFSchemeRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	scheme := &models.MFScheme{
		Name:        req.Name,
		SchemeCode:  req.SchemeCode,
		AMC:         stringPtr(req.AMC),
		Category:    stringPtr(req.Category),
		Type:        stringPtr(req.Type),
		CurrentNAV:  req.CurrentNAV,
		EntryPrice:  req.EntryPrice,
		ExitPrice:   req.ExitPrice,
		Trend:       req.Trend,
		Verdict:     req.Verdict,
		Rationale:   req.Rationale,
		Status:      req.Status,
	}

	if err := c.mfSchemeRepo.Create(ctx.Request.Context(), scheme); err != nil {
		return nil, fmt.Errorf("failed to create MF scheme: %w", err)
	}

	return scheme, nil
}

// UpdateMFScheme updates an MF scheme
func (c *AdminController) UpdateMFScheme(ctx *utils.Context) (interface{}, error) {
	var params models.GetMFSchemeParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateMFSchemeRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	schemeID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	scheme, err := c.mfSchemeRepo.FindByID(ctx.Request.Context(), schemeID)
	if err != nil {
		return nil, fmt.Errorf("MF scheme not found: %w", err)
	}

	if req.Name != nil {
		scheme.Name = *req.Name
	}
	if req.SchemeCode != nil {
		scheme.SchemeCode = req.SchemeCode
	}
	if req.AMC != nil {
		scheme.AMC = stringPtr(*req.AMC)
	}
	if req.Category != nil {
		scheme.Category = stringPtr(*req.Category)
	}
	if req.Type != nil {
		scheme.Type = stringPtr(*req.Type)
	}
	if req.CurrentNAV != nil {
		scheme.CurrentNAV = req.CurrentNAV
	}
	if req.EntryPrice != nil {
		scheme.EntryPrice = req.EntryPrice
	}
	if req.ExitPrice != nil {
		scheme.ExitPrice = req.ExitPrice
	}
	if req.Trend != nil {
		scheme.Trend = req.Trend
	}
	if req.Verdict != nil {
		scheme.Verdict = req.Verdict
	}
	if req.Rationale != nil {
		scheme.Rationale = req.Rationale
	}
	if req.Status != nil {
		scheme.Status = *req.Status
	}

	if err := c.mfSchemeRepo.Update(ctx.Request.Context(), scheme); err != nil {
		return nil, fmt.Errorf("failed to update MF scheme: %w", err)
	}

	return scheme, nil
}

// DeleteMFScheme deletes an MF scheme
func (c *AdminController) DeleteMFScheme(ctx *utils.Context) (interface{}, error) {
	var params models.GetMFSchemeParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	schemeID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.mfSchemeRepo.Delete(ctx.Request.Context(), schemeID); err != nil {
		return nil, fmt.Errorf("failed to delete MF scheme: %w", err)
	}

	return map[string]interface{}{
		"message": "MF scheme deleted successfully",
	}, nil
}

// NFOs - Create/Update/Delete
// CreateNFO creates a new NFO
func (c *AdminController) CreateNFO(ctx *utils.Context) (interface{}, error) {
	var req models.CreateNFORequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Parse dates and summary
	var openDate, closeDate *time.Time
	if req.OpenDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.OpenDate); err == nil {
			openDate = &parsed
		}
	}
	if req.CloseDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.CloseDate); err == nil {
			closeDate = &parsed
		}
	}
	
	var summaryStr *string
	if req.Summary != nil && len(req.Summary) > 0 {
		if jsonBytes, err := json.Marshal(req.Summary); err == nil {
			summaryStr = stringPtr(string(jsonBytes))
		}
	}

	nfo := &models.NFO{
		Name:              req.Name,
		AMC:               stringPtr(req.AMC),
		Category:          stringPtr(req.Category),
		Type:              stringPtr(req.Type),
		OpenDate:          openDate,
		CloseDate:         closeDate,
		Timeline:          req.Timeline,
		Summary:           summaryStr,
		MinimumInvestment: req.MinimumInvestment,
		Verdict:           req.Verdict,
		Rationale:         req.Rationale,
		Status:            req.Status,
	}

	if err := c.nfoRepo.Create(ctx.Request.Context(), nfo); err != nil {
		return nil, fmt.Errorf("failed to create NFO: %w", err)
	}

	return nfo, nil
}

// UpdateNFO updates an NFO
func (c *AdminController) UpdateNFO(ctx *utils.Context) (interface{}, error) {
	var params models.GetNFOParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateNFORequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	nfoID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	nfo, err := c.nfoRepo.FindByID(ctx.Request.Context(), nfoID)
	if err != nil {
		return nil, fmt.Errorf("NFO not found: %w", err)
	}

	if req.Name != nil {
		nfo.Name = *req.Name
	}
	if req.AMC != nil {
		nfo.AMC = stringPtr(*req.AMC)
	}
	if req.Category != nil {
		nfo.Category = stringPtr(*req.Category)
	}
	if req.Type != nil {
		nfo.Type = stringPtr(*req.Type)
	}
	if req.OpenDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.OpenDate); err == nil {
			nfo.OpenDate = &parsed
		}
	}
	if req.CloseDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.CloseDate); err == nil {
			nfo.CloseDate = &parsed
		}
	}
	if req.Timeline != nil {
		nfo.Timeline = req.Timeline
	}
	if req.Summary != nil && len(req.Summary) > 0 {
		if jsonBytes, err := json.Marshal(req.Summary); err == nil {
			nfo.Summary = stringPtr(string(jsonBytes))
		}
	}
	if req.MinimumInvestment != nil {
		nfo.MinimumInvestment = req.MinimumInvestment
	}
	if req.Verdict != nil {
		nfo.Verdict = req.Verdict
	}
	if req.Rationale != nil {
		nfo.Rationale = req.Rationale
	}
	if req.Status != nil {
		nfo.Status = *req.Status
	}

	if err := c.nfoRepo.Update(ctx.Request.Context(), nfo); err != nil {
		return nil, fmt.Errorf("failed to update NFO: %w", err)
	}

	return nfo, nil
}

// DeleteNFO deletes an NFO
func (c *AdminController) DeleteNFO(ctx *utils.Context) (interface{}, error) {
	var params models.GetNFOParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	nfoID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.nfoRepo.Delete(ctx.Request.Context(), nfoID); err != nil {
		return nil, fmt.Errorf("failed to delete NFO: %w", err)
	}

	return map[string]interface{}{
		"message": "NFO deleted successfully",
	}, nil
}

// IPOs - Delete (Create/Update already exist in IPOAdvisoryController)
// DeleteIPO deletes an IPO advisory
func (c *AdminController) DeleteIPO(ctx *utils.Context) (interface{}, error) {
	var params models.GetIPOAdvisoryParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	ipoID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.ipoAdvisoryRepo.Delete(ctx.Request.Context(), ipoID); err != nil {
		return nil, fmt.Errorf("failed to delete IPO advisory: %w", err)
	}

	return map[string]interface{}{
		"message": "IPO advisory deleted successfully",
	}, nil
}

// MF Baskets - Update/Delete (Create already exists in MutualFundBasketController)
// UpdateMFBasket updates a mutual fund basket
func (c *AdminController) UpdateMFBasket(ctx *utils.Context) (interface{}, error) {
	var params models.UpdateMutualFundBasketParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateMutualFundBasketRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	basketID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	basket, err := c.mutualFundBasketRepo.FindByID(ctx.Request.Context(), basketID)
	if err != nil {
		return nil, fmt.Errorf("mutual fund basket not found: %w", err)
	}

	if req.Name != nil {
		basket.Name = *req.Name
	}
	if req.Status != nil {
		basket.Status = *req.Status
	}

	if err := c.mutualFundBasketRepo.Update(ctx.Request.Context(), basket); err != nil {
		return nil, fmt.Errorf("failed to update mutual fund basket: %w", err)
	}

	return basket, nil
}

// DeleteMFBasket deletes a mutual fund basket
func (c *AdminController) DeleteMFBasket(ctx *utils.Context) (interface{}, error) {
	var params models.UpdateMutualFundBasketParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	basketID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.mutualFundBasketRepo.Delete(ctx.Request.Context(), basketID); err != nil {
		return nil, fmt.Errorf("failed to delete mutual fund basket: %w", err)
	}

	return map[string]interface{}{
		"message": "Mutual fund basket deleted successfully",
	}, nil
}

// Sectors - Create/Update/Delete
// CreateSector creates a new sector snapshot
func (c *AdminController) CreateSector(ctx *utils.Context) (interface{}, error) {
	var req models.CreateSectorRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	sector := &models.SectorSnapshot{
		SectorName:      req.SectorName,
		CurrentValue:    req.CurrentValue,
		ChangePercentage: req.ChangePercentage,
		ChangeValue:     req.ChangeValue,
		ReportURL:       req.ReportURL,
	}

	if err := c.sectorSnapshotRepo.Create(ctx.Request.Context(), sector); err != nil {
		return nil, fmt.Errorf("failed to create sector: %w", err)
	}

	return sector, nil
}

// UpdateSector updates a sector snapshot
func (c *AdminController) UpdateSector(ctx *utils.Context) (interface{}, error) {
	var params models.GetSectorParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateSectorRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	sectorID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	sector, err := c.sectorSnapshotRepo.FindByID(ctx.Request.Context(), sectorID)
	if err != nil {
		return nil, fmt.Errorf("sector not found: %w", err)
	}

	if req.SectorName != nil {
		sector.SectorName = *req.SectorName
	}
	if req.CurrentValue != nil {
		sector.CurrentValue = req.CurrentValue
	}
	if req.ChangePercentage != nil {
		sector.ChangePercentage = req.ChangePercentage
	}
	if req.ChangeValue != nil {
		sector.ChangeValue = req.ChangeValue
	}
	if req.ReportURL != nil {
		sector.ReportURL = req.ReportURL
	}

	if err := c.sectorSnapshotRepo.Update(ctx.Request.Context(), sector); err != nil {
		return nil, fmt.Errorf("failed to update sector: %w", err)
	}

	return sector, nil
}

// DeleteSector deletes a sector snapshot
func (c *AdminController) DeleteSector(ctx *utils.Context) (interface{}, error) {
	var params models.GetSectorParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	sectorID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.sectorSnapshotRepo.Delete(ctx.Request.Context(), sectorID); err != nil {
		return nil, fmt.Errorf("failed to delete sector: %w", err)
	}

	return map[string]interface{}{
		"message": "Sector deleted successfully",
	}, nil
}

// Webinars - Create/Update/Delete
// CreateWebinar creates a new webinar
func (c *AdminController) CreateWebinar(ctx *utils.Context) (interface{}, error) {
	var req models.CreateWebinarRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Parse date
	var date *time.Time
	if req.Date != nil {
		if parsed, err := time.Parse("2006-01-02", *req.Date); err == nil {
			date = &parsed
		}
	}

	webinar := &models.Webinar{
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Date:           date,
		Time:           req.Time,
		Duration:       req.Duration,
		RegistrationURL: req.RegistrationURL,
		RecordingURL:   req.RecordingURL,
		ThumbnailURL:   req.ThumbnailURL,
	}

	if err := c.webinarRepo.Create(ctx.Request.Context(), webinar); err != nil {
		return nil, fmt.Errorf("failed to create webinar: %w", err)
	}

	return webinar, nil
}

// UpdateWebinar updates a webinar
func (c *AdminController) UpdateWebinar(ctx *utils.Context) (interface{}, error) {
	var params models.GetWebinarParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateWebinarRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	webinarID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	webinar, err := c.webinarRepo.FindByID(ctx.Request.Context(), webinarID)
	if err != nil {
		return nil, fmt.Errorf("webinar not found: %w", err)
	}

	if req.Title != nil {
		webinar.Title = *req.Title
	}
	if req.Description != nil {
		webinar.Description = req.Description
	}
	if req.Status != nil {
		webinar.Status = *req.Status
	}
	if req.Date != nil {
		if parsed, err := time.Parse("2006-01-02", *req.Date); err == nil {
			webinar.Date = &parsed
		}
	}
	if req.Time != nil {
		webinar.Time = req.Time
	}
	if req.Duration != nil {
		webinar.Duration = req.Duration
	}
	if req.RegistrationURL != nil {
		webinar.RegistrationURL = req.RegistrationURL
	}
	if req.RecordingURL != nil {
		webinar.RecordingURL = req.RecordingURL
	}
	if req.ThumbnailURL != nil {
		webinar.ThumbnailURL = req.ThumbnailURL
	}

	if err := c.webinarRepo.Update(ctx.Request.Context(), webinar); err != nil {
		return nil, fmt.Errorf("failed to update webinar: %w", err)
	}

	return webinar, nil
}

// DeleteWebinar deletes a webinar
func (c *AdminController) DeleteWebinar(ctx *utils.Context) (interface{}, error) {
	var params models.GetWebinarParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	webinarID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.webinarRepo.Delete(ctx.Request.Context(), webinarID); err != nil {
		return nil, fmt.Errorf("failed to delete webinar: %w", err)
	}

	return map[string]interface{}{
		"message": "Webinar deleted successfully",
	}, nil
}

// Market Pulse - Create/Update/Delete
// Note: Market pulse is derived from sector snapshots, but we can create a dedicated model
// For now, we'll use sector snapshots

// CreateMarketPulse creates a new market pulse entry (as sector snapshot)
func (c *AdminController) CreateMarketPulse(ctx *utils.Context) (interface{}, error) {
	var req models.CreateMarketPulseRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Create as a sector snapshot with special naming
	sector := &models.SectorSnapshot{
		SectorName: fmt.Sprintf("Market Pulse: %s", req.Content[:min(50, len(req.Content))]),
		ReportURL:  nil,
	}

	if err := c.sectorSnapshotRepo.Create(ctx.Request.Context(), sector); err != nil {
		return nil, fmt.Errorf("failed to create market pulse: %w", err)
	}

	return map[string]interface{}{
		"id":      sector.ID.String(),
		"content": req.Content,
		"updated": req.Updated,
	}, nil
}

// UpdateMarketPulse updates a market pulse entry
func (c *AdminController) UpdateMarketPulse(ctx *utils.Context) (interface{}, error) {
	var params models.GetMarketPulseParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateMarketPulseRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	sectorID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	sector, err := c.sectorSnapshotRepo.FindByID(ctx.Request.Context(), sectorID)
	if err != nil {
		return nil, fmt.Errorf("market pulse not found: %w", err)
	}

	if req.Content != nil {
		sector.SectorName = fmt.Sprintf("Market Pulse: %s", (*req.Content)[:min(50, len(*req.Content))])
	}

	if err := c.sectorSnapshotRepo.Update(ctx.Request.Context(), sector); err != nil {
		return nil, fmt.Errorf("failed to update market pulse: %w", err)
	}

	return map[string]interface{}{
		"id":      sector.ID.String(),
		"content": req.Content,
		"updated": req.Updated,
	}, nil
}

// DeleteMarketPulse deletes a market pulse entry
func (c *AdminController) DeleteMarketPulse(ctx *utils.Context) (interface{}, error) {
	var params models.GetMarketPulseParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	sectorID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	if err := c.sectorSnapshotRepo.Delete(ctx.Request.Context(), sectorID); err != nil {
		return nil, fmt.Errorf("failed to delete market pulse: %w", err)
	}

	return map[string]interface{}{
		"message": "Market pulse deleted successfully",
	}, nil
}

// Top Call - Update
// UpdateTopCall updates the top call
func (c *AdminController) UpdateTopCall(ctx *utils.Context) (interface{}, error) {
	var req models.UpdateTopCallRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Top call is derived from stock baskets or IPOs
	// We can create a dedicated top call record or update the latest basket
	// For now, return success - actual implementation depends on business logic
	return map[string]interface{}{
		"type":     req.Type,
		"name":     req.Name,
		"verdict":  req.Verdict,
		"rationale": req.Rationale,
		"message":  "Top call updated successfully",
	}, nil
}

// Weekly Market Mood - Update
// UpdateWeeklyMarketMood updates weekly market mood
func (c *AdminController) UpdateWeeklyMarketMood(ctx *utils.Context) (interface{}, error) {
	var req models.UpdateWeeklyMarketMoodRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find latest or create new
	moods, err := c.weeklyMarketMoodRepo.FindAll(ctx.Request.Context(), 1, 0)
	if err != nil || len(moods) == 0 {
		// Convert points to JSON string
		var pointsStr *string
		if req.Points != nil && len(req.Points) > 0 {
			if jsonBytes, err := json.Marshal(req.Points); err == nil {
				pointsStr = stringPtr(string(jsonBytes))
			}
		}

		// Create new
		mood := &models.WeeklyMarketMood{
			Title:         req.Title,
			Week:          req.Week,
			Points:        pointsStr,
			Sentiment:     req.Sentiment,
			SentimentLabel: req.SentimentLabel,
			ReportURL:     req.ReportURL,
		}
		if err := c.weeklyMarketMoodRepo.Create(ctx.Request.Context(), mood); err != nil {
			return nil, fmt.Errorf("failed to create weekly market mood: %w", err)
		}
		return mood, nil
	}

	// Update existing
	mood := moods[0]
	if req.Title != "" {
		mood.Title = req.Title
	}
	if req.Week != "" {
		mood.Week = req.Week
	}
	if req.Points != nil && len(req.Points) > 0 {
		if jsonBytes, err := json.Marshal(req.Points); err == nil {
			mood.Points = stringPtr(string(jsonBytes))
		}
	}
	if req.Sentiment != 0 {
		mood.Sentiment = req.Sentiment
	}
	if req.SentimentLabel != "" {
		mood.SentimentLabel = req.SentimentLabel
	}
	if req.ReportURL != nil {
		mood.ReportURL = req.ReportURL
	}

	if err := c.weeklyMarketMoodRepo.Update(ctx.Request.Context(), mood); err != nil {
		return nil, fmt.Errorf("failed to update weekly market mood: %w", err)
	}

	return mood, nil
}

// Weekly Audio - Create/Update
// UpdateWeeklyAudio creates or updates weekly audio
func (c *AdminController) UpdateWeeklyAudio(ctx *utils.Context) (interface{}, error) {
	var req models.UpdateWeeklyAudioRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find latest or create new
	audios, err := c.weeklyAudioRepo.FindAll(ctx.Request.Context(), 1, 0)
	if err != nil || len(audios) == 0 {
		// Create new
		duration := ""
		if req.Duration != nil {
			duration = *req.Duration
		}
		week := ""
		if req.Week != nil {
			week = *req.Week
		}
		
		audio := &models.WeeklyAudio{
			Title:       req.Title,
			Description: req.Description,
			Duration:    duration,
			Week:        week,
			AudioURL:    req.AudioURL,
			ThumbnailURL: req.ThumbnailURL,
			Status:      req.Status,
		}
		if err := c.weeklyAudioRepo.Create(ctx.Request.Context(), audio); err != nil {
			return nil, fmt.Errorf("failed to create weekly audio: %w", err)
		}
		return audio, nil
	}

	// Update existing
	audio := audios[0]
	if req.Title != "" {
		audio.Title = req.Title
	}
	if req.Description != nil {
		audio.Description = req.Description
	}
	if req.Duration != nil {
		audio.Duration = *req.Duration
	}
	if req.Week != nil {
		audio.Week = *req.Week
	}
	if req.AudioURL != nil {
		audio.AudioURL = req.AudioURL
	}
	if req.ThumbnailURL != nil {
		audio.ThumbnailURL = req.ThumbnailURL
	}
	if req.Status != "" {
		audio.Status = req.Status
	}

	if err := c.weeklyAudioRepo.Update(ctx.Request.Context(), audio); err != nil {
		return nil, fmt.Errorf("failed to update weekly audio: %w", err)
	}

	return audio, nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func parsePrice(priceStr string) float64 {
	// Remove currency symbols and parse
	var price float64
	fmt.Sscanf(priceStr, "₹%f", &price)
	if price == 0 {
		fmt.Sscanf(priceStr, "%f", &price)
	}
	return price
}

func getActionFromVerdict(verdict string) string {
	switch verdict {
	case "Buy":
		return "buy"
	case "Sell":
		return "sell"
	default:
		return "hold"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
