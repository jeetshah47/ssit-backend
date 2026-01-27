package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	secret := "test-secret-key"
	expiry := 24 * time.Hour
	refreshExpiry := 168 * time.Hour

	service := NewService(secret, expiry, refreshExpiry)

	require.NotNil(t, service)
	assert.Equal(t, []byte(secret), service.secret)
	assert.Equal(t, expiry, service.expiry)
	assert.Equal(t, refreshExpiry, service.refreshExpiry)
}

func TestService_GenerateToken(t *testing.T) {
	service := NewService("test-secret-key", 24*time.Hour, 168*time.Hour)
	userID := "123e4567-e89b-12d3-a456-426614174000"
	email := "test@example.com"

	token, err := service.GenerateToken(userID, email)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Validate the token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "equitywala-backend", claims.Issuer)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(1*time.Second)))
}

func TestService_GenerateRefreshToken(t *testing.T) {
	service := NewService("test-secret-key", 24*time.Hour, 168*time.Hour)
	userID := "123e4567-e89b-12d3-a456-426614174000"
	email := "test@example.com"

	refreshToken, err := service.GenerateRefreshToken(userID, email)

	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)

	// Validate the refresh token
	claims, err := service.ValidateToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "equitywala-backend", claims.Issuer)
	// Refresh token should have longer expiry
	assert.True(t, claims.ExpiresAt.Time.After(time.Now().Add(24*time.Hour)))
}

func TestService_ValidateToken(t *testing.T) {
	service := NewService("test-secret-key", 24*time.Hour, 168*time.Hour)
	userID := "123e4567-e89b-12d3-a456-426614174000"
	email := "test@example.com"

	tests := []struct {
		name        string
		setupToken  func() string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, claims *Claims)
	}{
		{
			name: "validates valid token",
			setupToken: func() string {
				token, _ := service.GenerateToken(userID, email)
				return token
			},
			wantErr: false,
			validate: func(t *testing.T, claims *Claims) {
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
			},
		},
		{
			name: "returns error for invalid token",
			setupToken: func() string {
				return "invalid.token.here"
			},
			wantErr:     true,
			errContains: "",
		},
		{
			name: "returns error for token with wrong secret",
			setupToken: func() string {
				otherService := NewService("different-secret", 24*time.Hour, 168*time.Hour)
				token, _ := otherService.GenerateToken(userID, email)
				return token
			},
			wantErr:     true,
			errContains: "",
		},
		{
			name: "returns error for empty token",
			setupToken: func() string {
				return ""
			},
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := service.ValidateToken(token)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
				if tt.validate != nil {
					tt.validate(t, claims)
				}
			}
		})
	}
}

func TestService_ValidateToken_ExpiredToken(t *testing.T) {
	// Create a service with very short expiry
	service := NewService("test-secret-key", 1*time.Nanosecond, 168*time.Hour)
	userID := "123e4567-e89b-12d3-a456-426614174000"
	email := "test@example.com"

	token, err := service.GenerateToken(userID, email)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	claims, err := service.ValidateToken(token)
	require.Error(t, err)
	assert.Nil(t, claims)
	// Should be a validation error for expired token
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestClaims_Structure(t *testing.T) {
	service := NewService("test-secret-key", 24*time.Hour, 168*time.Hour)
	userID := "123e4567-e89b-12d3-a456-426614174000"
	email := "test@example.com"

	token, err := service.GenerateToken(userID, email)
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)

	// Verify all claim fields
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "equitywala-backend", claims.Issuer)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
	assert.True(t, claims.ExpiresAt.Time.After(claims.IssuedAt.Time))
}

func TestService_DifferentUsers(t *testing.T) {
	service := NewService("test-secret-key", 24*time.Hour, 168*time.Hour)

	user1ID := "user-1-id"
	user1Email := "user1@example.com"
	user2ID := "user-2-id"
	user2Email := "user2@example.com"

	token1, err := service.GenerateToken(user1ID, user1Email)
	require.NoError(t, err)

	token2, err := service.GenerateToken(user2ID, user2Email)
	require.NoError(t, err)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)

	// Validate both tokens
	claims1, err := service.ValidateToken(token1)
	require.NoError(t, err)
	assert.Equal(t, user1ID, claims1.UserID)
	assert.Equal(t, user1Email, claims1.Email)

	claims2, err := service.ValidateToken(token2)
	require.NoError(t, err)
	assert.Equal(t, user2ID, claims2.UserID)
	assert.Equal(t, user2Email, claims2.Email)
}

