package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"time"

	subCommands "github.com/equitywala/backend/internal/application/subscription/commands"
	userCommands "github.com/equitywala/backend/internal/application/user/commands"
	"github.com/equitywala/backend/internal/domain/auth"
	"github.com/equitywala/backend/internal/domain/email"
	"github.com/equitywala/backend/internal/domain/user"
	"github.com/equitywala/backend/internal/interfaces/http/api"
	"github.com/equitywala/backend/internal/interfaces/http/dto"
	"github.com/equitywala/backend/internal/shared/errors"
	"github.com/equitywala/backend/internal/shared/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo            user.Repository
	otpRepo             auth.OTPRepository
	sessionRepo         auth.SessionRepository
	createUserCmd        *userCommands.CreateUserHandler
	selectPaymentPlanCmd *subCommands.SelectPaymentPlanHandler
	jwtService           *jwt.Service
	emailService         email.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo user.Repository,
	otpRepo auth.OTPRepository,
	sessionRepo auth.SessionRepository,
	createUserCmd *userCommands.CreateUserHandler,
	selectPaymentPlanCmd *subCommands.SelectPaymentPlanHandler,
	jwtService *jwt.Service,
	emailService email.Service,
) *AuthHandler {
	return &AuthHandler{
		userRepo:            userRepo,
		otpRepo:             otpRepo,
		sessionRepo:         sessionRepo,
		createUserCmd:        createUserCmd,
		selectPaymentPlanCmd: selectPaymentPlanCmd,
		jwtService:           jwtService,
		emailService:         emailService,
	}
}

// Signup handles user registration
func (h *AuthHandler) Signup(ctx *api.Context) (interface{}, error) {
	var req dto.SignupRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Create user command (password optional for step-by-step signup)
	cmd := userCommands.CreateUserCommand{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password, // Can be empty for initial signup
	}

	// Execute command
	if err := h.createUserCmd.Handle(ctx.Request.Context(), cmd); err != nil {
		return nil, err
	}

	// Find created user
	createdUser, err := h.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find created user: %w", err)
	}

	// Generate and send OTP
	otpCode := generateOTP()
	otpHash := hashOTP(otpCode)

	otp := auth.NewOTP(
		createdUser.ID,
		req.Email,
		otpHash,
		"email_verification",
		time.Now().Add(5*time.Minute),
	)

	if err := h.otpRepo.Create(ctx.Request.Context(), otp); err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send OTP via email service asynchronously (non-blocking)
	// Email sending is done in a goroutine so it doesn't block the HTTP response
	if h.emailService != nil {
		// Try to use async method if available, otherwise use goroutine
		if asyncSvc, ok := h.emailService.(interface {
			SendOTPEmailAsync(toEmail, toName, otpCode string, expiresInMinutes int)
		}); ok {
			asyncSvc.SendOTPEmailAsync(createdUser.Email, createdUser.Name, otpCode, 5)
		} else {
			// Fallback: use goroutine with sync method
			go func(email, name, code string) {
				if err := h.emailService.SendOTPEmail(email, name, code, 5); err != nil {
					// Error is logged by the email service, don't fail the request
				}
			}(createdUser.Email, createdUser.Name, otpCode)
		}
	}

	return dto.SignupResponse{
		UserID:                  createdUser.ID.String(),
		Email:                   createdUser.Email,
		RequiresOTPVerification: true,
		Message:                 "OTP sent to your email address",
	}, nil
}

