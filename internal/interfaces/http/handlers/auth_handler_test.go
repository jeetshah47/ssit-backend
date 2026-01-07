package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/equitywala/backend/internal/application/user/commands"
	"github.com/equitywala/backend/internal/domain/auth"
	"github.com/equitywala/backend/internal/domain/email"
	"github.com/equitywala/backend/internal/domain/user"
	postgresAuth "github.com/equitywala/backend/internal/infrastructure/database/postgres/auth"
	postgresUser "github.com/equitywala/backend/internal/infrastructure/database/postgres/user"
	"github.com/equitywala/backend/internal/interfaces/http/api"
	"github.com/equitywala/backend/internal/interfaces/http/dto"
	"github.com/equitywala/backend/internal/shared/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// mockEmailService is a mock implementation of email.Service for testing
type mockEmailService struct{}

// Ensure mockEmailService implements email.Service
var _ email.Service = (*mockEmailService)(nil)

func (m *mockEmailService) SendOTPEmail(toEmail, toName, otpCode string, expiresInMinutes int) error {
	// Mock implementation - always succeeds
	return nil
}

// testHashOTP duplicates the hashOTP function from auth_handler.go for testing
func testHashOTP(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

// Helper function to create test context
func createTestContext(method, url string, body interface{}) *api.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return api.NewContext(c)
}

// setupTestUser creates a test user in the database
func setupTestUser(t *testing.T, db *gorm.DB, email, name, password string, status string, emailVerified bool) *user.User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	testUser := user.NewUser(email, name, string(hashedPassword))
	testUser.Status = status
	testUser.EmailVerified = emailVerified

	userRepo := postgresUser.NewRepository(db)
	err = userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	return testUser
}

// setupTestOTP creates a test OTP in the database
func setupTestOTP(t *testing.T, db *gorm.DB, userID uuid.UUID, email, code string, otpType string, expiresAt time.Time) *auth.OTP {
	otpHash := testHashOTP(code)
	otp := auth.NewOTP(userID, email, otpHash, otpType, expiresAt)

	otpRepo := postgresAuth.NewOTPRepository(db)
	err := otpRepo.Create(context.Background(), otp)
	require.NoError(t, err)

	return otp
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.LoginRequest
		setupDB     func(*testing.T, *gorm.DB) *user.User
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result interface{})
	}{
		{
			name: "successfully logs in user",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				return setupTestUser(t, db, "test@example.com", "Test User", "password123", "active", true)
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				response, ok := result.(dto.LoginResponse)
				require.True(t, ok)
				assert.NotEmpty(t, response.Token)
				assert.NotEmpty(t, response.RefreshToken)
				assert.NotNil(t, response.User)
				assert.Equal(t, "test@example.com", response.User.Email)
			},
		},
		{
			name: "returns error when user not found",
			request: dto.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				// No user created
				return nil
			},
			wantErr:     true,
			errContains: "user not found",
		},
		{
			name: "returns error when password is incorrect",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				return setupTestUser(t, db, "test@example.com", "Test User", "password123", "active", true)
			},
			wantErr:     true,
			errContains: "invalid password",
		},
		{
			name: "returns error when user is not active",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				return setupTestUser(t, db, "test@example.com", "Test User", "password123", "suspended", true)
			},
			wantErr:     true,
			errContains: "user account is not active",
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

			userRepo, otpRepo, sessionRepo := setupTestRepositories(db)
			createUserCmd := commands.NewCreateUserHandler(userRepo)
			jwtService := jwt.NewService("test-secret", 24*time.Hour, 168*time.Hour)

			tt.setupDB(t, db)

			emailSvc := &mockEmailService{}
			// For tests, we can pass nil for selectPaymentPlanCmd since these tests don't use it
			handler := NewAuthHandler(userRepo, otpRepo, sessionRepo, createUserCmd, nil, jwtService, emailSvc)
			ctx := createTestContext("POST", "/auth/login", tt.request)

			result, err := handler.Login(ctx)

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

