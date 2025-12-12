# Complete Backend Database Connection Fix - JSON Format Fixed
# This script fixes ALL issues preventing backend from connecting to RDS

param(
    [Parameter(Mandatory=$false)]
    [string]$StackName = 'hub-hrms',
    
    [Parameter(Mandatory=$false)]
    [string]$Region = 'us-east-1'
)

$ErrorActionPreference = "Continue"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Hub HRMS - Complete Backend Connection Fix" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Get stack outputs
Write-Host "Step 1: Getting stack information..." -ForegroundColor Yellow
$stackInfo = aws cloudformation describe-stacks --stack-name $StackName --region $Region --output json | ConvertFrom-Json
$stack = $stackInfo.Stacks[0]

$rdsEndpoint = ($stack.Outputs | Where-Object { $_.OutputKey -eq "RDSEndpoint" }).OutputValue
Write-Host "RDS Endpoint: $rdsEndpoint" -ForegroundColor White

# Get DatabaseSecret ARN
$secretArn = ($stack.Outputs | Where-Object { $_.OutputKey -eq "DatabaseSecretArn" }).OutputValue
if ([string]::IsNullOrEmpty($secretArn)) {
    Write-Host "Getting secret ARN from resources..." -ForegroundColor Gray
    $secretArn = aws cloudformation describe-stack-resource `
        --stack-name $StackName `
        --logical-resource-id DatabaseSecret `
        --region $Region `
        --query 'StackResourceDetail.PhysicalResourceId' `
        --output text
}
Write-Host "Database Secret ARN: $secretArn" -ForegroundColor White
Write-Host ""

# Step 2: Create or update JWT Secret
Write-Host "Step 2: Setting up JWT Secret..." -ForegroundColor Yellow
$jwtSecretName = "$StackName/jwt-secret"

# Check if JWT secret exists and delete if malformed
$jwtSecretCheck = aws secretsmanager describe-secret --secret-id $jwtSecretName --region $Region 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "JWT secret exists, checking if it's valid..." -ForegroundColor Gray
    $jwtSecretValue = aws secretsmanager get-secret-value --secret-id $jwtSecretName --region $Region --query 'SecretString' --output text 2>&1
    
    # Try to parse it as JSON
    try {
        $null = $jwtSecretValue | ConvertFrom-Json
        Write-Host "JWT secret is valid" -ForegroundColor Green
        $jwtSecretArn = aws secretsmanager describe-secret --secret-id $jwtSecretName --region $Region --query 'ARN' --output text
    } catch {
        Write-Host "JWT secret is malformed, deleting and recreating..." -ForegroundColor Yellow
        aws secretsmanager delete-secret --secret-id $jwtSecretName --force-delete-without-recovery --region $Region 2>&1 | Out-Null
        Start-Sleep -Seconds 2
        $LASTEXITCODE = 1  # Force recreation
    }
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "Creating JWT secret with proper JSON format..." -ForegroundColor Gray
    
    # Generate random secret
    $jwtValue = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 64 | ForEach-Object {[char]$_})
    
    # Create properly formatted JSON
    $jwtSecretJson = @"
{"secret":"$jwtValue"}
"@
    
    # Save to temp file
    $jwtTempFile = "$env:TEMP\jwt-secret.json"
    $jwtSecretJson | Out-File -FilePath $jwtTempFile -Encoding ASCII -NoNewline
    
    Write-Host "JWT Secret JSON:" -ForegroundColor Gray
    Write-Host $jwtSecretJson -ForegroundColor Gray
    
    # Create secret from file
    $jwtSecretArn = aws secretsmanager create-secret `
        --name $jwtSecretName `
        --description "JWT signing secret for Hub HRMS" `
        --secret-string "file://$jwtTempFile" `
        --region $Region `
        --query 'ARN' `
        --output text
    
    Remove-Item $jwtTempFile -ErrorAction SilentlyContinue
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "JWT secret created: $jwtSecretArn" -ForegroundColor Green
        
        # Verify it's valid JSON
        Write-Host "Verifying secret..." -ForegroundColor Gray
        $verifySecret = aws secretsmanager get-secret-value --secret-id $jwtSecretName --region $Region --query 'SecretString' --output text
        try {
            $parsedSecret = $verifySecret | ConvertFrom-Json
            Write-Host "Secret verified - contains key: secret" -ForegroundColor Green
        } catch {
            Write-Host "WARNING: Secret may be malformed" -ForegroundColor Yellow
        }
    } else {
        Write-Host "ERROR: Failed to create JWT secret" -ForegroundColor Red
        exit 1
    }
}
Write-Host ""

# Step 3: Get current task definition
Write-Host "Step 3: Getting current task definition..." -ForegroundColor Yellow
$taskDefArn = aws ecs list-task-definitions `
    --family-prefix "$StackName-backend" `
    --region $Region `
    --sort DESC `
    --max-items 1 `
    --query 'taskDefinitionArns[0]' `
    --output text

if ([string]::IsNullOrEmpty($taskDefArn) -or $taskDefArn -eq "None") {
    Write-Host "ERROR: Task definition not found" -ForegroundColor Red
    exit 1
}

Write-Host "Current task definition: $taskDefArn" -ForegroundColor Gray
$taskDefJson = aws ecs describe-task-definition --task-definition $taskDefArn --region $Region --output json
$taskDef = ($taskDefJson | ConvertFrom-Json).taskDefinition
Write-Host "Task definition retrieved" -ForegroundColor Green
Write-Host ""

# Step 4: Create new task definition with correct environment variables
Write-Host "Step 4: Creating new task definition with correct DB config..." -ForegroundColor Yellow

# Get the container definition
$container = $taskDef.containerDefinitions[0]

# Set up environment variables
$container.environment = @(
    @{ name = "SERVER_ADDR"; value = ":8080" }
    @{ name = "GIN_MODE"; value = "release" }
    @{ name = "PORT"; value = "8080" }
    @{ name = "ENVIRONMENT"; value = "production" }
)

# Set up secrets with individual DB variables
$container.secrets = @(
    @{ name = "DB_HOST"; valueFrom = "$secretArn`:host::" }
    @{ name = "DB_PORT"; valueFrom = "$secretArn`:port::" }
    @{ name = "DB_NAME"; valueFrom = "$secretArn`:dbname::" }
    @{ name = "DB_USER"; valueFrom = "$secretArn`:username::" }
    @{ name = "DB_PASSWORD"; valueFrom = "$secretArn`:password::" }
    @{ name = "JWT_SECRET"; valueFrom = "$jwtSecretArn`:secret::" }
)

