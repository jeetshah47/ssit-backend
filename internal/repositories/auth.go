package repositories

import (
	"context"
	"time"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/common/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OTPRepo defines the interface for OTP data persistence
type OTPRepo interface {
	// Create creates a new OTP
	Create(ctx context.Context, otp *models.OTP) error

	// FindByID finds an OTP by ID
	FindByID(ctx context.Context, id uuid.UUID) (*models.OTP, error)

	// FindByUserIDAndType finds the latest OTP for a user and type
	FindByUserIDAndType(ctx context.Context, userID uuid.UUID, otpType string) (*models.OTP, error)

	// FindByEmailAndType finds the latest OTP for an email and type
	FindByEmailAndType(ctx context.Context, email, otpType string) (*models.OTP, error)

	// Update updates an existing OTP
	Update(ctx context.Context, otp *models.OTP) error
}

// otpRepo implements the OTP repository interface
type otpRepo struct {
	*RepoContext
}

// NewOTPRepo creates a new OTP repository
func NewOTPRepo(ctx *RepoContext) OTPRepo {
	return &otpRepo{RepoContext: ctx}
}

// Create creates a new OTP
func (r *otpRepo) Create(ctx context.Context, otp *models.OTP) error {
	if err := r.db.WithContext(ctx).Create(otp).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to create OTP", err)
	}
	return nil
}

// FindByID finds an OTP by ID
func (r *otpRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.OTP, error) {
	var otp models.OTP
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&otp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("OTP_NOT_FOUND", "OTP not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find OTP", err)
	}
	return &otp, nil
}

// FindByUserIDAndType finds the latest OTP for a user and type
func (r *otpRepo) FindByUserIDAndType(ctx context.Context, userID uuid.UUID, otpType string) (*models.OTP, error) {
	var otp models.OTP
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, otpType).
		Order("created_at DESC").
		First(&otp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("OTP_NOT_FOUND", "OTP not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find OTP", err)
	}
	return &otp, nil
}

// FindByEmailAndType finds the latest OTP for an email and type
func (r *otpRepo) FindByEmailAndType(ctx context.Context, email, otpType string) (*models.OTP, error) {
	var otp models.OTP
	if err := r.db.WithContext(ctx).
		Where("email = ? AND type = ?", email, otpType).
		Order("created_at DESC").
		First(&otp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("OTP_NOT_FOUND", "OTP not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find OTP", err)
	}
	return &otp, nil
}

// Update updates an existing OTP
func (r *otpRepo) Update(ctx context.Context, otp *models.OTP) error {
	if err := r.db.WithContext(ctx).Save(otp).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to update OTP", err)
	}
	return nil
}

// SessionRepo defines the interface for session data persistence
type SessionRepo interface {
	// Create creates a new session
	Create(ctx context.Context, session *models.Session) error

	// FindByID finds a session by ID
	FindByID(ctx context.Context, id uuid.UUID) (*models.Session, error)

	// FindByTokenHash finds a session by token hash
	FindByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error)

	// FindActiveByUserID finds all active sessions for a user
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *models.Session) error

	// Revoke revokes a session
	Revoke(ctx context.Context, id uuid.UUID) error

	// RevokeAllByUserID revokes all sessions for a user (except current session)
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID, excludeSessionID uuid.UUID) error
}

// sessionRepo implements the session repository interface
type sessionRepo struct {
	*RepoContext
}

// NewSessionRepo creates a new session repository
func NewSessionRepo(ctx *RepoContext) SessionRepo {
	return &sessionRepo{RepoContext: ctx}
}

// Create creates a new session
func (r *sessionRepo) Create(ctx context.Context, session *models.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to create session", err)
	}
	return nil
}

// FindByID finds a session by ID
func (r *sessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("SESSION_NOT_FOUND", "session not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find session", err)
	}
	return &session, nil
}

// FindByTokenHash finds a session by token hash
func (r *sessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).
		Where("token_hash = ? AND is_active = true", tokenHash).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("SESSION_NOT_FOUND", "session not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find session", err)
	}
	return &session, nil
}

// FindActiveByUserID finds all active sessions for a user
func (r *sessionRepo) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	var sessions []models.Session
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true", userID).
		Order("last_active_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find sessions", err)
	}

	result := make([]*models.Session, len(sessions))
	for i := range sessions {
		result[i] = &sessions[i]
	}
	return result, nil
}

// Update updates an existing session
func (r *sessionRepo) Update(ctx context.Context, session *models.Session) error {
	if err := r.db.WithContext(ctx).Save(session).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to update session", err)
	}
	return nil
}

// Revoke revokes a session
func (r *sessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": now,
		}).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to revoke session", err)
	}
	return nil
}

// RevokeAllByUserID revokes all sessions for a user (except current session)
func (r *sessionRepo) RevokeAllByUserID(ctx context.Context, userID uuid.UUID, excludeSessionID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("user_id = ? AND id != ? AND is_active = true", userID, excludeSessionID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": now,
		}).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to revoke sessions", err)
	}
	return nil
}

