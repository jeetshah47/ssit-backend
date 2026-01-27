package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentPlanSelectionRepo defines the interface for payment plan selection data persistence
type PaymentPlanSelectionRepo interface {
	// Create creates a new payment plan selection
	Create(ctx context.Context, selection *models.PaymentPlanSelection) error

	// FindByUserID finds a payment plan selection by user ID
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.PaymentPlanSelection, error)

	// Update updates an existing payment plan selection
	Update(ctx context.Context, selection *models.PaymentPlanSelection) error

	// Delete deletes a payment plan selection by user ID
	Delete(ctx context.Context, userID uuid.UUID) error
}

// paymentPlanSelectionRepo implements the payment plan selection repository interface
type paymentPlanSelectionRepo struct {
	*RepoContext
}

// NewPaymentPlanSelectionRepo creates a new payment plan selection repository
func NewPaymentPlanSelectionRepo(ctx *RepoContext) PaymentPlanSelectionRepo {
	return &paymentPlanSelectionRepo{RepoContext: ctx}
}

// Create creates a new payment plan selection
func (r *paymentPlanSelectionRepo) Create(ctx context.Context, selection *models.PaymentPlanSelection) error {
	if err := r.db.WithContext(ctx).Create(selection).Error; err != nil {
		return fmt.Errorf("failed to create payment plan selection: %w", err)
	}
	return nil
}

// FindByUserID finds a payment plan selection by user ID
func (r *paymentPlanSelectionRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.PaymentPlanSelection, error) {
	var selection models.PaymentPlanSelection
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&selection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, not error, if not found
		}
		return nil, fmt.Errorf("failed to find payment plan selection: %w", err)
	}
	return &selection, nil
}

// Update updates an existing payment plan selection
func (r *paymentPlanSelectionRepo) Update(ctx context.Context, selection *models.PaymentPlanSelection) error {
	if err := r.db.WithContext(ctx).Save(selection).Error; err != nil {
		return fmt.Errorf("failed to update payment plan selection: %w", err)
	}
	return nil
}

// Delete deletes a payment plan selection by user ID
func (r *paymentPlanSelectionRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.PaymentPlanSelection{}).Error; err != nil {
		return fmt.Errorf("failed to delete payment plan selection: %w", err)
	}
	return nil
}

