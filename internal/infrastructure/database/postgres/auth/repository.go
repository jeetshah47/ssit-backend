package auth

import (
	"context"
	"time"

	"github.com/equitywala/backend/internal/domain/auth"
	"github.com/equitywala/backend/internal/shared/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OTPRepository implements the OTP repository interface
type OTPRepository struct {
	db *gorm.DB
}

// NewOTPRepository creates a new OTP repository
func NewOTPRepository(db *gorm.DB) auth.OTPRepository {
	return &OTPRepository{db: db}
}

// Create creates a new OTP
func (r *OTPRepository) Create(ctx context.Context, otp *auth.OTP) error {
	model := otpToModel(otp)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to create OTP", err)
	}
	*otp = *modelToOTP(model)
	return nil
}

// FindByID finds an OTP by ID
func (r *OTPRepository) FindByID(ctx context.Context, id uuid.UUID) (*auth.OTP, error) {
	var model OTPModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("OTP_NOT_FOUND", "OTP not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find OTP", err)
	}
	return modelToOTP(&model), nil
}

// FindByUserIDAndType finds the latest OTP for a user and type
func (r *OTPRepository) FindByUserIDAndType(ctx context.Context, userID uuid.UUID, otpType string) (*auth.OTP, error) {
	var model OTPModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, otpType).
		Order("created_at DESC").
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("OTP_NOT_FOUND", "OTP not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find OTP", err)
	}
	return modelToOTP(&model), nil
}

// FindByEmailAndType finds the latest OTP for an email and type
func (r *OTPRepository) FindByEmailAndType(ctx context.Context, email, otpType string) (*auth.OTP, error) {
	var model OTPModel
	if err := r.db.WithContext(ctx).
		Where("email = ? AND type = ?", email, otpType).
		Order("created_at DESC").
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("OTP_NOT_FOUND", "OTP not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find OTP", err)
	}
	return modelToOTP(&model), nil
}

// Update updates an existing OTP
func (r *OTPRepository) Update(ctx context.Context, otp *auth.OTP) error {
	model := otpToModel(otp)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to update OTP", err)
	}
	*otp = *modelToOTP(model)
	return nil
}

// SessionRepository implements the session repository interface
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) auth.SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(ctx context.Context, session *auth.Session) error {
	model := sessionToModel(session)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to create session", err)
	}
	*session = *modelToSession(model)
	return nil
}

// FindByID finds a session by ID
func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*auth.Session, error) {
	var model SessionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("SESSION_NOT_FOUND", "session not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find session", err)
	}
	return modelToSession(&model), nil
}

// FindByTokenHash finds a session by token hash
func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*auth.Session, error) {
	var model SessionModel
	if err := r.db.WithContext(ctx).
		Where("token_hash = ? AND is_active = true", tokenHash).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("SESSION_NOT_FOUND", "session not found")
		}
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find session", err)
	}
	return modelToSession(&model), nil
}

// FindActiveByUserID finds all active sessions for a user
func (r *SessionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*auth.Session, error) {
	var models []SessionModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true", userID).
		Order("last_active_at DESC").
		Find(&models).Error; err != nil {
		return nil, errors.WrapDomainError("DATABASE_ERROR", "failed to find sessions", err)
	}

	sessions := make([]*auth.Session, len(models))
	for i, m := range models {
		sessions[i] = modelToSession(&m)
	}
	return sessions, nil
}

// Update updates an existing session
func (r *SessionRepository) Update(ctx context.Context, session *auth.Session) error {
	model := sessionToModel(session)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to update session", err)
	}
	*session = *modelToSession(model)
	return nil
}

// Revoke revokes a session
func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&SessionModel{}).
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
func (r *SessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID, excludeSessionID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&SessionModel{}).
		Where("user_id = ? AND id != ? AND is_active = true", userID, excludeSessionID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": now,
		}).Error; err != nil {
		return errors.WrapDomainError("DATABASE_ERROR", "failed to revoke sessions", err)
	}
	return nil
}

