package subscription

import (
	"time"

	"github.com/google/uuid"
)

// PaymentPlanSelection represents a user's payment plan selection during signup
type PaymentPlanSelection struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	PackageID  uuid.UUID
	SelectedAt time.Time
	Status     string // 'pending', 'confirmed', 'cancelled'
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewPaymentPlanSelection creates a new payment plan selection
func NewPaymentPlanSelection(userID, packageID uuid.UUID) *PaymentPlanSelection {
	now := time.Now()
	return &PaymentPlanSelection{
		ID:         uuid.New(),
		UserID:     userID,
		PackageID:  packageID,
		SelectedAt: now,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Confirm marks the selection as confirmed
func (p *PaymentPlanSelection) Confirm() {
	p.Status = "confirmed"
	p.UpdatedAt = time.Now()
}

// Cancel marks the selection as cancelled
func (p *PaymentPlanSelection) Cancel() {
	p.Status = "cancelled"
	p.UpdatedAt = time.Now()
}

