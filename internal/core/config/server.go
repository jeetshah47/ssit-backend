package config

import "time"

type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func newServerConfig() ServerConfig {
	return ServerConfig{
		Port:        getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("SERVER_ENV", "development"),
	}
}

