# Runtime stage - optimized for pre-built binaries
# Build the binary locally first using: .\build.ps1 linux
# Use build argument for platform to avoid warning
ARG TARGETPLATFORM=linux/arm64
FROM --platform=${TARGETPLATFORM} alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates libc6-compat curl

WORKDIR /app

# Copy pre-built binary from build folder
# This is much faster than building inside Docker (avoids emulation)
COPY build/server .

# Verify the binary is ARM64 (if file command is available)
# Note: This is a safety check - the binary should already be ARM64 from local build
RUN if command -v file >/dev/null 2>&1; then \
        file /app/server | grep -q "ARM aarch64" || (echo "WARNING: Binary architecture check failed" && exit 1); \
    fi

# Expose port
EXPOSE 8080

# Run the server
CMD ["./server"]

