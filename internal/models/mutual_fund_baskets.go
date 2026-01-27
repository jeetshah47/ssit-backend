package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MutualFundBasket represents the database model for mutual fund baskets
type MutualFundBasket struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	AdvisoryTypeID   uuid.UUID  `gorm:"type:uuid;not null" json:"advisoryTypeId"`
	BasketType       string     `gorm:"type:varchar(50);not null" json:"basketType"` // general, index_fund, sectoral, nfo_review
	Name             string     `gorm:"type:varchar(255);not null" json:"name"`
	Description      *string    `gorm:"type:text" json:"description,omitempty"`
	MaxSchemes       int        `gorm:"default:10;not null" json:"maxSchemes"`
	ReportURL        *string    `gorm:"type:text" json:"reportURL,omitempty"`
	ReportFileName   *string    `gorm:"type:varchar(255)" json:"reportFileName,omitempty"`
	ReportUploadedAt *time.Time `json:"reportUploadedAt,omitempty"`
	ReportUploadedBy *uuid.UUID `gorm:"type:uuid" json:"reportUploadedBy,omitempty"`
	Status           string     `gorm:"type:varchar(50);default:draft;not null" json:"status"`
	PublishedBy      *uuid.UUID `gorm:"type:uuid" json:"publishedBy,omitempty"`
	PublishedAt      *time.Time `json:"publishedAt,omitempty"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	AdvisoryType AdvisoryType           `gorm:"foreignKey:AdvisoryTypeID" json:"advisoryType,omitempty"`
	Items        []MutualFundBasketItem `gorm:"foreignKey:MutualFundBasketID" json:"items"`
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
	ID                 uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	MutualFundBasketID uuid.UUID `gorm:"type:uuid;not null" json:"mutualFundBasketId"`
	SchemeName         string    `gorm:"type:varchar(255);not null" json:"schemeName"`
	SchemeCode         *string   `gorm:"type:varchar(50)" json:"schemeCode,omitempty"`
	EntryPrice         float64   `gorm:"type:decimal(10,2);not null" json:"entryPrice"`
	ExitPrice          *float64  `gorm:"type:decimal(10,2)" json:"exitPrice,omitempty"`
	CurrentNAV         *float64  `gorm:"type:decimal(10,4)" json:"currentNAV,omitempty"`
	Trend              *string   `gorm:"type:varchar(50)" json:"trend,omitempty"` // positive, negative, neutral
	Status             string    `gorm:"type:varchar(50);default:active;not null" json:"status"`
	DisplayOrder       int       `gorm:"default:0;not null" json:"displayOrder"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	MutualFundBasket MutualFundBasket `gorm:"foreignKey:MutualFundBasketID" json:"-"`
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
