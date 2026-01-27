package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subscription represents the database model for subscriptions
type Subscription struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key"`

	// User and Package
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	PackageID uuid.UUID `gorm:"type:uuid;not null"`

	// Pricing
	Price    float64 `gorm:"type:decimal(10,2);not null"`
	Currency string  `gorm:"type:varchar(3);not null;default:'INR'"`

	// Access Type
	AccessType  string  `gorm:"type:varchar(50);not null"`
	AccessReason *string `gorm:"type:text"`

	// Validity
	StartsAt time.Time `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsActive bool      `gorm:"default:true;not null"`

	// Status
	Status string `gorm:"type:varchar(50);not null;default:'active'"`

	// Payment Reference
	PaymentID *uuid.UUID `gorm:"type:uuid"`

	// Metadata
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	CancelledAt *time.Time
}

// TableName specifies the table name
func (Subscription) TableName() string {
	return "subscriptions"
}

// BeforeCreate hook to set UUID if not set
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsActiveSubscription returns true if subscription is currently active
func (s *Subscription) IsActiveSubscription() bool {
	now := time.Now()
	return s.IsActive && s.Status == "active" && now.After(s.StartsAt) && now.Before(s.ExpiresAt)
}

// Activate marks the subscription as active
func (s *Subscription) Activate() {
	s.Status = "active"
	s.IsActive = true
	s.UpdatedAt = time.Now()
}

// Cancel marks the subscription as cancelled
func (s *Subscription) Cancel() {
	s.Status = "cancelled"
	s.IsActive = false
	now := time.Now()
	s.CancelledAt = &now
	s.UpdatedAt = now
}

