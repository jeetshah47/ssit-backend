package email

// Service defines the interface for email operations
type Service interface {
	// SendOTPEmail sends an OTP verification email to the user
	SendOTPEmail(toEmail, toName, otpCode string, expiresInMinutes int) error
}

