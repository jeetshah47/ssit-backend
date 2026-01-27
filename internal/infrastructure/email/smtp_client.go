package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/equitywala/backend/internal/common/logger"
	"github.com/equitywala/backend/internal/core/config"
)

// SMTPClient wraps the SMTP email client
type SMTPClient struct {
	config *config.EmailConfig
	logger logger.Logger
}

// NewSMTPClient creates a new SMTP client
func NewSMTPClient(cfg *config.EmailConfig, log logger.Logger) (*SMTPClient, error) {
	// Validate SMTP configuration
	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP_HOST must be configured")
	}
	if cfg.SMTPPort == 0 {
		return nil, fmt.Errorf("SMTP_PORT must be configured")
	}
	if cfg.SMTPUsername == "" {
		return nil, fmt.Errorf("SMTP_USERNAME must be configured")
	}
	if cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD must be configured")
	}
	if cfg.FromEmail == "" {
		return nil, fmt.Errorf("SMTP_FROM_EMAIL must be configured")
	}

	return &SMTPClient{
		config: cfg,
		logger: log,
	}, nil
}

// SendEmail sends an email using SMTP
func (c *SMTPClient) SendEmail(ctx context.Context, toEmail, toName, subject, htmlBody string) error {
	// Validate email addresses
	fromAddr, err := mail.ParseAddress(fmt.Sprintf("%s <%s>", c.config.FromName, c.config.FromEmail))
	if err != nil {
		return fmt.Errorf("invalid from email address: %w", err)
	}

	toAddr, err := mail.ParseAddress(toEmail)
	if err != nil {
		return fmt.Errorf("invalid to email address: %w", err)
	}

	// Create the email message
	message, err := c.buildMessage(fromAddr, toAddr, subject, htmlBody)
	if err != nil {
		return fmt.Errorf("failed to build email message: %w", err)
	}

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", c.config.SMTPHost, c.config.SMTPPort)

	// Create authentication
	auth := smtp.PlainAuth("", c.config.SMTPUsername, c.config.SMTPPassword, c.config.SMTPHost)

	// Determine if we need TLS
	// Port 587 uses STARTTLS, port 465 uses TLS directly
	useTLS := c.config.SMTPPort == 465

	var client *smtp.Client
	if useTLS {
		// TLS connection for port 465
		tlsConfig := &tls.Config{
			ServerName: c.config.SMTPHost,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, c.config.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// Regular connection for port 587 (STARTTLS)
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}

		// Check if STARTTLS is supported
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: c.config.SMTPHost,
			}
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %w", err)
			}
		}
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipient
	if err = client.Mail(fromAddr.Address); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = client.Rcpt(toAddr.Address); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send the email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}
	_, err = writer.Write(message)
	if err != nil {
		writer.Close()
		return fmt.Errorf("failed to write email body: %w", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	// Quit the connection
	if err = client.Quit(); err != nil {
		// Log but don't fail - email might have been sent
		c.logger.Warn("Failed to quit SMTP connection gracefully", "error", err)
	}

	c.logger.Info("Email sent successfully via SMTP", "to", toEmail)
	return nil
}

// buildMessage creates a properly formatted MIME email message
func (c *SMTPClient) buildMessage(from, to *mail.Address, subject, htmlBody string) ([]byte, error) {
	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from.String()))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to.String()))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", encodeSubject(subject)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")

	// Body (HTML content)
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n")

	return buf.Bytes(), nil
}

// encodeSubject encodes the subject line to handle special characters
func encodeSubject(subject string) string {
	// Simple encoding - if subject contains non-ASCII, use MIME encoding
	if strings.ContainsAny(subject, "=?") {
		// Already encoded or contains special chars
		return subject
	}
	// For now, return as-is. For full MIME encoding, use mime.QEncoding
	return subject
}
