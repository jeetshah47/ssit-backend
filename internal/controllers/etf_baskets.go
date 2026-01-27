package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// ETFBasketController handles ETF basket endpoints
type ETFBasketController struct {
	getETFBasketService   *services.GetETFBasketService
	listETFBasketsService *services.ListETFBasketsService
}

// NewETFBasketController creates a new ETF basket controller
func NewETFBasketController(
	getETFBasketService *services.GetETFBasketService,
	listETFBasketsService *services.ListETFBasketsService,
) *ETFBasketController {
	return &ETFBasketController{
		getETFBasketService:   getETFBasketService,
		listETFBasketsService: listETFBasketsService,
	}
}

// GetETFBasket retrieves an ETF basket by ID
func (c *ETFBasketController) GetETFBasket(ctx *utils.Context) (interface{}, error) {
	var params models.GetETFBasketParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	basketID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := services.GetETFBasketQuery{
		BasketID: basketID,
	}

	result, err := c.getETFBasketService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Basket, nil
}

// ListETFBaskets lists ETF baskets
func (c *ETFBasketController) ListETFBaskets(ctx *utils.Context) (interface{}, error) {
	query := services.ListETFBasketsQuery{
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

	result, err := c.listETFBasketsService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Baskets, nil
}
