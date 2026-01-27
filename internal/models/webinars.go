package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Webinar represents the database model for webinars
type Webinar struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	Title           string     `gorm:"type:varchar(255);not null"`
	Description     *string    `gorm:"type:text"`
	Status           string     `gorm:"type:varchar(50);default:upcoming;not null"` // Live, Upcoming, Recorded
	Date             *time.Time
	Time             *string    `gorm:"type:varchar(50)"` // HH:MM format
	Duration         *string    `gorm:"type:varchar(50)"` // e.g., "60 minutes"
	RegistrationURL  *string    `gorm:"type:text"`
	RecordingURL     *string    `gorm:"type:text"`
	ThumbnailURL     *string    `gorm:"type:text"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (Webinar) TableName() string {
	return "webinars"
}

// BeforeCreate hook to set UUID if not set
func (w *Webinar) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
