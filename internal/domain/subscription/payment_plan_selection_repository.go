package subscription

import (
	"context"

	"github.com/google/uuid"
)

// PaymentPlanSelectionRepository handles payment plan selection persistence
type PaymentPlanSelectionRepository interface {
	Create(ctx context.Context, selection *PaymentPlanSelection) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*PaymentPlanSelection, error)
	Update(ctx context.Context, selection *PaymentPlanSelection) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

