# Generate a cryptographically secure 64-character API key
Add-Type -AssemblyName System.Web
$apiKey = [System.Web.Security.Membership]::GeneratePassword(64, 20)

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "Generated API Key:" -ForegroundColor Green
Write-Host $apiKey -ForegroundColor Yellow
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "SAVE THIS KEY - It cannot be retrieved later!" -ForegroundColor Red

# Save to file
$apiKey | Out-File -FilePath hub-hrms-api-key.txt -Encoding utf8
Write-Host "Also saved to: hub-hrms-api-key.txt" -ForegroundColor Green