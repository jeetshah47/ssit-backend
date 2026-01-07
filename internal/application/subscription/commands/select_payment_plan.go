package commands

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/application/subscription/queries"
	"github.com/equitywala/backend/internal/domain/subscription"
	"github.com/google/uuid"
)

// SelectPaymentPlanCommand represents the command
type SelectPaymentPlanCommand struct {
	UserID        uuid.UUID
	PlanName      string // 'standard', 'plus', 'premium'
	BillingPeriod string // 'quarterly', 'annual'
}

// SelectPaymentPlanHandler handles payment plan selection
type SelectPaymentPlanHandler struct {
	selectionRepo subscription.PaymentPlanSelectionRepository
	packageQuery  *queries.FindPricingPackageHandler
}

// NewSelectPaymentPlanHandler creates a new select payment plan handler
func NewSelectPaymentPlanHandler(
	selectionRepo subscription.PaymentPlanSelectionRepository,
	packageQuery *queries.FindPricingPackageHandler,
) *SelectPaymentPlanHandler {
	return &SelectPaymentPlanHandler{
		selectionRepo: selectionRepo,
		packageQuery:  packageQuery,
	}
}

// Handle executes the command
func (h *SelectPaymentPlanHandler) Handle(ctx context.Context, cmd SelectPaymentPlanCommand) error {
	// 1. Find pricing package
	packageResult, err := h.packageQuery.Handle(ctx, queries.FindPricingPackageQuery{
		PlanName:      cmd.PlanName,
		BillingPeriod: cmd.BillingPeriod,
	})
	if err != nil {
		return fmt.Errorf("failed to find pricing package: %w", err)
	}

	// 2. Check if user already has a selection (delete old one if exists)
	existing, err := h.selectionRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to check existing selection: %w", err)
	}
	if existing != nil {
		if err := h.selectionRepo.Delete(ctx, cmd.UserID); err != nil {
			return fmt.Errorf("failed to delete existing selection: %w", err)
		}
	}

	// 3. Create new selection
	selection := subscription.NewPaymentPlanSelection(cmd.UserID, packageResult.PackageID)

	// 4. Save selection
	if err := h.selectionRepo.Create(ctx, selection); err != nil {
		return fmt.Errorf("failed to create payment plan selection: %w", err)
	}

	return nil
}

