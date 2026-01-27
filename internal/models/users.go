package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the database model for users
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone        *string   `gorm:"type:varchar(20);uniqueIndex"`
	Name         string    `gorm:"type:varchar(255);not null"`
	PasswordHash string    `gorm:"type:varchar(255)"` // Can be empty for step-by-step signup

	// Additional Profile Fields
	DateOfBirth  *time.Time `gorm:"type:date"`
	City         *string    `gorm:"type:varchar(100)"`
	CustomerType *string    `gorm:"type:varchar(50)"` // 'Individual' or 'Non Individual'
	PAN          *string    `gorm:"type:varchar(10);uniqueIndex"`
	PANName      *string    `gorm:"type:varchar(255)"` // From PAN verification
	PANAddress   *string    `gorm:"type:text"`         // From PAN verification
	PANMobile    *string    `gorm:"type:varchar(20)"`  // From PAN verification

	// User Classification
	IsMfCustomer         bool    `gorm:"default:false;not null"`
	MfCustomerID         *string `gorm:"type:varchar(100);uniqueIndex"`
	ClassificationSource *string `gorm:"type:varchar(50)"`
	ClassificationNotes  *string `gorm:"type:text"`

	// Account Status
	Status string `gorm:"type:varchar(50);default:pending_verification;not null"`

	// Email Verification
	EmailVerified   bool       `gorm:"default:false;not null"`
	EmailVerifiedAt *time.Time

	// Profile Settings
	Username         *string `gorm:"type:varchar(100)"`
	Website          *string `gorm:"type:varchar(255)"`
	Bio              *string `gorm:"type:text"`
	JobTitle         *string `gorm:"type:varchar(255)"`
	ShowJobTitle     bool    `gorm:"default:false"`
	AlternativeEmail *string `gorm:"type:varchar(255)"`
	ProfilePhotoURL  *string `gorm:"type:text"`
	CoverPhotoURL    *string `gorm:"type:text"`

	// OAuth
	GoogleID      *string `gorm:"type:varchar(255);uniqueIndex"`
	OAuthProvider *string `gorm:"type:varchar(50);column:oauth_provider"`

	// Timestamps
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

// TableName specifies the table name
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to set UUID if not set
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// PANDetails represents PAN verification details
type PANDetails struct {
	Name    string
	Address string
	Mobile  string
}

// GetIsMfCustomer returns true if user is an MF customer
func (u *User) GetIsMfCustomer() bool {
	return u.IsMfCustomer
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == "active" || u.Status == "verified"
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Phone         *string `json:"phone,omitempty"`
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	IsMfCustomer  bool    `json:"isMfCustomer"`
	EmailVerified bool    `json:"emailVerified"`
	CreatedAt     string  `json:"createdAt"`
}

// GetUserParams represents path parameters for get user
type GetUserParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}

