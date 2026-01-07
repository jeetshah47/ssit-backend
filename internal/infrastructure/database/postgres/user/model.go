package user

import (
	"time"

	"github.com/equitywala/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserModel represents the database model for users
type UserModel struct {
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
	EmailVerified   bool `gorm:"default:false;not null"`
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
	OAuthProvider *string `gorm:"type:varchar(50)"`

	// Timestamps
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

// TableName specifies the table name
func (UserModel) TableName() string {
	return "users"
}

// BeforeCreate hook to set UUID if not set
func (m *UserModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// toEntity converts database model to domain entity
func toEntity(m *UserModel) *user.User {
	var panDetails *user.PANDetails
	if m.PAN != nil && m.PANName != nil {
		panDetails = &user.PANDetails{
			Name:    *m.PANName,
			Address: getStringValue(m.PANAddress),
			Mobile:  getStringValue(m.PANMobile),
		}
	}

	return &user.User{
		ID:                   m.ID,
		Email:                m.Email,
		Phone:                m.Phone,
		Name:                 m.Name,
		PasswordHash:         m.PasswordHash,
		DateOfBirth:          m.DateOfBirth,
		City:                 m.City,
		CustomerType:         m.CustomerType,
		PAN:                  m.PAN,
		PANDetails:           panDetails,
		IsMfCustomer:         m.IsMfCustomer,
		MfCustomerID:         m.MfCustomerID,
		ClassificationSource: getStringValue(m.ClassificationSource),
		ClassificationNotes:  m.ClassificationNotes,
		Status:               m.Status,
		EmailVerified:        m.EmailVerified,
		EmailVerifiedAt:      m.EmailVerifiedAt,
		Username:             m.Username,
		Website:              m.Website,
		Bio:                  m.Bio,
		JobTitle:             m.JobTitle,
		ShowJobTitle:         m.ShowJobTitle,
		AlternativeEmail:     m.AlternativeEmail,
		ProfilePhotoURL:      m.ProfilePhotoURL,
		CoverPhotoURL:        m.CoverPhotoURL,
		GoogleID:             m.GoogleID,
		OAuthProvider:        m.OAuthProvider,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
		DeletedAt:            m.DeletedAt,
	}
}

// toModel converts domain entity to database model
func toModel(u *user.User) *UserModel {
	var panName, panAddress, panMobile *string
	if u.PANDetails != nil {
		panName = &u.PANDetails.Name
		if u.PANDetails.Address != "" {
			panAddress = &u.PANDetails.Address
		}
		if u.PANDetails.Mobile != "" {
			panMobile = &u.PANDetails.Mobile
		}
	}

	return &UserModel{
		ID:                   u.ID,
		Email:                u.Email,
		Phone:                u.Phone,
		Name:                 u.Name,
		PasswordHash:         u.PasswordHash,
		DateOfBirth:          u.DateOfBirth,
		City:                 u.City,
		CustomerType:         u.CustomerType,
		PAN:                  u.PAN,
		PANName:              panName,
		PANAddress:           panAddress,
		PANMobile:            panMobile,
		IsMfCustomer:         u.IsMfCustomer,
		MfCustomerID:         u.MfCustomerID,
		ClassificationSource: stringPtr(u.ClassificationSource),
		ClassificationNotes:  u.ClassificationNotes,
		Status:               u.Status,
		EmailVerified:        u.EmailVerified,
		EmailVerifiedAt:      u.EmailVerifiedAt,
		Username:             u.Username,
		Website:              u.Website,
		Bio:                  u.Bio,
		JobTitle:             u.JobTitle,
		ShowJobTitle:         u.ShowJobTitle,
		AlternativeEmail:     u.AlternativeEmail,
		ProfilePhotoURL:      u.ProfilePhotoURL,
		CoverPhotoURL:        u.CoverPhotoURL,
		GoogleID:             u.GoogleID,
		OAuthProvider:        u.OAuthProvider,
		CreatedAt:            u.CreatedAt,
		UpdatedAt:            u.UpdatedAt,
		DeletedAt:            u.DeletedAt,
	}
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
