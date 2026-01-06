# Install Backend Tests Script
# Copies test files to appropriate directories in hub-hrms backend

param(
    [string]$BackendPath = ".\backend",
    [switch]$Help
)

$ErrorActionPreference = "Stop"

if ($Help) {
    Write-Host "Install Backend Tests"
    Write-Host ""
    Write-Host "Usage: .\install-tests.ps1 [-BackendPath <path>]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -BackendPath    Path to backend directory (default: .\backend)"
    Write-Host "  -Help           Show this help message"
    exit 0
}

Write-Host "Installing backend tests..." -ForegroundColor Blue

# Check if backend directory exists
if (-not (Test-Path $BackendPath)) {
    Write-Host "Error: Backend directory not found: $BackendPath" -ForegroundColor Red
    exit 1
}

# Check if go.mod exists
if (-not (Test-Path "$BackendPath\go.mod")) {
    Write-Host "Error: go.mod not found in $BackendPath" -ForegroundColor Red
    exit 1
}

# Test files mapping
$testFiles = @{
    "config_test.go"  = "internal\config"
    "models_test.go"  = "internal\models"
    "service_test.go" = "internal\service"
    "api_test.go"     = "internal\api"
}

Write-Host ""
Write-Host "Copying test files..." -ForegroundColor Yellow

foreach ($file in $testFiles.Keys) {
    $source = $file
    $destDir = Join-Path $BackendPath $testFiles[$file]
    $dest = Join-Path $destDir $file
    
    if (-not (Test-Path $source)) {
        Write-Host "  Warning: $source not found, skipping" -ForegroundColor Yellow
        continue
    }
    
    if (-not (Test-Path $destDir)) {
        New-Item -ItemType Directory -Path $destDir -Force | Out-Null
    }
    
    Copy-Item $source $dest -Force
    Write-Host "  Copied: $file -> $($testFiles[$file])" -ForegroundColor Green
}

Write-Host ""
Write-Host "Installing Go test dependencies..." -ForegroundColor Yellow

Push-Location $BackendPath

try {
    # Install testing dependencies
    Write-Host "  Installing testify..."
    go get -u github.com/stretchr/testify/assert 2>$null
    go get -u github.com/stretchr/testify/mock 2>$null
    go get -u github.com/stretchr/testify/suite 2>$null
    
    Write-Host "  Running go mod tidy..."
    go mod tidy
    
    Write-Host ""
    Write-Host "Test installation complete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. Run all tests: go test ./..." -ForegroundColor White
    Write-Host "2. Run with coverage: go test -coverprofile=coverage.out ./..." -ForegroundColor White
    Write-Host "3. View coverage: go tool cover -html=coverage.out" -ForegroundColor White
}
catch {
    Write-Host "Error installing dependencies: $_" -ForegroundColor Red
    exit 1
}
finally {
    Pop-Location
}

Write-Host ""
Write-Host "Test files installed successfully!" -ForegroundColor Green
