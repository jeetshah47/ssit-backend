package commands

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserCommand represents the input for creating a user
// Password is optional for step-by-step signup flow
type CreateUserCommand struct {
	Email    string
	Phone    *string
	Name     string
	Password string // Optional - can be empty for initial signup
}

// CreateUserHandler handles user creation
type CreateUserHandler struct {
	userRepo user.Repository
}

// NewCreateUserHandler creates a new command handler
func NewCreateUserHandler(userRepo user.Repository) *CreateUserHandler {
	return &CreateUserHandler{
		userRepo: userRepo,
	}
}

// Handle executes the command
func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) error {
	// 1. Validate command
	if err := h.validateCommand(cmd); err != nil {
		return err
	}

	// 2. Check if user already exists
	exists, err := h.userRepo.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return user.ErrUserAlreadyExists
	}

	// 3. Hash password if provided
	var passwordHash string
	if cmd.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(hash)
	}

	// 4. Create domain entity
	newUser := user.NewUser(cmd.Email, cmd.Name, passwordHash)
	if cmd.Phone != nil {
		newUser.Phone = cmd.Phone
	}

	// 5. TODO: Check MF customer database and set classification
	// This will be implemented when MF customer service is ready

	// 6. Persist via repository
	if err := h.userRepo.Create(ctx, newUser); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// validateCommand validates the command input
func (h *CreateUserHandler) validateCommand(cmd CreateUserCommand) error {
	if cmd.Email == "" {
		return user.ErrInvalidEmail
	}
	if cmd.Name == "" {
		return fmt.Errorf("name is required")
	}
	// Password is optional for step-by-step signup
	if cmd.Password != "" && len(cmd.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

