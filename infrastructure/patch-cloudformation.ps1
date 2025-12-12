# Hub HRMS Backend Fix - With Diagnostics
# Checks what exists and fixes what's needed

param(
    [Parameter(Mandatory=$false)]
    [string]$StackName = 'hub-hrms',
    
    [Parameter(Mandatory=$false)]
    [string]$Region = 'us-east-1'
)

$ErrorActionPreference = "Continue"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Hub HRMS - Diagnostics & Fix" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Get stack info
Write-Host "Checking CloudFormation stack..." -ForegroundColor Yellow
$stackInfo = aws cloudformation describe-stacks --stack-name $StackName --region $Region --output json 2>&1 | ConvertFrom-Json
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Stack not found: $StackName" -ForegroundColor Red
    exit 1
}

$stack = $stackInfo.Stacks[0]
Write-Host "Stack Status: $($stack.StackStatus)" -ForegroundColor $(if ($stack.StackStatus -like "*COMPLETE") { "Green" } else { "Yellow" })
Write-Host ""

# Check ECS cluster
Write-Host "Checking ECS cluster..." -ForegroundColor Yellow
$clusterName = "$StackName-cluster"
$clusterInfo = aws ecs describe-clusters --clusters $clusterName --region $Region --output json 2>&1 | ConvertFrom-Json

