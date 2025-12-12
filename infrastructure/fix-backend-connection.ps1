# Fix Hub HRMS Backend Database Connection
# This script updates the ECS task definition to use correct environment variables

param(
    [Parameter(Mandatory=$false)]
    [string]$StackName = 'hub-hrms',
    
    [Parameter(Mandatory=$false)]
    [string]$Region = 'us-east-1'
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Fixing Backend Database Connection" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get stack outputs
Write-Host "Getting stack information..." -ForegroundColor Yellow
$outputs = aws cloudformation describe-stacks --stack-name $StackName --region $Region --query 'Stacks[0].Outputs' --output json | ConvertFrom-Json
$rdsEndpoint = ($outputs | Where-Object { $_.OutputKey -eq "RDSEndpoint" }).OutputValue
$secretArn = ($outputs | Where-Object { $_.OutputKey -eq "DatabaseSecretArn" }).OutputValue

if ([string]::IsNullOrEmpty($secretArn)) {
    Write-Host "Getting secret ARN from resources..." -ForegroundColor Yellow
    $secretArn = aws cloudformation describe-stack-resources --stack-name $StackName --region $Region --logical-resource-id DatabaseSecret --query 'StackResources[0].PhysicalResourceId' --output text
}

Write-Host "RDS Endpoint: $rdsEndpoint" -ForegroundColor White
Write-Host "Secret ARN: $secretArn" -ForegroundColor White
Write-Host ""

# Step 1: Get current task definition
Write-Host "Step 1: Getting current task definition..." -ForegroundColor Yellow
$taskDefArn = aws ecs list-task-definitions --family-prefix "$StackName-backend" --region $Region --sort DESC --max-items 1 --query 'taskDefinitionArns[0]' --output text

if ([string]::IsNullOrEmpty($taskDefArn)) {
    Write-Host "ERROR: Task definition not found" -ForegroundColor Red
    exit 1
}

Write-Host "Current task: $taskDefArn" -ForegroundColor Gray

$taskDefJson = aws ecs describe-task-definition --task-definition $taskDefArn --region $Region --query 'taskDefinition' --output json
$taskDef = $taskDefJson | ConvertFrom-Json

Write-Host "✓ Task definition retrieved" -ForegroundColor Green

# Step 2: Create new task definition with correct environment variables
Write-Host ""
Write-Host "Step 2: Creating new task definition..." -ForegroundColor Yellow

# Build the container definition with correct env vars
$containerDef = $taskDef.containerDefinitions[0]

# Remove old environment variables
$newEnvironment = @()
foreach ($env in $containerDef.environment) {
    if ($env.name -ne "DATABASE_URL") {
        $newEnvironment += $env
    }
}
$containerDef.environment = $newEnvironment

# Set up the correct secrets
$containerDef.secrets = @(
    @{
        name = "DB_HOST"
        valueFrom = "$secretArn`:host::"
    },
    @{
        name = "DB_PORT"
        valueFrom = "$secretArn`:port::"
    },
    @{
        name = "DB_NAME"
        valueFrom = "$secretArn`:dbname::"
    },
    @{
        name = "DB_USER"
        valueFrom = "$secretArn`:username::"
    },
    @{
        name = "DB_PASSWORD"
        valueFrom = "$secretArn`:password::"
    }
)

# Create new task definition JSON
$newTaskDef = @{
    family = $taskDef.family
    networkMode = $taskDef.networkMode
    requiresCompatibilities = $taskDef.requiresCompatibilities
    cpu = $taskDef.cpu
    memory = $taskDef.memory
    executionRoleArn = $taskDef.executionRoleArn
    taskRoleArn = $taskDef.taskRoleArn
    containerDefinitions = @($containerDef)
}

# Write to temp file
$tempFile = "$env:TEMP\task-def-fixed.json"
$newTaskDef | ConvertTo-Json -Depth 10 | Set-Content $tempFile

Write-Host "New task definition created" -ForegroundColor Gray

# Register new task definition
Write-Host "Registering new task definition..." -ForegroundColor Gray
$registerOutput = aws ecs register-task-definition --cli-input-json "file://$tempFile" --region $Region --query 'taskDefinition.taskDefinitionArn' --output text

if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrEmpty($registerOutput)) {
    Write-Host "ERROR: Failed to register task definition" -ForegroundColor Red
    Write-Host "Check the temp file: $tempFile" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ New task definition registered: $registerOutput" -ForegroundColor Green

# Cleanup
Remove-Item $tempFile -ErrorAction SilentlyContinue

# Step 3: Update the ECS service
Write-Host ""
Write-Host "Step 3: Updating ECS service..." -ForegroundColor Yellow

$clusterName = "$StackName-cluster"
$serviceName = "$StackName-backend"

Write-Host "Cluster: $clusterName" -ForegroundColor Gray
Write-Host "Service: $serviceName" -ForegroundColor Gray

# Get new task def revision number
$newRevision = $registerOutput.Split(":")[-1]

# Update service with force new deployment
$updateOutput = aws ecs update-service `
    --cluster $clusterName `
    --service $serviceName `
    --task-definition $registerOutput `
    --force-new-deployment `
    --region $Region `
    --query 'service.serviceName' `
    --output text

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to update service" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Service update initiated" -ForegroundColor Green

# Step 4: Wait for service to stabilize
Write-Host ""
Write-Host "Step 4: Waiting for service to stabilize..." -ForegroundColor Yellow
Write-Host "This may take 2-3 minutes..." -ForegroundColor Gray
Write-Host ""

# Monitor deployment
$attempt = 0
$maxAttempts = 30
$stable = $false

while ($attempt -lt $maxAttempts -and -not $stable) {
    Start-Sleep -Seconds 10
    $attempt++
    
    # Get service status
    $serviceJson = aws ecs describe-services --cluster $clusterName --services $serviceName --region $Region --query 'services[0]' --output json
    $service = $serviceJson | ConvertFrom-Json
    
    $runningCount = $service.runningCount
    $desiredCount = $service.desiredCount
    $deployments = $service.deployments.Count
    
    Write-Host "[$attempt/$maxAttempts] Running: $runningCount/$desiredCount, Deployments: $deployments" -ForegroundColor Gray
    
    # Check if stable (only one deployment and running count matches desired)
    if ($deployments -eq 1 -and $runningCount -eq $desiredCount) {
        $stable = $true
    }
}

if ($stable) {
    Write-Host ""
    Write-Host "✓ Service is stable!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "⚠ Service update is taking longer than expected" -ForegroundColor Yellow
    Write-Host "Continue monitoring in the AWS Console" -ForegroundColor Yellow
}

# Step 5: Check logs
Write-Host ""
Write-Host "Step 5: Checking recent logs..." -ForegroundColor Yellow
Write-Host ""

Start-Sleep -Seconds 5

$logGroup = "/ecs/$StackName-backend"
$recentLogs = aws logs tail $logGroup --region $Region --since 2m 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host $recentLogs -ForegroundColor White
} else {
    Write-Host "Could not retrieve logs (service may still be starting)" -ForegroundColor Yellow
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Fix Applied!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Changes made:" -ForegroundColor Yellow
Write-Host "  ✓ Removed DATABASE_URL environment variable" -ForegroundColor Green
Write-Host "  ✓ Added DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD secrets" -ForegroundColor Green
Write-Host "  ✓ Force deployed new task definition" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Monitor logs:" -ForegroundColor White
Write-Host "     aws logs tail $logGroup --follow --region $Region" -ForegroundColor Cyan
Write-Host ""
Write-Host "  2. Check service health:" -ForegroundColor White
Write-Host "     aws ecs describe-services --cluster $clusterName --services $serviceName --region $Region" -ForegroundColor Cyan
Write-Host ""
Write-Host "  3. Test the API:" -ForegroundColor White
$albDns = ($outputs | Where-Object { $_.OutputKey -eq "ALBDNSName" }).OutputValue
if ($albDns) {
    Write-Host "     curl http://$albDns/api/health" -ForegroundColor Cyan
}
Write-Host ""