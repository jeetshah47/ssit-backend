package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// WebinarController handles webinar endpoints
type WebinarController struct {
	getWebinarService   *services.GetWebinarService
	listWebinarsService *services.ListWebinarsService
}

// NewWebinarController creates a new webinar controller
func NewWebinarController(
	getWebinarService *services.GetWebinarService,
	listWebinarsService *services.ListWebinarsService,
) *WebinarController {
	return &WebinarController{
		getWebinarService:   getWebinarService,
		listWebinarsService: listWebinarsService,
	}
}

// GetWebinar retrieves a webinar by ID
func (c *WebinarController) GetWebinar(ctx *utils.Context) (interface{}, error) {
	var params models.GetWebinarParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	webinarID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := services.GetWebinarQuery{
		WebinarID: webinarID,
	}

	result, err := c.getWebinarService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Webinar, nil
}

// ListWebinars lists webinars
func (c *WebinarController) ListWebinars(ctx *utils.Context) (interface{}, error) {
	query := services.ListWebinarsQuery{
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

	result, err := c.listWebinarsService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Webinars, nil
}