if ($clusterInfo.clusters.Count -gt 0) {
    $cluster = $clusterInfo.clusters[0]
    Write-Host "Cluster: $($cluster.clusterName)" -ForegroundColor Green
    Write-Host "  Running Tasks: $($cluster.runningTasksCount)" -ForegroundColor White
    Write-Host "  Pending Tasks: $($cluster.pendingTasksCount)" -ForegroundColor White
} else {
    Write-Host "ERROR: Cluster not found: $clusterName" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Check ECS service
Write-Host "Checking ECS service..." -ForegroundColor Yellow
$serviceName = "$StackName-backend"
$serviceInfo = aws ecs describe-services --cluster $clusterName --services $serviceName --region $Region --output json 2>&1 | ConvertFrom-Json

if ($serviceInfo.services.Count -gt 0) {
    $service = $serviceInfo.services[0]
    Write-Host "Service: $($service.serviceName)" -ForegroundColor Green
    Write-Host "  Status: $($service.status)" -ForegroundColor White
    Write-Host "  Desired: $($service.desiredCount)" -ForegroundColor White
    Write-Host "  Running: $($service.runningCount)" -ForegroundColor White
    Write-Host "  Task Definition: $($service.taskDefinition)" -ForegroundColor White
    
    # Get current task def ARN
    $currentTaskDefArn = $service.taskDefinition
} else {
    Write-Host "ERROR: Service not found: $serviceName" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Check task definitions
Write-Host "Checking task definitions..." -ForegroundColor Yellow
$taskDefs = aws ecs list-task-definitions --family-prefix "$StackName-backend" --region $Region --output json 2>&1 | ConvertFrom-Json

if ($taskDefs.taskDefinitionArns.Count -gt 0) {
    Write-Host "Found $($taskDefs.taskDefinitionArns.Count) task definition(s):" -ForegroundColor Green
    $taskDefs.taskDefinitionArns | ForEach-Object {
        Write-Host "  - $_" -ForegroundColor Gray
    }
    
    # Use the service's current task definition
    $taskDefArn = $currentTaskDefArn
    Write-Host ""
    Write-Host "Using current service task definition: $taskDefArn" -ForegroundColor Cyan
} else {
    Write-Host "ERROR: No task definitions found for family: $StackName-backend" -ForegroundColor Red
    Write-Host ""
    Write-Host "Available task definition families:" -ForegroundColor Yellow
    aws ecs list-task-definition-families --region $Region --output table
    exit 1
}
Write-Host ""

# Get secrets info
Write-Host "Checking secrets..." -ForegroundColor Yellow

# Database secret
$dbSecretArn = aws cloudformation describe-stack-resource `
    --stack-name $StackName `
    --logical-resource-id DatabaseSecret `
    --region $Region `
    --query 'StackResourceDetail.PhysicalResourceId' `
    --output text

Write-Host "Database Secret: $dbSecretArn" -ForegroundColor Green

# JWT secret
$jwtSecretName = "$StackName/jwt-secret"
$jwtSecretInfo = aws secretsmanager describe-secret --secret-id $jwtSecretName --region $Region 2>&1 | ConvertFrom-Json

if ($LASTEXITCODE -eq 0) {
    Write-Host "JWT Secret: $($jwtSecretInfo.ARN)" -ForegroundColor Green
    
    # Check if valid
    $jwtValue = aws secretsmanager get-secret-value --secret-id $jwtSecretName --region $Region --query 'SecretString' --output text
    try {
        $parsed = $jwtValue | ConvertFrom-Json
        if ($parsed.secret) {
            Write-Host "  Status: Valid JSON with 'secret' key" -ForegroundColor Green
        } else {
            Write-Host "  Status: Missing 'secret' key" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "  Status: Invalid JSON format" -ForegroundColor Red
    }
    
    $jwtSecretArn = $jwtSecretInfo.ARN
} else {
    Write-Host "JWT Secret: NOT FOUND - will create" -ForegroundColor Yellow
    
    # Create JWT secret
    $jwtValue = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 64 | ForEach-Object {[char]$_})
    $jwtSecretJson = "{`"secret`":`"$jwtValue`"}"
    
    $jwtSecretArn = aws secretsmanager create-secret `
        --name $jwtSecretName `
        --description "JWT signing secret for Hub HRMS" `
        --secret-string $jwtSecretJson `
        --region $Region `
        --query 'ARN' `
        --output text
    
    Write-Host "JWT Secret Created: $jwtSecretArn" -ForegroundColor Green
}
Write-Host ""

# Get current task definition details
Write-Host "Analyzing current task definition..." -ForegroundColor Yellow
$taskDefJson = aws ecs describe-task-definition --task-definition $taskDefArn --region $Region --output json
$taskDef = ($taskDefJson | ConvertFrom-Json).taskDefinition

$container = $taskDef.containerDefinitions[0]

Write-Host "Current Environment Variables:" -ForegroundColor White
if ($container.environment) {
    $container.environment | ForEach-Object {
        Write-Host "  - $($_.name) = $($_.value)" -ForegroundColor Gray
    }
} else {
    Write-Host "  (none)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Current Secrets:" -ForegroundColor White
if ($container.secrets) {
    $container.secrets | ForEach-Object {
        Write-Host "  - $($_.name) from $($_.valueFrom)" -ForegroundColor Gray
    }
} else {
    Write-Host "  (none)" -ForegroundColor Gray
}
Write-Host ""

# Check what needs to be fixed
$needsFix = $false
$issues = @()

# Check for individual DB variables
$hasDBHost = $container.secrets | Where-Object { $_.name -eq "DB_HOST" }
$hasDBPort = $container.secrets | Where-Object { $_.name -eq "DB_PORT" }
$hasDBName = $container.secrets | Where-Object { $_.name -eq "DB_NAME" }
$hasDBUser = $container.secrets | Where-Object { $_.name -eq "DB_USER" }
$hasDBPassword = $container.secrets | Where-Object { $_.name -eq "DB_PASSWORD" }
$hasJWTSecret = $container.secrets | Where-Object { $_.name -eq "JWT_SECRET" }

if (-not $hasDBHost) { $issues += "Missing DB_HOST"; $needsFix = $true }
if (-not $hasDBPort) { $issues += "Missing DB_PORT"; $needsFix = $true }
if (-not $hasDBName) { $issues += "Missing DB_NAME"; $needsFix = $true }
if (-not $hasDBUser) { $issues += "Missing DB_USER"; $needsFix = $true }
if (-not $hasDBPassword) { $issues += "Missing DB_PASSWORD"; $needsFix = $true }
if (-not $hasJWTSecret) { $issues += "Missing JWT_SECRET"; $needsFix = $true }

if ($needsFix) {
    Write-Host "Issues Found:" -ForegroundColor Red
    $issues | ForEach-Object {
        Write-Host "  ✗ $_" -ForegroundColor Red
    }
    Write-Host ""
    
    $confirm = Read-Host "Apply fix? (yes/no)"
    if ($confirm -ne "yes") {
        Write-Host "Aborted" -ForegroundColor Yellow
        exit 0
    }
    
    Write-Host ""
    Write-Host "Applying fix..." -ForegroundColor Yellow
    Write-Host ""
    
    # Create new task definition with correct config
    $container.environment = @(
        @{ name = "SERVER_ADDR"; value = ":8080" }
        @{ name = "GIN_MODE"; value = "release" }
        @{ name = "PORT"; value = "8080" }
        @{ name = "ENVIRONMENT"; value = "production" }
    )
    
    $container.secrets = @(
        @{ name = "DB_HOST"; valueFrom = "$dbSecretArn`:host::" }
        @{ name = "DB_PORT"; valueFrom = "$dbSecretArn`:port::" }
        @{ name = "DB_NAME"; valueFrom = "$dbSecretArn`:dbname::" }
        @{ name = "DB_USER"; valueFrom = "$dbSecretArn`:username::" }
        @{ name = "DB_PASSWORD"; valueFrom = "$dbSecretArn`:password::" }
        @{ name = "JWT_SECRET"; valueFrom = "$jwtSecretArn`:secret::" }
    )
    
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
    
    if ($taskDef.taskRoleArn) {
        $newTaskDef.taskRoleArn = $taskDef.taskRoleArn
    }
    
    $tempFile = "$env:TEMP\hub-hrms-taskdef-fixed.json"
    $newTaskDef | ConvertTo-Json -Depth 10 | Set-Content $tempFile
    
    Write-Host "Registering new task definition..." -ForegroundColor Cyan
    $newTaskDefArn = aws ecs register-task-definition `
        --cli-input-json "file://$tempFile" `
        --region $Region `
        --query 'taskDefinition.taskDefinitionArn' `
        --output text
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ New task definition: $newTaskDefArn" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed to register task definition" -ForegroundColor Red
        Write-Host "Task def saved to: $tempFile" -ForegroundColor Yellow
        exit 1
    }
    
    # Update IAM permissions
    Write-Host "Updating IAM permissions..." -ForegroundColor Cyan
    $executionRoleName = $taskDef.executionRoleArn.Split("/")[-1]
    
    $policyDoc = @{
        Version = "2012-10-17"
        Statement = @(
            @{
                Effect = "Allow"
                Action = @("secretsmanager:GetSecretValue")
                Resource = @($dbSecretArn, $jwtSecretArn)
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
    Write-Host "✓ IAM permissions updated" -ForegroundColor Green
    
    # Update service
    Write-Host "Updating ECS service..." -ForegroundColor Cyan
    aws ecs update-service `
        --cluster $clusterName `
        --service $serviceName `
        --task-definition $newTaskDefArn `
        --force-new-deployment `
        --region $Region 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Service updated" -ForegroundColor Green
    } else {
        Write-Host "✗ Service update failed" -ForegroundColor Red
        exit 1
    }
    
    Write-Host ""
    Write-Host "Waiting for new tasks to start (60 seconds)..." -ForegroundColor Yellow
    Start-Sleep -Seconds 60
    
    Write-Host ""
    Write-Host "Checking logs..." -ForegroundColor Cyan
    $logGroup = "/ecs/$StackName-backend"
    $logs = aws logs tail $logGroup --region $Region --since 2m --format short 2>&1
    
    if ($logs) {
        $logs | Select-Object -Last 20 | ForEach-Object {
            $line = $_.ToString()
            if ($line -match "ResourceInitializationError|unable to pull|invalid character|Failed|error|refused") {
                Write-Host $line -ForegroundColor Red
            } elseif ($line -match "Successfully|Starting|listening|Database pool|Server listening") {
                Write-Host $line -ForegroundColor Green
            } else {
                Write-Host $line
            }
        }
    }
    
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "Fix Applied!" -ForegroundColor Green
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host ""
    
} else {
    Write-Host "Configuration looks correct!" -ForegroundColor Green
    Write-Host "All required environment variables are present." -ForegroundColor Green
    Write-Host ""
    
    # Still check logs
    Write-Host "Checking recent logs..." -ForegroundColor Cyan
    $logGroup = "/ecs/$StackName-backend"
    $logs = aws logs tail $logGroup --region $Region --since 3m --format short 2>&1
    
    if ($logs) {
        $logs | Select-Object -Last 15 | ForEach-Object {
            $line = $_.ToString()
            if ($line -match "ResourceInitializationError|unable to pull|invalid character|Failed|error|refused") {
                Write-Host $line -ForegroundColor Red
            } elseif ($line -match "Successfully|Starting|listening|Database pool|Server listening") {
                Write-Host $line -ForegroundColor Green
            } else {
                Write-Host $line
            }
        }
    }
    Write-Host ""
}

# Show next steps
$albDns = ($stack.Outputs | Where-Object { $_.OutputKey -eq "ALBDNSName" }).OutputValue
if ($albDns) {
    Write-Host "Test backend:" -ForegroundColor Yellow
    Write-Host "  curl http://$albDns/api/health" -ForegroundColor Cyan
    Write-Host ""
}

Write-Host "Monitor logs:" -ForegroundColor Yellow
Write-Host "  aws logs tail /ecs/$StackName-backend --follow --region $Region" -ForegroundColor Cyan
Write-Host ""