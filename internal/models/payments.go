package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Payment represents the database model for payments
type Payment struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key"`

	// User and Subscription
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	SubscriptionID *uuid.UUID `gorm:"type:uuid"`

	// Amount
	Amount   float64 `gorm:"type:decimal(10,2);not null"`
	Currency string  `gorm:"type:varchar(3);not null;default:'INR'"`

	// Payment Method
	PaymentMethod string  `gorm:"type:varchar(50);not null"`
	PaymentGateway *string `gorm:"type:varchar(100)"`

	// Gateway Information
	GatewayTransactionID *string `gorm:"type:varchar(255)"`
	GatewayOrderID       *string `gorm:"type:varchar(255)"`
	GatewayPaymentID     *string `gorm:"type:varchar(255)"`

	// Status
	Status string `gorm:"type:varchar(50);not null;default:'pending'"`

	// Payment Details
	PaymentData  *PaymentData `gorm:"type:jsonb"`
	FailureReason *string     `gorm:"type:text"`

	// Refund Information
	RefundAmount       *float64   `gorm:"type:decimal(10,2)"`
	RefundTransactionID *string   `gorm:"type:varchar(255)"`
	RefundedBy         *uuid.UUID `gorm:"type:uuid"`
	RefundedAt         *time.Time
	RefundReason       *string `gorm:"type:text"`

	// Idempotency
	IdempotencyKey *string `gorm:"type:varchar(255);uniqueIndex"`

	// Metadata
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	CompletedAt *time.Time
}

// TableName specifies the table name
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate hook to set UUID if not set
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PaymentData represents additional payment data stored as JSONB
type PaymentData map[string]interface{}

// Value implements the driver.Valuer interface
func (pd PaymentData) Value() (driver.Value, error) {
	if pd == nil {
		return nil, nil
	}
	return json.Marshal(pd)
}

// Scan implements the sql.Scanner interface
func (pd *PaymentData) Scan(value interface{}) error {
	if value == nil {
		*pd = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), pd)
	}

	return json.Unmarshal(bytes, pd)
}

// State machine methods

// MarkAsCreated marks the payment as created
func (p *Payment) MarkAsCreated() {
	p.Status = "pending"
	p.UpdatedAt = time.Now()
}

// MarkAsAttempted marks the payment as attempted
func (p *Payment) MarkAsAttempted() {
	p.Status = "processing"
	p.UpdatedAt = time.Now()
}

// MarkAsVerified marks the payment as verified
func (p *Payment) MarkAsVerified() {
	p.Status = "success"
	p.UpdatedAt = time.Now()
	now := time.Now()
	p.CompletedAt = &now
}

// MarkAsConfirmed marks the payment as confirmed (via webhook)
func (p *Payment) MarkAsConfirmed() {
	p.Status = "success"
	p.UpdatedAt = time.Now()
	if p.CompletedAt == nil {
		now := time.Now()
		p.CompletedAt = &now
	}
}

// MarkAsFailed marks the payment as failed
func (p *Payment) MarkAsFailed(reason string) {
	p.Status = "failed"
	p.FailureReason = &reason
	p.UpdatedAt = time.Now()
}

// MarkAsCancelled marks the payment as cancelled
func (p *Payment) MarkAsCancelled() {
	p.Status = "cancelled"
	p.UpdatedAt = time.Now()
}

// IsSuccess returns true if payment is successful
func (p *Payment) IsSuccess() bool {
	return p.Status == "success"
}

// IsPending returns true if payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == "pending" || p.Status == "processing"
}

// IsFailed returns true if payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == "failed" || p.Status == "cancelled"
}

// CanTransitionToVerified returns true if payment can transition to verified state
func (p *Payment) CanTransitionToVerified() bool {
	return p.Status == "pending" || p.Status == "processing"
}

// CanTransitionToConfirmed returns true if payment can transition to confirmed state
func (p *Payment) CanTransitionToConfirmed() bool {
	return p.Status == "pending" || p.Status == "processing" || p.Status == "success"
}

// VerifyPaymentResponse represents the response for payment verification
type VerifyPaymentResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// WebhookResponse represents the response for webhook handling
type WebhookResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

