package queries

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

func TestNewGetUserHandler(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := NewGetUserHandler(mockRepo)

	require.NotNil(t, handler)
	assert.Equal(t, mockRepo, handler.userRepo)
}

func TestGetUserHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		query       GetUserQuery
		setupMocks  func(*MockUserRepository)
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result *GetUserResult)
	}{
		{
			name: "successfully retrieves user",
			query: GetUserQuery{
				UserID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			setupMocks: func(m *MockUserRepository) {
				expectedUser := user.NewUser("test@example.com", "Test User", "hash")
				expectedUser.ID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				m.On("FindByID", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(expectedUser, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, result *GetUserResult) {
				require.NotNil(t, result)
				require.NotNil(t, result.User)
				assert.Equal(t, "test@example.com", result.User.Email)
				assert.Equal(t, "Test User", result.User.Name)
			},
		},
		{
			name: "returns error when user not found",
			query: GetUserQuery{
				UserID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			setupMocks: func(m *MockUserRepository) {
				m.On("FindByID", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(nil, user.ErrUserNotFound)
			},
			wantErr:     true,
			errContains: "user not found",
		},
		{
			name: "returns error when repository fails",
			query: GetUserQuery{
				UserID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			setupMocks: func(m *MockUserRepository) {
				m.On("FindByID", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(nil, errors.New("database error"))
			},
			wantErr:     true,
			errContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewGetUserHandler(mockRepo)
			result, err := handler.Handle(context.Background(), tt.query)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

