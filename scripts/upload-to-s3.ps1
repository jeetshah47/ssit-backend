# S3 Upload Script for ssit-backend (PowerShell)
# Usage: .\scripts\upload-to-s3.ps1 <file-path> [options]

param(
    [Parameter(Position=0)]
    [string]$FilePath,
    
    [string]$Key,
    [string]$Prefix,
    [string]$ContentType,
    [string]$Bucket,
    [string]$Region
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

# Load .env file if it exists
$EnvFile = Join-Path $ProjectRoot ".env"
if (Test-Path $EnvFile) {
    Get-Content $EnvFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

if ([string]::IsNullOrEmpty($FilePath)) {
    Write-Host "S3 Upload Script"
    Write-Host ""
    Write-Host "Usage:"
    Write-Host "  .\scripts\upload-to-s3.ps1 <file-path> [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Key         S3 object key (optional, defaults to filename)"
    Write-Host "  -Prefix      S3 key prefix/folder (optional)"
    Write-Host "  -ContentType Content-Type header (optional, auto-detected)"
    Write-Host "  -Bucket      S3 bucket name (optional, uses S3_BUCKET_NAME env var)"
    Write-Host "  -Region      AWS region (optional, uses AWS_REGION env var)"
    Write-Host ""
    Write-Host "Environment Variables (can be set in .env file):"
    Write-Host "  AWS_ACCESS_KEY_ID      AWS Access Key ID (required)"
    Write-Host "  AWS_SECRET_ACCESS_KEY  AWS Secret Access Key (required)"
    Write-Host "  AWS_REGION             AWS Region (default: ap-south-1)"
    Write-Host "  S3_BUCKET_NAME         S3 Bucket Name (required)"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\scripts\upload-to-s3.ps1 .\logo.png"
    Write-Host "  .\scripts\upload-to-s3.ps1 .\document.pdf -Prefix uploads/documents"
    Write-Host "  .\scripts\upload-to-s3.ps1 .\image.jpg -Key custom-name.jpg"
    exit 1
}

$args = @("-file", $FilePath)

if (-not [string]::IsNullOrEmpty($Key)) {
    $args += @("-key", $Key)
}
if (-not [string]::IsNullOrEmpty($Prefix)) {
    $args += @("-prefix", $Prefix)
}
if (-not [string]::IsNullOrEmpty($ContentType)) {
    $args += @("-content-type", $ContentType)
}
if (-not [string]::IsNullOrEmpty($Bucket)) {
    $args += @("-bucket", $Bucket)
}
if (-not [string]::IsNullOrEmpty($Region)) {
    $args += @("-region", $Region)
}

Push-Location $ProjectRoot
try {
    & go run scripts/upload-to-s3.go @args
} finally {
    Pop-Location
}

