package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepo defines the interface for user data persistence
type UserRepo interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// FindByEmail finds a user by email (case-insensitive)
	FindByEmail(ctx context.Context, email string) (*models.User, error)

	// FindByPhone finds a user by phone number
	FindByPhone(ctx context.Context, phone string) (*models.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *models.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByEmail checks if a user exists with the given email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// userRepo implements the user repository interface
type userRepo struct {
	*RepoContext
}

// NewUserRepo creates a new user repository
func NewUserRepo(ctx *RepoContext) UserRepo {
	return &userRepo{RepoContext: ctx}
}

// Create creates a new user
func (r *userRepo) Create(ctx context.Context, u *models.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByID finds a user by ID
func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// FindByEmail finds a user by email (case-insensitive)
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?) AND deleted_at IS NULL", email).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// FindByPhone finds a user by phone number
func (r *userRepo) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("phone = ? AND deleted_at IS NULL", phone).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// Update updates an existing user
func (r *userRepo) Update(ctx context.Context, u *models.User) error {
	if err := r.db.WithContext(ctx).Save(u).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete soft deletes a user
func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).
		Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("LOWER(email) = LOWER(?) AND deleted_at IS NULL", email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}
