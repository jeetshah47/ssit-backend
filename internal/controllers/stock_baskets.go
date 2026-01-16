package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// StockBasketController handles stock basket endpoints
type StockBasketController struct {
	createStockBasketService *services.CreateStockBasketService
	getStockBasketService    *services.GetStockBasketService
	listStockBasketsService  *services.ListStockBasketsService
	updateStockBasketService *services.UpdateStockBasketService
}

// NewStockBasketController creates a new stock basket controller
func NewStockBasketController(
	createStockBasketService *services.CreateStockBasketService,
	getStockBasketService *services.GetStockBasketService,
	listStockBasketsService *services.ListStockBasketsService,
	updateStockBasketService *services.UpdateStockBasketService,
) *StockBasketController {
	return &StockBasketController{
		createStockBasketService: createStockBasketService,
		getStockBasketService:    getStockBasketService,
		listStockBasketsService:  listStockBasketsService,
		updateStockBasketService: updateStockBasketService,
	}
}

// CreateStockBasket creates a new stock basket
func (c *StockBasketController) CreateStockBasket(ctx *utils.Context) (interface{}, error) {
	var req models.CreateStockBasketRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Convert request to command
	items := make([]*services.CreateStockBasketItemCmd, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &services.CreateStockBasketItemCmd{
			StockName:      item.StockName,
			StockSymbol:    item.StockSymbol,
			CMP:            item.CMP,
			Target:         item.Target,
			StopLoss:       item.StopLoss,
			EntryRangeMin:  item.EntryRangeMin,
			EntryRangeMax:  item.EntryRangeMax,
			Action:        item.Action,
			RiskLevel:      item.RiskLevel,
			TimeHorizon:    item.TimeHorizon,
			Rationale:      item.Rationale,
			ReportURL:      item.ReportURL,
			FundamentalsURL: item.FundamentalsURL,
			DisplayOrder:   item.DisplayOrder,
			IsBulletIdea:   item.IsBulletIdea,
		})
	}

	userID, err := uuid.Parse(ctx.UserID)
	if err != nil {
		return nil, err
	}

	cmd := services.CreateStockBasketCmd{
		Name:         req.Name,
		Description:  req.Description,
		IsBulletIdea: req.IsBulletIdea,
		Items:        items,
		PublishedBy:  userID,
	}

	result, err := c.createStockBasketService.Execute(ctx.Request.Context(), cmd)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":      result.Basket.ID.String(),
		"name":    result.Basket.Name,
		"status":  result.Basket.Status,
		"message": "Stock basket created successfully",
	}, nil
}

// GetStockBasket retrieves a stock basket by ID
func (c *StockBasketController) GetStockBasket(ctx *utils.Context) (interface{}, error) {
	var params models.GetStockBasketParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	basketID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := services.GetStockBasketQuery{
		BasketID: basketID,
	}

	result, err := c.getStockBasketService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Basket, nil
}

// ListStockBaskets lists stock baskets
func (c *StockBasketController) ListStockBaskets(ctx *utils.Context) (interface{}, error) {
	query := services.ListStockBasketsQuery{
		Status: ctx.GetQuery("status"),
		Limit:  10,
		Offset: 0,
	}

	if limit := ctx.GetQuery("limit"); limit != "" {
		if parsedLimit := parseInt(limit); parsedLimit > 0 {
			query.Limit = parsedLimit
		}
	}
	if offset := ctx.GetQuery("offset"); offset != "" {
		if parsedOffset := parseInt(offset); parsedOffset >= 0 {
			query.Offset = parsedOffset
		}
	}

	result, err := c.listStockBasketsService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Baskets, nil
}

// UpdateStockBasket updates a stock basket
func (c *StockBasketController) UpdateStockBasket(ctx *utils.Context) (interface{}, error) {
	var params models.UpdateStockBasketParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateStockBasketRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	basketID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	cmd := services.UpdateStockBasketCmd{
		BasketID:    basketID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}

	result, err := c.updateStockBasketService.Execute(ctx.Request.Context(), cmd)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":      result.Basket.ID.String(),
		"message": "Stock basket updated successfully",
	}, nil
}
