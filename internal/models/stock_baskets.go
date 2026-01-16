package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockBasket represents the database model for stock baskets
type StockBasket struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key"`
	AdvisoryTypeID   uuid.UUID `gorm:"type:uuid;not null"`
	Name             string    `gorm:"type:varchar(255);not null"`
	Description      *string   `gorm:"type:text"`
	IsBulletIdea     bool      `gorm:"default:false;not null"`
	ReportURL        *string   `gorm:"type:text"`
	ReportFileName   *string   `gorm:"type:varchar(255)"`
	ReportUploadedAt *time.Time
	ReportUploadedBy *uuid.UUID `gorm:"type:uuid"`
	Status           string     `gorm:"type:varchar(50);default:draft;not null"`
	PublishedBy      *uuid.UUID `gorm:"type:uuid"`
	PublishedAt      *time.Time
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`

	// Relations
	AdvisoryType AdvisoryType      `gorm:"foreignKey:AdvisoryTypeID"`
	Items        []StockBasketItem `gorm:"foreignKey:StockBasketID"`
}

// TableName specifies the table name
func (StockBasket) TableName() string {
	return "stock_baskets"
}

// BeforeCreate hook to set UUID if not set
func (s *StockBasket) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsPublished checks if the basket is published
func (s *StockBasket) IsPublished() bool {
	return s.Status == "published" && s.PublishedAt != nil
}

// StockBasketItem represents the database model for stock basket items
type StockBasketItem struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key"`
	StockBasketID    uuid.UUID `gorm:"type:uuid;not null"`
	StockName        string    `gorm:"type:varchar(255);not null"`
	StockSymbol      *string   `gorm:"type:varchar(50)"`
	CMP              float64   `gorm:"type:decimal(10,2);not null"`
	Target           float64   `gorm:"type:decimal(10,2);not null"`
	StopLoss         *float64  `gorm:"type:decimal(10,2)"`
	EntryRangeMin    *float64  `gorm:"type:decimal(10,2)"`
	EntryRangeMax    *float64  `gorm:"type:decimal(10,2)"`
	Action           *string   `gorm:"type:varchar(10)"` // buy, sell
	RiskLevel        *string   `gorm:"type:varchar(50)"`
	TimeHorizon      *string   `gorm:"type:varchar(100)"`
	Rationale        *string   `gorm:"type:text"`
	ReportURL        *string   `gorm:"type:text"`
	FundamentalsURL  *string   `gorm:"type:text"`
	DisplayOrder     int       `gorm:"default:0;not null"`
	IsBulletIdea     bool      `gorm:"default:false;not null"`
	Status           string    `gorm:"type:varchar(50);default:active;not null"`
	CurrentPrice     *float64  `gorm:"type:decimal(10,2)"`
	TargetAchievedAt *time.Time
	StopLossHitAt    *time.Time
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`

	// Relations
	StockBasket StockBasket `gorm:"foreignKey:StockBasketID"`
}

// TableName specifies the table name
func (StockBasketItem) TableName() string {
	return "stock_basket_items"
}

// BeforeCreate hook to set UUID if not set
func (s *StockBasketItem) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
