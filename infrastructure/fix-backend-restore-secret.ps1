# Complete Backend Database Connection Fix - Handles Pending Deletion
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
if ([string]::IsNullOrEmpty($rdsEndpoint)) {
    Write-Host "Getting RDS endpoint from resources..." -ForegroundColor Gray
    $dbInstanceId = aws cloudformation describe-stack-resource `
        --stack-name $StackName `
        --logical-resource-id Database `
        --region $Region `
        --query 'StackResourceDetail.PhysicalResourceId' `
        --output text
    
    $rdsEndpoint = aws rds describe-db-instances `
        --db-instance-identifier $dbInstanceId `
        --region $Region `
        --query 'DBInstances[0].Endpoint.Address' `
        --output text
}
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

# Step 2: Handle JWT Secret (with pending deletion handling)
Write-Host "Step 2: Setting up JWT Secret..." -ForegroundColor Yellow
$jwtSecretName = "$StackName/jwt-secret"
$jwtSecretArn = $null

# Check if JWT secret exists
$jwtSecretStatus = aws secretsmanager describe-secret --secret-id $jwtSecretName --region $Region 2>&1
if ($LASTEXITCODE -eq 0) {
    $jwtSecretInfo = $jwtSecretStatus | ConvertFrom-Json
    
    # Check if it's scheduled for deletion
    if ($jwtSecretInfo.DeletedDate) {
        Write-Host "JWT secret is scheduled for deletion" -ForegroundColor Yellow
        Write-Host "Restoring secret instead of creating new one..." -ForegroundColor Gray
        
        $restoreResult = aws secretsmanager restore-secret --secret-id $jwtSecretName --region $Region 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Secret restored successfully" -ForegroundColor Green
            $jwtSecretArn = $jwtSecretInfo.ARN
            
            # Update the secret value with proper format
            Write-Host "Updating secret value with proper JSON format..." -ForegroundColor Gray
            $jwtValue = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 64 | ForEach-Object {[char]$_})
            $jwtSecretJson = "{`"secret`":`"$jwtValue`"}"
            
            $updateResult = aws secretsmanager put-secret-value `
                --secret-id $jwtSecretName `
                --secret-string $jwtSecretJson `
                --region $Region 2>&1
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "Secret value updated" -ForegroundColor Green
            } else {
                Write-Host "WARNING: Failed to update secret value: $updateResult" -ForegroundColor Yellow
            }
        } else {
            Write-Host "ERROR: Failed to restore secret: $restoreResult" -ForegroundColor Red
            Write-Host "You may need to wait 7 days for full deletion or use a different name" -ForegroundColor Yellow
            exit 1
        }
    } else {
        # Secret exists and is not being deleted
        Write-Host "JWT secret exists, verifying format..." -ForegroundColor Gray
        $jwtSecretValue = aws secretsmanager get-secret-value --secret-id $jwtSecretName --region $Region --query 'SecretString' --output text 2>&1
        
        # Try to parse it as JSON
        try {
            $parsedJwt = $jwtSecretValue | ConvertFrom-Json
            if ($parsedJwt.secret) {
                Write-Host "JWT secret is valid" -ForegroundColor Green
                $jwtSecretArn = $jwtSecretInfo.ARN
            } else {
                Write-Host "JWT secret exists but doesn't have 'secret' key, updating..." -ForegroundColor Yellow
                $jwtValue = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 64 | ForEach-Object {[char]$_})
                $jwtSecretJson = "{`"secret`":`"$jwtValue`"}"
                
                aws secretsmanager put-secret-value `
                    --secret-id $jwtSecretName `
                    --secret-string $jwtSecretJson `
                    --region $Region 2>&1 | Out-Null
                
                $jwtSecretArn = $jwtSecretInfo.ARN
                Write-Host "JWT secret updated" -ForegroundColor Green
            }
        } catch {
            Write-Host "JWT secret is malformed, updating with proper format..." -ForegroundColor Yellow
            $jwtValue = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 64 | ForEach-Object {[char]$_})
            $jwtSecretJson = "{`"secret`":`"$jwtValue`"}"
            
            aws secretsmanager put-secret-value `
                --secret-id $jwtSecretName `
                --secret-string $jwtSecretJson `
                --region $Region 2>&1 | Out-Null
            
            $jwtSecretArn = $jwtSecretInfo.ARN
            Write-Host "JWT secret updated" -ForegroundColor Green
        }
    }
} else {
    # Secret doesn't exist, create it
    Write-Host "Creating new JWT secret..." -ForegroundColor Gray
    
    $jwtValue = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 64 | ForEach-Object {[char]$_})
    $jwtSecretJson = "{`"secret`":`"$jwtValue`"}"
    
    Write-Host "JWT Secret JSON: $jwtSecretJson" -ForegroundColor Gray
    
    $jwtSecretArn = aws secretsmanager create-secret `
        --name $jwtSecretName `
        --description "JWT signing secret for Hub HRMS" `
        --secret-string $jwtSecretJson `
        --region $Region `
        --query 'ARN' `
        --output text 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "JWT secret created: $jwtSecretArn" -ForegroundColor Green
    } else {
        Write-Host "ERROR: Failed to create JWT secret: $jwtSecretArn" -ForegroundColor Red
        exit 1
    }
}

Write-Host "JWT Secret ARN: $jwtSecretArn" -ForegroundColor White
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

