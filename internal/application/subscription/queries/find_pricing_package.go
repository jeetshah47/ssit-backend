package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PricingPackageModel represents a pricing package from the database
type PricingPackageModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Name         string    `gorm:"type:varchar(255);not null"`
	DurationType string    `gorm:"type:varchar(50);not null"`
	Status       string    `gorm:"type:varchar(50);not null"`
	IsPublished  bool      `gorm:"default:false;not null"`
}

// TableName specifies the table name
func (PricingPackageModel) TableName() string {
	return "pricing_packages"
}

// FindPricingPackageQuery finds a pricing package by plan name and billing period
type FindPricingPackageQuery struct {
	PlanName      string // 'standard', 'plus', 'premium'
	BillingPeriod string // 'quarterly', 'annual'
}

// FindPricingPackageResult contains the found package
type FindPricingPackageResult struct {
	PackageID uuid.UUID
}

// FindPricingPackageHandler handles finding pricing packages
type FindPricingPackageHandler struct {
	db *gorm.DB
}

// NewFindPricingPackageHandler creates a new find pricing package handler
func NewFindPricingPackageHandler(db *gorm.DB) *FindPricingPackageHandler {
	return &FindPricingPackageHandler{db: db}
}

// Handle executes the query to find a pricing package
func (h *FindPricingPackageHandler) Handle(ctx context.Context, query FindPricingPackageQuery) (*FindPricingPackageResult, error) {
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
	var model PricingPackageModel
	if err := h.db.WithContext(ctx).
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
