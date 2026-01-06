# setup-test-harness.ps1
# Test Harness Setup Script for hub-hrms
# Sets up testing infrastructure for both frontend and backend

$ErrorActionPreference = "Stop"

Write-Host "🧪 Setting up test harness for hub-hrms..." -ForegroundColor Blue

# Check if we're in the right directory
if (-not (Test-Path "go.mod")) {
    Write-Host "Error: go.mod not found. Please run this script from the project root." -ForegroundColor Red
    exit 1
}

# Backend Testing Setup
Write-Host "`n📦 Setting up backend testing infrastructure..." -ForegroundColor Cyan

# Install Go testing dependencies
Write-Host "Installing Go testing packages..." -ForegroundColor Yellow
go get -u github.com/stretchr/testify/assert
go get -u github.com/stretchr/testify/mock
go get -u github.com/stretchr/testify/suite
go get -u github.com/DATA-DOG/go-sqlmock
go get -u github.com/golang/mock/gomock
go get -u github.com/jarcoal/httpmock

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to install Go dependencies" -ForegroundColor Red
    exit 1
}

# Create test directories
Write-Host "Creating test directory structure..." -ForegroundColor Yellow
$testDirs = @(
    "tests\unit",
    "tests\integration",
    "tests\e2e",
    "tests\fixtures",
    "tests\mocks"
)

foreach ($dir in $testDirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "  Created: $dir" -ForegroundColor Green
    }
}

# Frontend Testing Setup
Write-Host "`n📦 Setting up frontend testing infrastructure..." -ForegroundColor Cyan

# Check if frontend directory exists
if (Test-Path "frontend") {
    Push-Location frontend
    
    Write-Host "Installing frontend testing dependencies..." -ForegroundColor Yellow
    
    # Install testing dependencies
    npm install --save-dev `
        @testing-library/svelte `
        @testing-library/jest-dom `
        @testing-library/user-event `
        vitest `
        jsdom `
        @vitest/ui `
        @vitest/coverage-v8 `
        svelte-testing-library `
        msw
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Failed to install frontend dependencies" -ForegroundColor Red
        Pop-Location
        exit 1
    }
    
    Pop-Location
} else {
    Write-Host "Warning: frontend directory not found" -ForegroundColor Yellow
}

Write-Host "`n✅ Test harness setup complete!" -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Cyan
Write-Host "1. Run backend tests: .\scripts\test-backend.ps1" -ForegroundColor White
Write-Host "2. Run frontend tests: .\scripts\test-frontend.ps1" -ForegroundColor White
Write-Host "3. Run all tests: .\scripts\test-all.ps1" -ForegroundColor White
