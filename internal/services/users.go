package services

import (
	"context"
	"fmt"
	"time"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/equitywala/backend/internal/common/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserCmd represents the command for creating a user
type CreateUserCmd struct {
	Email    string
	Phone    *string
	Name     string
	Password string // Optional - can be empty for initial signup
}

// CreateUserCmdOutputData represents the output of CreateUser command
type CreateUserCmdOutputData struct {
	User *models.User
}

// CreateUserService handles user creation
type CreateUserService struct {
	userRepo repositories.UserRepo
}

// NewCreateUserService creates a new create user service
func NewCreateUserService(userRepo repositories.UserRepo) *CreateUserService {
	return &CreateUserService{
		userRepo: userRepo,
	}
}

// Execute executes the CreateUser command
func (s *CreateUserService) Execute(ctx context.Context, cmd CreateUserCmd) (*CreateUserCmdOutputData, error) {
	// 1. Validate command
	if err := s.validateCommand(cmd); err != nil {
		return nil, err
	}

	// 2. Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, errors.NewDomainError("USER_ALREADY_EXISTS", "user with this email already exists")
	}

	// 3. Hash password if provided
	var passwordHash string
	if cmd.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(hash)
	}

	// 4. Create user model
	now := time.Now()
	newUser := &models.User{
		ID:            uuid.New(),
		Email:         cmd.Email,
		Name:          cmd.Name,
		PasswordHash:  passwordHash,
		Phone:         cmd.Phone,
		IsMfCustomer:  false,
		Status:        "pending_verification",
		EmailVerified: false,
		ShowJobTitle:  false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 5. TODO: Check MF customer database and set classification
	// This will be implemented when MF customer service is ready

	// 6. Persist via repository
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &CreateUserCmdOutputData{User: newUser}, nil
}

// validateCommand validates the command input
func (s *CreateUserService) validateCommand(cmd CreateUserCmd) error {
	if cmd.Email == "" {
		return errors.NewDomainError("INVALID_EMAIL", "invalid email format")
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

// GetUserQuery represents the query for getting a user
type GetUserQuery struct {
	UserID uuid.UUID
}

// GetUserResult represents the output of GetUser query
type GetUserResult struct {
	User *models.User
}

// GetUserService handles user retrieval
type GetUserService struct {
	userRepo repositories.UserRepo
}

// NewGetUserService creates a new get user service
func NewGetUserService(userRepo repositories.UserRepo) *GetUserService {
	return &GetUserService{
		userRepo: userRepo,
	}
}

// Execute executes the GetUser query
func (s *GetUserService) Execute(ctx context.Context, query GetUserQuery) (*GetUserResult, error) {
	user, err := s.userRepo.FindByID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	return &GetUserResult{User: user}, nil
}

