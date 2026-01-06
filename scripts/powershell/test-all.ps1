# test-all.ps1
# Master Test Runner for hub-hrms
# Runs all test types: unit, integration, E2E, frontend

param(
    [switch]$All,
    [switch]$Unit,
    [switch]$Integration,
    [switch]$E2E,
    [switch]$Frontend,
    [switch]$Coverage,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

# Show help
if ($Help) {
    Write-Host "Master Test Runner for hub-hrms" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\test-all.ps1 [options]" -ForegroundColor White
    Write-Host ""
    Write-Host "Options:" -ForegroundColor Yellow
    Write-Host "  -All           Run all tests (unit, integration, e2e, frontend)" -ForegroundColor White
    Write-Host "  -Unit          Run unit tests only (default)" -ForegroundColor White
    Write-Host "  -Integration   Run integration tests" -ForegroundColor White
    Write-Host "  -E2E           Run E2E tests" -ForegroundColor White
    Write-Host "  -Frontend      Run frontend tests (default)" -ForegroundColor White
    Write-Host "  -Coverage      Generate coverage reports" -ForegroundColor White
    Write-Host "  -Help          Show this help message" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\test-all.ps1" -ForegroundColor White
    Write-Host "  .\test-all.ps1 -All -Coverage" -ForegroundColor White
    Write-Host "  .\test-all.ps1 -Integration -Coverage" -ForegroundColor White
    exit 0
}

# Default behavior: run unit and frontend tests
if (-not ($All -or $Unit -or $Integration -or $E2E -or $Frontend)) {
    $Unit = $true
    $Frontend = $true
}

# If -All is specified, enable everything
if ($All) {
    $Unit = $true
    $Integration = $true
    $E2E = $true
    $Frontend = $true
}

Write-Host "╔════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║     hub-hrms Test Suite                    ║" -ForegroundColor Blue
Write-Host "╚════════════════════════════════════════════╝" -ForegroundColor Blue
Write-Host ""

# Test counters
$totalTests = 0
$passedTests = 0
$failedTests = 0

# Track start time
$startTime = Get-Date

# Backend Unit Tests
if ($Unit) {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Yellow
    Write-Host "  Running Backend Unit Tests" -ForegroundColor Yellow
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Yellow
    Write-Host ""
    
    $totalTests++
    try {
        if ($Coverage) {
            & .\scripts\test-backend.ps1 -Coverage
        } else {
            & .\scripts\test-backend.ps1
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Backend unit tests passed" -ForegroundColor Green
            $passedTests++
        } else {
            throw "Tests failed"
        }
    } catch {
        Write-Host "❌ Backend unit tests failed" -ForegroundColor Red
        $failedTests++
    }
    Write-Host ""
}

# Backend Integration Tests
if ($Integration) {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Yellow
    Write-Host "  Running Backend Integration Tests" -ForegroundColor Yellow
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Yellow
    Write-Host ""
    
    $totalTests++
    
    # Start test database
    Write-Host "Starting test database..." -ForegroundColor Cyan
    docker-compose -f docker-compose.test.yml up -d db
    Start-Sleep -Seconds 5
    
    try {
        if ($Coverage) {
            & .\scripts\test-backend.ps1 -Integration -Coverage
        } else {
            & .\scripts\test-backend.ps1 -Integration
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Integration tests passed" -ForegroundColor Green
            $passedTests++
        } else {
            throw "Tests failed"
        }
    } catch {
        Write-Host "❌ Integration tests failed" -ForegroundColor Red
        $failedTests++
    } finally {
        # Cleanup test database
        Write-Host "Cleaning up test database..." -ForegroundColor Cyan
        docker-compose -f docker-compose.test.yml down
    }
    Write-Host ""
}

# Frontend Tests
if ($Frontend) {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Yellow
    Write-Host "  Running Frontend Tests" -ForegroundColor Yellow
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Yellow
    Write-Host ""
    
    $totalTests++
    try {
        if ($Coverage) {
            & .\scripts\test-frontend.ps1 -Coverage
        } else {
            & .\scripts\test-frontend.ps1
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Frontend tests passed" -ForegroundColor Green
            $passedTests++
        } else {
            throw "Tests failed"
        }
    } catch {
        Write-Host "❌ Frontend tests failed" -ForegroundColor Red
        $failedTests++
    }
    Write-Host ""
}

# E2E Tests
if ($E2E) {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Yellow
    Write-Host "  Running E2E Tests" -ForegroundColor Yellow
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Yellow
    Write-Host ""
    
    $totalTests++
    
    # Start the application
    Write-Host "Starting application for E2E tests..." -ForegroundColor Cyan
    docker-compose -f docker-compose.test.yml up -d
    Start-Sleep -Seconds 10
    
    # Wait for application to be ready
    Write-Host "Waiting for application to be ready..." -ForegroundColor Cyan
    $timeout = 60
    $elapsed = 0
    $ready = $false
    
    while ($elapsed -lt $timeout -and -not $ready) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:5173" -UseBasicParsing -TimeoutSec 2
            if ($response.StatusCode -eq 200) {
                $ready = $true
            }
        } catch {
            Start-Sleep -Seconds 2
            $elapsed += 2
        }
    }
    
    if ($ready) {
        try {
            & .\scripts\test-backend.ps1 -E2E
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ E2E tests passed" -ForegroundColor Green
                $passedTests++
            } else {
                throw "Tests failed"
            }
        } catch {
            Write-Host "❌ E2E tests failed" -ForegroundColor Red
            $failedTests++
        }
    } else {
        Write-Host "❌ Application failed to start" -ForegroundColor Red
        $failedTests++
    }
    
    # Cleanup
    Write-Host "Cleaning up..." -ForegroundColor Cyan
    docker-compose -f docker-compose.test.yml down
    Write-Host ""
}

