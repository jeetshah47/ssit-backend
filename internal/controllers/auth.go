package controllers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"time"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/equitywala/backend/internal/services"
	"github.com/equitywala/backend/internal/common/utils"
	emailInfra "github.com/equitywala/backend/internal/infrastructure/email"
	"github.com/equitywala/backend/internal/common/jwt"
	"github.com/equitywala/backend/internal/common/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles authentication endpoints
type AuthController struct {
	userRepo            repositories.UserRepo
	otpRepo             repositories.OTPRepo
	sessionRepo         repositories.SessionRepo
	createUserService   *services.CreateUserService
	selectPaymentPlanService *services.SelectPaymentPlanService
	jwtService          *jwt.Service
	emailService        emailInfra.Service
}

// NewAuthController creates a new auth controller
func NewAuthController(
	userRepo repositories.UserRepo,
	otpRepo repositories.OTPRepo,
	sessionRepo repositories.SessionRepo,
	createUserService *services.CreateUserService,
	selectPaymentPlanService *services.SelectPaymentPlanService,
	jwtService *jwt.Service,
	emailService emailInfra.Service,
) *AuthController {
	return &AuthController{
		userRepo:            userRepo,
		otpRepo:             otpRepo,
		sessionRepo:         sessionRepo,
		createUserService:   createUserService,
		selectPaymentPlanService: selectPaymentPlanService,
		jwtService:          jwtService,
		emailService:        emailService,
	}
}

// Signup handles user registration
func (c *AuthController) Signup(ctx *utils.Context) (interface{}, error) {
	var req models.SignupRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Create user command
	cmd := services.CreateUserCmd{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password, // Can be empty for initial signup
	}

	// Execute command
	result, err := c.createUserService.Execute(ctx.Request.Context(), cmd)
	if err != nil {
		return nil, err
	}

	createdUser := result.User

	// Generate and send OTP
	otpCode := generateOTP()
	otpHash := hashOTP(otpCode)

	now := time.Now()
	otp := &models.OTP{
		ID:          uuid.New(),
		UserID:      createdUser.ID,
		Email:       req.Email,
		CodeHash:    otpHash,
		Type:        "email_verification",
		ExpiresAt:   now.Add(5 * time.Minute),
		Used:        false,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   now,
	}

	if err := c.otpRepo.Create(ctx.Request.Context(), otp); err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send OTP via email service asynchronously (non-blocking)
	if c.emailService != nil {
		if asyncSvc, ok := c.emailService.(interface {
			SendOTPEmailAsync(toEmail, toName, otpCode string, expiresInMinutes int)
		}); ok {
			asyncSvc.SendOTPEmailAsync(createdUser.Email, createdUser.Name, otpCode, 5)
		} else {
			go func(email, name, code string) {
				if err := c.emailService.SendOTPEmail(email, name, code, 5); err != nil {
					// Error is logged by the email service, don't fail the request
				}
			}(createdUser.Email, createdUser.Name, otpCode)
		}
	}

	return models.SignupResponse{
		UserID:                  createdUser.ID.String(),
		Email:                   createdUser.Email,
		RequiresOTPVerification: true,
		Message:                 "OTP sent to your email address",
	}, nil
}

