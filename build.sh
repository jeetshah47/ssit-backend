#!/bin/bash

# Build script for Go application
# This script builds the Go binary and places it in the build folder
# Usage: ./build.sh [target]
#   target: linux (default), windows, or current (builds for current OS)

set -e  # Exit on error

TARGET=${1:-linux}

echo "Building Go application for: $TARGET"

# Create build directory if it doesn't exist
mkdir -p build

# Build the application based on target
case $TARGET in
    linux)
        echo "Building for Linux..."
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o build/server ./cmd/server
        OUTPUT="build/server"
        ;;
    windows)
        echo "Building for Windows..."
        CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o build/server.exe ./cmd/server
        OUTPUT="build/server.exe"
        ;;
    current)
        echo "Building for current OS..."
        CGO_ENABLED=1 go build -o build/server ./cmd/server
        OUTPUT="build/server"
        ;;
    *)
        echo "Unknown target: $TARGET"
        echo "Usage: ./build.sh [linux|windows|current]"
        exit 1
        ;;
esac

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "[OK] Build successful! Binary created at: $OUTPUT"
    ls -lh "$OUTPUT"
else
    echo "[ERROR] Build failed!"
    exit 1
fi