// Login handles user authentication
func (h *AuthHandler) Login(ctx *api.Context) (interface{}, error) {
	var req dto.LoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find user
	u, err := h.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, user.ErrInvalidPassword
	}

	// Check if user is active
	if !u.IsActive() {
		return nil, user.ErrUserNotActive
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(u.ID.String(), u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := h.jwtService.GenerateRefreshToken(u.ID.String(), u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session with token hash for tracking
	tokenHash := generateTokenHash()
	expiresAt := time.Now().Add(24 * time.Hour) // Session expiry (can be different from JWT expiry)

	userAgent := ctx.Request.UserAgent()
	session := auth.NewSession(u.ID, tokenHash, expiresAt)
	session.Device = getDeviceFromUserAgent(userAgent)
	session.IPAddress = getIPAddress(ctx)
	session.UserAgent = &userAgent

	if err := h.sessionRepo.Create(ctx.Request.Context(), session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return dto.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         toUserResponse(u),
	}, nil
}

// VerifyOTP handles OTP verification
func (h *AuthHandler) VerifyOTP(ctx *api.Context) (interface{}, error) {
	var req dto.OTPVerificationRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find OTP
	otp, err := h.otpRepo.FindByEmailAndType(ctx.Request.Context(), req.Email, "email_verification")
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
		h.otpRepo.Update(ctx.Request.Context(), otp)
		return nil, errors.NewDomainError("INVALID_OTP", "Invalid OTP code")
	}

	// Mark OTP as used
	otp.MarkAsUsed()
	if err := h.otpRepo.Update(ctx.Request.Context(), otp); err != nil {
		return nil, fmt.Errorf("failed to update OTP: %w", err)
	}

	// Find user and verify email
	u, err := h.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, err
	}

	u.VerifyEmail()
	if err := h.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// Don't generate token yet if password is not set (step-by-step signup)
	// User needs to complete profile and set password first
	if u.PasswordHash == "" {
		return dto.OTPVerificationResponse{
			Message: "Email verified successfully. Please complete your profile.",
			User:    toUserResponse(u),
		}, nil
	}

	// Generate JWT token if password is set
	token, err := h.jwtService.GenerateToken(u.ID.String(), u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session with token hash for tracking
	tokenHash := generateTokenHash()
	expiresAt := time.Now().Add(24 * time.Hour) // Session expiry
	session := auth.NewSession(u.ID, tokenHash, expiresAt)
	if err := h.sessionRepo.Create(ctx.Request.Context(), session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return dto.OTPVerificationResponse{
		Message: "Email verified successfully",
		Token:   token,
		User:    toUserResponse(u),
	}, nil
}

// ResendOTP handles OTP resend request
func (h *AuthHandler) ResendOTP(ctx *api.Context) (interface{}, error) {
	var req dto.ResendOTPRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Find user
	u, err := h.userRepo.FindByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	// Generate new OTP
	otpCode := generateOTP()
	otpHash := hashOTP(otpCode)

	otp := auth.NewOTP(
		u.ID,
		req.Email,
		otpHash,
		"email_verification",
		time.Now().Add(5*time.Minute),
	)

	if err := h.otpRepo.Create(ctx.Request.Context(), otp); err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send OTP via email service asynchronously (non-blocking)
	// Email sending is done in a goroutine so it doesn't block the HTTP response
	if h.emailService != nil {
		// Try to use async method if available, otherwise use goroutine
		if asyncSvc, ok := h.emailService.(interface {
			SendOTPEmailAsync(toEmail, toName, otpCode string, expiresInMinutes int)
		}); ok {
			asyncSvc.SendOTPEmailAsync(u.Email, u.Name, otpCode, 5)
		} else {
			// Fallback: use goroutine with sync method
			go func(email, name, code string) {
				if err := h.emailService.SendOTPEmail(email, name, code, 5); err != nil {
					// Error is logged by the email service, don't fail the request
				}
			}(u.Email, u.Name, otpCode)
		}
	}

	return dto.ResendOTPResponse{
		Message:   "OTP sent to your email address",
		ExpiresIn: 300,
	}, nil
}

// UpdateProfile handles profile update (mobile, DOB, city)
func (h *AuthHandler) UpdateProfile(ctx *api.Context) (interface{}, error) {
	var req dto.UpdateProfileRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, err
	}

	// Get user ID from context (set by auth middleware)
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Find user
	u, err := h.userRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		return nil, user.ErrUserNotFound
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
	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	}
	var city *string
	if req.City != "" {
		city = &req.City
	}

	u.UpdateProfile(phone, dob, city)

	if err := h.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return dto.UpdateProfileResponse{
		Message: "Profile updated successfully",
		User:    toUserResponse(u),
	}, nil
}

