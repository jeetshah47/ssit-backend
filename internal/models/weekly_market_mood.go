package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WeeklyMarketMood represents the database model for weekly market mood
type WeeklyMarketMood struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	Title           string     `gorm:"type:varchar(255);not null"`
	Week            string     `gorm:"type:varchar(100);not null"` // e.g., "Week of Oct 21-27, 2024"
	Points          *string    `gorm:"type:text"` // JSON array of points
	Sentiment       float64    `gorm:"type:decimal(3,2);not null;default:0.5"` // 0.0 to 1.0
	SentimentLabel  string     `gorm:"type:varchar(50);not null;default:Neutral"` // Bullish, Bearish, Neutral
	ReportURL       *string    `gorm:"type:text"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (WeeklyMarketMood) TableName() string {
	return "weekly_market_mood"
}

// BeforeCreate hook to set UUID if not set
func (w *WeeklyMarketMood) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
