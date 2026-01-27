package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// NFOController handles NFO endpoints
type NFOController struct {
	getNFOService   *services.GetNFOService
	listNFOsService *services.ListNFOsService
}

// NewNFOController creates a new NFO controller
func NewNFOController(
	getNFOService *services.GetNFOService,
	listNFOsService *services.ListNFOsService,
) *NFOController {
	return &NFOController{
		getNFOService:   getNFOService,
		listNFOsService: listNFOsService,
	}
}

// GetNFO retrieves an NFO by ID
func (c *NFOController) GetNFO(ctx *utils.Context) (interface{}, error) {
	var params models.GetNFOParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	nfoID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := services.GetNFOQuery{
		NFOID: nfoID,
	}

	result, err := c.getNFOService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.NFO, nil
}

// ListNFOs lists NFOs
func (c *NFOController) ListNFOs(ctx *utils.Context) (interface{}, error) {
	query := services.ListNFOsQuery{
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

	result, err := c.listNFOsService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.NFOs, nil
}
