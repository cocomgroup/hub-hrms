# HR Workflow System - AWS Deployment Script
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
    [string]$DBHost,
    
    [Parameter(Mandatory=$true)]
    [string]$DBPassword,
    
    [Parameter(Mandatory=$false)]
    [string]$DBUsername = 'postgres',
    
    [Parameter(Mandatory=$false)]
    [string]$DBName = 'hrmsdb',
    
    [Parameter(Mandatory=$false)]
    [string]$DBPort = '5432',
    
    [Parameter(Mandatory=$false)]
    [string]$CertificateArn = '',
    
    [Parameter(Mandatory=$false)]
    [string]$DomainName = '',
    
    [Parameter(Mandatory=$false)]
    [switch]$UseExistingVPC,
    
    [Parameter(Mandatory=$false)]
    [string]$VpcId = '',
    
    [Parameter(Mandatory=$false)]
    [string]$PublicSubnet1Id = '',
    
    [Parameter(Mandatory=$false)]
    [string]$PublicSubnet2Id = '',
    
    [Parameter(Mandatory=$false)]
    [string]$PrivateSubnet1Id = '',
    
    [Parameter(Mandatory=$false)]
    [string]$PrivateSubnet2Id = ''
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "HR System - AWS Deployment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Set default stack name
if ([string]::IsNullOrEmpty($StackName)) {
    $StackName = "hub-hrms"
}

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
$backendImageUri = "$ecrRepo/${backendRepoName}:latest"

Write-Host ""
Write-Host "Deployment Configuration:" -ForegroundColor Yellow
Write-Host "  Environment:  $Environment" -ForegroundColor White
Write-Host "  Region:       $Region" -ForegroundColor White
Write-Host "  Stack Name:   $StackName" -ForegroundColor White
Write-Host "  DB Host:      $DBHost" -ForegroundColor White
Write-Host "  DB Name:      $DBName" -ForegroundColor White
Write-Host ""

# Confirm deployment
$confirm = Read-Host "Continue with deployment? (yes/no)"
if ($confirm -ne "yes") {
    Write-Host "Deployment cancelled" -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 1: Creating ECR Repository" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "Checking for existing ECR repository: $backendRepoName..."

# Set ErrorActionPreference to Stop for the try block to catch the aws cli error
$ErrorActionPreference = "Stop"

try {
    # If the repository exists, this command will succeed. Output is piped to $null.
    # We explicitly suppress Write-Host output to prevent spam, but allow it to fail.
    aws ecr describe-repositories --repository-names $backendRepoName --region $Region | Out-Null
    Write-Host "ECR repository found. Skipping creation." -ForegroundColor Green
    $repoExists = $true

} catch {
    # The command failed, which means the repository likely does not exist.
    # A non-zero exit code in 'aws' is interpreted as a catchable error in PowerShell.
    Write-Host "ECR repository not found. Creating a new one..." -ForegroundColor Yellow

    # Now, attempt to create the repository
    try {
        aws ecr create-repository --repository-name $backendRepoName --region $Region --image-scanning-configuration scanOnPush=true --tags Key=Project,Value=HRM --output json | Out-Null
        Write-Host "ECR repository '$backendRepoName' created successfully." -ForegroundColor Green
        $repoExists = $true

    } catch {
        Write-Error "Failed to create ECR repository '$backendRepoName'. Error: $($_.Exception.Message)"
        exit 1 # Exit script on critical failure
    }
}

# Reset ErrorActionPreference after the try/catch block
$ErrorActionPreference = "Continue"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 2: Building Backend Docker Image" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Navigate to backend directory
$originalLocation = Get-Location
$backendPath = Join-Path (Split-Path $PSScriptRoot -Parent) "backend"

if (-not (Test-Path $backendPath)) {
    Write-Host "ERROR: Backend directory not found at $backendPath" -ForegroundColor Red
    exit 1
}

Set-Location $backendPath
Write-Host "Building Docker image..." -ForegroundColor Cyan

$null = docker build -t $backendRepoName . 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Docker build failed" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Docker image built" -ForegroundColor Green

# Tag image
docker tag ${backendRepoName}:latest $backendImageUri
Write-Host "OK: Image tagged" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 3: Pushing Image to ECR" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Login to ECR
Write-Host "Logging in to ECR..." -ForegroundColor Cyan
$null = Invoke-Expression "aws ecr get-login-password --region $Region | docker login --username AWS --password-stdin $ecrRepo" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: ECR login failed" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Logged in to ECR" -ForegroundColor Green

# Push image
Write-Host "Pushing image to ECR (this may take a few minutes)..." -ForegroundColor Cyan
$null = docker push $backendImageUri 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to push image" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Image pushed to ECR" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 4: Deploying CloudFormation Stack" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Generate JWT secret
$jwtSecret = [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))

# Build parameters
$cfParams = @(
    "EnvironmentName=$Environment"
    "DBHost=$DBHost"
    "DBPort=$DBPort"
    "DBName=$DBName"
    "DBUsername=$DBUsername"
    "DBPassword=$DBPassword"
    "JWTSecret=$jwtSecret"
    "BackendImageUri=$backendImageUri"
)

if ($CertificateArn) {
    $cfParams += "CertificateArn=$CertificateArn"
}

if ($DomainName) {
    $cfParams += "DomainName=$DomainName"
}

