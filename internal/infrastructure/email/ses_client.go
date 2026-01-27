package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/equitywala/backend/internal/core/config"
	"github.com/equitywala/backend/internal/common/logger"
)

// SESClient wraps the AWS SES API client
type SESClient struct {
	client *ses.Client
	config *config.EmailConfig
	logger logger.Logger
}

// NewSESClient creates a new AWS SES client
func NewSESClient(cfg *config.EmailConfig, log logger.Logger) (*SESClient, error) {
	// Validate AWS SES configuration
	if cfg.AWSAccessKeyID == "" || cfg.AWSSecretAccessKey == "" {
		return nil, fmt.Errorf("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be configured for AWS SES")
	}
	if cfg.AWSRegion == "" {
		return nil, fmt.Errorf("AWS_REGION must be configured for AWS SES")
	}
	if cfg.FromEmail == "" {
		return nil, fmt.Errorf("SMTP_FROM_EMAIL must be configured")
	}

	// Create AWS config with static credentials
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create SES client
	sesClient := ses.NewFromConfig(awsCfg)

	return &SESClient{
		client: sesClient,
		config: cfg,
		logger: log,
	}, nil
}

// SendEmail sends an email using AWS SES
func (c *SESClient) SendEmail(ctx context.Context, toEmail, toName, subject, htmlBody string) error {
	// Prepare the email destination
	destination := &types.Destination{
		ToAddresses: []string{toEmail},
	}

	// Prepare the message
	message := &types.Message{
		Subject: &types.Content{
			Data:    aws.String(subject),
			Charset: aws.String("UTF-8"),
		},
		Body: &types.Body{
			Html: &types.Content{
				Data:    aws.String(htmlBody),
				Charset: aws.String("UTF-8"),
			},
		},
	}

	// Prepare the email input
	input := &ses.SendEmailInput{
		Source:      aws.String(fmt.Sprintf("%s <%s>", c.config.FromName, c.config.FromEmail)),
		Destination: destination,
		Message:     message,
	}

	// Send the email
	_, err := c.client.SendEmail(ctx, input)
	if err != nil {
		// Provide helpful error messages for common AWS SES errors
		errorMsg := err.Error()
		var hint string
		
		// Check for specific AWS SES error types
		if strings.Contains(errorMsg, "Email address is not verified") || 
		   strings.Contains(errorMsg, "not verified") ||
		   strings.Contains(errorMsg, "Email address not verified") {
			hint = fmt.Sprintf("Email address '%s' is not verified in AWS SES. In sandbox mode, both sender and recipient emails must be verified. Verify the email in AWS SES Console or move out of sandbox mode.", toEmail)
		} else if strings.Contains(errorMsg, "InvalidParameterValue") || 
		          strings.Contains(errorMsg, "Invalid") {
			hint = fmt.Sprintf("Invalid email parameter. Ensure the 'From' email '%s' is verified in AWS SES.", c.config.FromEmail)
		} else if strings.Contains(errorMsg, "MessageRejected") {
			hint = "Email was rejected by AWS SES. Check if you're in sandbox mode (only verified emails allowed) or if there are sending limits."
		} else if strings.Contains(errorMsg, "InvalidClientTokenId") || 
		          strings.Contains(errorMsg, "SignatureDoesNotMatch") ||
		          strings.Contains(errorMsg, "InvalidAccessKeyId") {
			hint = "Invalid AWS credentials. Please check AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY."
		} else if strings.Contains(errorMsg, "Throttling") || 
		          strings.Contains(errorMsg, "Rate exceeded") {
			hint = "AWS SES rate limit exceeded. Please wait before sending more emails."
		} else {
			hint = fmt.Sprintf("AWS SES error. Check AWS credentials, region (%s), and ensure '%s' is verified in SES.", c.config.AWSRegion, c.config.FromEmail)
		}
		
		c.logger.Error("Failed to send email via AWS SES",
			"to", toEmail,
			"from", c.config.FromEmail,
			"error", err,
			"region", c.config.AWSRegion,
			"hint", hint,
		)
		
		// Return error with hint for better debugging
		return fmt.Errorf("failed to send email via AWS SES: %w. %s", err, hint)
	}

	c.logger.Info("Email sent successfully via AWS SES", "to", toEmail)
	return nil
}

