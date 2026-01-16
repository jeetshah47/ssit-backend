package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ETFBasket represents the database model for ETF baskets
type ETFBasket struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	AdvisoryTypeID  uuid.UUID  `gorm:"type:uuid;not null"`
	Name            string     `gorm:"type:varchar(255);not null"`
	Description     *string    `gorm:"type:text"`
	ReportURL       *string    `gorm:"type:text"`
	ReportFileName  *string    `gorm:"type:varchar(255)"`
	ReportUploadedAt *time.Time
	ReportUploadedBy *uuid.UUID `gorm:"type:uuid"`
	Status          string     `gorm:"type:varchar(50);default:draft;not null"`
	PublishedBy     *uuid.UUID `gorm:"type:uuid"`
	PublishedAt     *time.Time
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`

	// Relations
	AdvisoryType AdvisoryType    `gorm:"foreignKey:AdvisoryTypeID"`
	Items        []ETFBasketItem `gorm:"foreignKey:ETFBasketID"`
}

// TableName specifies the table name
func (ETFBasket) TableName() string {
	return "etf_baskets"
}

// BeforeCreate hook to set UUID if not set
func (e *ETFBasket) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// IsPublished checks if the basket is published
func (e *ETFBasket) IsPublished() bool {
	return e.Status == "published" && e.PublishedAt != nil
}

// ETFBasketItem represents the database model for ETF basket items
type ETFBasketItem struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key"`
	ETFBasketID   uuid.UUID  `gorm:"type:uuid;not null"`
	Name          string     `gorm:"type:varchar(255);not null"`
	Symbol        *string    `gorm:"type:varchar(50)"`
	CMP           float64    `gorm:"type:decimal(10,2);not null"`
	Target        float64    `gorm:"type:decimal(10,2);not null"`
	StopLoss      *float64   `gorm:"type:decimal(10,2)"`
	EntryRangeMin *float64   `gorm:"type:decimal(10,2)"`
	EntryRangeMax *float64   `gorm:"type:decimal(10,2)"`
	Action        *string    `gorm:"type:varchar(10)"` // buy, sell
	RiskLevel     *string    `gorm:"type:varchar(50)"`
	TimeHorizon   *string    `gorm:"type:varchar(100)"`
	Rationale     *string    `gorm:"type:text"`
	PDFLink       *string    `gorm:"type:text"`
	DisplayOrder  int        `gorm:"default:0;not null"`
	Status        string     `gorm:"type:varchar(50);default:active;not null"`
	CurrentPrice  *float64   `gorm:"type:decimal(10,2)"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime"`

	// Relations
	ETFBasket ETFBasket `gorm:"foreignKey:ETFBasketID"`
}

// TableName specifies the table name
func (ETFBasketItem) TableName() string {
	return "etf_basket_items"
}

// BeforeCreate hook to set UUID if not set
func (e *ETFBasketItem) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