# Calculate duration
$endTime = Get-Date
$duration = ($endTime - $startTime).TotalSeconds

# Print summary
Write-Host "╔════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║     Test Summary                           ║" -ForegroundColor Blue
Write-Host "╚════════════════════════════════════════════╝" -ForegroundColor Blue
Write-Host ""
Write-Host "Total test suites: $totalTests" -ForegroundColor White
Write-Host "Passed: $passedTests" -ForegroundColor Green
Write-Host "Failed: $failedTests" -ForegroundColor Red
Write-Host ("Duration: {0:N1}s" -f $duration) -ForegroundColor White
Write-Host ""

# Generate coverage report if requested
if ($Coverage) {
    Write-Host "📊 Coverage Reports:" -ForegroundColor Yellow
    Write-Host ""
    
    if (Test-Path "coverage") {
        Write-Host "Backend coverage reports:" -ForegroundColor Cyan
        if (Test-Path "coverage\unit.out") {
            go tool cover -func=coverage\unit.out | Select-Object -Last 1
        }
        
        if (Test-Path "coverage\integration.out") {
            go tool cover -func=coverage\integration.out | Select-Object -Last 1
        }
        
        Write-Host "`nView detailed report: coverage\coverage.html" -ForegroundColor White
    }
    
    if (Test-Path "frontend\coverage") {
        Write-Host ""
        Write-Host "Frontend coverage reports:" -ForegroundColor Cyan
        Write-Host "View detailed report: frontend\coverage\index.html" -ForegroundColor White
    }
    
    Write-Host ""
}

# Exit with appropriate code
if ($failedTests -eq 0) {
    Write-Host "╔════════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║     All Tests Passed! ✅                   ║" -ForegroundColor Green
    Write-Host "╚════════════════════════════════════════════╝" -ForegroundColor Green
    exit 0
} else {
    Write-Host "╔════════════════════════════════════════════╗" -ForegroundColor Red
    Write-Host "║     Some Tests Failed ❌                   ║" -ForegroundColor Red
    Write-Host "╚════════════════════════════════════════════╝" -ForegroundColor Red
    exit 1
}