// Login handles user authentication
func (c *AuthController) Login(ctx *utils.Context) (interface{}, error) {
	var req models.LoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find user
	u, err := c.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.NewDomainError("INVALID_PASSWORD", "invalid password")
	}

	// Check if user is active
	if !u.IsActive() {
		return nil, errors.NewDomainError("USER_NOT_ACTIVE", "user account is not active")
	}

	// Generate JWT token
	token, err := c.jwtService.GenerateToken(u.ID.String(), u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := c.jwtService.GenerateRefreshToken(u.ID.String(), u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session with token hash for tracking
	tokenHash := generateTokenHash()
	expiresAt := time.Now().Add(24 * time.Hour) // Session expiry

	userAgent := ctx.Request.UserAgent()
	now := time.Now()
	session := &models.Session{
		ID:           uuid.New(),
		UserID:       u.ID,
		TokenHash:    tokenHash,
		Device:       getDeviceFromUserAgent(userAgent),
		IPAddress:    getIPAddress(ctx),
		UserAgent:    &userAgent,
		IsActive:     true,
		LastActiveAt: now,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
	}

	if err := c.sessionRepo.Create(ctx.Request.Context(), session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         toUserResponse(u),
	}, nil
}

// VerifyOTP handles OTP verification
func (c *AuthController) VerifyOTP(ctx *utils.Context) (interface{}, error) {
	var req models.OTPVerificationRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find OTP
	otp, err := c.otpRepo.FindByEmailAndType(ctx.Request.Context(), req.Email, "email_verification")
	if err != nil {
		// Check if it's a domain error (OTP_NOT_FOUND)
		if domainErr, ok := err.(*errors.DomainError); ok && domainErr.Code == "OTP_NOT_FOUND" {
			return nil, errors.NewDomainError("INVALID_OTP", "Invalid or expired OTP")
		}
		return nil, err
	}

	// Check if already used
	if otp.IsUsed() {
		return nil, errors.NewDomainError("OTP_ALREADY_USED", "OTP already used")
	}

	// Check if expired
	if otp.IsExpired() {
		return nil, errors.NewDomainError("OTP_EXPIRED", "OTP expired")
	}

	// Check attempts
	if !otp.CanAttempt() {
		return nil, errors.NewDomainError("OTP_MAX_ATTEMPTS", "Maximum attempts exceeded")
	}

	// Verify OTP
	otpHash := hashOTP(req.OTP)
	if otp.CodeHash != otpHash {
		otp.IncrementAttempts()
		c.otpRepo.Update(ctx.Request.Context(), otp)
		return nil, errors.NewDomainError("INVALID_OTP", "Invalid OTP code")
	}

	// Mark OTP as used
	otp.MarkAsUsed()
	if err := c.otpRepo.Update(ctx.Request.Context(), otp); err != nil {
		return nil, fmt.Errorf("failed to update OTP: %w", err)
	}

	// Find user and verify email
	u, err := c.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, err
	}

	// Update user email verification status
	now := time.Now()
	u.EmailVerified = true
	u.EmailVerifiedAt = &now
	if u.Status == "pending_verification" {
		u.Status = "verified"
	}
	u.UpdatedAt = now

	if err := c.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// Don't generate token yet if password is not set (step-by-step signup)
	if u.PasswordHash == "" {
		return models.OTPVerificationResponse{
			Message: "Email verified successfully. Please complete your profile.",
			User:    toUserResponse(u),
		}, nil
	}

	// Generate JWT token if password is set
	token, err := c.jwtService.GenerateToken(u.ID.String(), u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session with token hash for tracking
	tokenHash := generateTokenHash()
	expiresAt := time.Now().Add(24 * time.Hour)
	sessionNow := time.Now()
	session := &models.Session{
		ID:           uuid.New(),
		UserID:       u.ID,
		TokenHash:    tokenHash,
		IsActive:     true,
		LastActiveAt: sessionNow,
		ExpiresAt:    expiresAt,
		CreatedAt:    sessionNow,
	}
	if err := c.sessionRepo.Create(ctx.Request.Context(), session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return models.OTPVerificationResponse{
		Message: "Email verified successfully",
		Token:   token,
		User:    toUserResponse(u),
	}, nil
}

// ResendOTP handles OTP resend request
func (c *AuthController) ResendOTP(ctx *utils.Context) (interface{}, error) {
	var req models.ResendOTPRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find user
	u, err := c.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
	}

	// Generate new OTP
	otpCode := generateOTP()
	otpHash := hashOTP(otpCode)

	now := time.Now()
	otp := &models.OTP{
		ID:          uuid.New(),
		UserID:      u.ID,
		Email:       req.Email,
		CodeHash:    otpHash,
		Type:        "email_verification",
		ExpiresAt:   now.Add(5 * time.Minute),
		Used:        false,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   now,
	}

	if err := c.otpRepo.Create(ctx.Request.Context(), otp); err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send OTP via email service asynchronously
	if c.emailService != nil {
		if asyncSvc, ok := c.emailService.(interface {
			SendOTPEmailAsync(toEmail, toName, otpCode string, expiresInMinutes int)
		}); ok {
			asyncSvc.SendOTPEmailAsync(u.Email, u.Name, otpCode, 5)
		} else {
			go func(email, name, code string) {
				if err := c.emailService.SendOTPEmail(email, name, code, 5); err != nil {
					// Error is logged by the email service
				}
			}(u.Email, u.Name, otpCode)
		}
	}

	return models.ResendOTPResponse{
		Message:   "OTP sent to your email address",
		ExpiresIn: 300,
	}, nil
}

// UpdateProfile handles profile update (mobile, DOB, city)
func (c *AuthController) UpdateProfile(ctx *utils.Context) (interface{}, error) {
	var req models.UpdateProfileRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Get user ID from context
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Find user
	u, err := c.userRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
	}

	// Parse date of birth if provided
	var dob *time.Time
	if req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		dob = &parsed
	}

	// Update profile
	if req.Phone != "" {
		u.Phone = &req.Phone
	}
	if req.City != "" {
		u.City = &req.City
	}
	if dob != nil {
		u.DateOfBirth = dob
	}
	u.UpdatedAt = time.Now()

	if err := c.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return models.UpdateProfileResponse{
		Message: "Profile updated successfully",
		User:    toUserResponse(u),
	}, nil
}

