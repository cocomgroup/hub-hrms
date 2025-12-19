# Deploy Backend to ECS
param(
    [string]$Region = "us-east-1",
    [string]$Cluster = "hub-hrms-cluster",
    [string]$Service = "hub-hrms-backend"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Deploying Backend to ECS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Get account ID
$ACCOUNT_ID = aws sts get-caller-identity --query Account --output text
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to get AWS account ID. Check AWS credentials." -ForegroundColor Red
    exit 1
}
Write-Host "Account ID: $ACCOUNT_ID" -ForegroundColor Green

# Check Docker is running
docker ps | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Build image
Write-Host ""
Write-Host "Building Docker image..." -ForegroundColor Yellow
Push-Location backend
docker build -t hub-hrms-backend:latest .
$buildResult = $LASTEXITCODE
Pop-Location

if ($buildResult -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Build successful" -ForegroundColor Green

# Tag image
Write-Host ""
Write-Host "Tagging image for ECR..." -ForegroundColor Yellow
$ECR_URI = "$ACCOUNT_ID.dkr.ecr.$Region.amazonaws.com/hub-hrms-backend:latest"
docker tag hub-hrms-backend:latest $ECR_URI
Write-Host "Tagged: $ECR_URI" -ForegroundColor Green

# Login to ECR
Write-Host ""
Write-Host "Logging in to ECR..." -ForegroundColor Yellow
aws ecr get-login-password --region $Region | docker login --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.$Region.amazonaws.com"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ECR login failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Logged in to ECR" -ForegroundColor Green

# Push image
Write-Host ""
Write-Host "Pushing image to ECR..." -ForegroundColor Yellow
docker push $ECR_URI
if ($LASTEXITCODE -ne 0) {
    Write-Host "Push failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Image pushed successfully" -ForegroundColor Green

# Force new deployment
Write-Host ""
Write-Host "Forcing new ECS deployment..." -ForegroundColor Yellow
aws ecs update-service --cluster $Cluster --service $Service --force-new-deployment --region $Region | Out-Null

if ($LASTEXITCODE -ne 0) {
    Write-Host "Deployment failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Backend deployment initiated!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

Write-Host ""
Write-Host "Monitoring deployment (Ctrl+C to stop)..." -ForegroundColor Cyan
Write-Host "This will take 2-5 minutes..." -ForegroundColor Yellow

# Wait for deployment to stabilize
aws ecs wait services-stable --cluster $Cluster --services $Service --region $Region

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Deployment complete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Check logs with:" -ForegroundColor Cyan
    Write-Host "aws logs tail /ecs/hub-hrms/backend --follow --region $Region" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "Deployment may have issues." -ForegroundColor Yellow
    Write-Host "Check service status with:" -ForegroundColor Yellow
    Write-Host "aws ecs describe-services --cluster $Cluster --services $Service --region $Region" -ForegroundColor Yellow
}
