package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdvisoryType represents the database model for advisory types
type AdvisoryType struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	DisplayName string    `gorm:"type:varchar(255);not null"`
	Description *string   `gorm:"type:text"`
	IsActive    bool      `gorm:"default:true;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (AdvisoryType) TableName() string {
	return "advisory_types"
}

// BeforeCreate hook to set UUID if not set
func (a *AdvisoryType) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
