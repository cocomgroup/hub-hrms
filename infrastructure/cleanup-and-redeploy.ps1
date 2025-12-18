# Quick Cleanup and Redeploy Script
# This deletes the failed stack and redeploys

param(
    [Parameter(Mandatory=$false)]
    [string]$StackName = "hub-hrms-stack",
    
    [Parameter(Mandatory=$false)]
    [string]$Region = "us-east-1"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Cleaning up failed deployment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Delete the stack
Write-Host "1. Deleting stack: $StackName" -ForegroundColor Yellow
aws cloudformation delete-stack --stack-name $StackName --region $Region

if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to initiate stack deletion" -ForegroundColor Red
    exit 1
}

Write-Host "   Stack deletion initiated" -ForegroundColor Green
Write-Host ""

# Step 2: Wait for deletion
Write-Host "2. Waiting for stack deletion to complete..." -ForegroundColor Yellow
Write-Host "   This may take 5-10 minutes" -ForegroundColor Gray

$maxWaitTime = 600  # 10 minutes
$elapsed = 0
$interval = 15

while ($elapsed -lt $maxWaitTime) {
    Start-Sleep -Seconds $interval
    $elapsed += $interval
    
    $status = aws cloudformation describe-stacks `
        --stack-name $StackName `
        --region $Region `
        --query 'Stacks[0].StackStatus' `
        --output text 2>&1
    
    if ($status -match "does not exist" -or $LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "   ✓ Stack deleted successfully!" -ForegroundColor Green
        break
    }
    
    Write-Host "   Current status: $status (${elapsed}s elapsed)" -ForegroundColor Gray
    
    if ($status -match "DELETE_FAILED") {
        Write-Host ""
        Write-Host "   ✗ Stack deletion failed!" -ForegroundColor Red
        Write-Host "   You may need to manually delete resources" -ForegroundColor Yellow
        exit 1
    }
}

if ($elapsed -ge $maxWaitTime) {
    Write-Host ""
    Write-Host "   ⚠ Timeout waiting for stack deletion" -ForegroundColor Yellow
    Write-Host "   Check AWS console for status" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Cleanup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Run your deploy script:" -ForegroundColor White
Write-Host "   .\deploy-aws.ps1 -Environment dev -Region $Region ..." -ForegroundColor Gray
Write-Host ""
Write-Host "2. Or delete the JWT secret manually:" -ForegroundColor White
Write-Host "   aws secretsmanager delete-secret --secret-id hub-hrms/jwt-secret --force-delete-without-recovery" -ForegroundColor Gray
Write-Host ""