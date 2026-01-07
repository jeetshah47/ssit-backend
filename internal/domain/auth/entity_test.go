package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOTP(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	codeHash := "hashed_code"
	otpType := "email_verification"
	expiresAt := time.Now().Add(5 * time.Minute)

	otp := NewOTP(userID, email, codeHash, otpType, expiresAt)

	require.NotNil(t, otp)
	assert.NotEmpty(t, otp.ID)
	assert.Equal(t, userID, otp.UserID)
	assert.Equal(t, email, otp.Email)
	assert.Equal(t, codeHash, otp.CodeHash)
	assert.Equal(t, otpType, otp.Type)
	assert.Equal(t, expiresAt, otp.ExpiresAt)
	assert.False(t, otp.Used)
	assert.Nil(t, otp.UsedAt)
	assert.Equal(t, 0, otp.Attempts)
	assert.Equal(t, 3, otp.MaxAttempts)
	assert.NotZero(t, otp.CreatedAt)
}

func TestOTP_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "returns true when OTP is expired",
			expiresAt: time.Now().Add(-1 * time.Minute),
			expected:  true,
		},
		{
			name:      "returns false when OTP is not expired",
			expiresAt: time.Now().Add(5 * time.Minute),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp := NewOTP(uuid.New(), "test@example.com", "hash", "email_verification", tt.expiresAt)
			result := otp.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOTP_IsUsed(t *testing.T) {
	tests := []struct {
		name     string
		used     bool
		expected bool
	}{
		{
			name:     "returns true when OTP is used",
			used:     true,
			expected: true,
		},
		{
			name:     "returns false when OTP is not used",
			used:     false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp := NewOTP(uuid.New(), "test@example.com", "hash", "email_verification", time.Now().Add(5*time.Minute))
			otp.Used = tt.used
			result := otp.IsUsed()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOTP_CanAttempt(t *testing.T) {
	tests := []struct {
		name      string
		attempts  int
		maxAttempts int
		expected  bool
	}{
		{
			name:       "returns true when attempts are below max",
			attempts:   1,
			maxAttempts: 3,
			expected:   true,
		},
		{
			name:       "returns false when attempts equal max",
			attempts:   3,
			maxAttempts: 3,
			expected:   false,
		},
		{
			name:       "returns false when attempts exceed max",
			attempts:   4,
			maxAttempts: 3,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp := NewOTP(uuid.New(), "test@example.com", "hash", "email_verification", time.Now().Add(5*time.Minute))
			otp.Attempts = tt.attempts
			otp.MaxAttempts = tt.maxAttempts
			result := otp.CanAttempt()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOTP_MarkAsUsed(t *testing.T) {
	otp := NewOTP(uuid.New(), "test@example.com", "hash", "email_verification", time.Now().Add(5*time.Minute))
	
	assert.False(t, otp.Used)
	assert.Nil(t, otp.UsedAt)

	otp.MarkAsUsed()

	assert.True(t, otp.Used)
	assert.NotNil(t, otp.UsedAt)
	assert.True(t, otp.UsedAt.Before(time.Now().Add(1*time.Second)))
	assert.True(t, otp.UsedAt.After(time.Now().Add(-1*time.Second)))
}

func TestOTP_IncrementAttempts(t *testing.T) {
	otp := NewOTP(uuid.New(), "test@example.com", "hash", "email_verification", time.Now().Add(5*time.Minute))
	
	assert.Equal(t, 0, otp.Attempts)

	otp.IncrementAttempts()
	assert.Equal(t, 1, otp.Attempts)

	otp.IncrementAttempts()
	assert.Equal(t, 2, otp.Attempts)

	otp.IncrementAttempts()
	assert.Equal(t, 3, otp.Attempts)
}

func TestNewSession(t *testing.T) {
	userID := uuid.New()
	tokenHash := "token_hash_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	session := NewSession(userID, tokenHash, expiresAt)

	require.NotNil(t, session)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, tokenHash, session.TokenHash)
	assert.True(t, session.IsActive)
	assert.NotZero(t, session.LastActiveAt)
	assert.Equal(t, expiresAt, session.ExpiresAt)
	assert.NotZero(t, session.CreatedAt)
	assert.Nil(t, session.RevokedAt)
}

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "returns true when session is expired",
			expiresAt: time.Now().Add(-1 * time.Hour),
			expected:  true,
		},
		{
			name:      "returns false when session is not expired",
			expiresAt: time.Now().Add(24 * time.Hour),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession(uuid.New(), "token_hash", tt.expiresAt)
			result := session.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSession_Revoke(t *testing.T) {
	session := NewSession(uuid.New(), "token_hash", time.Now().Add(24*time.Hour))
	
	assert.True(t, session.IsActive)
	assert.Nil(t, session.RevokedAt)

	session.Revoke()

	assert.False(t, session.IsActive)
	assert.NotNil(t, session.RevokedAt)
	assert.True(t, session.RevokedAt.Before(time.Now().Add(1*time.Second)))
	assert.True(t, session.RevokedAt.After(time.Now().Add(-1*time.Second)))
}

func TestSession_UpdateLastActive(t *testing.T) {
	session := NewSession(uuid.New(), "token_hash", time.Now().Add(24*time.Hour))
	originalLastActive := session.LastActiveAt

	// Wait a bit to ensure timestamp changes
	time.Sleep(10 * time.Millisecond)

	session.UpdateLastActive()

	assert.True(t, session.LastActiveAt.After(originalLastActive))
}

