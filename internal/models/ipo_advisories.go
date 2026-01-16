package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IPOAdvisory represents the database model for IPO advisories
type IPOAdvisory struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key"`
	AdvisoryTypeID   uuid.UUID `gorm:"type:uuid;not null"`
	IPOName          string    `gorm:"type:varchar(255);not null"`
	IPOSymbol        *string   `gorm:"type:varchar(50)"`
	GMP              *float64  `gorm:"type:decimal(10,2)"` // Grey Market Premium
	Suggestion       *string   `gorm:"type:varchar(255)"`
	LotSize          int       `gorm:"not null"`
	PriceBandMin     float64   `gorm:"type:decimal(10,2);not null"`
	PriceBandMax     float64   `gorm:"type:decimal(10,2);not null"`
	IssueDate        *string   `gorm:"type:varchar(100)"`
	IssueSize        *string   `gorm:"type:varchar(100)"`
	IPOTimetable     *string   `gorm:"type:text"`
	ReportURL        *string   `gorm:"type:text"`
	ReportFileName   *string   `gorm:"type:varchar(255)"`
	ReportUploadedAt *time.Time
	ReportUploadedBy *uuid.UUID `gorm:"type:uuid"`
	Status           string     `gorm:"type:varchar(50);default:upcoming;not null"`
	PublishedBy      *uuid.UUID `gorm:"type:uuid"`
	PublishedAt      *time.Time
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`

	// Relations
	AdvisoryType AdvisoryType `gorm:"foreignKey:AdvisoryTypeID"`
}

// TableName specifies the table name
func (IPOAdvisory) TableName() string {
	return "ipo_advisories"
}

// BeforeCreate hook to set UUID if not set
func (i *IPOAdvisory) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// IsPublished checks if the IPO advisory is published
func (i *IPOAdvisory) IsPublished() bool {
	return i.Status != "upcoming" && i.PublishedAt != nil
}
