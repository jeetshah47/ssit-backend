package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/equitywala/backend/internal/domain/subscription"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentPlanSelectionModel represents the database model for payment plan selections
type PaymentPlanSelectionModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	PackageID  uuid.UUID `gorm:"type:uuid;not null"`
	SelectedAt time.Time `gorm:"not null"`
	Status     string    `gorm:"type:varchar(50);not null;default:pending"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (PaymentPlanSelectionModel) TableName() string {
	return "user_payment_plan_selections"
}

// BeforeCreate hook to set UUID if not set
func (m *PaymentPlanSelectionModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// Repository implements the PaymentPlanSelectionRepository interface
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new payment plan selection repository
func NewRepository(db *gorm.DB) subscription.PaymentPlanSelectionRepository {
	return &Repository{db: db}
}

// Create creates a new payment plan selection
func (r *Repository) Create(ctx context.Context, selection *subscription.PaymentPlanSelection) error {
	model := toModel(selection)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create payment plan selection: %w", err)
	}
	*selection = *toEntity(model)
	return nil
}

// FindByUserID finds a payment plan selection by user ID
func (r *Repository) FindByUserID(ctx context.Context, userID uuid.UUID) (*subscription.PaymentPlanSelection, error) {
	var model PaymentPlanSelectionModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, not error, if not found
		}
		return nil, fmt.Errorf("failed to find payment plan selection: %w", err)
	}
	return toEntity(&model), nil
}

// Update updates an existing payment plan selection
func (r *Repository) Update(ctx context.Context, selection *subscription.PaymentPlanSelection) error {
	model := toModel(selection)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update payment plan selection: %w", err)
	}
	*selection = *toEntity(model)
	return nil
}

// Delete deletes a payment plan selection by user ID
func (r *Repository) Delete(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&PaymentPlanSelectionModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete payment plan selection: %w", err)
	}
	return nil
}

// toEntity converts database model to domain entity
func toEntity(m *PaymentPlanSelectionModel) *subscription.PaymentPlanSelection {
	return &subscription.PaymentPlanSelection{
		ID:         m.ID,
		UserID:     m.UserID,
		PackageID:  m.PackageID,
		SelectedAt: m.SelectedAt,
		Status:     m.Status,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// toModel converts domain entity to database model
func toModel(s *subscription.PaymentPlanSelection) *PaymentPlanSelectionModel {
	return &PaymentPlanSelectionModel{
		ID:         s.ID,
		UserID:     s.UserID,
		PackageID:  s.PackageID,
		SelectedAt: s.SelectedAt,
		Status:     s.Status,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

