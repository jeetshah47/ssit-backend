package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/equitywala/backend/internal/services"
)

// Client implements local file storage operations (replaces S3 for now)
type Client struct {
	baseDir string // Base directory for storing files (e.g., "tmp/advisory-reports")
	baseURL string // Base URL for serving files (e.g., "/api/v1/files")
}

// Ensure Client implements services.S3Client interface
var _ services.S3Client = (*Client)(nil)

// NewClient creates a new local file storage client
func NewClient(baseDir, baseURL string) (*Client, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Client{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

// UploadFile saves a file to local storage
func (c *Client) UploadFile(ctx context.Context, bucket, key string, file io.Reader, contentType string) (string, error) {
	// Construct full file path
	fullPath := filepath.Join(c.baseDir, key)
	
	// Create directory structure if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Copy file content
	_, err = io.Copy(outFile, file)
	if err != nil {
		os.Remove(fullPath) // Clean up on error
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Generate URL
	url := c.GetFileURL(bucket, key)
	return url, nil
}

// DeleteFile deletes a file from local storage
func (c *Client) DeleteFile(ctx context.Context, bucket, key string) error {
	fullPath := filepath.Join(c.baseDir, key)
	
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, consider it deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFileURL generates a URL for the file
// Note: bucket parameter is ignored for local storage
func (c *Client) GetFileURL(bucket, key string) string {
	// Generate local file URL
	// Format: {baseURL}/{key}
	// Example: /api/v1/files/advisory-reports/stock-baskets/{id}/report.pdf
	return fmt.Sprintf("%s/%s", c.baseURL, strings.ReplaceAll(key, "\\", "/"))
}

// ExtractKeyFromURL extracts the file key from a URL
func (c *Client) ExtractKeyFromURL(url string) string {
	// Extract key from URL format: /api/v1/files/advisory-reports/...
	if strings.HasPrefix(url, c.baseURL) {
		key := strings.TrimPrefix(url, c.baseURL)
		return strings.TrimPrefix(key, "/")
	}
	
	// Fallback: try to extract from any URL format
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "advisory-reports" && i+1 < len(parts) {
			return strings.Join(parts[i:], "/")
		}
	}
	
	return ""
}
