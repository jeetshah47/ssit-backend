package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/google/uuid"
)

// S3Client defines the interface for S3 operations
type S3Client interface {
	UploadFile(ctx context.Context, bucket, key string, file io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, bucket, key string) error
	GetFileURL(bucket, key string) string
}

// ReportUploadResult represents the result of a report upload
type ReportUploadResult struct {
	ReportURL     string
	ReportFileName string
	UploadedAt    time.Time
}

// UploadAdvisoryReportCmd represents the command for uploading an advisory report
type UploadAdvisoryReportCmd struct {
	AdvisoryID   uuid.UUID
	AdvisoryType string // stock_basket, etf_basket, ipo, mutual_fund_basket
	File         multipart.File
	FileName     string
	FileSize     int64
	UploadedBy   uuid.UUID
}

// UploadAdvisoryReportCmdOutputData represents the output
type UploadAdvisoryReportCmdOutputData struct {
	Result *ReportUploadResult
}

// UploadAdvisoryReportService handles advisory report uploads
type UploadAdvisoryReportService struct {
	s3Client S3Client
	baseDir  string // Base directory for file storage
}

// NewUploadAdvisoryReportService creates a new upload advisory report service
func NewUploadAdvisoryReportService(s3Client S3Client, baseDir string) *UploadAdvisoryReportService {
	return &UploadAdvisoryReportService{
		s3Client: s3Client,
		baseDir:  baseDir,
	}
}

// Execute executes the UploadAdvisoryReport command
func (s *UploadAdvisoryReportService) Execute(ctx context.Context, cmd UploadAdvisoryReportCmd) (*UploadAdvisoryReportCmdOutputData, error) {
	// Validate file
	if err := s.validateFile(cmd); err != nil {
		return nil, err
	}

	// Check if file storage client is configured
	if s.s3Client == nil {
		return nil, errors.NewDomainError("FILE_STORAGE_NOT_CONFIGURED", "file storage client is not configured")
	}

	// Generate file path key
	key := s.generateS3Key(cmd.AdvisoryType, cmd.AdvisoryID, cmd.FileName)

	// Upload to local storage
	reportURL, err := s.s3Client.UploadFile(ctx, s.baseDir, key, cmd.File, "application/pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	result := &ReportUploadResult{
		ReportURL:     reportURL,
		ReportFileName: cmd.FileName,
		UploadedAt:    time.Now(),
	}

	return &UploadAdvisoryReportCmdOutputData{Result: result}, nil
}

// validateFile validates the uploaded file
func (s *UploadAdvisoryReportService) validateFile(cmd UploadAdvisoryReportCmd) error {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(cmd.FileName))
	if ext != ".pdf" {
		return errors.NewDomainError("INVALID_FILE_TYPE", "only PDF files are allowed")
	}

	// Check file size (10MB max)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if cmd.FileSize > maxFileSize {
		return errors.NewDomainError("FILE_TOO_LARGE", "file size exceeds 10MB limit")
	}

	// Validate advisory type
	validTypes := map[string]bool{
		"stock_basket":      true,
		"etf_basket":        true,
		"ipo":               true,
		"mutual_fund_basket": true,
	}
	if !validTypes[cmd.AdvisoryType] {
		return errors.NewDomainError("INVALID_ADVISORY_TYPE", "invalid advisory type")
	}

	return nil
}

// generateS3Key generates a file path key for the file
func (s *UploadAdvisoryReportService) generateS3Key(advisoryType string, advisoryID uuid.UUID, fileName string) string {
	// Sanitize filename
	sanitizedFileName := s.sanitizeFileName(fileName)
	timestamp := time.Now().Format("20060102_150405")
	
	// Format: advisory-reports/{advisory_type}/{advisory_id}/{timestamp}_{filename}
	return fmt.Sprintf("advisory-reports/%s/%s/%s_%s", advisoryType, advisoryID.String(), timestamp, sanitizedFileName)
}

// sanitizeFileName sanitizes the filename for safe storage
func (s *UploadAdvisoryReportService) sanitizeFileName(fileName string) string {
	// Remove path components
	fileName = filepath.Base(fileName)
	
	// Replace spaces and special characters
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "(", "")
	fileName = strings.ReplaceAll(fileName, ")", "")
	fileName = strings.ReplaceAll(fileName, "[", "")
	fileName = strings.ReplaceAll(fileName, "]", "")
	
	// Keep only alphanumeric, dots, underscores, and hyphens
	var sanitized strings.Builder
	for _, r := range fileName {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			sanitized.WriteRune(r)
		}
	}
	
	result := sanitized.String()
	if result == "" {
		result = "report.pdf"
	}
	
	return result
}

// DeleteAdvisoryReportCmd represents the command for deleting an advisory report
type DeleteAdvisoryReportCmd struct {
	AdvisoryID   uuid.UUID
	AdvisoryType string
	ReportURL    string
}

// DeleteAdvisoryReportService handles advisory report deletion
type DeleteAdvisoryReportService struct {
	s3Client S3Client
	baseDir  string // Base directory for file storage
}

// NewDeleteAdvisoryReportService creates a new delete advisory report service
func NewDeleteAdvisoryReportService(s3Client S3Client, baseDir string) *DeleteAdvisoryReportService {
	return &DeleteAdvisoryReportService{
		s3Client: s3Client,
		baseDir:  baseDir,
	}
}

// Execute executes the DeleteAdvisoryReport command
func (s *DeleteAdvisoryReportService) Execute(ctx context.Context, cmd DeleteAdvisoryReportCmd) error {
	// Check if file storage client is configured
	if s.s3Client == nil {
		return errors.NewDomainError("FILE_STORAGE_NOT_CONFIGURED", "file storage client is not configured")
	}

	// Extract key from URL
	key := s.extractKeyFromURL(cmd.ReportURL)
	if key == "" {
		return errors.NewDomainError("INVALID_URL", "invalid report URL")
	}

	// Delete from local storage
	if err := s.s3Client.DeleteFile(ctx, s.baseDir, key); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// extractKeyFromURL extracts the file key from a URL
func (s *DeleteAdvisoryReportService) extractKeyFromURL(url string) string {
	// Extract key from URL format: /api/v1/files/advisory-reports/...
	// Or any URL containing advisory-reports
	parts := strings.Split(url, "/")
	
	// Find the advisory-reports part
	for i, part := range parts {
		if part == "advisory-reports" && i+1 < len(parts) {
			// Reconstruct the key
			return strings.Join(parts[i:], "/")
		}
	}
	
	return ""
}
