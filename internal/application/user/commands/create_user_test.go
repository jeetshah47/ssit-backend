package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/equitywala/backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of user.Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func TestNewCreateUserHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := NewCreateUserHandler(mockRepo)

	require.NotNil(t, handler)
	assert.Equal(t, mockRepo, handler.userRepo)
}

func TestCreateUserHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CreateUserCommand
		setupMocks  func(*MockUserRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "successfully creates user",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			setupMocks: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
				m.On("Create", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					return u.Email == "test@example.com" && u.Name == "Test User"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successfully creates user with phone",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
				Phone:    stringPtr("1234567890"),
			},
			setupMocks: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
				m.On("Create", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					return u.Email == "test@example.com" && u.Name == "Test User" && u.Phone != nil && *u.Phone == "1234567890"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "returns error when email is empty",
			cmd: CreateUserCommand{
				Email:    "",
				Name:     "Test User",
				Password: "password123",
			},
			setupMocks: func(m *MockUserRepository) {},
			wantErr:     true,
			errContains: "invalid email",
		},
		{
			name: "returns error when name is empty",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "",
				Password: "password123",
			},
			setupMocks: func(m *MockUserRepository) {},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "returns error when password is too short",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "short",
			},
			setupMocks: func(m *MockUserRepository) {},
			wantErr:     true,
			errContains: "password must be at least 8 characters",
		},
		{
			name: "returns error when user already exists",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			setupMocks: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "test@example.com").Return(true, nil)
			},
			wantErr:     true,
			errContains: "user with this email already exists",
		},
		{
			name: "returns error when ExistsByEmail fails",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			setupMocks: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, errors.New("database error"))
			},
			wantErr:     true,
			errContains: "failed to check user existence",
		},
		{
			name: "returns error when Create fails",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			setupMocks: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
				m.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			wantErr:     true,
			errContains: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewCreateUserHandler(mockRepo)
			err := handler.Handle(context.Background(), tt.cmd)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateUserHandler_validateCommand(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CreateUserCommand
		wantErr     bool
		errContains string
	}{
		{
			name: "valid command",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			cmd: CreateUserCommand{
				Email:    "",
				Name:     "Test User",
				Password: "password123",
			},
			wantErr:     true,
			errContains: "invalid email",
		},
		{
			name: "empty name",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "",
				Password: "password123",
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "short password",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "short",
			},
			wantErr:     true,
			errContains: "password must be at least 8 characters",
		},
		{
			name: "password exactly 8 characters",
			cmd: CreateUserCommand{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "12345678",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCreateUserHandler(new(MockUserRepository))
			err := handler.validateCommand(tt.cmd)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

