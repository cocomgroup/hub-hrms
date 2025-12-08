# Fix Admin Password - Generate Correct Hash
# This script generates a proper bcrypt hash and updates the database

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Fix Admin Password Hash" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "hash_password.go")) {
    Write-Host "Error: Run this script from backend/scripts directory" -ForegroundColor Red
    exit 1
}

# Check if Go is installed
try {
    $null = Get-Command go -ErrorAction Stop
} catch {
    Write-Host "Error: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

# Get database connection details
Write-Host "Enter database connection details:" -ForegroundColor Yellow
$DB_HOST = Read-Host "Host [localhost]"
if ([string]::IsNullOrWhiteSpace($DB_HOST)) { $DB_HOST = "localhost" }

$DB_PORT = Read-Host "Port [5432]"
if ([string]::IsNullOrWhiteSpace($DB_PORT)) { $DB_PORT = "5432" }

$DB_NAME = Read-Host "Database [drmsdb]"
if ([string]::IsNullOrWhiteSpace($DB_NAME)) { $DB_NAME = "drmsdb" }

$DB_USER = Read-Host "User [postgres]"
if ([string]::IsNullOrWhiteSpace($DB_USER)) { $DB_USER = "postgres" }

$SecurePassword = Read-Host "Database Password" -AsSecureString
$DB_PASSWORD = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecurePassword)
)

Write-Host ""
$USER_EMAIL = Read-Host "Admin user email [admin@cocomgroup.com]"
if ([string]::IsNullOrWhiteSpace($USER_EMAIL)) { $USER_EMAIL = "admin@cocomgroup.com" }

$USER_PASSWORD = Read-Host "New password for admin [admin123]"
if ([string]::IsNullOrWhiteSpace($USER_PASSWORD)) { $USER_PASSWORD = "admin123" }

Write-Host ""
Write-Host "Generating bcrypt hash..." -ForegroundColor Yellow

# Generate hash using Go
$hashOutput = & go run hash_password.go $USER_PASSWORD 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error generating hash:" -ForegroundColor Red
    Write-Host $hashOutput
    exit 1
}

# Extract hash from output (it's on line 3)
$lines = $hashOutput -split "`n"
$hash = ""
foreach ($line in $lines) {
    if ($line -match '^\$2[ab]\$') {
        $hash = $line.Trim()
        break
    }
}

if ([string]::IsNullOrWhiteSpace($hash)) {
    Write-Host "Error: Could not extract hash" -ForegroundColor Red
    Write-Host "Output was:" -ForegroundColor Yellow
    Write-Host $hashOutput
    exit 1
}

Write-Host "[OK] Hash generated successfully" -ForegroundColor Green
Write-Host "Hash: $hash" -ForegroundColor Gray
Write-Host ""

# Update database
Write-Host "Updating database..." -ForegroundColor Yellow
$env:PGPASSWORD = $DB_PASSWORD

$updateSQL = "UPDATE users SET password_hash = '$hash' WHERE email = '$USER_EMAIL';"

$result = $updateSQL | & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error updating database:" -ForegroundColor Red
    Write-Host $result
    $env:PGPASSWORD = $null
    exit 1
}

# Check if user was updated
$checkSQL = "SELECT COUNT(*) FROM users WHERE email = '$USER_EMAIL';"
$count = & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c $checkSQL 2>&1
$userExists = [int]($count -replace '\s','')

if ($userExists -eq 0) {
    Write-Host "WARNING: No user found with email: $USER_EMAIL" -ForegroundColor Yellow
    Write-Host "User may not exist. Run setup.ps1 first to create the user." -ForegroundColor Yellow
} else {
    Write-Host "[OK] Password updated successfully!" -ForegroundColor Green
}

$env:PGPASSWORD = $null

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Password Reset Complete!" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Email:    $USER_EMAIL" -ForegroundColor White
Write-Host "Password: $USER_PASSWORD" -ForegroundColor White
Write-Host ""
Write-Host "Try logging in now!" -ForegroundColor Green
Write-Host ""