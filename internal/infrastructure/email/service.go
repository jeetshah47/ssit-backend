package email

import (
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/equitywala/backend/internal/common/logger"
	"github.com/equitywala/backend/internal/core/config"
)

// Service defines the interface for email operations
type Service interface {
	// SendOTPEmail sends an OTP verification email to the user
	SendOTPEmail(toEmail, toName, otpCode string, expiresInMinutes int) error
}

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

// EmailService implements the email service interface
type EmailService struct {
	config     *config.Config
	logger     logger.Logger
	smtpClient *SMTPClient
	template   *template.Template
}

// NewService creates a new email service using SMTP
func NewService(cfg *config.Config, log logger.Logger) (Service, error) {
	// Validate email configuration - use SMTP
	var smtpClient *SMTPClient
	var err error

	// Log configuration status (without exposing secrets)
	log.Debug("Email service: Checking SMTP configuration",
		"hasSMTPHost", cfg.Email.SMTPHost != "",
		"hasSMTPUsername", cfg.Email.SMTPUsername != "",
		"hasSMTPPassword", cfg.Email.SMTPPassword != "",
		"smtpPort", cfg.Email.SMTPPort,
		"fromEmail", cfg.Email.FromEmail,
	)

	// Validate SMTP configuration
	if cfg.Email.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP_HOST must be configured")
	}
	if cfg.Email.SMTPPort == 0 {
		return nil, fmt.Errorf("SMTP_PORT must be configured")
	}
	if cfg.Email.SMTPUsername == "" {
		return nil, fmt.Errorf("SMTP_USERNAME must be configured")
	}
	if cfg.Email.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD must be configured")
	}

	// Create SMTP client
	log.Info("Email service: Attempting to initialize SMTP client",
		"smtpHost", cfg.Email.SMTPHost,
		"smtpPort", cfg.Email.SMTPPort,
		"fromEmail", cfg.Email.FromEmail,
	)
	smtpClient, err = NewSMTPClient(&cfg.Email, log)
	if err != nil {
		log.Error("Email service: Failed to initialize SMTP client",
			"error", err,
			"smtpHost", cfg.Email.SMTPHost,
			"smtpPort", cfg.Email.SMTPPort,
			"fromEmail", cfg.Email.FromEmail,
		)
		return nil, fmt.Errorf("failed to initialize SMTP client: %w", err)
	}
	log.Info("Email service: Using SMTP for email delivery",
		"smtpHost", cfg.Email.SMTPHost,
		"smtpPort", cfg.Email.SMTPPort,
		"fromEmail", cfg.Email.FromEmail,
	)

	// Load email template
	tmpl, err := loadTemplate()
	if err != nil {
		return nil, fmt.Errorf("failed to load email template: %w", err)
	}

	return &EmailService{
		config:     cfg,
		logger:     log,
		smtpClient: smtpClient,
		template:   tmpl,
	}, nil
}

// SendOTPEmail sends an OTP verification email (synchronous)
func (s *EmailService) SendOTPEmail(toEmail, toName, otpCode string, expiresInMinutes int) error {
	return s.sendOTPEmailSync(toEmail, toName, otpCode, expiresInMinutes)
}

// SendOTPEmailAsync sends an OTP verification email asynchronously using a goroutine
// This method doesn't block and errors are logged but not returned
func (s *EmailService) SendOTPEmailAsync(toEmail, toName, otpCode string, expiresInMinutes int) {
	go func() {
		if err := s.sendOTPEmailSync(toEmail, toName, otpCode, expiresInMinutes); err != nil {
			s.logger.Error("Failed to send OTP email asynchronously", "to", toEmail, "error", err)
		}
	}()
}

// sendOTPEmailSync is the internal synchronous implementation
func (s *EmailService) sendOTPEmailSync(toEmail, toName, otpCode string, expiresInMinutes int) error {
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

	// Send email using SMTP
	ctx := context.Background()
	subject := "Verify your email address - Equitywala"

	if err := s.smtpClient.SendEmail(ctx, toEmail, toName, subject, htmlBody.String()); err != nil {
		s.logger.Error("Failed to send email via SMTP",
			"to", toEmail,
			"error", err,
		)
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
