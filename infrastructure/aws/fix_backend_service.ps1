# Fix Backend Service Script
# This script diagnoses and fixes the missing backend service

param(
    [string]$StackName = "hub-hrms-dev",
    [string]$Region = "us-east-1"
)

Write-Host "==================================================" -ForegroundColor Blue
Write-Host "  Diagnosing Backend Service Issue" -ForegroundColor Blue
Write-Host "==================================================" -ForegroundColor Blue

# Check if stack exists
Write-Host "`nChecking stack status..." -ForegroundColor Cyan
$stack = aws cloudformation describe-stacks --stack-name $StackName --region $Region 2>&1 | ConvertFrom-Json
if ($LASTEXITCODE -ne 0) {
    Write-Host "Stack not found: $StackName" -ForegroundColor Red
    exit 1
}

Write-Host "Stack Status: $($stack.Stacks[0].StackStatus)" -ForegroundColor Yellow

# Check backend service resource
Write-Host "`nChecking BackendService resource..." -ForegroundColor Cyan
$backendResource = aws cloudformation describe-stack-resources --stack-name $StackName --region $Region --logical-resource-id BackendService 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "BackendService resource not found in stack" -ForegroundColor Red
} else {
    $resource = $backendResource | ConvertFrom-Json
    Write-Host "Resource Status: $($resource.StackResources[0].ResourceStatus)" -ForegroundColor Yellow
    if ($resource.StackResources[0].ResourceStatusReason) {
        Write-Host "Reason: $($resource.StackResources[0].ResourceStatusReason)" -ForegroundColor Yellow
    }
}

# Check backend task definition
Write-Host "`nChecking BackendTaskDefinition..." -ForegroundColor Cyan
$taskDefResource = aws cloudformation describe-stack-resources --stack-name $StackName --region $Region --logical-resource-id BackendTaskDefinition 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "BackendTaskDefinition resource not found in stack" -ForegroundColor Red
} else {
    $taskDef = $taskDefResource | ConvertFrom-Json
    Write-Host "TaskDefinition Status: $($taskDef.StackResources[0].ResourceStatus)" -ForegroundColor Green
    Write-Host "TaskDefinition ARN: $($taskDef.StackResources[0].PhysicalResourceId)" -ForegroundColor Green
}

# Get recent stack events
Write-Host "`nRecent stack events (last 20):" -ForegroundColor Cyan
aws cloudformation describe-stack-events --stack-name $StackName --region $Region --max-items 20 --query "StackEvents[?contains(LogicalResourceId, 'Backend')].[Timestamp,LogicalResourceId,ResourceStatus,ResourceStatusReason]" --output table

# Check if service exists in ECS
Write-Host "`nChecking ECS cluster for backend service..." -ForegroundColor Cyan
$clusterName = "$StackName-cluster"
$serviceName = "$StackName-backend"

$ecsService = aws ecs describe-services --cluster $clusterName --services $serviceName --region $Region 2>&1 | ConvertFrom-Json

if ($ecsService.failures) {
    Write-Host "Service not found in ECS cluster" -ForegroundColor Red
    Write-Host "Failure reason: $($ecsService.failures[0].reason)" -ForegroundColor Red
} else {
    Write-Host "Service found!" -ForegroundColor Green
    Write-Host "Desired tasks: $($ecsService.services[0].desiredCount)" -ForegroundColor Yellow
    Write-Host "Running tasks: $($ecsService.services[0].runningCount)" -ForegroundColor Yellow
    Write-Host "Status: $($ecsService.services[0].status)" -ForegroundColor Yellow
}

Write-Host "`n==================================================" -ForegroundColor Blue
Write-Host "  Recommended Actions" -ForegroundColor Blue
Write-Host "==================================================" -ForegroundColor Blue

Write-Host "`n1. If BackendService resource failed to create:" -ForegroundColor Cyan
Write-Host "   Run: .\infrastructure\aws\deploy.ps1 -Update -SkipBuild" -ForegroundColor White

Write-Host "`n2. If BackendTaskDefinition has issues:" -ForegroundColor Cyan
Write-Host "   Check CloudWatch logs at: /ecs/$StackName/backend" -ForegroundColor White

Write-Host "`n3. Check detailed CloudFormation events:" -ForegroundColor Cyan
Write-Host "   https://console.aws.amazon.com/cloudformation/home?region=$Region#/stacks/events?stackId=$StackName" -ForegroundColor White

Write-Host "`n4. View ECS cluster:" -ForegroundColor Cyan
Write-Host "   https://console.aws.amazon.com/ecs/v2/clusters/$clusterName/services?region=$Region" -ForegroundColor White

Write-Host ""