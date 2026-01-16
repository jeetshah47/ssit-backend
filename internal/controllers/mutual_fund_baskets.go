package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// MutualFundBasketController handles mutual fund basket endpoints
type MutualFundBasketController struct {
	createMutualFundBasketService *services.CreateMutualFundBasketService
	getMutualFundBasketService   *services.GetMutualFundBasketService
	listMutualFundBasketsService *services.ListMutualFundBasketsService
}

// NewMutualFundBasketController creates a new mutual fund basket controller
func NewMutualFundBasketController(
	createMutualFundBasketService *services.CreateMutualFundBasketService,
	getMutualFundBasketService *services.GetMutualFundBasketService,
	listMutualFundBasketsService *services.ListMutualFundBasketsService,
) *MutualFundBasketController {
	return &MutualFundBasketController{
		createMutualFundBasketService: createMutualFundBasketService,
		getMutualFundBasketService:   getMutualFundBasketService,
		listMutualFundBasketsService: listMutualFundBasketsService,
	}
}

// CreateMutualFundBasket creates a new mutual fund basket
func (c *MutualFundBasketController) CreateMutualFundBasket(ctx *utils.Context) (interface{}, error) {
	var req models.CreateMutualFundBasketRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Convert request to command
	items := make([]*services.CreateMutualFundBasketItemCmd, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &services.CreateMutualFundBasketItemCmd{
			SchemeName:   item.SchemeName,
			SchemeCode:   item.SchemeCode,
			EntryPrice:   item.EntryPrice,
			ExitPrice:    item.ExitPrice,
			CurrentNAV:   item.CurrentNAV,
			Trend:        item.Trend,
			DisplayOrder: item.DisplayOrder,
		})
	}

	userID, err := uuid.Parse(ctx.UserID)
	if err != nil {
		return nil, err
	}

	cmd := services.CreateMutualFundBasketCmd{
		BasketType:  req.BasketType,
		Name:        req.Name,
		Description: req.Description,
		MaxSchemes:  req.MaxSchemes,
		Items:       items,
		PublishedBy: userID,
	}

	result, err := c.createMutualFundBasketService.Execute(ctx.Request.Context(), cmd)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":      result.Basket.ID.String(),
		"message": "Mutual fund basket created successfully",
	}, nil
}

// GetMutualFundBasket retrieves a mutual fund basket by ID
func (c *MutualFundBasketController) GetMutualFundBasket(ctx *utils.Context) (interface{}, error) {
	var params models.GetMutualFundBasketParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	basketID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := services.GetMutualFundBasketQuery{
		BasketID: basketID,
	}

	result, err := c.getMutualFundBasketService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Basket, nil
}

// ListMutualFundBaskets lists mutual fund baskets
func (c *MutualFundBasketController) ListMutualFundBaskets(ctx *utils.Context) (interface{}, error) {
	query := services.ListMutualFundBasketsQuery{
		BasketType: ctx.GetQuery("basket_type"),
		Status:     ctx.GetQuery("status"),
		Limit:      10,
		Offset:     0,
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

	result, err := c.listMutualFundBasketsService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Baskets, nil
}
