package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/equitywala/backend/internal/application/user/queries"
	"github.com/equitywala/backend/internal/domain/user"
	postgresUser "github.com/equitywala/backend/internal/infrastructure/database/postgres/user"
	"github.com/equitywala/backend/internal/interfaces/http/api"
	"github.com/equitywala/backend/internal/interfaces/http/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestNewUserHandler(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userRepo := postgresUser.NewRepository(db)
	getUserQuery := queries.NewGetUserHandler(userRepo)
	handler := NewUserHandler(getUserQuery)

	require.NotNil(t, handler)
	assert.Equal(t, getUserQuery, handler.getUserQuery)
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		setupDB     func(*testing.T, *gorm.DB) *user.User
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result interface{})
	}{
		{
			name:   "successfully retrieves user",
			userID: "123e4567-e89b-12d3-a456-426614174000",
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				testUser := user.NewUser("test@example.com", "Test User", string(hashedPassword))
				testUser.ID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				testUser.Status = "active"
				testUser.EmailVerified = true

				userRepo := postgresUser.NewRepository(db)
				err := userRepo.Create(context.Background(), testUser)
				require.NoError(t, err)

				return testUser
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				response, ok := result.(*dto.UserResponse)
				require.True(t, ok)
				assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", response.ID)
				assert.Equal(t, "test@example.com", response.Email)
				assert.Equal(t, "Test User", response.Name)
				assert.Equal(t, "active", response.Status)
				assert.True(t, response.EmailVerified)
				assert.NotEmpty(t, response.CreatedAt)
			},
		},
		{
			name:   "returns error when user not found",
			userID: "123e4567-e89b-12d3-a456-426614174000",
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				// No user created
				return nil
			},
			wantErr:     true,
			errContains: "user not found",
		},
		{
			name:   "returns error when invalid UUID",
			userID: "invalid-uuid",
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				// No setup needed
				return nil
			},
			wantErr:     true,
			errContains: "",
		},
		{
			name:   "handles user with phone number",
			userID: "123e4567-e89b-12d3-a456-426614174000",
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				testUser := user.NewUser("test@example.com", "Test User", string(hashedPassword))
				testUser.ID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				phone := "1234567890"
				testUser.Phone = &phone

				userRepo := postgresUser.NewRepository(db)
				err := userRepo.Create(context.Background(), testUser)
				require.NoError(t, err)

				return testUser
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				response, ok := result.(*dto.UserResponse)
				require.True(t, ok)
				assert.NotNil(t, response.Phone)
				assert.Equal(t, "1234567890", *response.Phone)
			},
		},
		{
			name:   "handles MF customer user",
			userID: "123e4567-e89b-12d3-a456-426614174000",
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				testUser := user.NewUser("test@example.com", "Test User", string(hashedPassword))
				testUser.ID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				testUser.SetMfCustomer(true, stringPtr("MF123"), "database", "Verified customer")

				userRepo := postgresUser.NewRepository(db)
				err := userRepo.Create(context.Background(), testUser)
				require.NoError(t, err)

				return testUser
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				response, ok := result.(*dto.UserResponse)
				require.True(t, ok)
				assert.True(t, response.IsMfCustomer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer func() {
				sqlDB, _ := db.DB()
				if sqlDB != nil {
					sqlDB.Close()
				}
			}()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
			c.Request = req
			c.Params = gin.Params{
				{Key: "id", Value: tt.userID},
			}

			ctx := api.NewContext(c)

			tt.setupDB(t, db)

			userRepo := postgresUser.NewRepository(db)
			getUserQuery := queries.NewGetUserHandler(userRepo)
			handler := NewUserHandler(getUserQuery)

			result, err := handler.GetUser(ctx)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
