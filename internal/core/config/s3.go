package config

// S3Config holds file storage configuration (local for now, S3 later)
type S3Config struct {
	BaseDir string // Base directory for storing files locally
	BaseURL string // Base URL for serving files
}

// newS3Config creates a new file storage configuration
func newS3Config() S3Config {
	return S3Config{
		BaseDir: getEnv("ADVISORY_REPORTS_DIR", "tmp/advisory-reports"),
		BaseURL: getEnv("ADVISORY_REPORTS_BASE_URL", "/api/v1/files"),
	}
}
