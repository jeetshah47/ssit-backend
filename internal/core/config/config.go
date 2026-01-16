package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	MongoDB  MongoDBConfig
	Auth     AuthConfig
	Log      LogConfig
	CORS     CORSConfig
	Email    EmailConfig
	Paytm    PaytmConfig
	S3       S3Config
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type EmailConfig struct {
	// SMTP Configuration (legacy, kept for backward compatibility)
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	
	// AWS SES Configuration
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	
	// Common email configuration
	FromEmail    string
	FromName     string
	FrontendURL  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{
		Server: newServerConfig(),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "equitywala"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "equitywala_audit"),
		},
		Auth: newAuthConfig(),
		Log:  newLogConfig(),
		CORS: CORSConfig{
			AllowedOrigins: parseStringSlice(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		},
		Email: EmailConfig{
			// SMTP Configuration (legacy)
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     parseInt(getEnv("SMTP_PORT", "587")),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			// AWS SES Configuration
			AWSAccessKeyID:     strings.TrimSpace(getEnv("AWS_ACCESS_KEY_ID", "")),
			AWSSecretAccessKey: strings.TrimSpace(getEnv("AWS_SECRET_ACCESS_KEY", "")),
			AWSRegion:          strings.TrimSpace(getEnv("AWS_REGION", "ap-south-1")),
			// Common email configuration
			FromEmail:    getEnv("SMTP_FROM_EMAIL", "no-reply@equitywala.com"),
			FromName:     getEnv("SMTP_FROM_NAME", "Team Equitywala"),
			FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
		Paytm: newPaytmConfig(),
		S3:    newS3Config(),
	}

	// Parse timeouts
	if readTimeout := getEnv("SERVER_READ_TIMEOUT", "30s"); readTimeout != "" {
		if d, err := time.ParseDuration(readTimeout); err == nil {
			cfg.Server.ReadTimeout = d
		}
	}
	if writeTimeout := getEnv("SERVER_WRITE_TIMEOUT", "30s"); writeTimeout != "" {
		if d, err := time.ParseDuration(writeTimeout); err == nil {
			cfg.Server.WriteTimeout = d
		}
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	// Only validate JWT_SECRET in production
	if c.Server.Environment == "production" {
		if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "change-me-in-production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
		if c.Paytm.MerchantID == "" || c.Paytm.MerchantKey == "" {
			return fmt.Errorf("PAYTM_MERCHANT_ID and PAYTM_MERCHANT_KEY must be set in production")
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour // Default to 24 hours
	}
	return d
}

func parseStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	// Split by comma and trim spaces
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

