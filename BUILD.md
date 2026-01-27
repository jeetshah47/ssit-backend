# Build Instructions

This document explains how to build the application for Docker deployment.

## Overview

The build process has been separated from Docker. You need to build the Go binary locally first, then Docker will copy the pre-built binary.

## Build Scripts

### For Linux/Mac/WSL

Use the bash script:

```bash
# Build for Linux (default, for Docker)
./build.sh

# Or explicitly specify target
./build.sh linux    # For Docker (Linux)
./build.sh windows  # For Windows
./build.sh current  # For current OS
```

### For Windows (PowerShell)

Use the PowerShell script:

```powershell
# Build for Linux (default, for Docker)
.\build.ps1

# Or explicitly specify target
.\build.ps1 linux    # For Docker (Linux)
.\build.ps1 windows  # For Windows
.\build.ps1 current  # For current OS
```

## Build Process

### Option 1: Optimized Build (Recommended for ECS) ⚡

**This is the fastest method** - builds the Go binary locally using native cross-compilation, then copies it into Docker. This avoids slow emulation.

**For ECS ARM64 (Recommended):**

```powershell
# Windows PowerShell - Builds binary locally, then creates Docker image
.\build-docker.ps1 ssit-backend latest
```

```bash
# Linux/Mac/WSL
./build-docker.sh ssit-backend latest
```

**Why this is faster:**
- Go cross-compilation is native and fast (no emulation)
- Docker just copies the binary (no build step inside container)
- Avoids downloading dependencies and compiling inside emulated ARM64 environment

**Manual steps:**
```powershell
# 1. Build binary locally (fast)
$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="arm64"
go build -o build/server ./cmd/server

# 2. Build Docker image (just copies binary)
docker build --platform linux/arm64 -t ssit-backend:latest .
```

**For local development (AMD64):**

```bash
docker build --platform linux/amd64 -t ssit-backend:latest .
```

### Option 2: Build Inside Docker (Slower - uses emulation)

**Note:** This method is slower on Windows because it requires emulation to build ARM64 inside Docker.

If you prefer to build everything inside Docker:

1. **Build the binary**: Run the appropriate build script for your platform
   - The script will create a `build/` directory (if it doesn't exist)
   - The binary will be placed in `build/server` (Linux) or `build/server.exe` (Windows)

2. **Build Docker image**: After building the binary, build the Docker image:
   ```bash
   docker build -t ssit-backend .
   ```

3. **Run Docker container**:
   ```bash
   docker run -p 8080:8080 ssit-backend
   ```

## ECS Deployment

For AWS ECS deployment, you **must** build the image for the correct platform:

- **ARM64 (Graviton)**: Use `linux/arm64`
- **x86_64**: Use `linux/amd64`

The `build-docker.ps1` and `build-docker.sh` scripts automatically build for `linux/arm64` which is required for ECS ARM64 tasks.

**To push to ECR:**

```bash
# Tag the image
docker tag ssit-backend:latest <your-ecr-repo-url>:latest

# Push to ECR
docker push <your-ecr-repo-url>:latest
```

## Notes

- The `build/` folder is gitignored and should not be committed
- For Docker builds, always use the `linux` target (default)
- The build scripts use `CGO_ENABLED=1` which is required for PostgreSQL/MongoDB drivers
- Cross-compilation is handled automatically when building for Linux on Windows

## Troubleshooting

### ECS Error: "image Manifest does not contain descriptor matching platform 'linux/arm64'"

This error occurs when the Docker image was built for a different architecture than your ECS task requires.

**Solution:**
1. Use the `build-docker.ps1` or `build-docker.sh` script which builds for `linux/arm64`
2. Or manually build with: `docker buildx build --platform linux/arm64 -t ssit-backend:latest --load .`
3. Make sure your ECS task definition matches the image architecture

### Build fails with CGO errors
- The Dockerfile now uses `CGO_ENABLED=0` for static builds, which should work without C compiler
- If building locally, ensure you have a C compiler installed (gcc on Linux/Mac, MinGW on Windows)
- For Windows, you may need to install TDM-GCC or MinGW-w64

### Docker build fails with "file not found"
- If using Option 1 (build inside Docker), this shouldn't occur as the binary is built inside the container
- If using Option 2 (pre-built binary), make sure you've run the build script first
- Verify that `build/server` exists before building the Docker image

### Binary is too large
- This is normal for Go binaries
- The Dockerfile uses `CGO_ENABLED=0` for smaller static binaries
- Consider using `-ldflags="-s -w"` to reduce size (add to build command in Dockerfile)
