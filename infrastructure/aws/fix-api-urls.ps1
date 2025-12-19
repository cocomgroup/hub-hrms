# Fix All Double /api URLs in Frontend
# Run this from your project root directory

$files = @(
    "frontend/src/routes/Login.svelte",
    "frontend/src/routes/Employees.svelte",
    "frontend/src/routes/Workflows.svelte",
    "frontend/src/routes/WorkflowDetail.svelte"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Fixing Double /api URLs in Frontend" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$fixedCount = 0
$notFoundCount = 0

foreach ($file in $files) {
    if (Test-Path $file) {
        Write-Host "Processing: $file" -ForegroundColor Yellow
        
        $content = Get-Content $file -Raw
        $originalContent = $content
        
        # Replace ${import.meta.env.VITE_API_URL}/api/ with /api/
        $content = $content -replace '\$\{import\.meta\.env\.VITE_API_URL\}/api/', '/api/'
        
        if ($content -ne $originalContent) {
            Set-Content $file $content -NoNewline
            Write-Host "  Fixed!" -ForegroundColor Green
            $fixedCount++
        } else {
            Write-Host "  No changes needed" -ForegroundColor Gray
        }
    } else {
        Write-Host "Not found: $file" -ForegroundColor Red
        $notFoundCount++
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Summary" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "Files fixed: $fixedCount" -ForegroundColor Green
Write-Host "Files not found: $notFoundCount" -ForegroundColor $(if ($notFoundCount -gt 0) { "Yellow" } else { "Green" })
Write-Host ""

if ($fixedCount -gt 0) {
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "  1. cd frontend" -ForegroundColor Yellow
    Write-Host "  2. npm run build" -ForegroundColor Yellow
    Write-Host "  3. cd .." -ForegroundColor Yellow
    Write-Host "  4. .\deploy-frontend.ps1" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Or run this to rebuild and deploy automatically:" -ForegroundColor Cyan
    Write-Host "  cd frontend && npm run build && cd .. && .\deploy-frontend.ps1" -ForegroundColor Yellow
}