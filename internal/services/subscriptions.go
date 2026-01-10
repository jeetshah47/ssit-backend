package services

import (
	"context"
	"fmt"
	"time"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FindPricingPackageQuery finds a pricing package by plan name and billing period
type FindPricingPackageQuery struct {
	PlanName      string // 'standard', 'plus', 'premium'
	BillingPeriod string // 'quarterly', 'annual'
}

// FindPricingPackageResult contains the found package
type FindPricingPackageResult struct {
	PackageID uuid.UUID
}

// FindPricingPackageService handles finding pricing packages
type FindPricingPackageService struct {
	db *gorm.DB
}

// NewFindPricingPackageService creates a new find pricing package service
func NewFindPricingPackageService(db *gorm.DB) *FindPricingPackageService {
	return &FindPricingPackageService{db: db}
}

// Execute executes the query to find a pricing package
func (s *FindPricingPackageService) Execute(ctx context.Context, query FindPricingPackageQuery) (*FindPricingPackageResult, error) {
	// Map plan name to package name (capitalize first letter for matching)
	// Plan names: 'standard', 'plus', 'premium' -> Package names: 'Standard', 'Plus', 'Premium'
	var packageName string
	switch query.PlanName {
	case "standard":
		packageName = "Standard"
	case "plus":
		packageName = "Plus"
	case "premium":
		packageName = "Premium"
	default:
		return nil, fmt.Errorf("invalid plan name: %s", query.PlanName)
	}

	// Map billing period to duration_type
	var durationType string
	switch query.BillingPeriod {
	case "quarterly":
		durationType = "quarterly"
	case "annual":
		durationType = "annual"
	default:
		return nil, fmt.Errorf("invalid billing period: %s", query.BillingPeriod)
	}

	// Lookup pricing package by name and duration_type
	var model models.PricingPackageModel
	if err := s.db.WithContext(ctx).
		Where("LOWER(name) = LOWER(?) AND duration_type = ? AND status = ? AND is_published = ?",
			packageName, durationType, "active", true).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pricing package not found for plan '%s' with billing period '%s'", query.PlanName, query.BillingPeriod)
		}
		return nil, fmt.Errorf("failed to find pricing package: %w", err)
	}

	return &FindPricingPackageResult{
		PackageID: model.ID,
	}, nil
}

// SelectPaymentPlanCmd represents the command for selecting a payment plan
type SelectPaymentPlanCmd struct {
	UserID        uuid.UUID
	PlanName      string // 'standard', 'plus', 'premium'
	BillingPeriod string // 'quarterly', 'annual'
}

// SelectPaymentPlanCmdOutputData represents the output of SelectPaymentPlan command
type SelectPaymentPlanCmdOutputData struct {
	Selection *models.PaymentPlanSelection
}

// SelectPaymentPlanService handles payment plan selection
type SelectPaymentPlanService struct {
	selectionRepo  repositories.PaymentPlanSelectionRepo
	packageService *FindPricingPackageService
}

// NewSelectPaymentPlanService creates a new select payment plan service
func NewSelectPaymentPlanService(
	selectionRepo repositories.PaymentPlanSelectionRepo,
	packageService *FindPricingPackageService,
) *SelectPaymentPlanService {
	return &SelectPaymentPlanService{
		selectionRepo:  selectionRepo,
		packageService: packageService,
	}
}

