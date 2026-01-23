# S3 Infrastructure

This package provides S3 file upload functionality for the backend.

## Configuration

Add these environment variables to your `.env` file:

```bash
# AWS S3 Configuration
AWS_ACCESS_KEY_ID=your-access-key-id
AWS_SECRET_ACCESS_KEY=your-secret-access-key
AWS_REGION=ap-south-1
S3_BUCKET_NAME=your-bucket-name
S3_CDN_ENDPOINT=https://cdn.yourdomain.com  # Optional: CloudFront or CDN URL
```

## Usage

### Using the Script (CLI)

Upload a file to S3:

```bash
# Basic upload
./scripts/upload-to-s3.sh ./path/to/file.pdf

# Upload with custom key
./scripts/upload-to-s3.sh ./path/to/file.pdf -key custom-name.pdf

# Upload to a folder/prefix
./scripts/upload-to-s3.sh ./path/to/file.pdf -prefix uploads/documents

# Specify bucket and region
./scripts/upload-to-s3.sh ./path/to/file.pdf -bucket my-bucket -region us-east-1
```

For Windows (PowerShell):

```powershell
.\scripts\upload-to-s3.ps1 .\path\to\file.pdf
.\scripts\upload-to-s3.ps1 .\path\to\file.pdf -Prefix uploads/documents
```

### Using Go Directly

```bash
go run scripts/upload-to-s3.go -file ./path/to/file.pdf -prefix uploads
```

### Programmatic Usage

```go
package main

import (
    "context"
    "os"
    
    "github.com/equitywala/backend/internal/common/logger"
    "github.com/equitywala/backend/internal/core/config"
    "github.com/equitywala/backend/internal/infrastructure/s3"
)

func main() {
    cfg, _ := config.Load()
    log := logger.New(...)
    
    s3Client, err := s3.NewClient(&cfg.S3, log)
    if err != nil {
        panic(err)
    }
    
    file, _ := os.Open("./file.pdf")
    defer file.Close()
    
    result, err := s3Client.Upload(context.Background(), s3.UploadInput{
        Key:  "documents/file.pdf",
        Body: file,
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Uploaded to:", result.URL)
}
```

## Features

- **Upload**: Upload files to S3 with auto-detected content types
- **Delete**: Delete files from S3
- **Presigned URLs**: Generate presigned URLs for secure downloads
- **Upload Presigned URLs**: Generate presigned URLs for direct browser uploads
- **CDN Support**: Optional CDN endpoint configuration for public URLs

## API

### Client Methods

| Method | Description |
|--------|-------------|
| `Upload(ctx, input)` | Upload a file to S3 |
| `Delete(ctx, key)` | Delete a file from S3 |
| `GetPresignedURL(ctx, key, expiry)` | Get a presigned download URL |
| `GetUploadPresignedURL(ctx, key, contentType, expiry)` | Get a presigned upload URL |
| `GetObjectURL(key)` | Get the public URL for an object |
| `Exists(ctx, key)` | Check if an object exists |
| `GetBucketName()` | Get the configured bucket name |