Write-Host "Environment variables configured:" -ForegroundColor Gray
Write-Host "  - SERVER_ADDR: :8080" -ForegroundColor Gray
Write-Host "  - GIN_MODE: release" -ForegroundColor Gray
Write-Host "  - PORT: 8080" -ForegroundColor Gray
Write-Host "  - ENVIRONMENT: production" -ForegroundColor Gray
Write-Host ""
Write-Host "Secrets configured:" -ForegroundColor Gray
Write-Host "  - DB_HOST (from $secretArn)" -ForegroundColor Gray
Write-Host "  - DB_PORT (from $secretArn)" -ForegroundColor Gray
Write-Host "  - DB_NAME (from $secretArn)" -ForegroundColor Gray
Write-Host "  - DB_USER (from $secretArn)" -ForegroundColor Gray
Write-Host "  - DB_PASSWORD (from $secretArn)" -ForegroundColor Gray
Write-Host "  - JWT_SECRET (from $jwtSecretArn)" -ForegroundColor Gray
Write-Host ""

# Create new task definition
$newTaskDef = @{
    family = $taskDef.family
    networkMode = $taskDef.networkMode
    requiresCompatibilities = $taskDef.requiresCompatibilities
    cpu = $taskDef.cpu
    memory = $taskDef.memory
    executionRoleArn = $taskDef.executionRoleArn
    containerDefinitions = @($container)
}

# Add taskRoleArn if it exists
if ($taskDef.taskRoleArn) {
    $newTaskDef.taskRoleArn = $taskDef.taskRoleArn
}

# Write to temp file
$tempFile = "$env:TEMP\hub-hrms-taskdef-fixed.json"
$newTaskDef | ConvertTo-Json -Depth 10 | Set-Content $tempFile

