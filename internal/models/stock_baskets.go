package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockBasket represents the database model for stock baskets
type StockBasket struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	AdvisoryTypeID   uuid.UUID `gorm:"type:uuid;not null" json:"advisoryTypeId"`
	Name             string    `gorm:"type:varchar(255);not null" json:"name"`
	Description      *string   `gorm:"type:text" json:"description,omitempty"`
	IsBulletIdea     bool      `gorm:"default:false;not null" json:"isBulletIdea"`
	ReportURL        *string   `gorm:"type:text" json:"reportURL,omitempty"`
	ReportFileName   *string   `gorm:"type:varchar(255)" json:"reportFileName,omitempty"`
	ReportUploadedAt *time.Time `json:"reportUploadedAt,omitempty"`
	ReportUploadedBy *uuid.UUID `gorm:"type:uuid" json:"reportUploadedBy,omitempty"`
	Status           string     `gorm:"type:varchar(50);default:draft;not null" json:"status"`
	PublishedBy      *uuid.UUID `gorm:"type:uuid" json:"publishedBy,omitempty"`
	PublishedAt      *time.Time `json:"publishedAt,omitempty"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	AdvisoryType AdvisoryType      `gorm:"foreignKey:AdvisoryTypeID" json:"advisoryType,omitempty"`
	Items        []StockBasketItem `gorm:"foreignKey:StockBasketID" json:"items"`
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
	ID               uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StockBasketID    uuid.UUID `gorm:"type:uuid;not null" json:"stockBasketId"`
	StockID          uuid.UUID `gorm:"type:uuid;not null" json:"stockId"`
	CMP              float64   `gorm:"type:decimal(10,2);not null" json:"cmp"`
	Target           float64   `gorm:"type:decimal(10,2);not null" json:"target"`
	StopLoss         *float64  `gorm:"type:decimal(10,2)" json:"stopLoss,omitempty"`
	EntryRangeMin    *float64  `gorm:"type:decimal(10,2)" json:"entryRangeMin,omitempty"`
	EntryRangeMax    *float64  `gorm:"type:decimal(10,2)" json:"entryRangeMax,omitempty"`
	Action           *string   `gorm:"type:varchar(10)" json:"action,omitempty"` // buy, sell
	RiskLevel        *string   `gorm:"type:varchar(50)" json:"riskLevel,omitempty"`
	TimeHorizon      *string   `gorm:"type:varchar(100)" json:"timeHorizon,omitempty"`
	Rationale        *string   `gorm:"type:text" json:"rationale,omitempty"`
	ReportURL        *string   `gorm:"type:text" json:"reportURL,omitempty"`
	FundamentalsURL  *string   `gorm:"type:text" json:"fundamentalsURL,omitempty"`
	DisplayOrder     int       `gorm:"default:0;not null" json:"displayOrder"`
	IsBulletIdea     bool      `gorm:"default:false;not null" json:"isBulletIdea"`
	Status           string    `gorm:"type:varchar(50);default:active;not null" json:"status"`
	CurrentPrice     *float64  `gorm:"type:decimal(10,2)" json:"currentPrice,omitempty"`
	TargetAchievedAt *time.Time `json:"targetAchievedAt,omitempty"`
	StopLossHitAt    *time.Time `json:"stopLossHitAt,omitempty"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	StockBasket StockBasket `gorm:"foreignKey:StockBasketID" json:"-"`
	Stock       Stock       `gorm:"foreignKey:StockID" json:"stock,omitempty"`
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
