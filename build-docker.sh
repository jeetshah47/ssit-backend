#!/bin/bash

# Docker build script for ECS deployment
# This script builds a Docker image for ARM64 architecture (required for ECS)
# Usage: ./build-docker.sh [image-name] [tag]
#   Example: ./build-docker.sh ssit-backend latest

set -e  # Exit on error

IMAGE_NAME=${1:-ssit-backend}
TAG=${2:-latest}
PLATFORM=${3:-linux/arm64}

echo "Building Docker image for ECS..."
echo "  Image: ${IMAGE_NAME}:${TAG}"
echo "  Platform: ${PLATFORM}"
echo ""

# Check if docker buildx is available
if ! docker buildx version > /dev/null 2>&1; then
    echo "Warning: docker buildx not available, using standard docker build"
    echo "For ECS ARM64, it's recommended to use docker buildx"
    echo ""
    
    # Fallback to standard docker build with platform flag
    echo "Building with standard docker build..."
    docker build --platform "$PLATFORM" -t "${IMAGE_NAME}:${TAG}" .
else
    # Use buildx for better multi-platform support
    echo "Using docker buildx..."
    
    # Create and use a builder instance if it doesn't exist
    if ! docker buildx ls | grep -q "multiarch"; then
        echo "Creating buildx builder instance..."
        docker buildx create --name multiarch --use > /dev/null 2>&1
    else
        docker buildx use multiarch > /dev/null 2>&1
    fi
    
    # Build for ARM64 platform
    echo "Building Docker image..."
    docker buildx build \
        --platform "$PLATFORM" \
        --tag "${IMAGE_NAME}:${TAG}" \
        --load \
        .
fi

if [ $? -eq 0 ]; then
    echo ""
    echo "[OK] Docker image built successfully!"
    echo "  Image: ${IMAGE_NAME}:${TAG}"
    echo ""
    echo "To push to ECR:"
    echo "  docker tag ${IMAGE_NAME}:${TAG} <ecr-repo-url>:${TAG}"
    echo "  docker push <ecr-repo-url>:${TAG}"
else
    echo ""
    echo "[ERROR] Docker build failed!"
    exit 1
fi
