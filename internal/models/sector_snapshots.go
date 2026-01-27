package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SectorSnapshot represents the database model for sector snapshots
type SectorSnapshot struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	SectorName      string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	CurrentValue    *float64   `gorm:"type:decimal(10,2)"`
	ChangePercentage *float64   `gorm:"type:decimal(5,2)"`
	ChangeValue     *float64   `gorm:"type:decimal(10,2)"`
	ReportURL       *string    `gorm:"type:text"`
	LastUpdated     *time.Time
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (SectorSnapshot) TableName() string {
	return "sector_snapshots"
}

// BeforeCreate hook to set UUID if not set
func (s *SectorSnapshot) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
