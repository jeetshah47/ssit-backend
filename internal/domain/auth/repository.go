package auth

import (
	"context"

	"github.com/google/uuid"
)

// OTPRepository defines the interface for OTP data persistence
type OTPRepository interface {
	// Create creates a new OTP
	Create(ctx context.Context, otp *OTP) error

	// FindByID finds an OTP by ID
	FindByID(ctx context.Context, id uuid.UUID) (*OTP, error)

	// FindByUserIDAndType finds the latest OTP for a user and type
	FindByUserIDAndType(ctx context.Context, userID uuid.UUID, otpType string) (*OTP, error)

	// FindByEmailAndType finds the latest OTP for an email and type
	FindByEmailAndType(ctx context.Context, email, otpType string) (*OTP, error)

	// Update updates an existing OTP
	Update(ctx context.Context, otp *OTP) error
}

// SessionRepository defines the interface for session data persistence
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *Session) error

	// FindByID finds a session by ID
	FindByID(ctx context.Context, id uuid.UUID) (*Session, error)

	// FindByTokenHash finds a session by token hash
	FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error)

	// FindActiveByUserID finds all active sessions for a user
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *Session) error

	// Revoke revokes a session
	Revoke(ctx context.Context, id uuid.UUID) error

	// RevokeAllByUserID revokes all sessions for a user (except current session)
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID, excludeSessionID uuid.UUID) error
}

