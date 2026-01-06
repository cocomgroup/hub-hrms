# Test Harness Setup Script for hub-hrms Monorepo (PowerShell)
$ErrorActionPreference = "Stop"

Write-Host "ðŸ§ª Setting up test harness for hub-hrms monorepo..." -ForegroundColor Blue

# Detect project structure
$backendPath = if (Test-Path "backend\go.mod") { "backend" } elseif (Test-Path "go.mod") { "." } else { $null }
$frontendPath = if (Test-Path "frontend\package.json") { "frontend" } elseif (Test-Path "..\frontend\package.json") { "..\frontend" } else { $null }

if (-not $backendPath) {
    Write-Host "Error: Could not find backend directory with go.mod" -ForegroundColor Red
    Write-Host "Current location: $(Get-Location)" -ForegroundColor Yellow
    exit 1
}

# Backend Testing Setup
Write-Host "Setting up backend testing infrastructure..." -ForegroundColor Cyan
Write-Host "Backend path: $backendPath" -ForegroundColor Gray

Push-Location $backendPath

try {
    # Install Go testing dependencies
    Write-Host "Installing Go test dependencies..."
    go get -u github.com/stretchr/testify/assert
    go get -u github.com/stretchr/testify/mock
    go get -u github.com/stretchr/testify/suite
    go get -u github.com/DATA-DOG/go-sqlmock
    go get -u github.com/golang/mock/gomock
    go get -u github.com/jarcoal/httpmock

    # Create test directories
    Write-Host "Creating backend test directory structure..."
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
}
finally {
    Pop-Location
}

# Frontend Testing Setup
if ($frontendPath) {
    Write-Host "Setting up frontend testing infrastructure..." -ForegroundColor Cyan
    Write-Host "Frontend path: $frontendPath" -ForegroundColor Gray
    
    Push-Location $frontendPath
    
    try {
        Write-Host "Installing frontend test dependencies..."
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
        
        # Create frontend test directory
        if (-not (Test-Path "src\tests")) {
            New-Item -ItemType Directory -Path "src\tests" -Force | Out-Null
            Write-Host "  Created: src\tests" -ForegroundColor Green
        }
    }
    finally {
        Pop-Location
    }
}
else {
    Write-Host "Warning: frontend directory not found" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "âœ… Test harness setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:"
Write-Host "1. Run backend tests: .\scripts\powershell\test-backend.ps1"
Write-Host "2. Run frontend tests: .\scripts\powershell\test-frontend.ps1"
Write-Host "3. Run all tests: .\scripts\powershell\test-all.ps1"
