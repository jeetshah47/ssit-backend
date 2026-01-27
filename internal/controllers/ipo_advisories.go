package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// IPOAdvisoryController handles IPO advisory endpoints
type IPOAdvisoryController struct {
	createIPOAdvisoryService *services.CreateIPOAdvisoryService
	getIPOAdvisoryService    *services.GetIPOAdvisoryService
	listIPOAdvisoriesService  *services.ListIPOAdvisoriesService
	updateIPOAdvisoryService *services.UpdateIPOAdvisoryService
}

// NewIPOAdvisoryController creates a new IPO advisory controller
func NewIPOAdvisoryController(
	createIPOAdvisoryService *services.CreateIPOAdvisoryService,
	getIPOAdvisoryService *services.GetIPOAdvisoryService,
	listIPOAdvisoriesService *services.ListIPOAdvisoriesService,
	updateIPOAdvisoryService *services.UpdateIPOAdvisoryService,
) *IPOAdvisoryController {
	return &IPOAdvisoryController{
		createIPOAdvisoryService: createIPOAdvisoryService,
		getIPOAdvisoryService:    getIPOAdvisoryService,
		listIPOAdvisoriesService:  listIPOAdvisoriesService,
		updateIPOAdvisoryService: updateIPOAdvisoryService,
	}
}

// CreateIPOAdvisory creates a new IPO advisory
func (c *IPOAdvisoryController) CreateIPOAdvisory(ctx *utils.Context) (interface{}, error) {
	var req models.CreateIPOAdvisoryRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(ctx.UserID)
	if err != nil {
		return nil, err
	}

	cmd := services.CreateIPOAdvisoryCmd{
		IPOName:      req.IPOName,
		IPOSymbol:    req.IPOSymbol,
		GMP:          req.GMP,
		Suggestion:   req.Suggestion,
		LotSize:      req.LotSize,
		PriceBandMin: req.PriceBandMin,
		PriceBandMax: req.PriceBandMax,
		IssueDate:    req.IssueDate,
		IssueSize:    req.IssueSize,
		IPOTimetable: req.IPOTimetable,
		PublishedBy:  userID,
	}

	result, err := c.createIPOAdvisoryService.Execute(ctx.Request.Context(), cmd)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":      result.Advisory.ID.String(),
		"message": "IPO advisory created successfully",
	}, nil
}

// GetIPOAdvisory retrieves an IPO advisory by ID
func (c *IPOAdvisoryController) GetIPOAdvisory(ctx *utils.Context) (interface{}, error) {
	var params models.GetIPOAdvisoryParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	advisoryID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	query := services.GetIPOAdvisoryQuery{
		AdvisoryID: advisoryID,
	}

	result, err := c.getIPOAdvisoryService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Advisory, nil
}

// ListIPOAdvisories lists IPO advisories
func (c *IPOAdvisoryController) ListIPOAdvisories(ctx *utils.Context) (interface{}, error) {
	query := services.ListIPOAdvisoriesQuery{
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

	result, err := c.listIPOAdvisoriesService.Execute(ctx.Request.Context(), query)
	if err != nil {
		return nil, err
	}

	return result.Advisories, nil
}

// UpdateIPOAdvisory updates an IPO advisory
func (c *IPOAdvisoryController) UpdateIPOAdvisory(ctx *utils.Context) (interface{}, error) {
	var params models.UpdateIPOAdvisoryParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	var req models.UpdateIPOAdvisoryRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	advisoryID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	cmd := services.UpdateIPOAdvisoryCmd{
		AdvisoryID:   advisoryID,
		IPOName:      req.IPOName,
		GMP:          req.GMP,
		Suggestion:   req.Suggestion,
		Status:       req.Status,
		IssueDate:    req.IssueDate,
		IPOTimetable: req.IPOTimetable,
	}

	result, err := c.updateIPOAdvisoryService.Execute(ctx.Request.Context(), cmd)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":      result.Advisory.ID.String(),
		"message": "IPO advisory updated successfully",
	}, nil
}
