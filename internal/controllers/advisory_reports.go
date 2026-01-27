package controllers

import (
	"fmt"

	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/services"
	"github.com/google/uuid"
)

// AdvisoryReportController handles advisory report upload endpoints
type AdvisoryReportController struct {
	uploadReportService                    *services.UploadAdvisoryReportService
	deleteReportService                    *services.DeleteAdvisoryReportService
	updateStockBasketReportURLService      *services.UpdateStockBasketReportURLService
	updateIPOAdvisoryReportURLService      *services.UpdateIPOAdvisoryReportURLService
	updateMutualFundBasketReportURLService *services.UpdateMutualFundBasketReportURLService
}

// NewAdvisoryReportController creates a new advisory report controller
func NewAdvisoryReportController(
	uploadReportService *services.UploadAdvisoryReportService,
	deleteReportService *services.DeleteAdvisoryReportService,
	updateStockBasketReportURLService *services.UpdateStockBasketReportURLService,
	updateIPOAdvisoryReportURLService *services.UpdateIPOAdvisoryReportURLService,
	updateMutualFundBasketReportURLService *services.UpdateMutualFundBasketReportURLService,
) *AdvisoryReportController {
	return &AdvisoryReportController{
		uploadReportService:                    uploadReportService,
		deleteReportService:                    deleteReportService,
		updateStockBasketReportURLService:      updateStockBasketReportURLService,
		updateIPOAdvisoryReportURLService:      updateIPOAdvisoryReportURLService,
		updateMutualFundBasketReportURLService: updateMutualFundBasketReportURLService,
	}
}

// UploadStockBasketReport uploads a report for a stock basket
func (c *AdvisoryReportController) UploadStockBasketReport(ctx *utils.Context) (interface{}, error) {
	return c.uploadReport(ctx, "stock_basket")
}

// UploadETFReport uploads a report for an ETF basket
func (c *AdvisoryReportController) UploadETFReport(ctx *utils.Context) (interface{}, error) {
	return c.uploadReport(ctx, "etf_basket")
}

// UploadIPOReport uploads a report for an IPO advisory
func (c *AdvisoryReportController) UploadIPOReport(ctx *utils.Context) (interface{}, error) {
	return c.uploadReport(ctx, "ipo")
}

// UploadMutualFundReport uploads a report for a mutual fund basket
func (c *AdvisoryReportController) UploadMutualFundReport(ctx *utils.Context) (interface{}, error) {
	return c.uploadReport(ctx, "mutual_fund_basket")
}

// uploadReport handles the common upload logic
func (c *AdvisoryReportController) uploadReport(ctx *utils.Context, advisoryType string) (interface{}, error) {
	var params models.UploadReportParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	advisoryID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	// Get file from multipart form
	file, header, err := ctx.Request.FormFile("report")
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	defer file.Close()

	userID, err := uuid.Parse(ctx.UserID)
	if err != nil {
		return nil, err
	}

	cmd := services.UploadAdvisoryReportCmd{
		AdvisoryID:   advisoryID,
		AdvisoryType: advisoryType,
		File:         file,
		FileName:     header.Filename,
		FileSize:     header.Size,
		UploadedBy:   userID,
	}

	result, err := c.uploadReportService.Execute(ctx.Request.Context(), cmd)
	if err != nil {
		return nil, err
	}

	// Update the advisory record with report URL
	updateErr := c.updateAdvisoryReportURL(ctx, advisoryType, advisoryID, result.Result.ReportURL, result.Result.ReportFileName, userID)
	if updateErr != nil {
		// Log error but don't fail the upload
		// The file is already uploaded to S3, we can update the DB record later
	}

	return map[string]interface{}{
		"reportUrl":     result.Result.ReportURL,
		"reportFileName": result.Result.ReportFileName,
		"uploadedAt":    result.Result.UploadedAt,
	}, nil
}

// updateAdvisoryReportURL updates the advisory record with report URL
func (c *AdvisoryReportController) updateAdvisoryReportURL(ctx *utils.Context, advisoryType string, advisoryID uuid.UUID, reportURL, reportFileName string, userID uuid.UUID) error {
	switch advisoryType {
	case "stock_basket":
		cmd := services.UpdateStockBasketReportURLCmd{
			BasketID:        advisoryID,
			ReportURL:       reportURL,
			ReportFileName:  reportFileName,
			ReportUploadedBy: userID,
		}
		return c.updateStockBasketReportURLService.Execute(ctx.Request.Context(), cmd)
	case "ipo":
		cmd := services.UpdateIPOAdvisoryReportURLCmd{
			AdvisoryID:      advisoryID,
			ReportURL:       reportURL,
			ReportFileName:  reportFileName,
			ReportUploadedBy: userID,
		}
		return c.updateIPOAdvisoryReportURLService.Execute(ctx.Request.Context(), cmd)
	case "mutual_fund_basket":
		cmd := services.UpdateMutualFundBasketReportURLCmd{
			BasketID:         advisoryID,
			ReportURL:        reportURL,
			ReportFileName:   reportFileName,
			ReportUploadedBy: userID,
		}
		return c.updateMutualFundBasketReportURLService.Execute(ctx.Request.Context(), cmd)
	default:
		return fmt.Errorf("unsupported advisory type: %s", advisoryType)
	}
}

// DeleteStockBasketReport deletes a report from a stock basket
func (c *AdvisoryReportController) DeleteStockBasketReport(ctx *utils.Context) (interface{}, error) {
	return c.deleteReport(ctx, "stock_basket")
}

// DeleteETFReport deletes a report from an ETF basket
func (c *AdvisoryReportController) DeleteETFReport(ctx *utils.Context) (interface{}, error) {
	return c.deleteReport(ctx, "etf_basket")
}

// DeleteIPOReport deletes a report from an IPO advisory
func (c *AdvisoryReportController) DeleteIPOReport(ctx *utils.Context) (interface{}, error) {
	return c.deleteReport(ctx, "ipo")
}

// DeleteMutualFundReport deletes a report from a mutual fund basket
func (c *AdvisoryReportController) DeleteMutualFundReport(ctx *utils.Context) (interface{}, error) {
	return c.deleteReport(ctx, "mutual_fund_basket")
}

// deleteReport handles the common delete logic
func (c *AdvisoryReportController) deleteReport(ctx *utils.Context, advisoryType string) (interface{}, error) {
	var params models.DeleteReportParams
	if err := ctx.BindURI(&params); err != nil {
		return nil, err
	}

	advisoryID, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	// Get report URL from query or body
	reportURL := ctx.GetQuery("report_url")
	if reportURL == "" {
		var req models.DeleteReportRequest
		if err := ctx.BindJSON(&req); err == nil {
			reportURL = req.ReportURL
		}
	}

	if reportURL == "" {
		return nil, fmt.Errorf("report URL is required")
	}

	cmd := services.DeleteAdvisoryReportCmd{
		AdvisoryID:   advisoryID,
		AdvisoryType: advisoryType,
		ReportURL:    reportURL,
	}

	if err := c.deleteReportService.Execute(ctx.Request.Context(), cmd); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "Report deleted successfully",
	}, nil
}
