#!/bin/bash

# S3 Upload Script for ssit-backend
# Usage: ./scripts/upload-to-s3.sh <file-path> [options]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load .env file if it exists
if [ -f "$PROJECT_ROOT/.env" ]; then
    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        [[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue
        # Remove leading/trailing whitespace from key
        key=$(echo "$key" | xargs)
        # Remove surrounding quotes from value if present
        value=$(echo "$value" | sed -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'$//")
        export "$key=$value"
    done < "$PROJECT_ROOT/.env"
fi

if [ -z "$1" ]; then
    echo "S3 Upload Script"
    echo ""
    echo "Usage:"
    echo "  ./scripts/upload-to-s3.sh <file-path> [options]"
    echo ""
    echo "Options:"
    echo "  -key         S3 object key (optional, defaults to filename)"
    echo "  -prefix      S3 key prefix/folder (optional)"
    echo "  -content-type Content-Type header (optional, auto-detected)"
    echo "  -bucket      S3 bucket name (optional, uses S3_BUCKET_NAME env var)"
    echo "  -region      AWS region (optional, uses AWS_REGION env var)"
    echo ""
    echo "Environment Variables (can be set in .env file):"
    echo "  AWS_ACCESS_KEY_ID      AWS Access Key ID (required)"
    echo "  AWS_SECRET_ACCESS_KEY  AWS Secret Access Key (required)"
    echo "  AWS_REGION             AWS Region (default: ap-south-1)"
    echo "  S3_BUCKET_NAME         S3 Bucket Name (required)"
    echo ""
    echo "Examples:"
    echo "  ./scripts/upload-to-s3.sh ./logo.png"
    echo "  ./scripts/upload-to-s3.sh ./document.pdf -prefix uploads/documents"
    echo "  ./scripts/upload-to-s3.sh ./image.jpg -key custom-name.jpg"
    exit 1
fi

FILE_PATH="$1"
shift

cd "$PROJECT_ROOT"
go run scripts/upload-to-s3.go -file "$FILE_PATH" "$@"

