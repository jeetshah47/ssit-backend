package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OTP represents the database model for OTPs
type OTP struct {
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

func (OTP) TableName() string {
	return "otps"
}

// BeforeCreate hook to set UUID if not set
func (o *OTP) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
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

// Session represents the database model for sessions
type Session struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash        string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	RefreshTokenHash *string
	Device           *string    `gorm:"type:varchar(255)"`
	DeviceType       *string    `gorm:"type:varchar(50)"`
	UserAgent        *string    `gorm:"type:text"`
	IPAddress        *string    `gorm:"type:varchar(45)"`
	Location         *string    `gorm:"type:varchar(255)"`
	IsActive         bool       `gorm:"default:true;not null;index"`
	LastActiveAt     time.Time  `gorm:"autoCreateTime;autoUpdateTime"`
	ExpiresAt        time.Time  `gorm:"not null;index"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	RevokedAt        *time.Time
}

func (Session) TableName() string {
	return "sessions"
}

// BeforeCreate hook to set UUID if not set
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
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

// SignupRequest represents the signup request
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=1"`
	Password string `json:"password,omitempty"` // Optional for step-by-step signup
}

// SignupResponse represents the signup response
type SignupResponse struct {
	UserID                  string `json:"userId"`
	Email                   string `json:"email"`
	RequiresOTPVerification bool   `json:"requiresOTPVerification"`
	Message                 string `json:"message,omitempty"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token        string         `json:"token"`
	RefreshToken string         `json:"refreshToken,omitempty"`
	User         *UserResponse  `json:"user"`
	Progress     *SignupProgress `json:"progress,omitempty"` // Signup progress for journey tracking
}

// OTPVerificationRequest represents the OTP verification request
type OTPVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=4"` // 4-digit OTP
}

// SignupProgress represents the signup progress tracking
type SignupProgress struct {
	EmailVerified      bool `json:"emailVerified"`      // Step 1: Email verified
	ProfileCompleted   bool `json:"profileCompleted"`   // Step 2: Profile details (phone, DOB, city) filled
	PANVerified        bool `json:"panVerified"`        // Step 3: PAN verified
	PasswordSet        bool `json:"passwordSet"`        // Step 4: Password set
	PlanSelected       bool `json:"planSelected"`        // Step 5: Payment plan selected
	PaymentCompleted   bool `json:"paymentCompleted"`   // Step 6: Payment completed
	NextStep           string `json:"nextStep"`          // Suggested next step route
}

// OTPVerificationResponse represents the OTP verification response
type OTPVerificationResponse struct {
	Message string         `json:"message"`
	Token   string         `json:"token,omitempty"`
	User    *UserResponse  `json:"user,omitempty"`
	Progress *SignupProgress `json:"progress,omitempty"` // Signup progress tracking
}

// ResendOTPRequest represents the resend OTP request
type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResendOTPResponse represents the resend OTP response
type ResendOTPResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expiresIn"`
}

// UpdateProfileRequest represents the profile update request
type UpdateProfileRequest struct {
	Phone       string `json:"phone,omitempty"`
	DateOfBirth string `json:"dateOfBirth,omitempty"` // ISO 8601 format
	City        string `json:"city,omitempty"`
}

// UpdateProfileResponse represents the profile update response
type UpdateProfileResponse struct {
	Message string        `json:"message"`
	User    *UserResponse `json:"user"`
}

// VerifyPANRequest represents the PAN verification request
type VerifyPANRequest struct {
	PAN          string `json:"pan" binding:"required,min=10,max=10"`
	CustomerType string `json:"customerType" binding:"required"`
}

// VerifyPANResponse represents the PAN verification response
type VerifyPANResponse struct {
	Message    string      `json:"message"`
	PANDetails *PANDetails `json:"panDetails,omitempty"`
}

// SetPasswordRequest represents the set password request
type SetPasswordRequest struct {
	Email           string `json:"email" binding:"required,email"` // Email to identify user (they've verified OTP)
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

// SelectPaymentPlanRequest represents the payment plan selection request
type SelectPaymentPlanRequest struct {
	Plan          string `json:"plan" binding:"required,oneof=standard plus premium"`
	BillingPeriod string `json:"billingPeriod" binding:"required,oneof=quarterly annual"`
}

// SelectPaymentPlanResponse represents the payment plan selection response
type SelectPaymentPlanResponse struct {
	Message string        `json:"message"`
	User    *UserResponse `json:"user"`
}

