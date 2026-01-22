package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Service handles JWT token generation and validation
type Service struct {
	secret     []byte
	expiry     time.Duration
	refreshExpiry time.Duration
}

// NewService creates a new JWT service
func NewService(secret string, expiry, refreshExpiry time.Duration) *Service {
	return &Service{
		secret:        []byte(secret),
		expiry:        expiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateToken generates a new JWT token for a user
func (s *Service) GenerateToken(userID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "equitywala-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken generates a refresh token
func (s *Service) GenerateRefreshToken(userID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "equitywala-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secret, nil
	})

	if err != nil {
		// jwt/v5 automatically validates expiration and other claims
		// Return the error as-is (it will be jwt.ErrTokenExpired for expired tokens)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Additional manual validation for extra safety
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, jwt.ErrTokenExpired
		}
		// Check not before
		if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
			return nil, jwt.ErrTokenNotValidYet
		}
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

