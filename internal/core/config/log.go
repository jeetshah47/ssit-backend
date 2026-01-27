package config

type LogConfig struct {
	Level  string
	Format string
}

func newLogConfig() LogConfig {
	return LogConfig{
		Level:  getEnv("LOG_LEVEL", "info"),
		Format: getEnv("LOG_FORMAT", "json"),
	}
}