// VerifyPAN handles PAN verification
func (c *AuthController) VerifyPAN(ctx *utils.Context) (interface{}, error) {
	var req models.VerifyPANRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Get user ID from context
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Find user
	u, err := c.userRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
	}

	// Validate customer type
	if req.CustomerType != "Individual" && req.CustomerType != "Non Individual" {
		return nil, errors.NewDomainError("INVALID_CUSTOMER_TYPE", "Customer type must be 'Individual' or 'Non Individual'")
	}

	// Set customer type
	u.CustomerType = &req.CustomerType

	// TODO: Integrate with actual PAN verification service/API
	panDetails := &models.PANDetails{
		Name:    u.Name, // Use user's name as placeholder
		Address: "Address from PAN verification service",
		Mobile:  "",
	}
	if u.Phone != nil {
		panDetails.Mobile = *u.Phone
	}

	// Set PAN on user
	u.PAN = &req.PAN
	if panDetails.Name != "" {
		u.PANName = &panDetails.Name
	}
	if panDetails.Address != "" {
		u.PANAddress = &panDetails.Address
	}
	if panDetails.Mobile != "" {
		u.PANMobile = &panDetails.Mobile
	}
	u.UpdatedAt = time.Now()

	if err := c.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to update PAN: %w", err)
	}

	return models.VerifyPANResponse{
		Message: "PAN verified successfully",
		PANDetails: &models.PANDetails{
			Name:    panDetails.Name,
			Address: panDetails.Address,
			Mobile:  panDetails.Mobile,
		},
	}, nil
}

// SetPassword handles password setting
func (c *AuthController) SetPassword(ctx *utils.Context) (interface{}, error) {
	var req models.SetPasswordRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Get user ID from context
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Find user
	u, err := c.userRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set password
	u.PasswordHash = string(passwordHash)
	u.UpdatedAt = time.Now()

	if err := c.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	// Generate JWT token now that password is set
	token, err := c.jwtService.GenerateToken(u.ID.String(), u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session
	tokenHash := generateTokenHash()
	expiresAt := time.Now().Add(24 * time.Hour)
	now := time.Now()
	session := &models.Session{
		ID:           uuid.New(),
		UserID:       u.ID,
		TokenHash:    tokenHash,
		IsActive:     true,
		LastActiveAt: now,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
	}
	if err := c.sessionRepo.Create(ctx.Request.Context(), session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return models.OTPVerificationResponse{
		Message: "Password set successfully",
		Token:   token,
		User:    toUserResponse(u),
	}, nil
}

// SelectPaymentPlan handles payment plan selection during signup
func (c *AuthController) SelectPaymentPlan(ctx *utils.Context) (interface{}, error) {
	var req models.SelectPaymentPlanRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Get user ID from context
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Find user to return in response
	u, err := c.userRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
	}

	// Validate plan and billing period
	if req.Plan != "standard" && req.Plan != "plus" && req.Plan != "premium" {
		return nil, errors.NewDomainError("INVALID_PLAN", "Plan must be 'standard', 'plus', or 'premium'")
	}

	if req.BillingPeriod != "quarterly" && req.BillingPeriod != "annual" {
		return nil, errors.NewDomainError("INVALID_BILLING_PERIOD", "Billing period must be 'quarterly' or 'annual'")
	}

	// Use service to select payment plan
	cmd := services.SelectPaymentPlanCmd{
		UserID:        userID,
		PlanName:      req.Plan,
		BillingPeriod: req.BillingPeriod,
	}

	if _, err := c.selectPaymentPlanService.Execute(ctx.Request.Context(), cmd); err != nil {
		return nil, fmt.Errorf("failed to select payment plan: %w", err)
	}

	return models.SelectPaymentPlanResponse{
		Message: "Payment plan selected successfully",
		User:    toUserResponse(u),
	}, nil
}

// Helper functions
func generateOTP() string {
	// Generate 4-digit OTP (0000-9999)
	return fmt.Sprintf("%04d", mathrand.Intn(10000))
}

func hashOTP(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

func generateTokenHash() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getDeviceFromUserAgent(userAgent string) *string {
	if len(userAgent) > 0 {
		return &userAgent
	}
	return nil
}

func getIPAddress(ctx *utils.Context) *string {
	ip := ctx.ClientIP()
	return &ip
}

func toUserResponse(u *models.User) *models.UserResponse {
	return &models.UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Phone:         u.Phone,
		Name:          u.Name,
		Status:        u.Status,
		IsMfCustomer:  u.GetIsMfCustomer(),
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
	}
}