func TestAuthHandler_VerifyOTP(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.OTPVerificationRequest
		setupDB     func(*testing.T, *gorm.DB) (*user.User, *auth.OTP)
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result interface{})
	}{
		{
			name: "successfully verifies OTP",
			request: dto.OTPVerificationRequest{
				Email: "test@example.com",
				OTP:   "123456",
			},
			setupDB: func(t *testing.T, db *gorm.DB) (*user.User, *auth.OTP) {
				testUser := setupTestUser(t, db, "test@example.com", "Test User", "hash", "pending_verification", false)
				otp := setupTestOTP(t, db, testUser.ID, "test@example.com", "123456", "email_verification", time.Now().Add(5*time.Minute))
				return testUser, otp
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				response, ok := result.(dto.OTPVerificationResponse)
				require.True(t, ok)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, "Email verified successfully", response.Message)
			},
		},
		{
			name: "returns error when OTP not found",
			request: dto.OTPVerificationRequest{
				Email: "test@example.com",
				OTP:   "123456",
			},
			setupDB: func(t *testing.T, db *gorm.DB) (*user.User, *auth.OTP) {
				// No OTP created
				return nil, nil
			},
			wantErr:     true,
			errContains: "invalid or expired OTP",
		},
		{
			name: "returns error when OTP is already used",
			request: dto.OTPVerificationRequest{
				Email: "test@example.com",
				OTP:   "123456",
			},
			setupDB: func(t *testing.T, db *gorm.DB) (*user.User, *auth.OTP) {
				testUser := setupTestUser(t, db, "test@example.com", "Test User", "hash", "pending_verification", false)
				otp := setupTestOTP(t, db, testUser.ID, "test@example.com", "123456", "email_verification", time.Now().Add(5*time.Minute))
				otp.MarkAsUsed()
				otpRepo := postgresAuth.NewOTPRepository(db)
				otpRepo.Update(context.Background(), otp)
				return testUser, otp
			},
			wantErr:     true,
			errContains: "OTP already used",
		},
		{
			name: "returns error when OTP is expired",
			request: dto.OTPVerificationRequest{
				Email: "test@example.com",
				OTP:   "123456",
			},
			setupDB: func(t *testing.T, db *gorm.DB) (*user.User, *auth.OTP) {
				testUser := setupTestUser(t, db, "test@example.com", "Test User", "hash", "pending_verification", false)
				otp := setupTestOTP(t, db, testUser.ID, "test@example.com", "123456", "email_verification", time.Now().Add(-1*time.Minute))
				return testUser, otp
			},
			wantErr:     true,
			errContains: "OTP expired",
		},
		{
			name: "returns error when OTP is invalid",
			request: dto.OTPVerificationRequest{
				Email: "test@example.com",
				OTP:   "wrong",
			},
			setupDB: func(t *testing.T, db *gorm.DB) (*user.User, *auth.OTP) {
				testUser := setupTestUser(t, db, "test@example.com", "Test User", "hash", "pending_verification", false)
				otp := setupTestOTP(t, db, testUser.ID, "test@example.com", "123456", "email_verification", time.Now().Add(5*time.Minute))
				return testUser, otp
			},
			wantErr:     true,
			errContains: "invalid OTP",
		},
		{
			name: "returns error when maximum attempts exceeded",
			request: dto.OTPVerificationRequest{
				Email: "test@example.com",
				OTP:   "123456",
			},
			setupDB: func(t *testing.T, db *gorm.DB) (*user.User, *auth.OTP) {
				testUser := setupTestUser(t, db, "test@example.com", "Test User", "hash", "pending_verification", false)
				otp := setupTestOTP(t, db, testUser.ID, "test@example.com", "123456", "email_verification", time.Now().Add(5*time.Minute))
				otp.Attempts = 3 // Max attempts reached
				otpRepo := postgresAuth.NewOTPRepository(db)
				otpRepo.Update(context.Background(), otp)
				return testUser, otp
			},
			wantErr:     true,
			errContains: "maximum attempts exceeded",
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

			userRepo, otpRepo, sessionRepo := setupTestRepositories(db)
			createUserCmd := commands.NewCreateUserHandler(userRepo)
			jwtService := jwt.NewService("test-secret", 24*time.Hour, 168*time.Hour)

			tt.setupDB(t, db)

			emailSvc := &mockEmailService{}
			// For tests, we can pass nil for selectPaymentPlanCmd since these tests don't use it
			handler := NewAuthHandler(userRepo, otpRepo, sessionRepo, createUserCmd, nil, jwtService, emailSvc)
			ctx := createTestContext("POST", "/auth/verify-otp", tt.request)

			result, err := handler.VerifyOTP(ctx)

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

func TestAuthHandler_Signup(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.SignupRequest
		setupDB     func(*testing.T, *gorm.DB)
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result interface{})
	}{
		{
			name: "successfully signs up user",
			request: dto.SignupRequest{
				Email:    "newuser@example.com",
				Name:     "New User",
				Password: "password123",
			},
			setupDB: func(t *testing.T, db *gorm.DB) {
				// No existing user
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				response, ok := result.(dto.SignupResponse)
				require.True(t, ok)
				assert.NotEmpty(t, response.UserID)
				assert.Equal(t, "newuser@example.com", response.Email)
				assert.True(t, response.RequiresOTPVerification)
				assert.Contains(t, response.Message, "OTP sent to email")
			},
		},
		{
			name: "successfully signs up user with phone",
			request: dto.SignupRequest{
				Email:    "newuser@example.com",
				Name:     "New User",
				Password: "password123",
			},
			setupDB: func(t *testing.T, db *gorm.DB) {
				// No existing user
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				response, ok := result.(dto.SignupResponse)
				require.True(t, ok)
				assert.Equal(t, "newuser@example.com", response.Email)
			},
		},
		{
			name: "returns error when user already exists",
			request: dto.SignupRequest{
				Email:    "existing@example.com",
				Name:     "Existing User",
				Password: "password123",
			},
			setupDB: func(t *testing.T, db *gorm.DB) {
				setupTestUser(t, db, "existing@example.com", "Existing User", "password123", "active", true)
			},
			wantErr:     true,
			errContains: "user with this email already exists",
		},
		{
			name: "returns error when email validation fails",
			request: dto.SignupRequest{
				Email:    "invalid-email",
				Name:     "Test User",
				Password: "password123",
			},
			setupDB: func(t *testing.T, db *gorm.DB) {
				// No setup needed
			},
			wantErr: true,
		},
		{
			name: "returns error when password is too short",
			request: dto.SignupRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "short",
			},
			setupDB: func(t *testing.T, db *gorm.DB) {
				// No setup needed
			},
			wantErr: true,
		},
		{
			name: "returns error when name is empty",
			request: dto.SignupRequest{
				Email:    "test@example.com",
				Name:     "",
				Password: "password123",
			},
			setupDB: func(t *testing.T, db *gorm.DB) {
				// No setup needed
			},
			wantErr: true,
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

			userRepo, otpRepo, sessionRepo := setupTestRepositories(db)
			createUserCmd := commands.NewCreateUserHandler(userRepo)
			jwtService := jwt.NewService("test-secret", 24*time.Hour, 168*time.Hour)

			tt.setupDB(t, db)

			emailSvc := &mockEmailService{}
			// For tests, we can pass nil for selectPaymentPlanCmd since these tests don't use it
			handler := NewAuthHandler(userRepo, otpRepo, sessionRepo, createUserCmd, nil, jwtService, emailSvc)
			ctx := createTestContext("POST", "/auth/signup", tt.request)

			result, err := handler.Signup(ctx)

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

func TestAuthHandler_ResendOTP(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.ResendOTPRequest
		setupDB     func(*testing.T, *gorm.DB) *user.User
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result interface{})
	}{
		{
			name: "successfully resends OTP",
			request: dto.ResendOTPRequest{
				Email: "test@example.com",
			},
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				return setupTestUser(t, db, "test@example.com", "Test User", "hash", "pending_verification", false)
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				response, ok := result.(dto.ResendOTPResponse)
				require.True(t, ok)
				assert.Contains(t, response.Message, "OTP sent")
				assert.Equal(t, 300, response.ExpiresIn)
			},
		},
		{
			name: "returns error when user not found",
			request: dto.ResendOTPRequest{
				Email: "notfound@example.com",
			},
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				// No user created
				return nil
			},
			wantErr:     true,
			errContains: "user not found",
		},
		{
			name: "returns error when email validation fails",
			request: dto.ResendOTPRequest{
				Email: "invalid-email",
			},
			setupDB: func(t *testing.T, db *gorm.DB) *user.User {
				// No setup needed
				return nil
			},
			wantErr: true,
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

			userRepo, otpRepo, sessionRepo := setupTestRepositories(db)
			createUserCmd := commands.NewCreateUserHandler(userRepo)
			jwtService := jwt.NewService("test-secret", 24*time.Hour, 168*time.Hour)

			tt.setupDB(t, db)

			emailSvc := &mockEmailService{}
			// For tests, we can pass nil for selectPaymentPlanCmd since these tests don't use it
			handler := NewAuthHandler(userRepo, otpRepo, sessionRepo, createUserCmd, nil, jwtService, emailSvc)
			ctx := createTestContext("POST", "/auth/resend-otp", tt.request)

			result, err := handler.ResendOTP(ctx)

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
