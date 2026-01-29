package controllers

import (
	"context"
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
	"github.com/equitywala/backend/internal/common/logger"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles authentication endpoints
type AuthController struct {
	userRepo                 repositories.UserRepo
	otpRepo                  repositories.OTPRepo
	sessionRepo              repositories.SessionRepo
	paymentPlanSelectionRepo repositories.PaymentPlanSelectionRepo
	createUserService        *services.CreateUserService
	selectPaymentPlanService *services.SelectPaymentPlanService
	jwtService              *jwt.Service
	emailService            emailInfra.Service
	roleService             *services.RoleService
	logger                  logger.Logger
}

// NewAuthController creates a new auth controller
func NewAuthController(
	userRepo repositories.UserRepo,
	otpRepo repositories.OTPRepo,
	sessionRepo repositories.SessionRepo,
	paymentPlanSelectionRepo repositories.PaymentPlanSelectionRepo,
	createUserService *services.CreateUserService,
	selectPaymentPlanService *services.SelectPaymentPlanService,
	jwtService *jwt.Service,
	emailService emailInfra.Service,
	roleService *services.RoleService,
	logger logger.Logger,
) *AuthController {
	return &AuthController{
		userRepo:                 userRepo,
		otpRepo:                  otpRepo,
		sessionRepo:              sessionRepo,
		paymentPlanSelectionRepo: paymentPlanSelectionRepo,
		createUserService:        createUserService,
		selectPaymentPlanService: selectPaymentPlanService,
		jwtService:               jwtService,
		emailService:             emailService,
		roleService:              roleService,
		logger:                   logger,
	}
}

// Signup handles user registration
func (c *AuthController) Signup(ctx *utils.Context) (interface{}, error) {
	var req models.SignupRequest
	if err := ctx.BindJSON(&req); err != nil {
		c.logger.Error("Signup: Failed to bind request", "error", err)
		return nil, err
	}

	c.logger.Info("Signup: Processing signup request", "email", req.Email, "name", req.Name)

	// Check if user already exists
	exists, err := c.userRepo.ExistsByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		c.logger.Error("Signup: Failed to check user existence", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	
	c.logger.Debug("Signup: User existence check completed", "email", req.Email, "exists", exists)
	
	// If user exists, check if they've completed signup
	if exists {
		c.logger.Info("Signup: User already exists, checking signup status", "email", req.Email)
		
		existingUser, err := c.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
		if err != nil {
			c.logger.Error("Signup: Failed to find existing user", "email", req.Email, "error", err)
			return nil, fmt.Errorf("failed to find existing user: %w", err)
		}
		
		// Check if user has completed signup (has password and active subscription or payment)
		// If not, allow them to continue signup by resending OTP
		hasPassword := existingUser.PasswordHash != ""
		hasActiveSubscription := existingUser.Status == "active"
		
		c.logger.Debug("Signup: Existing user status", 
			"email", req.Email, 
			"userID", existingUser.ID.String(),
			"hasPassword", hasPassword,
			"hasActiveSubscription", hasActiveSubscription,
			"status", existingUser.Status)
		
		// If user has password and is active, they're fully signed up
		if hasPassword && hasActiveSubscription {
			c.logger.Warn("Signup: User already exists and is fully signed up", "email", req.Email, "userID", existingUser.ID.String())
			return nil, errors.NewDomainError("USER_ALREADY_EXISTS", "user with this email already exists")
		}
		
		// User exists but hasn't completed signup - resend OTP to continue
		c.logger.Info("Signup: User exists but signup incomplete, resending OTP", "email", req.Email, "userID", existingUser.ID.String())
		
		// Generate new OTP
		otpCode := generateOTP()
		c.logger.Debug("Signup: Generated OTP for existing user", "email", req.Email, "otpLength", len(otpCode))
		
		// Send OTP via email service
		if c.emailService == nil {
			c.logger.Error("Signup: Email service is not configured", "email", req.Email)
			return nil, fmt.Errorf("email service is not configured")
		}
		
		c.logger.Info("Signup: Sending OTP email to existing user", "email", req.Email, "name", existingUser.Name)
		if err := c.emailService.SendOTPEmail(req.Email, existingUser.Name, otpCode, 5); err != nil {
			c.logger.Error("Signup: Failed to send OTP email to existing user", 
				"email", req.Email, 
				"error", err)
			return nil, fmt.Errorf("failed to send OTP email: %w", err)
		}
		
		// Create new OTP record
		otpHash := hashOTP(otpCode)
		now := time.Now()
		otp := &models.OTP{
			ID:          uuid.New(),
			UserID:      existingUser.ID,
			Email:       req.Email,
			CodeHash:    otpHash,
			Type:        "email_verification",
			ExpiresAt:   now.Add(5 * time.Minute),
			Used:        false,
			Attempts:    0,
			MaxAttempts: 3,
			CreatedAt:   now,
		}
		
		c.logger.Debug("Signup: Creating OTP record for existing user", "email", req.Email, "otpID", otp.ID.String())
		if err := c.otpRepo.Create(ctx.Request.Context(), otp); err != nil {
			c.logger.Error("Signup: Failed to create OTP record for existing user", 
				"email", req.Email, 
				"otpID", otp.ID.String(),
				"error", err)
			return nil, fmt.Errorf("failed to create OTP: %w", err)
		}
		
		c.logger.Info("Signup: Successfully sent OTP to existing user", 
			"email", req.Email, 
			"userID", existingUser.ID.String(),
			"otpID", otp.ID.String())
		
		return models.SignupResponse{
			UserID:                  existingUser.ID.String(),
			Email:                   existingUser.Email,
			RequiresOTPVerification: true,
			Message:                 "OTP sent to your email address. Continue your signup.",
		}, nil
	}

	// New user signup flow
	c.logger.Info("Signup: Processing new user signup", "email", req.Email, "name", req.Name)

	// Generate OTP code first
	otpCode := generateOTP()
	c.logger.Debug("Signup: Generated OTP for new user", "email", req.Email, "otpLength", len(otpCode))

	// Send OTP via email service synchronously (blocking)
	// Only proceed with user creation if email is successfully sent
	if c.emailService == nil {
		c.logger.Error("Signup: Email service is not configured", "email", req.Email)
		return nil, fmt.Errorf("email service is not configured")
	}
	
	c.logger.Info("Signup: Sending OTP email to new user", "email", req.Email, "name", req.Name)
	if err := c.emailService.SendOTPEmail(req.Email, req.Name, otpCode, 5); err != nil {
		c.logger.Error("Signup: Failed to send OTP email to new user", 
			"email", req.Email, 
			"error", err)
		return nil, fmt.Errorf("failed to send OTP email: %w", err)
	}
	
	c.logger.Info("Signup: OTP email sent successfully, creating user", "email", req.Email)

	// Email sent successfully, now create user and OTP in database
	cmd := services.CreateUserCmd{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password, // Can be empty for initial signup
	}

	// Execute command to create user
	c.logger.Debug("Signup: Executing create user service", "email", req.Email)
	result, err := c.createUserService.Execute(ctx.Request.Context(), cmd)
	if err != nil {
		c.logger.Error("Signup: Failed to create user", "email", req.Email, "error", err)
		return nil, err
	}

	createdUser := result.User
	c.logger.Info("Signup: User created successfully", "email", req.Email, "userID", createdUser.ID.String())

	// Create OTP record in database
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

	c.logger.Debug("Signup: Creating OTP record for new user", "email", req.Email, "otpID", otp.ID.String())
	if err := c.otpRepo.Create(ctx.Request.Context(), otp); err != nil {
		c.logger.Error("Signup: Failed to create OTP record for new user", 
			"email", req.Email, 
			"userID", createdUser.ID.String(),
			"otpID", otp.ID.String(),
			"error", err)
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	c.logger.Info("Signup: Signup completed successfully", 
		"email", req.Email, 
		"userID", createdUser.ID.String(),
		"otpID", otp.ID.String())

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

	// Calculate signup progress for journey tracking
	progress := c.calculateSignupProgress(ctx.Request.Context(), u)

	// Get user roles
	userResponse := c.toUserResponse(ctx.Request.Context(), u)

	return models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         userResponse,
		Progress:     progress,
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

	// Calculate signup progress
	progress := c.calculateSignupProgress(ctx.Request.Context(), u)

	// Generate JWT token for all verified users (including those without password for onboarding)
	// This allows users to complete onboarding steps (profile, PAN, password, subscription)
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

	// Return response with token for all users (with or without password)
	// Users without password can use this token to complete onboarding steps
	message := "Email verified successfully"
	if u.PasswordHash == "" {
		message = "Email verified successfully. Please complete your profile."
	}

	// Get user roles
	userResponse := c.toUserResponse(ctx.Request.Context(), u)

	return models.OTPVerificationResponse{
		Message:  message,
		Token:    token,
		User:     userResponse,
		Progress: progress,
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

	// Get user roles
	userResponse := c.toUserResponse(ctx.Request.Context(), u)

	return models.UpdateProfileResponse{
		Message: "Profile updated successfully",
		User:    userResponse,
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
// This endpoint can be called without authentication if user has verified their email via OTP
func (c *AuthController) SetPassword(ctx *utils.Context) (interface{}, error) {
	var req models.SetPasswordRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	var u *models.User
	var err error

	// Try to get user ID from context (if authenticated)
	userIDStr, exists := ctx.Get("user_id")
	if exists {
		// User is authenticated, use user ID from context
		userID, parseErr := uuid.Parse(userIDStr.(string))
		if parseErr != nil {
			return nil, fmt.Errorf("invalid user ID: %w", parseErr)
		}
		u, err = c.userRepo.FindByID(ctx.Request.Context(), userID)
		if err != nil {
			return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
		}
	} else {
		// User is not authenticated, identify by email (must have verified OTP)
		if req.Email == "" {
			return nil, fmt.Errorf("email is required when not authenticated")
		}
		u, err = c.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
		if err != nil {
			return nil, errors.NewDomainError("USER_NOT_FOUND", "user not found")
		}
		// Verify that user has verified their email (required for password setting)
		if !u.EmailVerified {
			return nil, errors.NewDomainError("EMAIL_NOT_VERIFIED", "email must be verified before setting password")
		}
		// If password is already set, return progress to help user continue onboarding
		if u.PasswordHash != "" {
			// Calculate progress to determine next step
			progress := c.calculateSignupProgress(ctx.Request.Context(), u)
			
			// Generate token for authenticated access
			token, err := c.jwtService.GenerateToken(u.ID.String(), u.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to generate token: %w", err)
			}
			
			// Get user roles
			userResponse := c.toUserResponse(ctx.Request.Context(), u)

			// Return response with progress so frontend can navigate to next pending step
			return models.OTPVerificationResponse{
				Message:  "Password already set. Continuing onboarding...",
				Token:    token,
				User:     userResponse,
				Progress: progress,
			}, nil
		}
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

	// Calculate signup progress after password is set
	progress := c.calculateSignupProgress(ctx.Request.Context(), u)

	// Get user roles
	userResponse := c.toUserResponse(ctx.Request.Context(), u)

	return models.OTPVerificationResponse{
		Message:  "Password set successfully",
		Token:   token,
		User:    userResponse,
		Progress: progress,
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

	// Mark payment as completed by updating user status to active
	// This skips the payment gateway for now and marks the record as paid
	u.Status = "active"
	u.UpdatedAt = time.Now()
	if err := c.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	// Update payment plan selection status to completed
	planSelection, err := c.paymentPlanSelectionRepo.FindByUserID(ctx.Request.Context(), userID)
	if err == nil && planSelection != nil {
		planSelection.Status = "completed"
		planSelection.UpdatedAt = time.Now()
		if err := c.paymentPlanSelectionRepo.Update(ctx.Request.Context(), planSelection); err != nil {
			c.logger.Warn("Failed to update payment plan selection status", "error", err)
			// Don't fail the request if this update fails, user status is already updated
		}
	}

	// Get user roles
	userResponse := c.toUserResponse(ctx.Request.Context(), u)

	return models.SelectPaymentPlanResponse{
		Message: "Payment plan selected and activated successfully",
		User:    userResponse,
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

// calculateSignupProgress calculates the signup progress for a user
func (c *AuthController) calculateSignupProgress(ctx context.Context, u *models.User) *models.SignupProgress {
	progress := &models.SignupProgress{
		EmailVerified:    u.EmailVerified,
		ProfileCompleted: u.Phone != nil && u.DateOfBirth != nil && u.City != nil,
		PANVerified:      u.PAN != nil && u.PANName != nil,
		PasswordSet:      u.PasswordHash != "",
		PlanSelected:     false,
		PaymentCompleted: false,
		NextStep:         "",
	}

	// Check if payment plan is selected
	planSelection, err := c.paymentPlanSelectionRepo.FindByUserID(ctx, u.ID)
	if err == nil && planSelection != nil {
		progress.PlanSelected = true
	}

	// Check if payment is completed (has active subscription)
	// This would require checking subscriptions table, but for now we'll use status
	progress.PaymentCompleted = u.Status == "active"

	// Determine next step based on progress
	// Flow: OTP -> Password -> Profile -> PAN -> Pricing
	if !progress.PasswordSet {
		progress.NextStep = "/signup/password"
	} else if !progress.ProfileCompleted {
		progress.NextStep = "/signup/your-details"
	} else if !progress.PANVerified {
		progress.NextStep = "/signup/verify-pan"
	} else if !progress.PlanSelected {
		progress.NextStep = "/signup/pricing"
	} else if !progress.PaymentCompleted {
		progress.NextStep = "/signup/pricing" // Still need to complete payment
	} else {
		progress.NextStep = "/" // Dashboard
	}

	return progress
}

// toUserResponse converts a User model to UserResponse with roles
func (c *AuthController) toUserResponse(ctx context.Context, u *models.User) *models.UserResponse {
	userResponse := &models.UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Phone:         u.Phone,
		Name:          u.Name,
		Status:        u.Status,
		IsMfCustomer:  u.GetIsMfCustomer(),
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
	}

	// Get user roles if role service is available
	if c.roleService != nil {
		roles, err := c.roleService.GetUserRoles(ctx, u.ID)
		if err == nil {
			userResponse.Roles = roles

			// Get primary role
			primaryRole, err := c.roleService.GetPrimaryRole(ctx, u.ID)
			if err == nil {
				userResponse.PrimaryRole = primaryRole
			} else {
				c.logger.Warn("toUserResponse: GetPrimaryRole failed, defaulting to user", "user_id", u.ID, "error", err)
				userResponse.PrimaryRole = "user"
			}
		} else {
			c.logger.Warn("toUserResponse: GetUserRoles failed, defaulting to user", "user_id", u.ID, "error", err)
			userResponse.Roles = []string{"user"}
			userResponse.PrimaryRole = "user"
		}
	} else {
		c.logger.Warn("toUserResponse: role service is nil, defaulting to user", "user_id", u.ID)
		userResponse.Roles = []string{"user"}
		userResponse.PrimaryRole = "user"
	}

	return userResponse
}

