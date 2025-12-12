# HR Workflow System - AWS Deployment Script (CORRECTED)
# PowerShell

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet('dev', 'staging', 'prod')]
    [string]$Environment = 'dev',
    
    [Parameter(Mandatory=$false)]
    [string]$Region = 'us-east-1',
    
    [Parameter(Mandatory=$false)]
    [string]$StackName = '',
        
    [Parameter(Mandatory=$true)]
    [string]$DBPassword,
    
    [Parameter(Mandatory=$false)]
    [string]$DBUsername = 'postgres'
)

# Allow Docker stderr without failing
$ErrorActionPreference = "Continue"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "HR System - AWS Deployment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Set default stack name
if ([string]::IsNullOrEmpty($StackName)) {
    $StackName = "hub-hrms"
}

# Save original location
$originalLocation = Get-Location

# Check prerequisites
Write-Host "Checking prerequisites..." -ForegroundColor Green

if (-not (Get-Command aws -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: AWS CLI not found" -ForegroundColor Red
    Write-Host "Install from: https://aws.amazon.com/cli/" -ForegroundColor Yellow
    exit 1
}
Write-Host "OK: AWS CLI found" -ForegroundColor Green

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Docker not found" -ForegroundColor Red
    Write-Host "Install from: https://www.docker.com/get-started" -ForegroundColor Yellow
    exit 1
}
Write-Host "OK: Docker found" -ForegroundColor Green

# Verify AWS credentials
Write-Host ""
Write-Host "Verifying AWS credentials..." -ForegroundColor Green
$identity = aws sts get-caller-identity --query 'Account' --output text 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to verify AWS credentials" -ForegroundColor Red
    Write-Host "Run: aws configure" -ForegroundColor Yellow
    exit 1
}
Write-Host "OK: AWS Account: $identity" -ForegroundColor Green

$accountId = $identity
$ecrRepo = "$accountId.dkr.ecr.$Region.amazonaws.com"
$backendRepoName = "hub-hrms-backend"
$frontendRepoName = "hub-hrms-frontend"
$backendImageUri = "$ecrRepo/${backendRepoName}:latest"
$frontendImageUri = "$ecrRepo/${frontendRepoName}:latest"

Write-Host ""
Write-Host "Deployment Configuration:" -ForegroundColor Yellow
Write-Host "  Environment:  $Environment" -ForegroundColor White
Write-Host "  Region:       $Region" -ForegroundColor White
Write-Host "  Stack Name:   $StackName" -ForegroundColor White
Write-Host "  Backend URI:  $backendImageUri" -ForegroundColor White
Write-Host ""

