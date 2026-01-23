package s3

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/equitywala/backend/internal/common/logger"
	"github.com/equitywala/backend/internal/core/config"
)

type Client struct {
	s3Client *s3.Client
	config   *config.S3Config
	logger   logger.Logger
}

type UploadInput struct {
	Key         string
	Body        io.Reader
	ContentType string
	ACL         string
}

type UploadOutput struct {
	Key      string
	URL      string
	ETag     string
	UploadAt time.Time
}

func NewClient(cfg *config.S3Config, log logger.Logger) (*Client, error) {
	if cfg.AWSAccessKeyID == "" || cfg.AWSSecretAccessKey == "" {
		return nil, fmt.Errorf("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be configured for S3")
	}
	if cfg.AWSRegion == "" {
		return nil, fmt.Errorf("AWS_REGION must be configured for S3")
	}
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME must be configured")
	}

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

	s3Client := s3.NewFromConfig(awsCfg)

	return &Client{
		s3Client: s3Client,
		config:   cfg,
		logger:   log,
	}, nil
}

func (c *Client) Upload(ctx context.Context, input UploadInput) (*UploadOutput, error) {
	contentType := input.ContentType
	if contentType == "" {
		ext := filepath.Ext(input.Key)
		contentType = mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(c.config.BucketName),
		Key:         aws.String(input.Key),
		Body:        input.Body,
		ContentType: aws.String(contentType),
	}

	result, err := c.s3Client.PutObject(ctx, putInput)
	if err != nil {
		c.logger.Error("Failed to upload file to S3",
			"bucket", c.config.BucketName,
			"key", input.Key,
			"error", err,
		)
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	url := c.GetObjectURL(input.Key)

	c.logger.Info("File uploaded successfully to S3",
		"bucket", c.config.BucketName,
		"key", input.Key,
		"url", url,
	)

	etag := ""
	if result.ETag != nil {
		etag = *result.ETag
	}

	return &UploadOutput{
		Key:      input.Key,
		URL:      url,
		ETag:     etag,
		UploadAt: time.Now(),
	}, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(key),
	}

	_, err := c.s3Client.DeleteObject(ctx, deleteInput)
	if err != nil {
		c.logger.Error("Failed to delete file from S3",
			"bucket", c.config.BucketName,
			"key", key,
			"error", err,
		)
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	c.logger.Info("File deleted successfully from S3",
		"bucket", c.config.BucketName,
		"key", key,
	)

	return nil
}

func (c *Client) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3Client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

func (c *Client) GetUploadPresignedURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3Client)

	input := &s3.PutObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	request, err := presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate upload presigned URL: %w", err)
	}

	return request.URL, nil
}

func (c *Client) GetObjectURL(key string) string {
	if c.config.CDNEndpoint != "" {
		return fmt.Sprintf("%s/%s", c.config.CDNEndpoint, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.config.BucketName, c.config.AWSRegion, key)
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (c *Client) GetBucketName() string {
	return c.config.BucketName
}
