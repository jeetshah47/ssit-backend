# PowerShell script to truncate all tables in the database
# Usage: .\truncate-tables.ps1 [-Confirm] [-Tables "users,otps"] [-Verify] [-List]

param(
    [switch]$Confirm,
    [string]$Tables = "",
    [switch]$Verify,
    [switch]$List,
    [string]$EnvFile = ".env"
)

# Load environment variables from .env file if it exists
if (Test-Path $EnvFile) {
    Get-Content $EnvFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

# Get database connection parameters
$dbHost = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$dbPort = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
$dbPassword = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "postgres" }
$dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "equitywala" }
$dbSSLMode = if ($env:DB_SSLMODE) { $env:DB_SSLMODE } else { "disable" }

# Check if DATABASE_URL is set
if ($env:DATABASE_URL) {
    $databaseUrl = $env:DATABASE_URL
} else {
    $databaseUrl = "postgresql://${dbUser}:${dbPassword}@${dbHost}:${dbPort}/${dbName}?sslmode=${dbSSLMode}"
}

# Define all tables in truncation order (child tables first)
$allTables = @(
    "payment_webhooks",
    "payments",
    "voucher_redemptions",
    "subscriptions",
    "vouchers",
    "pricing_packages",
    "advisory_views",
    "advisories",
    "team_members",
    "notification_settings",
    "user_classification_logs",
    "kyc_records",
    "user_roles",
    "sessions",
    "otps",
    "roles",
    "users"
)

# Determine which tables to truncate
$tablesToTruncate = $allTables
if ($Tables -ne "") {
    $tablesToTruncate = $Tables -split "," | ForEach-Object { $_.Trim() }
}

# List tables if requested
if ($List) {
    Write-Host "Tables that would be truncated:" -ForegroundColor Cyan
    foreach ($table in $tablesToTruncate) {
        try {
            $count = psql -d $databaseUrl -t -c "SELECT COUNT(*) FROM $table;" 2>$null
            $count = $count.Trim()
            Write-Host "  - $table ($count rows)"
        } catch {
            Write-Host "  - $table (error checking count)"
        }
    }
    exit 0
}

# Safety check
if (-not $Confirm) {
    Write-Host "ERROR: Truncation requires confirmation flag" -ForegroundColor Red
    Write-Host "Usage: .\truncate-tables.ps1 -Confirm"
    Write-Host "Or: .\truncate-tables.ps1 -Confirm -Tables 'users,otps'"
    Write-Host "Use -List to see tables and row counts"
    exit 1
}

# Check if psql is available
$psqlPath = Get-Command psql -ErrorAction SilentlyContinue
if (-not $psqlPath) {
    Write-Host "ERROR: psql command not found. Please install PostgreSQL client tools." -ForegroundColor Red
    Write-Host "You can download it from: https://www.postgresql.org/download/" -ForegroundColor Yellow
    exit 1
}

# Create SQL script content
$sqlScript = @"
BEGIN;

"@

foreach ($table in $tablesToTruncate) {
    $sqlScript += "TRUNCATE TABLE $table CASCADE;`n"
}

$sqlScript += @"

COMMIT;
"@

# Write SQL to temporary file
$tempFile = [System.IO.Path]::GetTempFileName() + ".sql"
$sqlScript | Out-File -FilePath $tempFile -Encoding UTF8

try {
    Write-Host "Truncating tables..." -ForegroundColor Cyan
    
    # Execute SQL script
    $env:PGPASSWORD = $dbPassword
    $result = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -f $tempFile 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✓ All tables truncated successfully!" -ForegroundColor Green
    } else {
        Write-Host "`n✗ Error truncating tables:" -ForegroundColor Red
        Write-Host $result
        exit 1
    }
    
    # Verify if requested
    if ($Verify) {
        Write-Host "`nVerifying tables are empty..." -ForegroundColor Cyan
        $allEmpty = $true
        foreach ($table in $tablesToTruncate) {
            $count = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM $table;" 2>&1
            $count = $count.Trim()
            if ($count -eq "0" -or $count -eq "") {
                Write-Host "  ✓ $table : empty" -ForegroundColor Green
            } else {
                Write-Host "  ✗ $table : $count rows (not empty!)" -ForegroundColor Red
                $allEmpty = $false
            }
        }
        if ($allEmpty) {
            Write-Host "`n✓ All tables verified as empty!" -ForegroundColor Green
        } else {
            Write-Host "`n✗ Some tables are not empty!" -ForegroundColor Red
            exit 1
        }
    }
} finally {
    # Clean up temporary file
    if (Test-Path $tempFile) {
        Remove-Item $tempFile -Force
    }
    Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue
}

