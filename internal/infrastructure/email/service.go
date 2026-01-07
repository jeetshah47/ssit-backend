package email

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/equitywala/backend/internal/config"
	"github.com/equitywala/backend/internal/domain/email"
	"github.com/equitywala/backend/internal/shared/logger"
	"gopkg.in/mail.v2"
)

// TemplateData holds data for email templates
type TemplateData struct {
	RecipientName         string
	OTPCode               string
	OTPDigits             []string
	ExpiresInMinutes      int
	VerificationLink      string
	UnsubscribeLink       string
	ManagePreferencesLink string
	RecipientEmail        string
	CompanyName           string
	FromEmail             string
	FromName              string
}

// Service implements the email domain service
type Service struct {
	config   *config.Config
	logger   logger.Logger
	dialer   *mail.Dialer
	template *template.Template
}

// NewService creates a new email service
func NewService(cfg *config.Config, log logger.Logger) (email.Service, error) {
	// Validate email configuration
	if cfg.Email.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP_HOST is not configured")
	}
	if cfg.Email.SMTPUsername == "" || cfg.Email.SMTPPassword == "" {
		log.Warn("Email service: SMTP credentials not configured. Email sending will fail until credentials are set.")
	}

	// Create mail dialer
	dialer := mail.NewDialer(cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.SMTPUsername, cfg.Email.SMTPPassword)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	// Load email template
	tmpl, err := loadTemplate()
	if err != nil {
		return nil, fmt.Errorf("failed to load email template: %w", err)
	}

	return &Service{
		config:   cfg,
		logger:   log,
		dialer:   dialer,
		template: tmpl,
	}, nil
}

// SendOTPEmail sends an OTP verification email (synchronous)
func (s *Service) SendOTPEmail(toEmail, toName, otpCode string, expiresInMinutes int) error {
	return s.sendOTPEmailSync(toEmail, toName, otpCode, expiresInMinutes)
}

// SendOTPEmailAsync sends an OTP verification email asynchronously using a goroutine
// This method doesn't block and errors are logged but not returned
func (s *Service) SendOTPEmailAsync(toEmail, toName, otpCode string, expiresInMinutes int) {
	go func() {
		if err := s.sendOTPEmailSync(toEmail, toName, otpCode, expiresInMinutes); err != nil {
			s.logger.Error("Failed to send OTP email asynchronously", "to", toEmail, "error", err)
		}
	}()
}

// sendOTPEmailSync is the internal synchronous implementation
func (s *Service) sendOTPEmailSync(toEmail, toName, otpCode string, expiresInMinutes int) error {
	// Split OTP code into individual digits
	otpDigits := make([]string, len(otpCode))
	for i, char := range otpCode {
		otpDigits[i] = string(char)
	}

	// Prepare template data
	data := TemplateData{
		RecipientName:         toName,
		OTPCode:               otpCode,
		OTPDigits:             otpDigits,
		ExpiresInMinutes:      expiresInMinutes,
		VerificationLink:      fmt.Sprintf("%s/verify-email?code=%s", s.config.Email.FrontendURL, otpCode),
		UnsubscribeLink:       fmt.Sprintf("%s/unsubscribe", s.config.Email.FrontendURL),
		ManagePreferencesLink: fmt.Sprintf("%s/email-preferences", s.config.Email.FrontendURL),
		RecipientEmail:        toEmail,
		CompanyName:           "SSIT Finserv Private limited",
		FromEmail:             s.config.Email.FromEmail,
		FromName:              s.config.Email.FromName,
	}

	// Render HTML body
	var htmlBody strings.Builder
	if err := s.template.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Create email message
	msg := mail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.Email.FromName, s.config.Email.FromEmail))
	msg.SetHeader("To", toEmail)
	msg.SetHeader("Subject", "Verify your email address - Equitywala")
	msg.SetBody("text/html", htmlBody.String())

	// Send email
	if err := s.dialer.DialAndSend(msg); err != nil {
		// Provide more helpful error messages for common SMTP errors
		errorMsg := err.Error()
		if contains(errorMsg, "535") || contains(errorMsg, "Authentication") || contains(errorMsg, "Invalid credentials") {
			s.logger.Error("Failed to send email: SMTP authentication failed. Please check SMTP_USERNAME and SMTP_PASSWORD in your .env file",
				"to", toEmail,
				"error", err,
				"hint", "For Gmail, use an App Password (not your regular password). See internal/infrastructure/email/README.md")
		} else {
			s.logger.Error("Failed to send email", "to", toEmail, "error", err)
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("OTP email sent successfully", "to", toEmail)

	return nil
}

// loadTemplate loads the email template
func loadTemplate() (*template.Template, error) {
	// Try to load from templates directory
	templatePath := filepath.Join("templates", "email", "otp.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		// If file doesn't exist, use embedded template
		return template.New("otp").Parse(otpEmailTemplate)
	}
	return tmpl, nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
