package auth

import (
	"time"

	"github.com/equitywala/backend/internal/domain/auth"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OTPModel represents the database model for OTPs
type OTPModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Email       string    `gorm:"type:varchar(255);not null"`
	CodeHash    string    `gorm:"type:varchar(255);not null"`
	Type        string    `gorm:"type:varchar(50);not null"`
	ExpiresAt   time.Time `gorm:"not null;index"`
	Used        bool      `gorm:"default:false;not null"`
	UsedAt      *time.Time
	Attempts    int       `gorm:"default:0;not null"`
	MaxAttempts int       `gorm:"default:3;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func (OTPModel) TableName() string {
	return "otps"
}

// BeforeCreate hook to set UUID if not set
func (m *OTPModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// SessionModel represents the database model for sessions
type SessionModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash       string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	RefreshTokenHash *string
	Device          *string    `gorm:"type:varchar(255)"`
	DeviceType      *string    `gorm:"type:varchar(50)"`
	UserAgent       *string    `gorm:"type:text"`
	IPAddress       *string    `gorm:"type:varchar(45)"`
	Location        *string    `gorm:"type:varchar(255)"`
	IsActive        bool       `gorm:"default:true;not null;index"`
	LastActiveAt    time.Time  `gorm:"autoCreateTime;autoUpdateTime"`
	ExpiresAt       time.Time  `gorm:"not null;index"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	RevokedAt       *time.Time
}

func (SessionModel) TableName() string {
	return "sessions"
}

// BeforeCreate hook to set UUID if not set
func (m *SessionModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// Conversion functions
func otpToModel(o *auth.OTP) *OTPModel {
	return &OTPModel{
		ID:          o.ID,
		UserID:      o.UserID,
		Email:       o.Email,
		CodeHash:    o.CodeHash,
		Type:        o.Type,
		ExpiresAt:   o.ExpiresAt,
		Used:        o.Used,
		UsedAt:      o.UsedAt,
		Attempts:    o.Attempts,
		MaxAttempts: o.MaxAttempts,
		CreatedAt:   o.CreatedAt,
	}
}

func modelToOTP(m *OTPModel) *auth.OTP {
	return &auth.OTP{
		ID:          m.ID,
		UserID:      m.UserID,
		Email:       m.Email,
		CodeHash:    m.CodeHash,
		Type:        m.Type,
		ExpiresAt:   m.ExpiresAt,
		Used:        m.Used,
		UsedAt:      m.UsedAt,
		Attempts:    m.Attempts,
		MaxAttempts: m.MaxAttempts,
		CreatedAt:   m.CreatedAt,
	}
}

func sessionToModel(s *auth.Session) *SessionModel {
	return &SessionModel{
		ID:              s.ID,
		UserID:          s.UserID,
		TokenHash:       s.TokenHash,
		RefreshTokenHash: s.RefreshTokenHash,
		Device:          s.Device,
		DeviceType:      s.DeviceType,
		UserAgent:       s.UserAgent,
		IPAddress:       s.IPAddress,
		Location:        s.Location,
		IsActive:        s.IsActive,
		LastActiveAt:    s.LastActiveAt,
		ExpiresAt:       s.ExpiresAt,
		CreatedAt:       s.CreatedAt,
		RevokedAt:       s.RevokedAt,
	}
}

func modelToSession(m *SessionModel) *auth.Session {
	return &auth.Session{
		ID:              m.ID,
		UserID:          m.UserID,
		TokenHash:       m.TokenHash,
		RefreshTokenHash: m.RefreshTokenHash,
		Device:          m.Device,
		DeviceType:      m.DeviceType,
		UserAgent:       m.UserAgent,
		IPAddress:       m.IPAddress,
		Location:        m.Location,
		IsActive:        m.IsActive,
		LastActiveAt:    m.LastActiveAt,
		ExpiresAt:       m.ExpiresAt,
		CreatedAt:       m.CreatedAt,
		RevokedAt:       m.RevokedAt,
	}
}

