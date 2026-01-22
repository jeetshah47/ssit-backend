package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WeeklyAudio represents the database model for weekly audio content
type WeeklyAudio struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	Title           string     `gorm:"type:varchar(255);not null"`
	Description     *string    `gorm:"type:text"`
	Duration         string     `gorm:"type:varchar(50);not null"` // e.g., "45:30"
	Week             string     `gorm:"type:varchar(100);not null"` // e.g., "Week of Oct 21-27, 2024"
	AudioURL         *string    `gorm:"type:text"`
	ThumbnailURL     *string    `gorm:"type:text"`
	PostedAt         *time.Time
	Status           string     `gorm:"type:varchar(50);default:published;not null"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (WeeklyAudio) TableName() string {
	return "weekly_audio"
}

// BeforeCreate hook to set UUID if not set
func (w *WeeklyAudio) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
