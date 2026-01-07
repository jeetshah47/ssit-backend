package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		userName     string
		passwordHash string
		wantErr      bool
		validate     func(t *testing.T, user *User)
	}{
		{
			name:         "creates user with valid data",
			email:        "test@example.com",
			userName:     "Test User",
			passwordHash: "hashed_password",
			wantErr:      false,
			validate: func(t *testing.T, user *User) {
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "Test User", user.Name)
				assert.Equal(t, "hashed_password", user.PasswordHash)
				assert.False(t, user.IsMfCustomer)
				assert.Equal(t, "pending_verification", user.Status)
				assert.False(t, user.EmailVerified)
				assert.False(t, user.ShowJobTitle)
				assert.NotZero(t, user.CreatedAt)
				assert.NotZero(t, user.UpdatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.email, tt.userName, tt.passwordHash)
			require.NotNil(t, user)
			tt.validate(t, user)
		})
	}
}

func TestUser_VerifyEmail(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		expectedStatus string
	}{
		{
			name:           "verifies email and updates status from pending_verification",
			initialStatus:  "pending_verification",
			expectedStatus: "verified",
		},
		{
			name:           "verifies email but keeps existing status",
			initialStatus:  "active",
			expectedStatus: "active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser("test@example.com", "Test User", "hash")
			user.Status = tt.initialStatus
			beforeUpdate := user.UpdatedAt

			// Wait a bit to ensure UpdatedAt changes
			time.Sleep(10 * time.Millisecond)

			user.VerifyEmail()

			assert.True(t, user.EmailVerified)
			assert.NotNil(t, user.EmailVerifiedAt)
			assert.Equal(t, tt.expectedStatus, user.Status)
			assert.True(t, user.UpdatedAt.After(beforeUpdate))
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "returns true for active status",
			status:   "active",
			expected: true,
		},
		{
			name:     "returns true for verified status",
			status:   "verified",
			expected: true,
		},
		{
			name:     "returns false for pending_verification status",
			status:   "pending_verification",
			expected: false,
		},
		{
			name:     "returns false for suspended status",
			status:   "suspended",
			expected: false,
		},
		{
			name:     "returns false for deleted status",
			status:   "deleted",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser("test@example.com", "Test User", "hash")
			user.Status = tt.status

			result := user.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_GetIsMfCustomer(t *testing.T) {
	tests := []struct {
		name         string
		isMfCustomer bool
		expected     bool
	}{
		{
			name:         "returns true when user is MF customer",
			isMfCustomer: true,
			expected:     true,
		},
		{
			name:         "returns false when user is not MF customer",
			isMfCustomer: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser("test@example.com", "Test User", "hash")
			user.IsMfCustomer = tt.isMfCustomer

			result := user.GetIsMfCustomer()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_SetMfCustomer(t *testing.T) {
	tests := []struct {
		name          string
		isMf          bool
		mfCustomerID  *string
		source        string
		notes         string
		validate      func(t *testing.T, user *User)
	}{
		{
			name:         "sets MF customer with all fields",
			isMf:         true,
			mfCustomerID: stringPtr("MF123"),
			source:       "database",
			notes:        "Verified customer",
			validate: func(t *testing.T, u *User) {
				assert.True(t, u.IsMfCustomer)
				assert.Equal(t, "MF123", *u.MfCustomerID)
				assert.Equal(t, "database", u.ClassificationSource)
				assert.Equal(t, "Verified customer", *u.ClassificationNotes)
				assert.True(t, u.UpdatedAt.After(u.CreatedAt))
			},
		},
		{
			name:         "sets MF customer without notes",
			isMf:         true,
			mfCustomerID: stringPtr("MF456"),
			source:       "manual",
			notes:        "",
			validate: func(t *testing.T, user *User) {
				assert.True(t, user.IsMfCustomer)
				assert.Nil(t, user.ClassificationNotes)
			},
		},
		{
			name:         "removes MF customer status",
			isMf:         false,
			mfCustomerID: nil,
			source:       "manual",
			notes:        "",
			validate: func(t *testing.T, user *User) {
				assert.False(t, user.IsMfCustomer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUser("test@example.com", "Test User", "hash")
			_ = u.UpdatedAt // Capture before update

			// Wait a bit to ensure UpdatedAt changes
			time.Sleep(10 * time.Millisecond)

			u.SetMfCustomer(tt.isMf, tt.mfCustomerID, tt.source, tt.notes)

			tt.validate(t, u)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

