package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the domain
type User struct {
	ID           uuid.UUID
	Email        string
	Phone        *string
	Name         string
	PasswordHash string

	// Additional Profile Fields
	DateOfBirth  *time.Time
	City         *string
	CustomerType *string // 'Individual' or 'Non Individual'
	PAN          *string
	PANDetails   *PANDetails

	// User Classification (Critical for MF vs Paid logic)
	IsMfCustomer         bool
	MfCustomerID         *string
	ClassificationSource string
	ClassificationNotes  *string

	// Account Status
	Status string // pending_verification, verified, active, suspended, deleted

	// Email Verification
	EmailVerified   bool
	EmailVerifiedAt *time.Time

	// Profile Settings
	Username         *string
	Website          *string
	Bio              *string
	JobTitle         *string
	ShowJobTitle     bool
	AlternativeEmail *string
	ProfilePhotoURL  *string
	CoverPhotoURL    *string

	// OAuth
	GoogleID      *string
	OAuthProvider *string

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// PANDetails represents PAN verification details
type PANDetails struct {
	Name    string
	Address string
	Mobile  string
}

// NewUser creates a new user entity
// passwordHash can be empty for initial signup (password set later)
func NewUser(email, name, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:            uuid.New(),
		Email:         email,
		Name:          name,
		PasswordHash:  passwordHash,
		IsMfCustomer:  false,
		Status:        "pending_verification",
		EmailVerified: false,
		ShowJobTitle:  false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// SetPassword updates the user's password hash
func (u *User) SetPassword(passwordHash string) {
	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now()
}

// UpdateProfile updates user profile fields
func (u *User) UpdateProfile(phone *string, dateOfBirth *time.Time, city *string) {
	if phone != nil {
		u.Phone = phone
	}
	if dateOfBirth != nil {
		u.DateOfBirth = dateOfBirth
	}
	if city != nil {
		u.City = city
	}
	u.UpdatedAt = time.Now()
}

// SetPAN sets the PAN and PAN details
func (u *User) SetPAN(pan string, details *PANDetails) {
	u.PAN = &pan
	u.PANDetails = details
	u.UpdatedAt = time.Now()
}

// VerifyEmail marks the user's email as verified
func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerified = true
	u.EmailVerifiedAt = &now
	if u.Status == "pending_verification" {
		u.Status = "verified"
	}
	u.UpdatedAt = now
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == "active" || u.Status == "verified"
}

// GetIsMfCustomer returns true if user is an MF customer
func (u *User) GetIsMfCustomer() bool {
	return u.IsMfCustomer
}

// SetMfCustomer sets the MF customer status
func (u *User) SetMfCustomer(isMf bool, mfCustomerID *string, source, notes string) {
	u.IsMfCustomer = isMf
	u.MfCustomerID = mfCustomerID
	u.ClassificationSource = source
	if notes != "" {
		u.ClassificationNotes = &notes
	}
	u.UpdatedAt = time.Now()
}

// SetCustomerType sets the customer type (Individual or Non Individual)
func (u *User) SetCustomerType(customerType string) {
	u.CustomerType = &customerType
	u.UpdatedAt = time.Now()
}