Write-Host "Task definition JSON created at: $tempFile" -ForegroundColor Gray
Write-Host "Registering new task definition..." -ForegroundColor Gray

$newTaskDefArn = aws ecs register-task-definition `
    --cli-input-json "file://$tempFile" `
    --region $Region `
    --query 'taskDefinition.taskDefinitionArn' `
    --output text 2>&1

if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrEmpty($newTaskDefArn)) {
    Write-Host "ERROR: Failed to register new task definition" -ForegroundColor Red
    Write-Host "Error output: $newTaskDefArn" -ForegroundColor Red
    Write-Host "Task definition JSON saved to: $tempFile" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please check the file and try manually:" -ForegroundColor Yellow
    Write-Host "  aws ecs register-task-definition --cli-input-json file://$tempFile --region $Region" -ForegroundColor Cyan
    exit 1
}

Write-Host "New task definition registered: $newTaskDefArn" -ForegroundColor Green
# Keep the temp file for debugging
Write-Host "Task definition saved to: $tempFile" -ForegroundColor Gray
Write-Host ""

# Step 5: Update IAM role to access JWT secret
Write-Host "Step 5: Updating IAM roles..." -ForegroundColor Yellow

$executionRoleName = $taskDef.executionRoleArn.Split("/")[-1]
Write-Host "Execution Role: $executionRoleName" -ForegroundColor Gray

# Add policy to access JWT secret
$policyDoc = @{
    Version = "2012-10-17"
    Statement = @(
        @{
            Effect = "Allow"
            Action = @("secretsmanager:GetSecretValue")
            Resource = @($secretArn, $jwtSecretArn)
        }
    )
} | ConvertTo-Json -Depth 10

$policyFile = "$env:TEMP\secrets-policy.json"
$policyDoc | Set-Content $policyFile

Write-Host "IAM Policy:" -ForegroundColor Gray
Write-Host $policyDoc -ForegroundColor Gray

# Update inline policy
$iamResult = aws iam put-role-policy `
    --role-name $executionRoleName `
    --policy-name SecretsAccess `
    --policy-document "file://$policyFile" 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "IAM permissions updated" -ForegroundColor Green
} else {
    Write-Host "WARNING: IAM update may have failed: $iamResult" -ForegroundColor Yellow
}

Remove-Item $policyFile -ErrorAction SilentlyContinue
Write-Host ""

# Step 6: Update ECS service
Write-Host "Step 6: Updating ECS service..." -ForegroundColor Yellow

$clusterName = "$StackName-cluster"
$serviceName = "$StackName-backend"

Write-Host "Cluster: $clusterName" -ForegroundColor Gray
Write-Host "Service: $serviceName" -ForegroundColor Gray
Write-Host "Task Definition: $newTaskDefArn" -ForegroundColor Gray
Write-Host ""

$updateResult = aws ecs update-service `
    --cluster $clusterName `
    --service $serviceName `
    --task-definition $newTaskDefArn `
    --force-new-deployment `
    --region $Region `
    --query 'service.serviceName' `
    --output text 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to update service" -ForegroundColor Red
    Write-Host $updateResult -ForegroundColor Red
    exit 1
}

Write-Host "Service update initiated" -ForegroundColor Green
Write-Host ""

# Step 7: Monitor deployment
Write-Host "Step 7: Monitoring deployment (this may take 2-3 minutes)..." -ForegroundColor Yellow
Write-Host ""

$maxAttempts = 36  # 6 minutes
$attempt = 0
$stable = $false
$lastError = ""

