package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IPOAdvisory represents the database model for IPO advisories
type IPOAdvisory struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;column:id"`
	AdvisoryTypeID   uuid.UUID `gorm:"type:uuid;not null;column:advisory_type_id"`
	IPOName          string    `gorm:"type:varchar(255);not null;column:ipo_name"`
	IPOSymbol        *string   `gorm:"type:varchar(50);column:ipo_symbol"`
	GMP              *float64  `gorm:"type:decimal(10,2);column:gmp"` // Grey Market Premium
	Suggestion       *string   `gorm:"type:varchar(255);column:suggestion"`
	LotSize          int       `gorm:"not null;column:lot_size"`
	PriceBandMin     float64   `gorm:"type:decimal(10,2);not null;column:price_band_min"`
	PriceBandMax     float64   `gorm:"type:decimal(10,2);not null;column:price_band_max"`
	IssueDate        *string   `gorm:"type:varchar(100);column:issue_date"`
	IssueSize        *string   `gorm:"type:varchar(100);column:issue_size"`
	IPOTimetable     *string   `gorm:"type:text;column:ipo_timetable"`
	ReportURL        *string   `gorm:"type:text;column:report_url"`
	ReportFileName   *string   `gorm:"type:varchar(255);column:report_file_name"`
	ReportUploadedAt *time.Time `gorm:"column:report_uploaded_at"`
	ReportUploadedBy *uuid.UUID `gorm:"type:uuid;column:report_uploaded_by"`
	Status           string     `gorm:"type:varchar(50);default:upcoming;not null;column:status"`
	PublishedBy      *uuid.UUID `gorm:"type:uuid;column:published_by"`
	PublishedAt      *time.Time `gorm:"column:published_at"`
	CreatedAt        time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime;column:updated_at"`

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