if ($UseExistingVPC) {
    if ([string]::IsNullOrEmpty($VpcId)) {
        Write-Host "ERROR: VpcId required when using existing VPC" -ForegroundColor Red
        Set-Location $originalLocation
        exit 1
    }
    $cfParams += "VpcId=$VpcId"
    $cfParams += "PublicSubnet1Id=$PublicSubnet1Id"
    $cfParams += "PublicSubnet2Id=$PublicSubnet2Id"
    $cfParams += "PrivateSubnet1Id=$PrivateSubnet1Id"
    $cfParams += "PrivateSubnet2Id=$PrivateSubnet2Id"
}

Write-Host "Deploying CloudFormation stack (this will take 15-20 minutes)..." -ForegroundColor Cyan
$templatePath = Join-Path $PSScriptRoot "cloudformation-stack.yaml"

$null = aws cloudformation deploy --template-file $templatePath --stack-name $StackName --parameter-overrides $cfParams --capabilities CAPABILITY_NAMED_IAM --region $Region 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: CloudFormation deployment failed" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}

Write-Host "OK: CloudFormation stack deployed" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 5: Building & Deploying Frontend" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

# Get outputs from CloudFormation
Write-Host "Getting CloudFormation outputs..." -ForegroundColor Cyan
$outputsJson = aws cloudformation describe-stacks --stack-name $StackName --region $Region --query 'Stacks[0].Outputs' --output json
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to get stack outputs" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}

$outputs = $outputsJson | ConvertFrom-Json
$frontendBucket = ($outputs | Where-Object { $_.OutputKey -eq "FrontendBucketName" }).OutputValue
$cloudfrontId = ($outputs | Where-Object { $_.OutputKey -eq "CloudFrontDistributionId" }).OutputValue
$albDns = ($outputs | Where-Object { $_.OutputKey -eq "ALBDNSName" }).OutputValue
$appUrl = ($outputs | Where-Object { $_.OutputKey -eq "ApplicationURL" }).OutputValue

Write-Host "  Frontend Bucket: $frontendBucket" -ForegroundColor White
Write-Host "  CloudFront ID:   $cloudfrontId" -ForegroundColor White
Write-Host "  ALB DNS:         $albDns" -ForegroundColor White

# Build frontend
Write-Host ""
Write-Host "Building frontend..." -ForegroundColor Cyan
$frontendPath = Join-Path (Split-Path $PSScriptRoot -Parent) "frontend"

if (-not (Test-Path $frontendPath)) {
    Write-Host "ERROR: Frontend directory not found at $frontendPath" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}

Set-Location $frontendPath

# Update API endpoint
$apiEndpoint = if ($DomainName) { "https://$DomainName/api" } else { "https://$appUrl/api" }
Write-Host "  API Endpoint: $apiEndpoint" -ForegroundColor White

@"
VITE_API_URL=$apiEndpoint
"@ | Out-File -FilePath ".env.production" -Encoding utf8

npm run build 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Frontend build failed" -ForegroundColor Red
    Write-Host "Run 'npm install' in the frontend directory first" -ForegroundColor Yellow
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Frontend built" -ForegroundColor Green

# Deploy to S3
Write-Host "Deploying frontend to S3..." -ForegroundColor Cyan
$null = aws s3 sync dist/ s3://$frontendBucket/ --delete --region $Region 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: S3 sync failed" -ForegroundColor Red
    Set-Location $originalLocation
    exit 1
}
Write-Host "OK: Frontend deployed to S3" -ForegroundColor Green

# Invalidate CloudFront cache
Write-Host "Invalidating CloudFront cache..." -ForegroundColor Cyan
aws cloudfront create-invalidation --distribution-id $cloudfrontId --paths "/*" 2>&1 | Out-Null
Write-Host "OK: CloudFront cache invalidated" -ForegroundColor Green

Set-Location $originalLocation

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "DEPLOYMENT COMPLETE!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Deployment Summary:" -ForegroundColor Yellow
Write-Host "  Environment:        $Environment" -ForegroundColor White
Write-Host "  Stack Name:         $StackName" -ForegroundColor White
Write-Host "  Region:             $Region" -ForegroundColor White
Write-Host ""
Write-Host "Application URLs:" -ForegroundColor Yellow
Write-Host "  Application:        $appUrl" -ForegroundColor White
Write-Host "  Backend (ALB):      https://$albDns" -ForegroundColor White
Write-Host ""
Write-Host "AWS Resources:" -ForegroundColor Yellow
Write-Host "  Frontend Bucket:    $frontendBucket" -ForegroundColor White
Write-Host "  CloudFront ID:      $cloudfrontId" -ForegroundColor White
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Create admin user in database" -ForegroundColor White
Write-Host "  2. Test application at: $appUrl" -ForegroundColor White
Write-Host "  3. Configure DNS if using custom domain" -ForegroundColor White
Write-Host ""
Write-Host "Useful Commands:" -ForegroundColor Yellow
Write-Host "  View logs:    aws logs tail /ecs/$Environment-hrms-backend --follow" -ForegroundColor White
Write-Host "  Delete stack: aws cloudformation delete-stack --stack-name $StackName" -ForegroundColor White
Write-Host ""
Write-Host "Deployment successful!" -ForegroundColor Green
