package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for user data persistence
type Repository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)

	// FindByEmail finds a user by email (case-insensitive)
	FindByEmail(ctx context.Context, email string) (*User, error)

	// FindByPhone finds a user by phone number
	FindByPhone(ctx context.Context, phone string) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByEmail checks if a user exists with the given email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

