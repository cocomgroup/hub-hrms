# Backend Test Runner for hub-hrms (PowerShell - Windows)
# Runs all backend tests with coverage reporting

param(
    [switch]$Coverage,
    [switch]$Verbose,
    [switch]$Integration,
    [switch]$E2E,
    [switch]$Race,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

if ($Help) {
    Write-Host "Backend Test Runner for hub-hrms"
    Write-Host ""
    Write-Host "Usage: .\test-backend.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Coverage      Generate coverage reports"
    Write-Host "  -Verbose       Show verbose test output"
    Write-Host "  -Integration   Run integration tests"
    Write-Host "  -E2E          Run E2E tests"
    Write-Host "  -Race         Enable race detection (requires CGO)"
    Write-Host "  -Help         Show this help message"
    Write-Host ""
    Write-Host "Note: Race detection requires CGO. Set CGO_ENABLED=1 to use -Race"
    exit 0
}

# Detect backend directory
$backendPath = "."
if (Test-Path "backend\go.mod") {
    $backendPath = "backend"
    Write-Host "Detected monorepo structure, using backend directory" -ForegroundColor Cyan
} elseif (-not (Test-Path "go.mod")) {
    Write-Host "Error: go.mod not found. Run from project root or backend directory." -ForegroundColor Red
    exit 1
}

Write-Host "Running backend test suite..." -ForegroundColor Blue
Write-Host "Backend path: $backendPath" -ForegroundColor Gray

# Check if CGO is available for race detection
$cgoEnabled = $env:CGO_ENABLED -eq "1"
if ($Race -and -not $cgoEnabled) {
    Write-Host "Warning: Race detection requires CGO. Set CGO_ENABLED=1 or install GCC." -ForegroundColor Yellow
    Write-Host "Continuing without race detection..." -ForegroundColor Yellow
    $Race = $false
}

# Change to backend directory
Push-Location $backendPath

try {
    # Create coverage directory
    if (-not (Test-Path "coverage")) {
        New-Item -ItemType Directory -Path "coverage" | Out-Null
    }

    # Unit Tests
    Write-Host ""
    Write-Host "Running unit tests..." -ForegroundColor Yellow

    $testArgs = @()
    if ($Verbose) {
        $testArgs += "-v"
    }
    
    if ($Race) {
        $testArgs += "-race"
    }

    if ($Coverage) {
        $testArgs += "-coverprofile=coverage\unit.out"
        $testArgs += "-covermode=atomic"
    }

    $testArgs += "./..."
    $testArgs += "-short"

    Write-Host "Running: go test $($testArgs -join ' ')" -ForegroundColor Gray
    & go test @testArgs
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Unit tests failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "Unit tests passed" -ForegroundColor Green

    # Integration Tests
    if ($Integration) {
        Write-Host ""
        Write-Host "Running integration tests..." -ForegroundColor Yellow
        
        $integrationArgs = @("-v")
        
        if ($Race) {
            $integrationArgs += "-race"
        }
        
        if ($Coverage) {
            $integrationArgs += "-coverprofile=coverage\integration.out"
            $integrationArgs += "-covermode=atomic"
        }
        
        $integrationArgs += "./tests/integration/..."
        
        Write-Host "Running: go test $($integrationArgs -join ' ')" -ForegroundColor Gray
        & go test @integrationArgs
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Integration tests failed" -ForegroundColor Red
            exit 1
        }
        Write-Host "Integration tests passed" -ForegroundColor Green
    }

    # E2E Tests
    if ($E2E) {
        Write-Host ""
        Write-Host "Running E2E tests..." -ForegroundColor Yellow
        
        $e2eArgs = @("-v")
        if ($Race) {
            $e2eArgs += "-race"
        }
        $e2eArgs += "./tests/e2e/..."
        
        Write-Host "Running: go test $($e2eArgs -join ' ')" -ForegroundColor Gray
        & go test @e2eArgs
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "E2E tests failed" -ForegroundColor Red
            exit 1
        }
        Write-Host "E2E tests passed" -ForegroundColor Green
    }

    # Generate coverage report
    if ($Coverage) {
        Write-Host ""
        Write-Host "Generating coverage report..." -ForegroundColor Yellow
        
        if (Test-Path "coverage\unit.out") {
            go tool cover -html=coverage\unit.out -o coverage\coverage.html
            
            Write-Host ""
            Write-Host "Coverage Summary:" -ForegroundColor Cyan
            $coverageOutput = go tool cover -func=coverage\unit.out
            $coverageOutput | Select-Object -Last 1
            
            Write-Host ""
            Write-Host "Coverage report: $backendPath\coverage\coverage.html" -ForegroundColor Green
            
            # Open in browser (optional)
            # Start-Process "$backendPath\coverage\coverage.html"
        }
    }

    Write-Host ""
    Write-Host "Backend tests complete!" -ForegroundColor Green
}
catch {
    Write-Host ""
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}
finally {
    Pop-Location
}