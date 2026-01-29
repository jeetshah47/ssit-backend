package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Stock represents a master list entry (used for selection in bullets/recommendations)
type Stock struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Symbol    *string   `gorm:"type:varchar(50)" json:"symbol,omitempty"`
	Exchange  *string   `gorm:"type:varchar(50)" json:"exchange,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName specifies the table name
func (Stock) TableName() string {
	return "stocks"
}

// BeforeCreate hook to set UUID if not set
func (s *Stock) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
