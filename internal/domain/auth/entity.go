package auth

import (
	"time"

	"github.com/google/uuid"
)

// OTP represents an OTP entity
type OTP struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Email      string
	CodeHash   string // Hashed OTP code
	Type       string // email_verification, password_reset, login
	ExpiresAt  time.Time
	Used       bool
	UsedAt     *time.Time
	Attempts   int
	MaxAttempts int
	CreatedAt  time.Time
}

// NewOTP creates a new OTP entity
func NewOTP(userID uuid.UUID, email, codeHash, otpType string, expiresAt time.Time) *OTP {
	return &OTP{
		ID:          uuid.New(),
		UserID:      userID,
		Email:       email,
		CodeHash:    codeHash,
		Type:        otpType,
		ExpiresAt:   expiresAt,
		Used:        false,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
	}
}

// IsExpired checks if the OTP has expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsUsed checks if the OTP has been used
func (o *OTP) IsUsed() bool {
	return o.Used
}

// CanAttempt checks if more attempts are allowed
func (o *OTP) CanAttempt() bool {
	return o.Attempts < o.MaxAttempts
}

// MarkAsUsed marks the OTP as used
func (o *OTP) MarkAsUsed() {
	now := time.Now()
	o.Used = true
	o.UsedAt = &now
}

// IncrementAttempts increments the attempt counter
func (o *OTP) IncrementAttempts() {
	o.Attempts++
}

// Session represents a user session
type Session struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	TokenHash       string
	RefreshTokenHash *string
	Device          *string
	DeviceType      *string // desktop, mobile, tablet
	UserAgent       *string
	IPAddress       *string
	Location        *string
	IsActive        bool
	LastActiveAt    time.Time
	ExpiresAt       time.Time
	CreatedAt       time.Time
	RevokedAt       *time.Time
}

// NewSession creates a new session
func NewSession(userID uuid.UUID, tokenHash string, expiresAt time.Time) *Session {
	return &Session{
		ID:           uuid.New(),
		UserID:       userID,
		TokenHash:    tokenHash,
		IsActive:     true,
		LastActiveAt: time.Now(),
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Revoke revokes the session
func (s *Session) Revoke() {
	now := time.Now()
	s.IsActive = false
	s.RevokedAt = &now
}

// UpdateLastActive updates the last active timestamp
func (s *Session) UpdateLastActive() {
	s.LastActiveAt = time.Now()
}

