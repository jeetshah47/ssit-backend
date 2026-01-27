package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NFO represents the database model for New Fund Offers
type NFO struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	Name            string     `gorm:"type:varchar(255);not null"`
	AMC             *string    `gorm:"type:varchar(255)"`
	Category        *string    `gorm:"type:varchar(100)"`
	Type            *string    `gorm:"type:varchar(50)"` // Equity, Debt, Hybrid, etc.
	OpenDate        *time.Time
	CloseDate       *time.Time
	Timeline        *string    `gorm:"type:text"`
	Summary         *string    `gorm:"type:text"` // JSON array of summary points
	MinimumInvestment *float64 `gorm:"type:decimal(10,2)"`
	Verdict          *string    `gorm:"type:varchar(50)"` // Subscribe, Wait, Avoid
	Rationale        *string    `gorm:"type:text"`
	ReportURL        *string    `gorm:"type:text"`
	Status           string     `gorm:"type:varchar(50);default:upcoming;not null"` // upcoming, open, closed
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (NFO) TableName() string {
	return "nfos"
}

// BeforeCreate hook to set UUID if not set
func (n *NFO) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// IsOpen checks if the NFO is currently open
func (n *NFO) IsOpen() bool {
	if n.Status != "open" {
		return false
	}
	now := time.Now()
	if n.OpenDate != nil && n.CloseDate != nil {
		return now.After(*n.OpenDate) && now.Before(*n.CloseDate)
	}
	return false
}