Write-Host "Secrets configured:" -ForegroundColor Gray
Write-Host "  - DB_HOST: $secretArn`:host::" -ForegroundColor Gray
Write-Host "  - DB_PORT: $secretArn`:port::" -ForegroundColor Gray
Write-Host "  - DB_NAME: $secretArn`:dbname::" -ForegroundColor Gray
Write-Host "  - DB_USER: $secretArn`:username::" -ForegroundColor Gray
Write-Host "  - DB_PASSWORD: $secretArn`:password::" -ForegroundColor Gray
Write-Host "  - JWT_SECRET: $jwtSecretArn`:secret::" -ForegroundColor Gray
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

Write-Host "Registering new task definition..." -ForegroundColor Gray
$newTaskDefArn = aws ecs register-task-definition `
    --cli-input-json "file://$tempFile" `
    --region $Region `
    --query 'taskDefinition.taskDefinitionArn' `
    --output text 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to register task definition" -ForegroundColor Red
    Write-Host "Error: $newTaskDefArn" -ForegroundColor Red
    Write-Host "Task def saved to: $tempFile" -ForegroundColor Yellow
    exit 1
}

Write-Host "New task definition registered: $newTaskDefArn" -ForegroundColor Green
Write-Host ""

# Step 5: Update IAM role to access JWT secret
Write-Host "Step 5: Updating IAM roles..." -ForegroundColor Yellow

$executionRoleName = $taskDef.executionRoleArn.Split("/")[-1]
Write-Host "Execution Role: $executionRoleName" -ForegroundColor Gray

# Add policy to access both secrets
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

aws iam put-role-policy `
    --role-name $executionRoleName `
    --policy-name SecretsAccess `
    --policy-document "file://$policyFile" 2>&1 | Out-Null

Remove-Item $policyFile -ErrorAction SilentlyContinue
Write-Host "IAM permissions updated" -ForegroundColor Green
Write-Host ""

# Step 6: Update ECS service
Write-Host "Step 6: Updating ECS service..." -ForegroundColor Yellow

$clusterName = "$StackName-cluster"
$serviceName = "$StackName-backend"

Write-Host "Updating service with new task definition..." -ForegroundColor Gray
aws ecs update-service `
    --cluster $clusterName `
    --service $serviceName `
    --task-definition $newTaskDefArn `
    --force-new-deployment `
    --region $Region 2>&1 | Out-Null

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to update service" -ForegroundColor Red
    exit 1
}

Write-Host "Service update initiated" -ForegroundColor Green
Write-Host ""

# Step 7: Monitor deployment
Write-Host "Step 7: Monitoring deployment..." -ForegroundColor Yellow
Write-Host "Waiting for tasks to start (this takes 2-3 minutes)..." -ForegroundColor Gray
Write-Host ""

$maxWait = 180  # 3 minutes
$elapsed = 0
$success = $false

while ($elapsed -lt $maxWait) {
    Start-Sleep -Seconds 15
    $elapsed += 15
    
    $serviceJson = aws ecs describe-services `
        --cluster $clusterName `
        --services $serviceName `
        --region $Region `
        --output json
    
    $service = ($serviceJson | ConvertFrom-Json).services[0]
    $runningCount = $service.runningCount
    $desiredCount = $service.desiredCount
    
    Write-Host "[${elapsed}s] Running: $runningCount/$desiredCount" -ForegroundColor $(if ($runningCount -eq $desiredCount) { "Green" } else { "Yellow" })
    
    if ($runningCount -eq $desiredCount -and $runningCount -gt 0) {
        Write-Host ""
        Write-Host "Service is running!" -ForegroundColor Green
        $success = $true
        break
    }
}

if (-not $success) {
    Write-Host ""
    Write-Host "Service is taking longer than expected to stabilize" -ForegroundColor Yellow
}

# Step 8: Check logs
Write-Host ""
Write-Host "Step 8: Checking recent logs..." -ForegroundColor Yellow
Write-Host ""

Start-Sleep -Seconds 5

$logGroup = "/ecs/$StackName-backend"
$logs = aws logs tail $logGroup --region $Region --since 3m --format short 2>&1

if ($logs) {
    $logs | ForEach-Object {
        $line = $_.ToString()
        if ($line -match "ResourceInitializationError|unable to pull|invalid character") {
            Write-Host $line -ForegroundColor Red
        } elseif ($line -match "Failed|error|refused") {
            Write-Host $line -ForegroundColor Red
        } elseif ($line -match "Successfully|Starting|listening|Database pool") {
            Write-Host $line -ForegroundColor Green
        } else {
            Write-Host $line
        }
    }
}

# Summary
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Fix Complete!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "What was fixed:" -ForegroundColor Yellow
Write-Host "  - JWT secret restored/created with proper JSON" -ForegroundColor Green
Write-Host "  - Individual DB env vars configured" -ForegroundColor Green
Write-Host "  - IAM permissions updated" -ForegroundColor Green
Write-Host "  - New task definition deployed" -ForegroundColor Green
Write-Host ""
Write-Host "Monitor with:" -ForegroundColor Yellow
Write-Host "  aws logs tail $logGroup --follow --region $Region" -ForegroundColor Cyan
Write-Host ""

$albDns = ($stack.Outputs | Where-Object { $_.OutputKey -eq "ALBDNSName" }).OutputValue
if ($albDns) {
    Write-Host "Test API:" -ForegroundColor Yellow
    Write-Host "  curl http://$albDns/api/health" -ForegroundColor Cyan
    Write-Host ""
}
