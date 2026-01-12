package email

import (
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/equitywala/backend/internal/core/config"
	"github.com/equitywala/backend/internal/common/logger"
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
	config   *config.Config
	logger   logger.Logger
	sesClient *SESClient
	template *template.Template
}

// NewService creates a new email service using AWS SES
func NewService(cfg *config.Config, log logger.Logger) (Service, error) {
	// Validate email configuration - prefer AWS SES over SMTP
	var sesClient *SESClient
	var err error

	// Log configuration status (without exposing secrets)
	log.Debug("Email service: Checking configuration",
		"hasAWSAccessKeyID", cfg.Email.AWSAccessKeyID != "",
		"hasAWSSecretAccessKey", cfg.Email.AWSSecretAccessKey != "",
		"awsRegion", cfg.Email.AWSRegion,
		"fromEmail", cfg.Email.FromEmail,
		"hasSMTPHost", cfg.Email.SMTPHost != "",
	)

	// Try to create AWS SES client first
	if cfg.Email.AWSAccessKeyID != "" && cfg.Email.AWSSecretAccessKey != "" {
		log.Info("Email service: Attempting to initialize AWS SES client",
			"awsRegion", cfg.Email.AWSRegion,
			"fromEmail", cfg.Email.FromEmail,
		)
		sesClient, err = NewSESClient(&cfg.Email, log)
		if err != nil {
			log.Error("Email service: Failed to initialize AWS SES client",
				"error", err,
				"awsRegion", cfg.Email.AWSRegion,
				"fromEmail", cfg.Email.FromEmail,
			)
			return nil, fmt.Errorf("failed to initialize AWS SES client: %w", err)
		}
		log.Info("Email service: Using AWS SES for email delivery",
			"awsRegion", cfg.Email.AWSRegion,
			"fromEmail", cfg.Email.FromEmail,
		)
	} else {
		// Fallback to SMTP if AWS credentials are not provided
		log.Warn("Email service: AWS SES credentials not configured",
			"hasAWSAccessKeyID", cfg.Email.AWSAccessKeyID != "",
			"hasAWSSecretAccessKey", cfg.Email.AWSSecretAccessKey != "",
			"hasSMTPHost", cfg.Email.SMTPHost != "",
		)
		if cfg.Email.SMTPHost == "" {
			return nil, fmt.Errorf("neither AWS SES credentials nor SMTP_HOST is configured. Please configure AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_REGION for AWS SES")
		}
		// Note: SMTP support is deprecated but kept for backward compatibility
		return nil, fmt.Errorf("SMTP is deprecated. Please configure AWS SES credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION)")
	}

	// Load email template
	tmpl, err := loadTemplate()
	if err != nil {
		return nil, fmt.Errorf("failed to load email template: %w", err)
	}

	return &EmailService{
		config:    cfg,
		logger:    log,
		sesClient: sesClient,
		template:  tmpl,
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

	// Send email using AWS SES
	ctx := context.Background()
	subject := "Verify your email address - Equitywala"
	
	if err := s.sesClient.SendEmail(ctx, toEmail, toName, subject, htmlBody.String()); err != nil {
		s.logger.Error("Failed to send email via AWS SES",
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

