package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ETFBasket represents the database model for ETF baskets
type ETFBasket struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	AdvisoryTypeID  uuid.UUID  `gorm:"type:uuid;not null" json:"advisoryTypeId"`
	Name            string     `gorm:"type:varchar(255);not null" json:"name"`
	Description     *string    `gorm:"type:text" json:"description,omitempty"`
	ReportURL       *string    `gorm:"type:text" json:"reportURL,omitempty"`
	ReportFileName  *string    `gorm:"type:varchar(255)" json:"reportFileName,omitempty"`
	ReportUploadedAt *time.Time `json:"reportUploadedAt,omitempty"`
	ReportUploadedBy *uuid.UUID `gorm:"type:uuid" json:"reportUploadedBy,omitempty"`
	Status          string     `gorm:"type:varchar(50);default:draft;not null" json:"status"`
	PublishedBy     *uuid.UUID `gorm:"type:uuid" json:"publishedBy,omitempty"`
	PublishedAt     *time.Time `json:"publishedAt,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	AdvisoryType AdvisoryType    `gorm:"foreignKey:AdvisoryTypeID" json:"advisoryType,omitempty"`
	Items        []ETFBasketItem `gorm:"foreignKey:ETFBasketID" json:"items"`
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
	ID            uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	ETFBasketID   uuid.UUID  `gorm:"type:uuid;not null" json:"etfBasketId"`
	Name          string     `gorm:"type:varchar(255);not null" json:"name"`
	Symbol        *string    `gorm:"type:varchar(50)" json:"symbol,omitempty"`
	CMP           float64    `gorm:"type:decimal(10,2);not null" json:"cmp"`
	Target        float64    `gorm:"type:decimal(10,2);not null" json:"target"`
	StopLoss      *float64   `gorm:"type:decimal(10,2)" json:"stopLoss,omitempty"`
	EntryRangeMin *float64   `gorm:"type:decimal(10,2)" json:"entryRangeMin,omitempty"`
	EntryRangeMax *float64   `gorm:"type:decimal(10,2)" json:"entryRangeMax,omitempty"`
	Action        *string    `gorm:"type:varchar(10)" json:"action,omitempty"` // buy, sell
	RiskLevel     *string    `gorm:"type:varchar(50)" json:"riskLevel,omitempty"`
	TimeHorizon   *string    `gorm:"type:varchar(100)" json:"timeHorizon,omitempty"`
	Rationale     *string    `gorm:"type:text" json:"rationale,omitempty"`
	PDFLink       *string    `gorm:"type:text" json:"pdfLink,omitempty"`
	DisplayOrder  int        `gorm:"default:0;not null" json:"displayOrder"`
	Status        string     `gorm:"type:varchar(50);default:active;not null" json:"status"`
	CurrentPrice  *float64   `gorm:"type:decimal(10,2)" json:"currentPrice,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	ETFBasket ETFBasket `gorm:"foreignKey:ETFBasketID" json:"-"`
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
