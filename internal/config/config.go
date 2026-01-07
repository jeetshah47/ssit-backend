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
}

type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
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

type AuthConfig struct {
	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration
}

// GetJWTSecret returns the JWT secret
func (a AuthConfig) GetJWTSecret() string {
	return a.JWTSecret
}

// GetJWTExpiry returns the JWT expiry duration
func (a AuthConfig) GetJWTExpiry() time.Duration {
	return a.JWTExpiry
}

// GetJWTRefreshExpiry returns the JWT refresh expiry duration
func (a AuthConfig) GetJWTRefreshExpiry() time.Duration {
	return a.JWTRefreshExpiry
}

type LogConfig struct {
	Level  string
	Format string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	FrontendURL  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Environment: getEnv("SERVER_ENV", "development"),
		},
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
		Auth: AuthConfig{
			JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
			JWTExpiry:        parseDuration(getEnv("JWT_EXPIRY", "24h")),
			JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseStringSlice(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     parseInt(getEnv("SMTP_PORT", "587")),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("SMTP_FROM_EMAIL", "no-reply@equitywala.com"),
			FromName:     getEnv("SMTP_FROM_NAME", "Team Equitywala"),
			FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
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