// VerifyPAN handles PAN verification
func (h *AuthHandler) VerifyPAN(ctx *api.Context) (interface{}, error) {
	var req dto.VerifyPANRequest
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
	u, err := h.userRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	// Validate customer type
	if req.CustomerType != "Individual" && req.CustomerType != "Non Individual" {
		return nil, errors.NewDomainError("INVALID_CUSTOMER_TYPE", "Customer type must be 'Individual' or 'Non Individual'")
	}

	// Set customer type
	u.SetCustomerType(req.CustomerType)

	// TODO: Integrate with actual PAN verification service/API
	// For now, return mock data based on PAN
	// In production, this should call an external PAN verification service
	panDetails := &user.PANDetails{
		Name:    u.Name, // Use user's name as placeholder
		Address: "Address from PAN verification service",
		Mobile:  "",
	}
	if u.Phone != nil {
		panDetails.Mobile = *u.Phone
	}

	// Set PAN on user
	u.SetPAN(req.PAN, panDetails)

	if err := h.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to update PAN: %w", err)
	}

	return dto.VerifyPANResponse{
		Message: "PAN verified successfully",
		PANDetails: &dto.PANDetails{
			Name:    panDetails.Name,
			Address: panDetails.Address,
			Mobile:  panDetails.Mobile,
		},
	}, nil
}

// SetPassword handles password setting
func (h *AuthHandler) SetPassword(ctx *api.Context) (interface{}, error) {
	var req dto.SetPasswordRequest
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
	u, err := h.userRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set password
	u.SetPassword(string(passwordHash))

	if err := h.userRepo.Update(ctx.Request.Context(), u); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	// Generate JWT token now that password is set
	token, err := h.jwtService.GenerateToken(u.ID.String(), u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session
	tokenHash := generateTokenHash()
	expiresAt := time.Now().Add(24 * time.Hour)
	session := auth.NewSession(u.ID, tokenHash, expiresAt)
	if err := h.sessionRepo.Create(ctx.Request.Context(), session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return dto.OTPVerificationResponse{
		Message: "Password set successfully",
		Token:   token,
		User:    toUserResponse(u),
	}, nil
}

// SelectPaymentPlan handles payment plan selection during signup
func (h *AuthHandler) SelectPaymentPlan(ctx *api.Context) (interface{}, error) {
	var req dto.SelectPaymentPlanRequest
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
	u, err := h.userRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	// Validate plan and billing period (validation already done by binding, but double-check)
	if req.Plan != "standard" && req.Plan != "plus" && req.Plan != "premium" {
		return nil, errors.NewDomainError("INVALID_PLAN", "Plan must be 'standard', 'plus', or 'premium'")
	}

	if req.BillingPeriod != "quarterly" && req.BillingPeriod != "annual" {
		return nil, errors.NewDomainError("INVALID_BILLING_PERIOD", "Billing period must be 'quarterly' or 'annual'")
	}

	// Use command handler to select payment plan
	if h.selectPaymentPlanCmd == nil {
		return nil, fmt.Errorf("select payment plan command handler not initialized")
	}

	if err := h.selectPaymentPlanCmd.Handle(ctx.Request.Context(), subCommands.SelectPaymentPlanCommand{
		UserID:        userID,
		PlanName:      req.Plan,
		BillingPeriod: req.BillingPeriod,
	}); err != nil {
		return nil, fmt.Errorf("failed to select payment plan: %w", err)
	}

	return dto.SelectPaymentPlanResponse{
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
	// Simple device detection (can be enhanced)
	if len(userAgent) > 0 {
		return &userAgent
	}
	return nil
}

func getIPAddress(ctx *api.Context) *string {
	ip := ctx.ClientIP()
	return &ip
}

func toUserResponse(u *user.User) *dto.UserResponse {
	return &dto.UserResponse{
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
