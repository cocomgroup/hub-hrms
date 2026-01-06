# test-frontend.ps1
# Frontend Test Runner for hub-hrms
# Runs all frontend tests with coverage and watch mode

param(
    [switch]$Coverage,
    [switch]$Watch,
    [switch]$UI,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

# Show help
if ($Help) {
    Write-Host "Frontend Test Runner for hub-hrms" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\test-frontend.ps1 [options]" -ForegroundColor White
    Write-Host ""
    Write-Host "Options:" -ForegroundColor Yellow
    Write-Host "  -Coverage    Generate coverage reports" -ForegroundColor White
    Write-Host "  -Watch       Run tests in watch mode" -ForegroundColor White
    Write-Host "  -UI          Start test UI" -ForegroundColor White
    Write-Host "  -Help        Show this help message" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\test-frontend.ps1" -ForegroundColor White
    Write-Host "  .\test-frontend.ps1 -Coverage" -ForegroundColor White
    Write-Host "  .\test-frontend.ps1 -Watch" -ForegroundColor White
    exit 0
}

Write-Host "🧪 Running frontend test suite..." -ForegroundColor Blue

# Check if frontend directory exists
if (-not (Test-Path "frontend")) {
    Write-Host "Error: frontend directory not found" -ForegroundColor Red
    exit 1
}

Push-Location frontend

try {
    if ($Watch) {
        Write-Host "`n👀 Running tests in watch mode..." -ForegroundColor Yellow
        npm run test:watch
    } elseif ($UI) {
        Write-Host "`n🎨 Starting test UI..." -ForegroundColor Yellow
        npm run test:ui
    } elseif ($Coverage) {
        Write-Host "`n📊 Running tests with coverage..." -ForegroundColor Yellow
        npm run test:coverage
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "`n✅ Coverage report generated: frontend\coverage\index.html" -ForegroundColor Green
        }
    } else {
        Write-Host "`n📝 Running all tests..." -ForegroundColor Yellow
        npm run test
    }
    
    if ($LASTEXITCODE -ne 0) {
        throw "Frontend tests failed"
    }
    
    Write-Host "`n✅ Frontend tests complete!" -ForegroundColor Green
    
} catch {
    Write-Host "`n❌ Frontend tests failed" -ForegroundColor Red
    Pop-Location
    exit 1
} finally {
    Pop-Location
}