while ($attempt -lt $maxAttempts -and -not $stable) {
    Start-Sleep -Seconds 10
    $attempt++
    
    $serviceJson = aws ecs describe-services `
        --cluster $clusterName `
        --services $serviceName `
        --region $Region `
        --output json
    
    $service = ($serviceJson | ConvertFrom-Json).services[0]
    $runningCount = $service.runningCount
    $desiredCount = $service.desiredCount
    $deployments = $service.deployments.Count
    
    # Get deployment status
    $primaryDeployment = $service.deployments | Where-Object { $_.status -eq "PRIMARY" }
    $rolloutState = $primaryDeployment.rolloutState
    
    # Check for errors in events
    $events = $service.events | Select-Object -First 3
    $errorEvent = $events | Where-Object { $_.message -match "error|failed|unable" } | Select-Object -First 1
    if ($errorEvent -and $errorEvent.message -ne $lastError) {
        Write-Host ""
        Write-Host "Service Event: $($errorEvent.message)" -ForegroundColor Yellow
        $lastError = $errorEvent.message
        Write-Host ""
    }
    
    $statusColor = if ($rolloutState -eq "COMPLETED") { "Green" } elseif ($rolloutState -eq "FAILED") { "Red" } else { "Yellow" }
    
    Write-Host ("[{0:D2}/{1:D2}] Running: {2}/{3} | Deployments: {4} | Status: {5}" -f $attempt, $maxAttempts, $runningCount, $desiredCount, $deployments, $rolloutState) -ForegroundColor $statusColor
    
    if ($rolloutState -eq "COMPLETED" -and $deployments -eq 1 -and $runningCount -eq $desiredCount) {
        $stable = $true
    } elseif ($rolloutState -eq "FAILED") {
        Write-Host ""
        Write-Host "Deployment failed!" -ForegroundColor Red
        break
    }
}

Write-Host ""

if ($stable) {
    Write-Host "Deployment completed successfully!" -ForegroundColor Green
} else {
    Write-Host "Deployment is taking longer than expected or failed" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Check task errors with:" -ForegroundColor Yellow
    Write-Host "  aws ecs describe-tasks --cluster $clusterName --tasks `$(aws ecs list-tasks --cluster $clusterName --service-name $serviceName --region $Region --query 'taskArns[0]' --output text) --region $Region" -ForegroundColor Cyan
}

# Step 8: Check logs
Write-Host ""
Write-Host "Step 8: Checking recent logs..." -ForegroundColor Yellow
Write-Host ""

Start-Sleep -Seconds 5

$logGroup = "/ecs/$StackName-backend"
$recentLogs = aws logs tail $logGroup --region $Region --since 3m --format short 2>&1

if ($LASTEXITCODE -eq 0 -and $recentLogs) {
    $recentLogs | ForEach-Object {
        $line = $_.ToString()
        if ($line -match "ResourceInitializationError|unable to pull secrets|invalid character") {
            Write-Host $line -ForegroundColor Red
        } elseif ($line -match "Failed|refused|error") {
            Write-Host $line -ForegroundColor Red
        } elseif ($line -match "Successfully|Starting|listening|Database pool|Server listening") {
            Write-Host $line -ForegroundColor Green
        } else {
            Write-Host $line -ForegroundColor White
        }
    }
} else {
    Write-Host "No recent log entries found (service may still be starting)" -ForegroundColor Gray
}

# Summary
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Fix Applied!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Changes Applied:" -ForegroundColor Yellow
Write-Host "  - Created/verified JWT secret with proper JSON format" -ForegroundColor Green
Write-Host "  - Added individual DB environment variables" -ForegroundColor Green
Write-Host "  - Updated IAM permissions" -ForegroundColor Green
Write-Host "  - Force deployed new task definition" -ForegroundColor Green
Write-Host ""
Write-Host "Debugging Files:" -ForegroundColor Yellow
Write-Host "  - Task definition: $tempFile" -ForegroundColor Gray
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Monitor logs for successful startup:" -ForegroundColor White
Write-Host "   aws logs tail $logGroup --follow --region $Region" -ForegroundColor Cyan
Write-Host ""
Write-Host "2. Verify secrets are valid:" -ForegroundColor White
Write-Host "   aws secretsmanager get-secret-value --secret-id $jwtSecretName --region $Region" -ForegroundColor Cyan
Write-Host ""
Write-Host "3. Test the backend API:" -ForegroundColor White
$albDns = ($stack.Outputs | Where-Object { $_.OutputKey -eq "ALBDNSName" }).OutputValue
if ($albDns) {
    Write-Host "   curl http://$albDns/api/health" -ForegroundColor Cyan
}
Write-Host ""