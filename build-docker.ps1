# Docker build script for ECS deployment (PowerShell)
# This script builds a Docker image for ARM64 architecture (required for ECS)
# and optionally tags and pushes to AWS ECR with automatic authentication.
#
# Usage: .\build-docker.ps1 [image-name] [tag] [ecr-repo-url] [-Push] [-NoPush]
#
# Examples:
#   .\build-docker.ps1 ssit-backend latest
#     - Just builds the image locally
#
#   .\build-docker.ps1 ssit-backend latest 128977214845.dkr.ecr.ap-south-1.amazonaws.com/equitywala-server
#     - Builds, tags, authenticates with ECR, and pushes (auto-push enabled)
#
#   .\build-docker.ps1 ssit-backend latest 128977214845.dkr.ecr.ap-south-1.amazonaws.com/equitywala-server -NoPush
#     - Builds and tags, but doesn't push
#
# Requirements:
#   - AWS Tools for PowerShell (preferred): Install-Module -Name AWS.Tools.ECR
#   - OR AWS CLI: https://aws.amazon.com/cli/
#   - Docker Desktop with ARM64 support

param(
    [string]$ImageName = "ssit-backend",
    [string]$Tag = "latest",
    [string]$Platform = "linux/arm64",
    [string]$EcrRepoUrl = "",
    [switch]$Push = $false,
    [switch]$NoPush = $false
)

$ErrorActionPreference = "Stop"

Write-Host "Building Docker image for ECS..." -ForegroundColor Cyan
Write-Host "  Image: ${ImageName}:${Tag}" -ForegroundColor Gray
Write-Host "  Platform: ${Platform}" -ForegroundColor Gray
if ($EcrRepoUrl) {
    Write-Host "  ECR Repo: ${EcrRepoUrl}" -ForegroundColor Gray
    if ($NoPush) {
        Write-Host "  Auto-push: Disabled (tag only)" -ForegroundColor Gray
    } else {
        Write-Host "  Auto-push: Enabled" -ForegroundColor Gray
    }
}
Write-Host ""

# Step 1: Build the Go binary locally first (fast native cross-compilation)
Write-Host "Step 1: Building Go binary for ARM64 Linux..." -ForegroundColor Yellow
Write-Host "  This uses native Go cross-compilation (fast, no emulation)" -ForegroundColor Gray

$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "arm64"

# Create build directory if it doesn't exist
if (-not (Test-Path -Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
}

# Build the binary
go build -o build/server ./cmd/server

if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Failed to build Go binary!" -ForegroundColor Red
    exit 1
}

Write-Host "[OK] Go binary built successfully" -ForegroundColor Green

# Verify binary exists
if (-not (Test-Path -Path "build/server")) {
    Write-Host "[ERROR] Binary not found at build/server!" -ForegroundColor Red
    exit 1
}

# Verify binary architecture (if file command is available)
Write-Host "  Verifying binary architecture..." -ForegroundColor Gray
$fileCheck = Get-Command file -ErrorAction SilentlyContinue
if ($fileCheck) {
    $fileOutput = file build/server 2>&1
    if ($fileOutput -match "ARM|aarch64|ARM64") {
        Write-Host "  [OK] Binary is ARM64" -ForegroundColor Green
    } else {
        Write-Host "  [WARNING] Binary architecture check inconclusive" -ForegroundColor Yellow
        Write-Host "    Output: $fileOutput" -ForegroundColor Gray
    }
} else {
    Write-Host "  [INFO] 'file' command not available, skipping architecture check" -ForegroundColor Gray
}

Write-Host ""

# Step 2: Build Docker image (just copies the binary, very fast)
Write-Host "Step 2: Building Docker image..." -ForegroundColor Yellow
Write-Host "  This just copies the pre-built binary (no compilation needed)" -ForegroundColor Gray
Write-Host "  Platform: $Platform" -ForegroundColor Gray
Write-Host ""

