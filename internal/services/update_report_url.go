package services

import (
	"context"
	"fmt"
	"time"

	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// UpdateStockBasketReportURLCmd represents the command for updating a stock basket report URL
type UpdateStockBasketReportURLCmd struct {
	BasketID        uuid.UUID
	ReportURL       string
	ReportFileName  string
	ReportUploadedBy uuid.UUID
}

// UpdateStockBasketReportURLService handles updating report URL in stock basket
type UpdateStockBasketReportURLService struct {
	stockBasketRepo repositories.StockBasketRepo
}

// NewUpdateStockBasketReportURLService creates a new update stock basket report URL service
func NewUpdateStockBasketReportURLService(stockBasketRepo repositories.StockBasketRepo) *UpdateStockBasketReportURLService {
	return &UpdateStockBasketReportURLService{
		stockBasketRepo: stockBasketRepo,
	}
}

// Execute executes the UpdateStockBasketReportURL command
func (s *UpdateStockBasketReportURLService) Execute(ctx context.Context, cmd UpdateStockBasketReportURLCmd) error {
	basket, err := s.stockBasketRepo.FindByID(ctx, cmd.BasketID)
	if err != nil {
		return err
	}

	now := time.Now()
	basket.ReportURL = &cmd.ReportURL
	basket.ReportFileName = &cmd.ReportFileName
	basket.ReportUploadedAt = &now
	basket.ReportUploadedBy = &cmd.ReportUploadedBy

	if err := s.stockBasketRepo.Update(ctx, basket); err != nil {
		return fmt.Errorf("failed to update stock basket report URL: %w", err)
	}

	return nil
}

// UpdateIPOAdvisoryReportURLCmd represents the command for updating an IPO advisory report URL
type UpdateIPOAdvisoryReportURLCmd struct {
	AdvisoryID      uuid.UUID
	ReportURL       string
	ReportFileName  string
	ReportUploadedBy uuid.UUID
}

// UpdateIPOAdvisoryReportURLService handles updating report URL in IPO advisory
type UpdateIPOAdvisoryReportURLService struct {
	ipoAdvisoryRepo repositories.IPOAdvisoryRepo
}

// NewUpdateIPOAdvisoryReportURLService creates a new update IPO advisory report URL service
func NewUpdateIPOAdvisoryReportURLService(ipoAdvisoryRepo repositories.IPOAdvisoryRepo) *UpdateIPOAdvisoryReportURLService {
	return &UpdateIPOAdvisoryReportURLService{
		ipoAdvisoryRepo: ipoAdvisoryRepo,
	}
}

// Execute executes the UpdateIPOAdvisoryReportURL command
func (s *UpdateIPOAdvisoryReportURLService) Execute(ctx context.Context, cmd UpdateIPOAdvisoryReportURLCmd) error {
	advisory, err := s.ipoAdvisoryRepo.FindByID(ctx, cmd.AdvisoryID)
	if err != nil {
		return err
	}

	now := time.Now()
	advisory.ReportURL = &cmd.ReportURL
	advisory.ReportFileName = &cmd.ReportFileName
	advisory.ReportUploadedAt = &now
	advisory.ReportUploadedBy = &cmd.ReportUploadedBy

	if err := s.ipoAdvisoryRepo.Update(ctx, advisory); err != nil {
		return fmt.Errorf("failed to update IPO advisory report URL: %w", err)
	}

	return nil
}

// UpdateMutualFundBasketReportURLCmd represents the command for updating a mutual fund basket report URL
type UpdateMutualFundBasketReportURLCmd struct {
	BasketID         uuid.UUID
	ReportURL        string
	ReportFileName   string
	ReportUploadedBy uuid.UUID
}

// UpdateMutualFundBasketReportURLService handles updating report URL in mutual fund basket
type UpdateMutualFundBasketReportURLService struct {
	mutualFundBasketRepo repositories.MutualFundBasketRepo
}

// NewUpdateMutualFundBasketReportURLService creates a new update mutual fund basket report URL service
func NewUpdateMutualFundBasketReportURLService(mutualFundBasketRepo repositories.MutualFundBasketRepo) *UpdateMutualFundBasketReportURLService {
	return &UpdateMutualFundBasketReportURLService{
		mutualFundBasketRepo: mutualFundBasketRepo,
	}
}

// Execute executes the UpdateMutualFundBasketReportURL command
func (s *UpdateMutualFundBasketReportURLService) Execute(ctx context.Context, cmd UpdateMutualFundBasketReportURLCmd) error {
	basket, err := s.mutualFundBasketRepo.FindByID(ctx, cmd.BasketID)
	if err != nil {
		return err
	}

	now := time.Now()
	basket.ReportURL = &cmd.ReportURL
	basket.ReportFileName = &cmd.ReportFileName
	basket.ReportUploadedAt = &now
	basket.ReportUploadedBy = &cmd.ReportUploadedBy

	if err := s.mutualFundBasketRepo.Update(ctx, basket); err != nil {
		return fmt.Errorf("failed to update mutual fund basket report URL: %w", err)
	}

	return nil
}
