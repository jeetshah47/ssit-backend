package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func main() {
	filePath := flag.String("file", "", "Path to the file to upload (required)")
	key := flag.String("key", "", "S3 object key (optional, defaults to filename)")
	prefix := flag.String("prefix", "", "S3 key prefix/folder (optional)")
	contentType := flag.String("content-type", "", "Content-Type header (optional, auto-detected if not provided)")
	bucket := flag.String("bucket", "", "S3 bucket name (optional, uses S3_BUCKET_NAME env var if not provided)")
	region := flag.String("region", "", "AWS region (optional, uses AWS_REGION env var if not provided)")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Error: -file flag is required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run scripts/upload-to-s3.go -file <path> [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  -file        Path to the file to upload (required)")
		fmt.Println("  -key         S3 object key (optional, defaults to filename)")
		fmt.Println("  -prefix      S3 key prefix/folder (optional)")
		fmt.Println("  -content-type Content-Type header (optional, auto-detected)")
		fmt.Println("  -bucket      S3 bucket name (optional, uses S3_BUCKET_NAME env var)")
		fmt.Println("  -region      AWS region (optional, uses AWS_REGION env var)")
		fmt.Println("\nExamples:")
		fmt.Println("  go run scripts/upload-to-s3.go -file ./logo.png")
		fmt.Println("  go run scripts/upload-to-s3.go -file ./document.pdf -prefix uploads/documents")
		fmt.Println("  go run scripts/upload-to-s3.go -file ./image.jpg -key custom-name.jpg -bucket my-bucket")
		os.Exit(1)
	}

	_ = godotenv.Load()

	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := *region
	if awsRegion == "" {
		awsRegion = os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = "ap-south-1"
		}
	}
	bucketName := *bucket
	if bucketName == "" {
		bucketName = os.Getenv("S3_BUCKET_NAME")
	}

	if accessKeyID == "" || secretAccessKey == "" {
		fmt.Println("Error: AWS credentials not configured")
		fmt.Println("Please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables")
		os.Exit(1)
	}
	if bucketName == "" {
		fmt.Println("Error: S3 bucket name not configured")
		fmt.Println("Please set S3_BUCKET_NAME environment variable or use -bucket flag")
		os.Exit(1)
	}

	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		fmt.Printf("Error: File not found: %s\n", *filePath)
		os.Exit(1)
	}

	file, err := os.Open(*filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		os.Exit(1)
	}

	objectKey := *key
	if objectKey == "" {
		objectKey = filepath.Base(*filePath)
	}
	if *prefix != "" {
		objectKey = strings.TrimSuffix(*prefix, "/") + "/" + objectKey
	}

	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(awsRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		fmt.Printf("Error loading AWS config: %v\n", err)
		os.Exit(1)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	ct := *contentType
	if ct == "" {
		ct = detectContentType(*filePath)
	}

	fmt.Printf("Uploading file to S3...\n")
	fmt.Printf("  File: %s\n", *filePath)
	fmt.Printf("  Size: %d bytes\n", fileInfo.Size())
	fmt.Printf("  Bucket: %s\n", bucketName)
	fmt.Printf("  Key: %s\n", objectKey)
	fmt.Printf("  Content-Type: %s\n", ct)
	fmt.Printf("  Region: %s\n", awsRegion)
	fmt.Println()

	startTime := time.Now()

	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(ct),
	}

	result, err := s3Client.PutObject(ctx, putInput)
	if err != nil {
		fmt.Printf("Error uploading file: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(startTime)
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, awsRegion, objectKey)

	fmt.Println("Upload successful!")
	fmt.Printf("  Duration: %v\n", duration)
	if result.ETag != nil {
		fmt.Printf("  ETag: %s\n", *result.ETag)
	}
	fmt.Printf("  URL: %s\n", url)
}

func detectContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	contentTypes := map[string]string{
		".html": "text/html",
		".htm":  "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain",
		".csv":  "text/csv",

		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".bmp":  "image/bmp",

		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".ogg":  "audio/ogg",
		".wav":  "audio/wav",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",

		".zip": "application/zip",
		".tar": "application/x-tar",
		".gz":  "application/gzip",
		".rar": "application/vnd.rar",
		".7z":  "application/x-7z-compressed",

		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".otf":   "font/otf",
		".eot":   "application/vnd.ms-fontobject",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}