# Confirm deployment
$confirm = Read-Host "Continue with deployment? (yes/no)"
if ($confirm -ne "yes") {
    Write-Host "Deployment cancelled" -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 1: Creating ECR Repositories" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Create backend repository
Write-Host "Checking for backend ECR repository: $backendRepoName..."
$repoCheck = aws ecr describe-repositories --repository-names $backendRepoName --region $Region 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "Backend ECR repository found. Skipping creation." -ForegroundColor Green
} else {
    Write-Host "Creating backend ECR repository..." -ForegroundColor Yellow
    $null = aws ecr create-repository `
        --repository-name $backendRepoName `
        --region $Region `
        --image-scanning-configuration scanOnPush=true `
        --tags Key=Project,Value=HRMS `
        --output json 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Backend ECR repository created." -ForegroundColor Green
    } else {
        Write-Host "ERROR: Failed to create backend repository" -ForegroundColor Red
        exit 1
    }
}

# Create frontend repository
Write-Host "Checking for frontend ECR repository: $frontendRepoName..."
$repoCheck = aws ecr describe-repositories --repository-names $frontendRepoName --region $Region 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "Frontend ECR repository found. Skipping creation." -ForegroundColor Green
} else {
    Write-Host "Creating frontend ECR repository..." -ForegroundColor Yellow
    $null = aws ecr create-repository `
        --repository-name $frontendRepoName `
        --region $Region `
        --image-scanning-configuration scanOnPush=true `
        --tags Key=Project,Value=HRMS `
        --output json 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Frontend ECR repository created." -ForegroundColor Green
    } else {
        Write-Host "ERROR: Failed to create frontend repository" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 2: Building Backend Docker Image" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

$backendPath = Join-Path $PSScriptRoot "..\backend"
if (-not (Test-Path $backendPath)) {
    Write-Host "ERROR: Backend directory not found at $backendPath" -ForegroundColor Red
    exit 1
}

Set-Location $backendPath

Write-Host "Building Docker image (this may take a few minutes)..."
Write-Host "Command: docker build -t $backendRepoName ." -ForegroundColor Gray

# Run Docker build and capture output but don't fail on stderr
$buildProcess = Start-Process -FilePath "docker" -ArgumentList "build","-t","$backendRepoName","." -NoNewWindow -Wait -PassThru
if ($buildProcess.ExitCode -ne 0) {
    Write-Host "ERROR: Docker build failed with exit code $($buildProcess.ExitCode)" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Docker image built" -ForegroundColor Green

# Tag image
Write-Host "Tagging image for ECR..."
$tagProcess = Start-Process -FilePath "docker" -ArgumentList "tag","${backendRepoName}:latest","$backendImageUri" -NoNewWindow -Wait -PassThru
if ($tagProcess.ExitCode -ne 0) {
    Write-Host "ERROR: Docker tag failed" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Image tagged" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 3: Pushing Backend Image to ECR" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Login to ECR
Write-Host "Logging in to ECR..."
$loginCmd = "aws ecr get-login-password --region $Region | docker login --username AWS --password-stdin $ecrRepo"
$loginProcess = Start-Process -FilePath "powershell" -ArgumentList "-Command",$loginCmd -NoNewWindow -Wait -PassThru
if ($loginProcess.ExitCode -ne 0) {
    Write-Host "ERROR: ECR login failed" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Logged in to ECR" -ForegroundColor Green

# Push image
Write-Host "Pushing image to ECR (this may take several minutes)..."
Write-Host "Image: $backendImageUri" -ForegroundColor Gray
$pushProcess = Start-Process -FilePath "docker" -ArgumentList "push","$backendImageUri" -NoNewWindow -Wait -PassThru
if ($pushProcess.ExitCode -ne 0) {
    Write-Host "ERROR: Docker push failed" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Backend image pushed to ECR" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 4: Creating Frontend Placeholder" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Check if frontend image already exists
Write-Host "Checking for existing frontend image..."
$imageCheck = aws ecr describe-images --repository-name $frontendRepoName --region $Region --image-ids imageTag=latest 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "Frontend image already exists. Skipping creation." -ForegroundColor Green
} else {
    Write-Host "Creating placeholder frontend image..." -ForegroundColor Yellow
    
    # Create temporary directory and Dockerfile
    $tempDir = New-Item -ItemType Directory -Force -Path "$env:TEMP\hub-hrms-frontend-temp"
    $frontendDockerfile = @"
FROM nginx:alpine
RUN echo '<html><body><h1>Frontend Coming Soon</h1><p>Backend API is running at /api</p></body></html>' > /usr/share/nginx/html/index.html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
"@
    $frontendDockerfile | Out-File -FilePath "$tempDir\Dockerfile" -Encoding ASCII -NoNewline
    
    # Build frontend placeholder
    Push-Location $tempDir
    $frontendBuildProcess = Start-Process -FilePath "docker" -ArgumentList "build","-t","$frontendRepoName","." -NoNewWindow -Wait -PassThru
    if ($frontendBuildProcess.ExitCode -eq 0) {
        Write-Host "OK: Frontend placeholder built" -ForegroundColor Green
        
        # Tag and push
        $null = docker tag ${frontendRepoName}:latest $frontendImageUri 2>&1
        $frontendPushProcess = Start-Process -FilePath "docker" -ArgumentList "push","$frontendImageUri" -NoNewWindow -Wait -PassThru
        if ($frontendPushProcess.ExitCode -eq 0) {
            Write-Host "OK: Frontend placeholder pushed" -ForegroundColor Green
        } else {
            Write-Host "WARNING: Frontend push failed, but continuing..." -ForegroundColor Yellow
        }
    } else {
        Write-Host "WARNING: Frontend build failed, but continuing..." -ForegroundColor Yellow
    }
    Pop-Location
    
    # Cleanup
    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
}

Set-Location $originalLocation

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 5: Deploying CloudFormation Stack" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Build parameters
$cfParams = @(
    "EnvironmentName=$StackName"
    "DBUsername=$DBUsername"
    "DBPassword=$DBPassword"
    "BackendImageUri=$backendImageUri"
    "FrontendImageUri=$frontendImageUri"
)

Write-Host "Stack Name: $StackName" -ForegroundColor White
Write-Host "Parameters:" -ForegroundColor White
Write-Host "  EnvironmentName: $StackName" -ForegroundColor Gray
Write-Host "  DBUsername: $DBUsername" -ForegroundColor Gray
Write-Host "  DBPassword: ***" -ForegroundColor Gray
Write-Host "  BackendImageUri: $backendImageUri" -ForegroundColor Gray
Write-Host "  FrontendImageUri: $frontendImageUri" -ForegroundColor Gray
Write-Host ""

$templatePath = Join-Path $PSScriptRoot "cloudformation-stack.yaml"
if (-not (Test-Path $templatePath)) {
    Write-Host "ERROR: CloudFormation template not found at: $templatePath" -ForegroundColor Red
    exit 1
}

Write-Host "Deploying CloudFormation stack (this will take 15-20 minutes)..." -ForegroundColor Cyan
Write-Host "Template: $templatePath" -ForegroundColor Gray
Write-Host ""
Write-Host "Monitor progress at:" -ForegroundColor Gray
Write-Host "https://console.aws.amazon.com/cloudformation/home?region=$Region" -ForegroundColor Cyan
Write-Host ""

# Deploy with error handling
$ErrorActionPreference = "Stop"
try {
    $deployOutput = aws cloudformation deploy `
        --template-file $templatePath `
        --stack-name $StackName `
        --parameter-overrides $cfParams `
        --capabilities CAPABILITY_NAMED_IAM `
        --region $Region 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "CloudFormation deployment failed"
    }
    
    Write-Host "OK: CloudFormation stack deployed" -ForegroundColor Green
    
} catch {
    Write-Host ""
    Write-Host "ERROR: CloudFormation deployment failed" -ForegroundColor Red
    Write-Host ""
    Write-Host "Error details:" -ForegroundColor Yellow
    Write-Host $deployOutput -ForegroundColor Red
    Write-Host ""
    Write-Host "To diagnose, run:" -ForegroundColor Yellow
    Write-Host "  aws cloudformation describe-stack-events --stack-name $StackName --region $Region --max-items 20" -ForegroundColor Cyan
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 6: Getting Stack Outputs" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Get outputs
$outputsJson = aws cloudformation describe-stacks `
    --stack-name $StackName `
    --region $Region `
    --query 'Stacks[0].Outputs' `
    --output json

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to get stack outputs" -ForegroundColor Red
    exit 1
}

$outputs = $outputsJson | ConvertFrom-Json
$appUrl = ($outputs | Where-Object { $_.OutputKey -eq "ApplicationURL" }).OutputValue
$albDns = ($outputs | Where-Object { $_.OutputKey -eq "ALBDNSName" }).OutputValue
$rdsEndpoint = ($outputs | Where-Object { $_.OutputKey -eq "RDSEndpoint" }).OutputValue
$frontendBucket = ($outputs | Where-Object { $_.OutputKey -eq "FrontendBucketName" }).OutputValue
$cloudfrontId = ($outputs | Where-Object { $_.OutputKey -eq "CloudFrontDistributionId" }).OutputValue

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Deployment Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Application URL:" -ForegroundColor Yellow
Write-Host "  $appUrl" -ForegroundColor Cyan
Write-Host ""
Write-Host "Backend API (ALB):" -ForegroundColor Yellow
Write-Host "  http://$albDns" -ForegroundColor Cyan
Write-Host ""
Write-Host "Database:" -ForegroundColor Yellow
Write-Host "  Endpoint: $rdsEndpoint" -ForegroundColor Cyan
Write-Host "  Database: hrmsdb" -ForegroundColor Gray
Write-Host "  Username: $DBUsername" -ForegroundColor Gray
Write-Host ""
Write-Host "Frontend:" -ForegroundColor Yellow
Write-Host "  S3 Bucket: $frontendBucket" -ForegroundColor Cyan
Write-Host "  CloudFront: $cloudfrontId" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Test the backend API:" -ForegroundColor White
Write-Host "   curl http://$albDns/api/health" -ForegroundColor Cyan
Write-Host ""
Write-Host "2. View backend logs:" -ForegroundColor White
Write-Host "   aws logs tail /ecs/hub-hrms-backend --follow --region $Region" -ForegroundColor Cyan
Write-Host ""
Write-Host "3. Build and deploy frontend:" -ForegroundColor White
Write-Host "   cd ..\frontend" -ForegroundColor Cyan
Write-Host "   npm install && npm run build" -ForegroundColor Cyan
Write-Host "   aws s3 sync dist/ s3://$frontendBucket" -ForegroundColor Cyan
Write-Host "   aws cloudfront create-invalidation --distribution-id $cloudfrontId --paths '/*'" -ForegroundColor Cyan
Write-Host ""
Write-Host "4. Run database migrations:" -ForegroundColor White
Write-Host "   # Connect to: $rdsEndpoint" -ForegroundColor Cyan
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Deployment script completed successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan