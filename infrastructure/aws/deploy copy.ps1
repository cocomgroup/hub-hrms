# HRMS CloudFormation Deployment Script (PowerShell)
# This script automates the deployment of the HRMS application to AWS

[CmdletBinding()]
param(
    [string]$StackName = "hrms-prod",
    [string]$Region = "us-east-1",
    [switch]$SkipImages,
    [string]$JWTSecret,
    [switch]$Help
)

# Configuration
$DBUsername = "postgres"
$DBPassword = "postgresql123!"
$DBName = "hrmsdb"
$Environment = "production"

# Colors for output
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to display help
function Show-Help {
    Write-Host @"
HRMS CloudFormation Deployment Script

Usage: .\deploy.ps1 [options]

Options:
  -StackName <name>    CloudFormation stack name (default: hrms-prod)
  -Region <region>     AWS region (default: us-east-1)
  -SkipImages          Skip building and pushing Docker images
  -JWTSecret <secret>  Use specific JWT secret
  -Help                Show this help message

Examples:
  .\deploy.ps1
  .\deploy.ps1 -StackName my-hrms -Region us-west-2
  .\deploy.ps1 -SkipImages
"@
    exit 0
}

if ($Help) {
    Show-Help
}

# Function to check prerequisites
function Test-Prerequisites {
    Write-Info "Checking prerequisites..."
    
    # Check AWS CLI
    try {
        $null = aws --version 2>&1
    } catch {
        Write-Error "AWS CLI is not installed. Please install it first."
        Write-Host "Download from: https://aws.amazon.com/cli/"
        exit 1
    }
    
    # Check Docker
    try {
        $null = docker --version 2>&1
    } catch {
        Write-Error "Docker is not installed. Please install it first."
        Write-Host "Download from: https://www.docker.com/products/docker-desktop"
        exit 1
    }
    
    # Check AWS credentials
    try {
        $null = aws sts get-caller-identity 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "AWS credentials not configured"
        }
    } catch {
        Write-Error "AWS credentials are not configured. Run 'aws configure' first."
        exit 1
    }
    
    Write-Success "All prerequisites met"
}

# Function to get AWS account ID
function Get-AccountId {
    $accountId = aws sts get-caller-identity --query Account --output text 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to get AWS account ID"
        exit 1
    }
    return $accountId.Trim()
}

# Function to create ECR repositories if they don't exist
function New-ECRRepositories {
    Write-Info "Checking ECR repositories..."
    
    $accountId = Get-AccountId
    $script:ECRReposExist = $false
    
    # Check backend repository
    $backendRepo = aws ecr describe-repositories --repository-names "$StackName-backend" --region $Region 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Info "Backend ECR repository already exists"
        $script:ECRReposExist = $true
    } else {
        Write-Info "Creating backend ECR repository..."
        aws ecr create-repository `
            --repository-name "$StackName-backend" `
            --region $Region `
            --image-scanning-configuration scanOnPush=true | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Created backend ECR repository"
        } else {
            Write-Error "Failed to create backend ECR repository"
            exit 1
        }
    }
    
    # Check frontend repository
    $frontendRepo = aws ecr describe-repositories --repository-names "$StackName-frontend" --region $Region 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Info "Frontend ECR repository already exists"
        $script:ECRReposExist = $true
    } else {
        Write-Info "Creating frontend ECR repository..."
        aws ecr create-repository `
            --repository-name "$StackName-frontend" `
            --region $Region `
            --image-scanning-configuration scanOnPush=true | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Created frontend ECR repository"
        } else {
            Write-Error "Failed to create frontend ECR repository"
            exit 1
        }
    }
}

# Function to build and push Docker images
function Build-AndPushImages {
    Write-Info "Building and pushing Docker images..."
    
    $accountId = Get-AccountId
    $ecrUri = "$accountId.dkr.ecr.$Region.amazonaws.com"
    
    # Login to ECR
    Write-Info "Logging in to ECR..."
    $password = aws ecr get-login-password --region $Region
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to get ECR login password"
        exit 1
    }
    
    $password | docker login --username AWS --password-stdin $ecrUri
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to login to ECR"
        exit 1
    }
    
    # Build and push backend
    Write-Info "Building backend image..."
    Push-Location backend
    
    docker build -t "$StackName-backend:latest" .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build backend image"
        Pop-Location
        exit 1
    }
    
    docker tag "$StackName-backend:latest" "$ecrUri/$StackName-backend:latest"
    
    Write-Info "Pushing backend image..."
    docker push "$ecrUri/$StackName-backend:latest"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to push backend image"
        Pop-Location
        exit 1
    }
    
    Write-Success "Backend image pushed successfully"
    Pop-Location
    
    # Build and push frontend
    Write-Info "Building frontend image..."
    Push-Location frontend
    
    docker build -t "$StackName-frontend:latest" .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build frontend image"
        Pop-Location
        exit 1
    }
    
    docker tag "$StackName-frontend:latest" "$ecrUri/$StackName-frontend:latest"
    
    Write-Info "Pushing frontend image..."
    docker push "$ecrUri/$StackName-frontend:latest"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to push frontend image"
        Pop-Location
        exit 1
    }
    
    Write-Success "Frontend image pushed successfully"
    Pop-Location
}

