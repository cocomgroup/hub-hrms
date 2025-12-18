# ============================================
# SIMPLE FRONTEND DIAGNOSTIC
# Copy and paste these commands ONE AT A TIME
# ============================================

# STEP 1: Find log group name
Write-Host "`n=== STEP 1: Find Log Group ===" -ForegroundColor Cyan
aws logs describe-log-groups --region us-east-1 --query "logGroups[?contains(logGroupName, 'frontend')].logGroupName" --output table

# After you see the log group name above, use it in the next command
# Replace LOG_GROUP_NAME_HERE with the actual name

Write-Host "`n=== STEP 2: View Logs (REPLACE LOG_GROUP_NAME) ===" -ForegroundColor Cyan
Write-Host "aws logs tail /ecs/hub-hrms-frontend --region us-east-1 --since 10m" -ForegroundColor Yellow
Write-Host "OR" -ForegroundColor Gray
Write-Host "aws logs tail /aws/ecs/hub-hrms-frontend --region us-east-1 --since 10m" -ForegroundColor Yellow

# STEP 3: Check target health
Write-Host "`n=== STEP 3: Check Target Health ===" -ForegroundColor Cyan
$tgJson = aws elbv2 describe-target-groups --region us-east-1 --output json
$tgData = $tgJson | ConvertFrom-Json
$frontendTg = $tgData.TargetGroups | Where-Object { $_.TargetGroupName -like "*frontend*" }

if ($frontendTg) {
    Write-Host "Target Group Found: $($frontendTg.TargetGroupName)" -ForegroundColor Green
    
    $healthJson = aws elbv2 describe-target-health --target-group-arn $frontendTg.TargetGroupArn --region us-east-1 --output json
    $healthData = $healthJson | ConvertFrom-Json
    
    foreach ($t in $healthData.TargetHealthDescriptions) {
        $ip = $t.Target.Id
        $port = $t.Target.Port
        $state = $t.TargetHealth.State
        
        if ($state -eq "healthy") {
            Write-Host "  HEALTHY: $ip`:$port" -ForegroundColor Green
        } else {
            Write-Host "  UNHEALTHY: $ip`:$port" -ForegroundColor Red
            if ($t.TargetHealth.Reason) {
                Write-Host "    Reason: $($t.TargetHealth.Reason)" -ForegroundColor Yellow
            }
        }
    }
} else {
    Write-Host "No frontend target group found" -ForegroundColor Red
}

# STEP 4: Get frontend URL
Write-Host "`n=== STEP 4: Frontend URL ===" -ForegroundColor Cyan
$albJson = aws elbv2 describe-load-balancers --region us-east-1 --output json
$albData = $albJson | ConvertFrom-Json
$alb = $albData.LoadBalancers | Select-Object -First 1

if ($alb) {
    Write-Host "http://$($alb.DNSName)" -ForegroundColor Cyan
}

Write-Host "`n=== DONE ===" -ForegroundColor Green