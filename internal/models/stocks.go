package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Stock represents the master list of stocks
type Stock struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null;index" json:"name"`
	Symbol    string    `gorm:"type:varchar(50);not null;index" json:"symbol"`
	Exchange  string    `gorm:"type:varchar(20);not null" json:"exchange"`
	Sector    *string   `gorm:"type:varchar(100);index" json:"sector,omitempty"`
	IsActive  bool      `gorm:"default:true;not null" json:"isActive"`
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