# Function to generate JWT secret
function New-JWTSecret {
    $bytes = New-Object byte[] 32
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $rng.GetBytes($bytes)
    $secret = [Convert]::ToBase64String($bytes)
    $rng.Dispose()
    return $secret
}

# Function to deploy CloudFormation stack
function Deploy-Stack {
    Write-Info "Deploying CloudFormation stack..."
    
    # Generate JWT secret if not set
    if (-not $JWTSecret) {
        $script:JWTSecret = New-JWTSecret
        Write-Info "Generated JWT secret"
    }
    
    # Check if stack exists
    $stackExists = aws cloudformation describe-stacks --stack-name $StackName --region $Region 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Warning "Stack already exists. Updating..."
        $operation = "update-stack"
    } else {
        Write-Info "Creating new stack..."
        $operation = "create-stack"
    }
    
     # Get account ID
    $accountId = Get-AccountId
    
    # Determine if we need to create ECR repos
    $createECR = if ($script:ECRReposExist) { "false" } else { "true" }

    # Deploy stack
    try {
        aws cloudformation $operation `
            --stack-name $StackName `
            --template-body file://infrastructure/aws/cloudformation-stack.yaml `
            --parameters  `
                ParameterKey=Environment,ParameterValue="$Environment" `
                ParameterKey=DBUsername,ParameterValue="$DBUsername" `
                ParameterKey=DBPassword,ParameterValue="$DBPassword" `
                ParameterKey=DBName,ParameterValue="$DBName" `
                ParameterKey=JWTSecret,ParameterValue="$JWTSecret" `
                ParameterKey=CreateECRRepositories,ParameterValue="$createECR" `
                ParameterKey=BackendImageURI,ParameterValue="$accountId.dkr.ecr.us-east-1.amazonaws.com/hub-hrms-backend:latest" `
                ParameterKey=FrontendImageURI,ParameterValue="$accountId.dkr.ecr.us-east-1.amazonaws.com/hub-hrms-frontend:latest" `
            --capabilities CAPABILITY_NAMED_IAM `
            --region $Region
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to initiate stack operation"
            Remove-Item $parametersFile -ErrorAction SilentlyContinue
            exit 1
        }
        
        Write-Success "CloudFormation stack deployment initiated"
    } finally {
        Remove-Item $parametersFile -ErrorAction SilentlyContinue
    }
}

# Function to wait for stack completion
function Wait-ForStack {
    Write-Info "Waiting for stack to complete..."
    Write-Info "This may take 10-15 minutes..."
    
    $startTime = Get-Date
    $timeout = 30 # minutes
    
    while ($true) {
        Start-Sleep -Seconds 30
        
        $stack = aws cloudformation describe-stacks `
            --stack-name $StackName `
            --region $Region `
            --query 'Stacks[0].StackStatus' `
            --output text 2>&1
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to get stack status"
            exit 1
        }
        
        $stack = $stack.Trim()
        
        Write-Host "Current status: $stack" -ForegroundColor Cyan
        
        if ($stack -match "COMPLETE$") {
            Write-Success "Stack operation completed successfully"
            break
        } elseif ($stack -match "FAILED$" -or $stack -match "ROLLBACK") {
            Write-Error "Stack operation failed"
            Write-Info "Check CloudFormation console for details"
            exit 1
        }
        
        $elapsed = (Get-Date) - $startTime
        if ($elapsed.TotalMinutes -gt $timeout) {
            Write-Error "Stack operation timed out after $timeout minutes"
            exit 1
        }
    }
}

# Function to get stack outputs
function Get-StackOutputs {
    Write-Info "Retrieving stack outputs..."
    
    $outputs = aws cloudformation describe-stacks `
        --stack-name $StackName `
        --region $Region `
        --query 'Stacks[0].Outputs[*].[OutputKey,OutputValue]' `
        --output text
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to get stack outputs"
        exit 1
    }
    
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "Stack Outputs" -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host $outputs
    Write-Host ""
    
    # Get and display the application URL
    $appUrl = aws cloudformation describe-stacks `
        --stack-name $StackName `
        --region $Region `
        --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' `
        --output text
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Application URL: $appUrl"
    }
}

# Function to display next steps
function Show-NextSteps {
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Green
    Write-Host "Deployment Complete!" -ForegroundColor Green
    Write-Host "======================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "1. Wait a few minutes for ECS services to start"
    Write-Host "2. Access your application at the URL above"
    Write-Host "3. Initialize the database with migrations"
    Write-Host "4. Create an admin user using scripts in backend/scripts/"
    Write-Host ""
    Write-Host "For database initialization, run: .\init-database.ps1"
    Write-Host "For detailed instructions, see: DEPLOYMENT_GUIDE.md"
    Write-Host ""
}

# Main deployment flow
function Main {
    Write-Info "Starting HRMS deployment..."
    Write-Info "Stack Name: $StackName"
    Write-Info "Region: $Region"
    Write-Host ""
    
    try {
        Test-Prerequisites
        
        if (-not $SkipImages) {
            New-ECRRepositories
            Build-AndPushImages
        } else {
            Write-Warning "Skipping image build and push"
        }
        
        Deploy-Stack
        Wait-ForStack
        Get-StackOutputs
        Show-NextSteps
        
        Write-Success "Deployment completed successfully!"
    } catch {
        Write-Error "Deployment failed: $_"
        exit 1
    }
}

# Run main function
Main
