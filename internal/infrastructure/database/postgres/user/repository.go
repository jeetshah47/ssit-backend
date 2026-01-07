package user

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository implements the user repository interface
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) user.Repository {
	return &Repository{db: db}
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, u *user.User) error {
	model := toModel(u)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	*u = *toEntity(model)
	return nil
}

// FindByID finds a user by ID
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return toEntity(&model), nil
}

// FindByEmail finds a user by email (case-insensitive)
func (r *Repository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?) AND deleted_at IS NULL", email).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return toEntity(&model), nil
}

// FindByPhone finds a user by phone number
func (r *Repository) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).
		Where("phone = ? AND deleted_at IS NULL", phone).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return toEntity(&model), nil
}

// Update updates an existing user
func (r *Repository) Update(ctx context.Context, u *user.User) error {
	model := toModel(u)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	*u = *toEntity(model)
	return nil
}

// Delete soft deletes a user
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).
		Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("LOWER(email) = LOWER(?) AND deleted_at IS NULL", email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}

