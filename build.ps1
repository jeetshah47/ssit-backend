# Build script for Go application (PowerShell)
# This script builds the Go binary and places it in the build folder
# Usage: .\build.ps1 [target]
#   target: linux (default for Docker), windows, or current
#   Example: .\build.ps1 linux    (for Docker)
#            .\build.ps1 windows  (for Windows)
#            .\build.ps1 current (for current OS)

param(
    [string]$Target = "linux"
)

$ErrorActionPreference = "Stop"

Write-Host "Building Go application for: $Target" -ForegroundColor Cyan

# Create build directory if it doesn't exist
if (-not (Test-Path -Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
    Write-Host "Created build directory" -ForegroundColor Green
}

# Build the application based on target
$env:CGO_ENABLED = "0"

switch ($Target.ToLower()) {
    "linux" {
        Write-Host "Building for Linux (Docker)..." -ForegroundColor Yellow
        $env:GOOS = "linux"
        $env:GOARCH = "arm64"
        go build -o build/server ./cmd/server
        $OutputFile = "build/server"
    }
    "windows" {
        Write-Host "Building for Windows..." -ForegroundColor Yellow
        go build -o build/server.exe ./cmd/server
        $OutputFile = "build/server.exe"
    }
    "current" {
        Write-Host "Building for current OS..." -ForegroundColor Yellow
        go build -o build/server.exe ./cmd/server
        $OutputFile = "build/server.exe"
    }
    default {
        Write-Host "Unknown target: $Target" -ForegroundColor Red
        Write-Host "Usage: .\build.ps1 [linux|windows|current]" -ForegroundColor Yellow
        exit 1
    }
}

# Check if build was successful
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Build successful! Binary created at: $OutputFile" -ForegroundColor Green
    
    # Display file info
    if (Test-Path -Path $OutputFile) {
        $fileInfo = Get-Item $OutputFile
        Write-Host "File size: $([math]::Round($fileInfo.Length / 1MB, 2)) MB" -ForegroundColor Gray
        Write-Host "Created: $($fileInfo.CreationTime)" -ForegroundColor Gray
    }
} else {
    Write-Host "[ERROR] Build failed!" -ForegroundColor Red
    exit 1
}
