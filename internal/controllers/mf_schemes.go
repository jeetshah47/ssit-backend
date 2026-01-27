package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// MFSchemeController handles MF scheme endpoints
type MFSchemeController struct {
	getMFSchemeService   *services.GetMFSchemeService
	listMFSchemesService *services.ListMFSchemesService
}

// NewMFSchemeController creates a new MF scheme controller
func NewMFSchemeController(
	getMFSchemeService *services.GetMFSchemeService,
	listMFSchemesService *services.ListMFSchemesService,
) *MFSchemeController {
	return &MFSchemeController{
		getMFSchemeService:   getMFSchemeService,
		listMFSchemesService: listMFSchemesService,
	}
}

// GetMFScheme retrieves an MF scheme by ID
func (c *MFSchemeController) GetMFScheme(ctx *utils.Context) (interface{}, error) {
	var params models.GetMFSchemeParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	schemeID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := services.GetMFSchemeQuery{
		SchemeID: schemeID,
	}

	result, err := c.getMFSchemeService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Scheme, nil
}

// ListMFSchemes lists MF schemes
func (c *MFSchemeController) ListMFSchemes(ctx *utils.Context) (interface{}, error) {
	query := services.ListMFSchemesQuery{
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

	result, err := c.listMFSchemesService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Schemes, nil
}
