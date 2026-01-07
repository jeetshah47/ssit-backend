package dto

// SignupRequest represents the signup request
// Password is optional for initial signup (can be set later)
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=1"`
	Password string `json:"password,omitempty"` // Optional for step-by-step signup
}

// SignupResponse represents the signup response
type SignupResponse struct {
	UserID                  string `json:"userId"`
	Email                   string `json:"email"`
	RequiresOTPVerification bool   `json:"requiresOTPVerification"`
	Message                 string `json:"message,omitempty"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token        string        `json:"token"`
	RefreshToken string        `json:"refreshToken,omitempty"`
	User         *UserResponse `json:"user"`
}

// OTPVerificationRequest represents the OTP verification request
type OTPVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=4"` // Changed to 4-digit OTP
}

// OTPVerificationResponse represents the OTP verification response
type OTPVerificationResponse struct {
	Message string        `json:"message"`
	Token   string        `json:"token,omitempty"`
	User    *UserResponse `json:"user,omitempty"`
}

// ResendOTPRequest represents the resend OTP request
type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResendOTPResponse represents the resend OTP response
type ResendOTPResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expiresIn"`
}

// UpdateProfileRequest represents the profile update request
type UpdateProfileRequest struct {
	Phone       string `json:"phone,omitempty"`
	DateOfBirth string `json:"dateOfBirth,omitempty"` // ISO 8601 format
	City        string `json:"city,omitempty"`
}

// UpdateProfileResponse represents the profile update response
type UpdateProfileResponse struct {
	Message string        `json:"message"`
	User    *UserResponse `json:"user"`
}

// VerifyPANRequest represents the PAN verification request
type VerifyPANRequest struct {
	PAN          string `json:"pan" binding:"required,min=10,max=10"`
	CustomerType string `json:"customerType" binding:"required"`
}

// VerifyPANResponse represents the PAN verification response
type VerifyPANResponse struct {
	Message    string      `json:"message"`
	PANDetails *PANDetails `json:"panDetails,omitempty"`
}

// PANDetails represents PAN verification details
type PANDetails struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Mobile  string `json:"mobile"`
}

// SetPasswordRequest represents the set password request
type SetPasswordRequest struct {
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

// SelectPaymentPlanRequest represents the payment plan selection request
type SelectPaymentPlanRequest struct {
	Plan          string `json:"plan" binding:"required,oneof=standard plus premium"`
	BillingPeriod string `json:"billingPeriod" binding:"required,oneof=quarterly annual"`
}

// SelectPaymentPlanResponse represents the payment plan selection response
type SelectPaymentPlanResponse struct {
	Message string        `json:"message"`
	User    *UserResponse `json:"user"`
}
