package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MFScheme represents the database model for mutual fund schemes
type MFScheme struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	Name            string     `gorm:"type:varchar(255);not null"`
	SchemeCode      *string    `gorm:"type:varchar(50)"`
	AMC             *string    `gorm:"type:varchar(255)"`
	Category        *string    `gorm:"type:varchar(100)"`
	Type            *string    `gorm:"type:varchar(50)"` // Equity, Debt, Hybrid, etc.
	CurrentNAV      *float64   `gorm:"type:decimal(10,4)"`
	EntryPrice      *float64   `gorm:"type:decimal(10,4)"`
	ExitPrice       *float64   `gorm:"type:decimal(10,4)"`
	Trend            *string    `gorm:"type:varchar(50)"` // positive, negative, neutral
	Verdict          *string    `gorm:"type:varchar(50)"` // Buy, Hold, Sell
	Rationale        *string    `gorm:"type:text"`
	ReportURL        *string    `gorm:"type:text"`
	Status           string     `gorm:"type:varchar(50);default:active;not null"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (MFScheme) TableName() string {
	return "mf_schemes"
}

// BeforeCreate hook to set UUID if not set
func (m *MFScheme) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
