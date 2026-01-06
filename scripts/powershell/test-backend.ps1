# test-backend.ps1
# Backend Test Runner for hub-hrms
# Runs all backend tests with coverage reporting

param(
    [switch]$Coverage,
    [switch]$Verbose,
    [switch]$Integration,
    [switch]$E2E,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

# Show help
if ($Help) {
    Write-Host "Backend Test Runner for hub-hrms" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\test-backend.ps1 [options]" -ForegroundColor White
    Write-Host ""
    Write-Host "Options:" -ForegroundColor Yellow
    Write-Host "  -Coverage        Generate coverage reports" -ForegroundColor White
    Write-Host "  -Verbose         Show verbose output" -ForegroundColor White
    Write-Host "  -Integration     Run integration tests" -ForegroundColor White
    Write-Host "  -E2E             Run E2E tests" -ForegroundColor White
    Write-Host "  -Help            Show this help message" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\test-backend.ps1" -ForegroundColor White
    Write-Host "  .\test-backend.ps1 -Coverage" -ForegroundColor White
    Write-Host "  .\test-backend.ps1 -Integration -Coverage" -ForegroundColor White
    exit 0
}

Write-Host "🧪 Running backend test suite..." -ForegroundColor Blue

# Create coverage directory
if (-not (Test-Path "coverage")) {
    New-Item -ItemType Directory -Path "coverage" -Force | Out-Null
}

# Build test command arguments
$testArgs = @()
if ($Verbose) {
    $testArgs += "-v"
}
$testArgs += "-race"

# Unit Tests
Write-Host "`n📝 Running unit tests..." -ForegroundColor Yellow

if ($Coverage) {
    $testArgs += "-coverprofile=coverage\unit.out"
    $testArgs += "-covermode=atomic"
}

$testArgs += "./..."
$testArgs += "-short"

try {
    & go test @testArgs
    if ($LASTEXITCODE -ne 0) {
        throw "Unit tests failed"
    }
    Write-Host "✅ Unit tests passed" -ForegroundColor Green
} catch {
    Write-Host "❌ Unit tests failed" -ForegroundColor Red
    exit 1
}

# Integration Tests
if ($Integration) {
    Write-Host "`n📝 Running integration tests..." -ForegroundColor Yellow
    
    $integrationArgs = @("-v", "-race")
    if ($Coverage) {
        $integrationArgs += "-coverprofile=coverage\integration.out"
        $integrationArgs += "-covermode=atomic"
    }
    $integrationArgs += "./tests/integration/..."
    
    try {
        & go test @integrationArgs
        if ($LASTEXITCODE -ne 0) {
            throw "Integration tests failed"
        }
        Write-Host "✅ Integration tests passed" -ForegroundColor Green
    } catch {
        Write-Host "❌ Integration tests failed" -ForegroundColor Red
        exit 1
    }
}

# E2E Tests
if ($E2E) {
    Write-Host "`n📝 Running E2E tests..." -ForegroundColor Yellow
    
    try {
        go test -v -race ./tests/e2e/...
        if ($LASTEXITCODE -ne 0) {
            throw "E2E tests failed"
        }
        Write-Host "✅ E2E tests passed" -ForegroundColor Green
    } catch {
        Write-Host "❌ E2E tests failed" -ForegroundColor Red
        exit 1
    }
}

# Generate coverage report
if ($Coverage) {
    Write-Host "`n📊 Generating coverage report..." -ForegroundColor Yellow
    
    if (Test-Path "coverage\unit.out") {
        go tool cover -html=coverage\unit.out -o coverage\coverage.html
        
        Write-Host "Coverage summary:" -ForegroundColor Cyan
        go tool cover -func=coverage\unit.out | Select-Object -Last 1
        
        Write-Host "`n✅ Coverage report generated: coverage\coverage.html" -ForegroundColor Green
    }
}

Write-Host "`n✅ Backend tests complete!" -ForegroundColor Green
