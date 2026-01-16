package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MutualFundBasket represents the database model for mutual fund baskets
type MutualFundBasket struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key"`
	AdvisoryTypeID   uuid.UUID `gorm:"type:uuid;not null"`
	BasketType       string    `gorm:"type:varchar(50);not null"` // general, index_fund, sectoral, nfo_review
	Name             string    `gorm:"type:varchar(255);not null"`
	Description      *string   `gorm:"type:text"`
	MaxSchemes       int       `gorm:"default:10;not null"`
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
	AdvisoryType AdvisoryType           `gorm:"foreignKey:AdvisoryTypeID"`
	Items        []MutualFundBasketItem `gorm:"foreignKey:MutualFundBasketID"`
}

// TableName specifies the table name
func (MutualFundBasket) TableName() string {
	return "mutual_fund_baskets"
}

// BeforeCreate hook to set UUID if not set
func (m *MutualFundBasket) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// IsPublished checks if the basket is published
func (m *MutualFundBasket) IsPublished() bool {
	return m.Status == "published" && m.PublishedAt != nil
}

// MutualFundBasketItem represents the database model for mutual fund basket items
type MutualFundBasketItem struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key"`
	MutualFundBasketID uuid.UUID `gorm:"type:uuid;not null"`
	SchemeName         string    `gorm:"type:varchar(255);not null"`
	SchemeCode         *string   `gorm:"type:varchar(50)"`
	EntryPrice         float64   `gorm:"type:decimal(10,2);not null"`
	ExitPrice          *float64  `gorm:"type:decimal(10,2)"`
	CurrentNAV         *float64  `gorm:"type:decimal(10,4)"`
	Trend              *string   `gorm:"type:varchar(50)"` // positive, negative, neutral
	Status             string    `gorm:"type:varchar(50);default:active;not null"`
	DisplayOrder       int       `gorm:"default:0;not null"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`

	// Relations
	MutualFundBasket MutualFundBasket `gorm:"foreignKey:MutualFundBasketID"`
}

// TableName specifies the table name
func (MutualFundBasketItem) TableName() string {
	return "mutual_fund_basket_items"
}

// BeforeCreate hook to set UUID if not set
func (m *MutualFundBasketItem) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