# Always use buildx for proper ARM64 support and manifest creation
# Check if buildx is available
docker buildx version | Out-Null
$buildxAvailable = $LASTEXITCODE -eq 0

if ($buildxAvailable) {
    Write-Host "Using docker buildx for proper ARM64 support..." -ForegroundColor Green
    
    # Initialize buildx if needed
    docker buildx inspect | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Initializing buildx builder..." -ForegroundColor Yellow
        docker buildx create --use --name arm64-builder --driver docker-container --platform linux/arm64 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            # Try simpler builder creation
            docker buildx create --use --name arm64-builder 2>&1 | Out-Null
        }
    }
    
    # Build with buildx - this ensures proper ARM64 manifest
    # Pass TARGETPLATFORM build arg to Dockerfile
    docker buildx build `
        --platform $Platform `
        --build-arg TARGETPLATFORM=$Platform `
        --tag "${ImageName}:${Tag}" `
        --load `
        --progress plain `
        .
    $buildExitCode = $LASTEXITCODE
} else {
    Write-Host "Buildx not available, using standard docker build..." -ForegroundColor Yellow
    Write-Host "  Note: For proper ARM64 support, buildx is recommended" -ForegroundColor Gray
    Write-Host ""
    
    # Fallback to standard docker build
    docker build --platform $Platform --build-arg TARGETPLATFORM=$Platform -t "${ImageName}:${Tag}" .
    $buildExitCode = $LASTEXITCODE
}

# If build failed, exit
if ($buildExitCode -ne 0) {
    Write-Host ""
    Write-Host "[ERROR] Docker build failed!" -ForegroundColor Red
    Write-Host "  Make sure Docker Desktop supports ARM64 builds" -ForegroundColor Yellow
    Write-Host "  You may need to enable experimental features or install buildx" -ForegroundColor Yellow
    exit 1
}

