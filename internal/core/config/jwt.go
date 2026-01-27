package config

import "time"

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

func newAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:        parseDuration(getEnv("JWT_EXPIRY", "24h")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
	}
}