// Execute executes the SelectPaymentPlan command
func (s *SelectPaymentPlanService) Execute(ctx context.Context, cmd SelectPaymentPlanCmd) (*SelectPaymentPlanCmdOutputData, error) {
	// 1. Find pricing package
	packageResult, err := s.packageService.Execute(ctx, FindPricingPackageQuery{
		PlanName:      cmd.PlanName,
		BillingPeriod: cmd.BillingPeriod,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find pricing package: %w", err)
	}

	// 2. Check if user already has a selection (delete old one if exists)
	existing, err := s.selectionRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing selection: %w", err)
	}
	if existing != nil {
		if err := s.selectionRepo.Delete(ctx, cmd.UserID); err != nil {
			return nil, fmt.Errorf("failed to delete existing selection: %w", err)
		}
	}

	// 3. Create new selection
	now := time.Now()
	selection := &models.PaymentPlanSelection{
		ID:         uuid.New(),
		UserID:     cmd.UserID,
		PackageID:  packageResult.PackageID,
		SelectedAt: now,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 4. Save selection
	if err := s.selectionRepo.Create(ctx, selection); err != nil {
		return nil, fmt.Errorf("failed to create payment plan selection: %w", err)
	}

	return &SelectPaymentPlanCmdOutputData{Selection: selection}, nil
}

// CreateSubscriptionCmd represents the command for creating a subscription
type CreateSubscriptionCmd struct {
	UserID       uuid.UUID
	PackageID    uuid.UUID
	PaymentID    *uuid.UUID
	Price        float64
	Currency     string
	DurationDays int
}

// CreateSubscriptionResult represents the result of creating a subscription
type CreateSubscriptionResult struct {
	Subscription *models.Subscription
}

// CreateSubscriptionService handles subscription creation
type CreateSubscriptionService struct {
	db *gorm.DB
}

// NewCreateSubscriptionService creates a new create subscription service
func NewCreateSubscriptionService(db *gorm.DB) *CreateSubscriptionService {
	return &CreateSubscriptionService{db: db}
}

// Execute creates a subscription
func (s *CreateSubscriptionService) Execute(ctx context.Context, cmd CreateSubscriptionCmd) (*CreateSubscriptionResult, error) {
	now := time.Now()
	expiresAt := now.AddDate(0, 0, cmd.DurationDays)

	subscription := &models.Subscription{
		ID:         uuid.New(),
		UserID:     cmd.UserID,
		PackageID:  cmd.PackageID,
		PaymentID:  cmd.PaymentID,
		Price:      cmd.Price,
		Currency:   cmd.Currency,
		AccessType: "paid",
		StartsAt:   now,
		ExpiresAt:  expiresAt,
		IsActive:   true,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.db.WithContext(ctx).Create(subscription).Error; err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return &CreateSubscriptionResult{Subscription: subscription}, nil
}

// ListPricingPackagesService handles listing all active pricing packages
type ListPricingPackagesService struct {
	db *gorm.DB
}

// NewListPricingPackagesService creates a new list pricing packages service
func NewListPricingPackagesService(db *gorm.DB) *ListPricingPackagesService {
	return &ListPricingPackagesService{db: db}
}

// PricingPackageResponse represents a pricing package response
type PricingPackageResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	DurationDays int     `json:"durationDays"`
	DurationType string  `json:"durationType"` // 'quarterly', 'annual'
	AccessLevel  string  `json:"accessLevel"`
	Status       string  `json:"status"`
	IsPublished  bool    `json:"isPublished"`
}

// Execute lists all active and published pricing packages
func (s *ListPricingPackagesService) Execute(ctx context.Context) ([]*PricingPackageResponse, error) {
	var packages []models.PricingPackageModel
	
	// Fetch all active and published pricing packages
	if err := s.db.WithContext(ctx).
		Where("status = ? AND is_published = ?", "active", true).
		Order("name ASC, duration_type ASC").
		Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("failed to list pricing packages: %w", err)
	}

	// Convert to response format
	responses := make([]*PricingPackageResponse, 0, len(packages))
	for _, pkg := range packages {
		responses = append(responses, &PricingPackageResponse{
			ID:           pkg.ID.String(),
			Name:         pkg.Name,
			Description:  pkg.Description,
			Price:        pkg.Price,
			Currency:     pkg.Currency,
			DurationDays: pkg.DurationDays,
			DurationType: pkg.DurationType,
			AccessLevel:  pkg.AccessLevel,
			Status:       pkg.Status,
			IsPublished:  pkg.IsPublished,
		})
	}

	return responses, nil
}

// ExecuteWithTx creates a subscription within a transaction
func (s *CreateSubscriptionService) ExecuteWithTx(ctx context.Context, tx *gorm.DB, cmd CreateSubscriptionCmd) (*CreateSubscriptionResult, error) {
	now := time.Now()
	expiresAt := now.AddDate(0, 0, cmd.DurationDays)

	subscription := &models.Subscription{
		ID:         uuid.New(),
		UserID:     cmd.UserID,
		PackageID:  cmd.PackageID,
		PaymentID:  cmd.PaymentID,
		Price:      cmd.Price,
		Currency:   cmd.Currency,
		AccessType: "paid",
		StartsAt:   now,
		ExpiresAt:  expiresAt,
		IsActive:   true,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := tx.WithContext(ctx).Create(subscription).Error; err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return &CreateSubscriptionResult{Subscription: subscription}, nil
}