# Check final result
if ($buildExitCode -eq 0) {
    Write-Host ""
    Write-Host "[OK] Docker image built successfully!" -ForegroundColor Green
    Write-Host "  Image: ${ImageName}:${Tag}" -ForegroundColor Gray
    
    # Verify image platform
    Write-Host ""
    Write-Host "  Verifying image platform..." -ForegroundColor Gray
    $inspectOutput = docker inspect "${ImageName}:${Tag}" --format '{{.Architecture}}' 2>&1
    if ($LASTEXITCODE -eq 0) {
        if ($inspectOutput -match "arm64|aarch64") {
            Write-Host "  [OK] Image architecture: $inspectOutput (ARM64)" -ForegroundColor Green
        } else {
            Write-Host "  [WARNING] Image architecture: $inspectOutput (Expected: arm64)" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  [WARNING] Could not verify image architecture" -ForegroundColor Yellow
    }
    
    # Note: manifest inspect only works for registry images, not local images
    # The manifest will be created when pushing to ECR
    Write-Host "  [INFO] Manifest will be created when pushing to registry" -ForegroundColor Gray
    
    Write-Host ""
    
    # Step 3: Tag and push to ECR if ECR URL is provided
    if ($EcrRepoUrl) {
        $ecrImageTag = "${EcrRepoUrl}:${Tag}"
        
        # Extract registry URL from ECR repo URL (everything before the last /)
        $ecrRegistry = $EcrRepoUrl -replace '/[^/]+$', ''
        
        Write-Host "Step 3: Tagging image for ECR..." -ForegroundColor Yellow
        docker tag "${ImageName}:${Tag}" $ecrImageTag
        $tagExitCode = $LASTEXITCODE
        
        if ($tagExitCode -eq 0) {
            Write-Host "[OK] Image tagged successfully" -ForegroundColor Green
            Write-Host "  ECR Tag: $ecrImageTag" -ForegroundColor Gray
            Write-Host ""
            
            # Push to ECR if -Push flag is set, or auto-push when ECR URL provided (unless -NoPush)
            $shouldPush = $Push -or ($EcrRepoUrl -and -not $NoPush)
            if ($shouldPush) {
                Write-Host "Step 4: Authenticating with ECR..." -ForegroundColor Yellow
                
                # Try AWS Tools for PowerShell first
                $awsToolsAvailable = $false
                try {
                    $null = Get-Command Get-ECRLoginCommand -ErrorAction Stop
                    $awsToolsAvailable = $true
                } catch {
                    $awsToolsAvailable = $false
                }
                
                if ($awsToolsAvailable) {
                    Write-Host "  Using AWS Tools for PowerShell..." -ForegroundColor Gray
                    try {
                        (Get-ECRLoginCommand).Password | docker login --username AWS --password-stdin $ecrRegistry
                        $authExitCode = $LASTEXITCODE
                    } catch {
                        Write-Host "[ERROR] Failed to authenticate with ECR using AWS Tools for PowerShell" -ForegroundColor Red
                        Write-Host "  Error: $_" -ForegroundColor Gray
                        $authExitCode = 1
                    }
                } else {
                    # Fallback to AWS CLI
                    Write-Host "  AWS Tools for PowerShell not available, trying AWS CLI..." -ForegroundColor Gray
                    $awsCliAvailable = $false
                    try {
                        $null = aws --version 2>&1
                        $awsCliAvailable = $true
                    } catch {
                        $awsCliAvailable = $false
                    }
                    
                    if ($awsCliAvailable) {
                        # Extract region from ECR URL (e.g., ap-south-1 from dkr.ecr.ap-south-1.amazonaws.com)
                        if ($ecrRegistry -match '\.ecr\.([^.]+)\.amazonaws\.com') {
                            $awsRegion = $matches[1]
                        } else {
                            $awsRegion = "ap-south-1"  # Default region
                        }
                        
                        aws ecr get-login-password --region $awsRegion | docker login --username AWS --password-stdin $ecrRegistry
                        $authExitCode = $LASTEXITCODE
                    } else {
                        Write-Host "[ERROR] Neither AWS Tools for PowerShell nor AWS CLI is available!" -ForegroundColor Red
                        Write-Host "  Please install one of the following:" -ForegroundColor Yellow
                        Write-Host "    1. AWS Tools for PowerShell: Install-Module -Name AWS.Tools.ECR" -ForegroundColor Gray
                        Write-Host "    2. AWS CLI: https://aws.amazon.com/cli/" -ForegroundColor Gray
                        Write-Host ""
                        Write-Host "  Or manually authenticate:" -ForegroundColor Yellow
                        Write-Host "    (Get-ECRLoginCommand).Password | docker login --username AWS --password-stdin $ecrRegistry" -ForegroundColor Gray
                        exit 1
                    }
                }
                
                if ($authExitCode -eq 0) {
                    Write-Host "[OK] Authenticated with ECR successfully" -ForegroundColor Green
                    Write-Host ""
                    
                    Write-Host "Step 5: Pushing image to ECR with correct ARM64 manifest..." -ForegroundColor Yellow
                    Write-Host "  Using buildx to ensure correct platform manifest..." -ForegroundColor Gray
                    Write-Host "  This may take a few minutes..." -ForegroundColor Gray
                    Write-Host ""
                    
                    # Use buildx build --push to ensure correct ARM64 manifest
                    # Docker will use cached layers, so this is fast
                    # This is the most reliable way to ensure the manifest is correct
                    docker buildx build `
                        --platform $Platform `
                        --build-arg TARGETPLATFORM=$Platform `
                        --tag $ecrImageTag `
                        --push `
                        --progress plain `
                        .
                    $pushExitCode = $LASTEXITCODE
                    
                    # Fallback to standard push if buildx push fails
                    if ($pushExitCode -ne 0) {
                        Write-Host ""
                        Write-Host "  Buildx push failed, trying standard docker push..." -ForegroundColor Yellow
                        Write-Host "  Warning: Standard push may not include correct platform manifest" -ForegroundColor Yellow
                        Write-Host ""
                        docker push $ecrImageTag
                        $pushExitCode = $LASTEXITCODE
                    }
                    
                    if ($pushExitCode -eq 0) {
                        Write-Host ""
                        Write-Host "[OK] Image pushed to ECR successfully!" -ForegroundColor Green
                        Write-Host "  ECR Image: $ecrImageTag" -ForegroundColor Gray
                        Write-Host ""
                        
                        # Verify the manifest
                        Write-Host "  Verifying image manifest in ECR..." -ForegroundColor Gray
                        docker manifest inspect $ecrImageTag 2>&1 | Out-Null
                        if ($LASTEXITCODE -eq 0) {
                            $manifestOutput = docker manifest inspect $ecrImageTag 2>&1
                            if ($manifestOutput -match "linux/arm64|architecture.*arm64") {
                                Write-Host "  [OK] Image manifest includes linux/arm64" -ForegroundColor Green
                            } else {
                                Write-Host "  [WARNING] Could not verify ARM64 in manifest" -ForegroundColor Yellow
                                Write-Host "    You may need to rebuild and push with buildx" -ForegroundColor Yellow
                            }
                        }
                        
                        Write-Host ""
                        Write-Host "  Image is ready for ECS deployment!" -ForegroundColor Green
                        Write-Host "  Platform: linux/arm64" -ForegroundColor Gray
                        Write-Host "  Task Definition: equitywala-server-task" -ForegroundColor Gray
                    } else {
                        Write-Host ""
                        Write-Host "[ERROR] Failed to push image to ECR!" -ForegroundColor Red
                        Write-Host "  Check your AWS credentials and ECR permissions." -ForegroundColor Yellow
                        Write-Host ""
                        Write-Host "  Manual push command:" -ForegroundColor Cyan
                        Write-Host "    docker push $ecrImageTag" -ForegroundColor Gray
                        exit 1
                    }
                } else {
                    Write-Host ""
                    Write-Host "[ERROR] Failed to authenticate with ECR!" -ForegroundColor Red
                    Write-Host "  Please authenticate manually:" -ForegroundColor Yellow
                    Write-Host "    (Get-ECRLoginCommand).Password | docker login --username AWS --password-stdin $ecrRegistry" -ForegroundColor Gray
                    Write-Host "  Or using AWS CLI:" -ForegroundColor Yellow
                    Write-Host "    aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin $ecrRegistry" -ForegroundColor Gray
                    exit 1
                }
            } else {
                Write-Host ""
                Write-Host "To push to ECR, first authenticate:" -ForegroundColor Cyan
                Write-Host "  (Get-ECRLoginCommand).Password | docker login --username AWS --password-stdin $ecrRegistry" -ForegroundColor Gray
                Write-Host "Then push:" -ForegroundColor Cyan
                Write-Host "  docker push $ecrImageTag" -ForegroundColor Gray
            }
        } else {
            Write-Host ""
            Write-Host "[ERROR] Failed to tag image!" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "To tag and push to ECR:" -ForegroundColor Cyan
        Write-Host "  docker tag ${ImageName}:${Tag} <ecr-repo-url>:${Tag}" -ForegroundColor Gray
        Write-Host "  docker push <ecr-repo-url>:${Tag}" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Or use the script with ECR URL (auto-pushes):" -ForegroundColor Cyan
        Write-Host "  .\build-docker.ps1 ${ImageName} ${Tag} <ecr-repo-url>" -ForegroundColor Gray
        Write-Host "  Example: .\build-docker.ps1 ${ImageName} ${Tag} 128977214845.dkr.ecr.ap-south-1.amazonaws.com/equitywala-server" -ForegroundColor Gray
    }
} else {
    Write-Host ""
    Write-Host "[ERROR] Docker build failed!" -ForegroundColor Red
    exit 1
}
