package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentPlanSelection represents the database model for payment plan selections
type PaymentPlanSelection struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	PackageID  uuid.UUID `gorm:"type:uuid;not null"`
	SelectedAt time.Time `gorm:"not null"`
	Status     string    `gorm:"type:varchar(50);not null;default:pending"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (PaymentPlanSelection) TableName() string {
	return "user_payment_plan_selections"
}

// BeforeCreate hook to set UUID if not set
func (p *PaymentPlanSelection) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
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

// PricingPackageModel represents a pricing package from the database
type PricingPackageModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Name         string    `gorm:"type:varchar(255);not null"`
	Description  *string   `gorm:"type:text"`
	Price        float64   `gorm:"type:decimal(10,2);not null"`
	Currency     string    `gorm:"type:varchar(3);not null;default:'INR'"`
	DurationDays int       `gorm:"type:integer;not null"`
	DurationType string    `gorm:"type:varchar(50);not null"`
	AccessLevel  string    `gorm:"type:varchar(50);not null;default:'standard'"`
	Status       string    `gorm:"type:varchar(50);not null"`
	IsPublished  bool      `gorm:"default:false;not null"`
}

// TableName specifies the table name
func (PricingPackageModel) TableName() string {
	return "pricing_packages"
}

