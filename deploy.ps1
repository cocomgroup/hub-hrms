# Hub-HRMS CloudFormation Deployment Script (PowerShell)
# VERSION: CORRECTED-2024-01-12-v5
# Automated deployment to AWS with Docker image building and pushing

[CmdletBinding()]
param(
    [string]$StackName = "hub-hrms-dev",
    [string]$Region = "us-east-1",
    [switch]$SkipImages,
    [switch]$SkipBuild,
    [string]$JWTSecret,
    [string]$BankInfoKey,
    [string]$DBPassword,
    [string]$CertificateArn = "",
    [string]$DomainName = "",
    [switch]$Help,
    [switch]$Update,
    [switch]$Delete,
    [switch]$Status
)

# Configuration
$DBUsername = "postgres"
$DBName = "hrmsdb"
$Environment = "development"

# Derived values
$BackendECRRepo = "$StackName-backend"
$FrontendECRRepo = "$StackName-frontend"
$CloudFormationTemplate = "cloudformation-stack.yaml"

# Helper Functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
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

function Write-Step {
    param([string]$Message)
    Write-Host ""
    Write-Host "==================================================" -ForegroundColor Blue
    Write-Host "  $Message" -ForegroundColor Blue
    Write-Host "==================================================" -ForegroundColor Blue
}

function Show-Help {
    Write-Host ""
    Write-Host "Hub-HRMS CloudFormation Deployment Script"
    Write-Host "=========================================="
    Write-Host ""
    Write-Host "Usage: .\deploy.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -StackName NAME      Stack name (default: hub-hrms-dev)"
    Write-Host "  -Region REGION       AWS region (default: us-east-1)"
    Write-Host "  -SkipImages          Skip building and pushing Docker images"
    Write-Host "  -SkipBuild           Skip building images"
    Write-Host "  -JWTSecret SECRET    JWT secret"
    Write-Host "  -BankInfoKey KEY     32-character encryption key"
    Write-Host "  -DBPassword PASS     Database password"
    Write-Host "  -CertificateArn ARN  ACM certificate ARN for HTTPS"
    Write-Host "  -DomainName DOMAIN   Domain name"
    Write-Host "  -Update              Update existing stack"
    Write-Host "  -Delete              Delete stack"
    Write-Host "  -Status              Show stack status"
    Write-Host "  -Help                Show this help"
    Write-Host ""
    exit 0
}

function Test-Prerequisites {
    Write-Step "Checking Prerequisites"
    
    $missing = @()
    
    # Check AWS CLI
    try {
        $awsVersion = aws --version 2>&1
        Write-Success "AWS CLI found: $awsVersion"
    } catch {
        $missing += "AWS CLI"
    }
    
    # Check Docker (only if not skipping images)
    if (-not $SkipImages -and -not $SkipBuild) {
        try {
            $dockerVersion = docker --version 2>&1
            Write-Success "Docker found: $dockerVersion"
        } catch {
            $missing += "Docker"
        }
    }
    
    # Check AWS credentials
    try {
        $identity = aws sts get-caller-identity 2>&1 | ConvertFrom-Json
        Write-Success "AWS Account: $($identity.Account)"
        Write-Success "AWS User: $($identity.Arn)"
    } catch {
        $missing += "AWS credentials"
    }
    
    if ($missing.Count -gt 0) {
        Write-Error "Missing prerequisites: $($missing -join ', ')"
        Write-Host "'nInstall missing components:"
        if ($missing -contains "AWS CLI") {
            Write-Host "  AWS CLI: https://aws.amazon.com/cli/"
        }
        if ($missing -contains "Docker") {
            Write-Host "  Docker: https://www.docker.com/products/docker-desktop"
        }
        if ($missing -contains "AWS credentials") {
            Write-Host "  Run: aws configure"
        }
        exit 1
    }
    
    Write-Success "All prerequisites met!"
}

function Get-RandomString {
    param([int]$Length = 32)
    $chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    -join ((1..$Length) | ForEach-Object { $chars[(Get-Random -Maximum $chars.Length)] })
}

function New-ECRRepository {
    param([string]$RepoName)
    
    Write-Info "Checking ECR repository: $RepoName"
    
    try {
        $repo = aws ecr describe-repositories --repository-names $RepoName --region $Region 2>&1 | ConvertFrom-Json
        Write-Success "Repository exists: $RepoName"
        return $repo.repositories[0].repositoryUri
    } catch {
        Write-Info "Creating ECR repository: $RepoName"
        $repo = aws ecr create-repository --repository-name $RepoName --region $Region --image-scanning-configuration scanOnPush=true | ConvertFrom-Json
        Write-Success "Repository created: $RepoName"
        return $repo.repository.repositoryUri
    }
}

function Build-AndPushImage {
    param(
        [string]$ImageName,
        [string]$DockerfilePath,
        [string]$ContextPath,
        [string]$ECRUri
    )
    
    Write-Info "Building $ImageName image..."
    
    # Build image (redirect output to null to avoid capturing in return value)
    docker build -t $ImageName -f $DockerfilePath $ContextPath 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build $ImageName image"
        exit 1
    }
    Write-Success "Built $ImageName image"
    
    # Tag for ECR
    $imageTag = "${ECRUri}:latest"
    docker tag $ImageName $imageTag 2>&1 | Out-Null
    Write-Success "Tagged as $imageTag"
    
    # Login to ECR
    Write-Info "Logging into ECR..."
    $loginCmd = aws ecr get-login-password --region $Region
    $loginCmd | docker login --username AWS --password-stdin $ECRUri.Split('/')[0] 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to login to ECR"
        exit 1
    }
    
    # Push to ECR
    Write-Info "Pushing $ImageName to ECR..."
    docker push $imageTag 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to push $ImageName image"
        exit 1
    }
    Write-Success "Pushed $ImageName to ECR"
    
    # Explicitly return only the image tag
    Write-Output $imageTag
}

function Deploy-Stack {
    param(
        [string]$Action = "create"
    )
    
    Write-Step "$(if ($Action -eq 'create') { 'Deploying' } else { 'Updating' }) CloudFormation Stack"
    
    # Build parameters as JSON array
    $parametersArray = @(
        @{
            ParameterKey = "Environment"
            ParameterValue = $Environment
        },
        @{
            ParameterKey = "DBUsername"
            ParameterValue = $DBUsername
        },
        @{
            ParameterKey = "DBPassword"
            ParameterValue = $script:DBPassword
        },
        @{
            ParameterKey = "DBName"
            ParameterValue = $DBName
        },
        @{
            ParameterKey = "JWTSecret"
            ParameterValue = $script:JWTSecret
        },
        @{
            ParameterKey = "BankInfoEncryptionKey"
            ParameterValue = $script:BankInfoKey
        }
    )
    
    if ($script:BackendImageURI) {
        $parametersArray += @{
            ParameterKey = "BackendImageURI"
            ParameterValue = $script:BackendImageURI
        }
    }
    
    if ($script:FrontendImageURI) {
        $parametersArray += @{
            ParameterKey = "FrontendImageURI"
            ParameterValue = $script:FrontendImageURI
        }
    }
    
    if ($CertificateArn) {
        $parametersArray += @{
            ParameterKey = "CertificateArn"
            ParameterValue = $CertificateArn
        }
    }
    
    if ($DomainName) {
        $parametersArray += @{
            ParameterKey = "DomainName"
            ParameterValue = $DomainName
        }
    }
    
    # Write parameters to temporary JSON file (without BOM)
    $tempParamsFile = "cf-params-temp.json"
    $jsonContent = $parametersArray | ConvertTo-Json -Depth 10
    [System.IO.File]::WriteAllText($tempParamsFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))
    
    # Create or update stack
    $verb = if ($Action -eq "create") { "create-stack" } else { "update-stack" }
    
    try {
        $stackId = aws cloudformation $verb `
            --stack-name $StackName `
            --template-body file://$CloudFormationTemplate `
            --parameters file://$tempParamsFile `
            --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM `
            --region $Region `
            --output json 2>&1
        
        if ($LASTEXITCODE -ne 0) {
            if ($stackId -match "No updates are to be performed") {
                Write-Warning "No changes to deploy"
                return $true
            }
            throw $stackId
        }
        
        $stackIdJson = $stackId | ConvertFrom-Json
        Write-Success "Stack $Action initiated: $($stackIdJson.StackId)"
        
        # Wait for stack to complete
        Write-Info "Waiting for stack to complete (this may take 10-15 minutes)..."
        Write-Info "You can check progress in the AWS Console:"
        Write-Host "  https://console.aws.amazon.com/cloudformation/home?region=$Region#/stacks" -ForegroundColor Blue
        
        $waitEvent = if ($Action -eq "create") { "stack-create-complete" } else { "stack-update-complete" }
        
        aws cloudformation wait $waitEvent --stack-name $StackName --region $Region
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Stack $Action completed successfully!"
            # Clean up temp params file
            if (Test-Path $tempParamsFile) {
                Remove-Item $tempParamsFile -Force
            }
            return $true
        } else {
            Write-Error "Stack $Action failed or timed out"
            Show-StackEvents
            # Clean up temp params file
            if (Test-Path $tempParamsFile) {
                Remove-Item $tempParamsFile -Force
            }
            return $false
        }
        
    } catch {
        Write-Error "Failed to $Action stack: $_"
        Show-StackEvents
        # Clean up temp params file
        if (Test-Path $tempParamsFile) {
            Remove-Item $tempParamsFile -Force
        }
        return $false
    }
}

function Show-StackEvents {
    Write-Info "Recent stack events:"
    $events = aws cloudformation describe-stack-events --stack-name $StackName --region $Region --max-items 10 2>&1
    if ($LASTEXITCODE -eq 0) {
        $events | ConvertFrom-Json | Select-Object -ExpandProperty StackEvents | 
            Format-Table Timestamp, ResourceStatus, ResourceType, ResourceStatusReason -AutoSize
    }
}

function Get-StackOutputs {
    Write-Step "Stack Outputs"
    
    try {
        $stack = aws cloudformation describe-stacks --stack-name $StackName --region $Region | ConvertFrom-Json
        $outputs = $stack.Stacks[0].Outputs
        
        Write-Host "'nStack Outputs:" -ForegroundColor Green
        Write-Host "=" * 60 -ForegroundColor Green
        
        foreach ($output in $outputs) {
            Write-Host "$($output.OutputKey):" -ForegroundColor Cyan -NoNewline
            Write-Host " $($output.OutputValue)" -ForegroundColor White
            if ($output.Description) {
                Write-Host "  $($output.Description)" -ForegroundColor Gray
            }
        }
        
        Write-Host "=" * 60 -ForegroundColor Green
        
        # Extract and display application URL
        $appURL = ($outputs | Where-Object { $_.OutputKey -eq "ApplicationURL" }).OutputValue
        if ($appURL) {
            Write-Host "'n Application URL: $appURL" -ForegroundColor Yellow
            Write-Host "'n Deployment complete! Your Hub-HRMS application is ready." -ForegroundColor Green
        }
        
    } catch {
        Write-Warning "Could not retrieve stack outputs: $_"
    }
}

function Show-StackStatus {
    try {
        $stack = aws cloudformation describe-stacks --stack-name $StackName --region $Region | ConvertFrom-Json
        $stackInfo = $stack.Stacks[0]
        
        Write-Host "'nStack Status" -ForegroundColor Cyan
        Write-Host "=" * 60
        Write-Host "Stack Name:   $($stackInfo.StackName)"
        Write-Host "Status:       $($stackInfo.StackStatus)" -ForegroundColor $(
            switch ($stackInfo.StackStatus) {
                { $_ -match "COMPLETE" } { "Green" }
                { $_ -match "PROGRESS" } { "Yellow" }
                { $_ -match "FAILED" } { "Red" }
                default { "White" }
            }
        )
        Write-Host "Created:      $($stackInfo.CreationTime)"
        if ($stackInfo.LastUpdatedTime) {
            Write-Host "Last Updated: $($stackInfo.LastUpdatedTime)"
        }
        Write-Host "=" * 60
        
        Get-StackOutputs
        
    } catch {
        Write-Error "Stack not found: $StackName"
        exit 1
    }
}

function Remove-Stack {
    Write-Warning "This will delete the entire stack including:"
    Write-Host "  - ECS Cluster and Services"
    Write-Host "  - RDS Database (snapshot will be created)"
    Write-Host "  - Load Balancer and Target Groups"
    Write-Host "  - VPC and Networking"
    Write-Host "  - S3 Buckets (must be empty)"
    Write-Host "  - DynamoDB Tables"
    
    $confirmation = Read-Host "'nAre you sure you want to delete stack '$StackName'? (yes/no)"
    
    if ($confirmation -ne "yes") {
        Write-Info "Deletion cancelled"
        exit 0
    }
    
    Write-Info "Deleting stack: $StackName"
    
    try {
        aws cloudformation delete-stack --stack-name $StackName --region $Region
        Write-Success "Stack deletion initiated"
        
        Write-Info "Waiting for stack deletion to complete..."
        aws cloudformation wait stack-delete-complete --stack-name $StackName --region $Region
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Stack deleted successfully!"
        } else {
            Write-Error "Stack deletion failed or timed out"
            Show-StackEvents
        }
    } catch {
        Write-Error "Failed to delete stack: $_"
    }
}

function Save-Secrets {
    param(
        [string]$JWTSecret,
        [string]$BankInfoKey,
        [string]$DBPassword
    )
    
    $secretsFile = "secrets-$StackName.txt"
    $dateStr = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $content = "Hub-HRMS Deployment Secrets`n"
    $content += "Generated: $dateStr`n"
    $content += "Stack: $StackName`n"
    $content += "Region: $Region`n`n"
    $content += "WARNING: KEEP THIS FILE SECURE - DO NOT COMMIT TO VERSION CONTROL`n`n"
    $content += "JWT_SECRET=$JWTSecret`n"
    $content += "BANK_INFO_ENCRYPTION_KEY=$BankInfoKey`n"
    $content += "DB_PASSWORD=$DBPassword`n"
    
    $content | Out-File -FilePath $secretsFile -Encoding UTF8
    Write-Success "Secrets saved to: $secretsFile"
    Write-Warning "Keep this file secure and do not commit to version control!"
}

# ==========================================
# Main Script
# ==========================================

if ($Help) {
    Show-Help
}

if ($Status) {
    Show-StackStatus
    exit 0
}

if ($Delete) {
    Remove-Stack
    exit 0
}

# Check prerequisites
Test-Prerequisites

# Generate secrets if not provided
if (-not $JWTSecret) {
    $script:JWTSecret = Get-RandomString -Length 64
    Write-Info "Generated JWT secret (64 characters)"
} else {
    if ($JWTSecret.Length -lt 32) {
        Write-Error "JWT secret must be at least 32 characters"
        exit 1
    }
    $script:JWTSecret = $JWTSecret
}

if (-not $BankInfoKey) {
    $script:BankInfoKey = Get-RandomString -Length 32
    Write-Info "Generated bank info encryption key (32 characters)"
} else {
    if ($BankInfoKey.Length -ne 32) {
        Write-Error "Bank info encryption key must be exactly 32 characters"
        exit 1
    }
    $script:BankInfoKey = $BankInfoKey
}

if (-not $DBPassword) {
    $script:DBPassword = Get-RandomString -Length 16
    Write-Info "Generated database password (16 characters)"
} else {
    if ($DBPassword.Length -lt 8) {
        Write-Error "Database password must be at least 8 characters"
        exit 1
    }
    $script:DBPassword = $DBPassword
}

# Save secrets
Save-Secrets -JWTSecret $script:JWTSecret -BankInfoKey $script:BankInfoKey -DBPassword $script:DBPassword

# Build and push Docker images
if (-not $SkipImages -and -not $SkipBuild) {
    Write-Step "Building and Pushing Docker Images"
    
    # Create ECR repositories
    $backendECRUri = New-ECRRepository -RepoName $BackendECRRepo
    $frontendECRUri = New-ECRRepository -RepoName $FrontendECRRepo
    
    # Check if backend and frontend directories exist
    if (-not (Test-Path "backend")) {
        Write-Error "Backend directory not found. Extract backend.tar first."
        exit 1
    }
    
    if (-not (Test-Path "frontend")) {
        Write-Error "Frontend directory not found. Extract frontend.tar first."
        exit 1
    }
    
    # Build and push backend
    $script:BackendImageURI = Build-AndPushImage `
        -ImageName "hub-hrms-backend" `
        -DockerfilePath "backend/Dockerfile" `
        -ContextPath "backend" `
        -ECRUri $backendECRUri
    
    # Build and push frontend
    $script:FrontendImageURI = Build-AndPushImage `
        -ImageName "hub-hrms-frontend" `
        -DockerfilePath "frontend/Dockerfile" `
        -ContextPath "frontend" `
        -ECRUri $frontendECRUri
    
    Write-Success "All images built and pushed successfully!"
} elseif ($SkipBuild) {
    Write-Info "Skipping image build - using existing images in ECR"
    
    # Get existing image URIs
    $accountId = (aws sts get-caller-identity --query Account --output text)
    $script:BackendImageURI = "$accountId.dkr.ecr.$Region.amazonaws.com/${BackendECRRepo}:latest"
    $script:FrontendImageURI = "$accountId.dkr.ecr.$Region.amazonaws.com/${FrontendECRRepo}:latest"
    
    Write-Info "Using backend image: $script:BackendImageURI"
    Write-Info "Using frontend image: $script:FrontendImageURI"
}

# Deploy CloudFormation stack
if ($Update) {
    $success = Deploy-Stack -Action "update"
} else {
    # Check if stack already exists
    try {
        aws cloudformation describe-stacks --stack-name $StackName --region $Region 2>&1 | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Warning "Stack already exists. Use -Update to update it."
            exit 1
        }
    } catch {}
    
    $success = Deploy-Stack -Action "create"
}

if ($success) {
    Get-StackOutputs
    
    Write-Host "`n" + ("=" * 60) -ForegroundColor Green
    Write-Host "   Hub-HRMS Deployment Complete! " -ForegroundColor Green
    Write-Host ("=" * 60) -ForegroundColor Green
    
    Write-Host "`nNext Steps:" -ForegroundColor Cyan
    Write-Host "  1. Wait 3-5 minutes for services to fully start and register with load balancer"
    Write-Host "  2. Backend will run database migrations (check CloudWatch logs if needed)"
    Write-Host "  3. Visit your application URL above once all services are healthy"
    Write-Host "  4. Set up DNS if using custom domain"
    
    if ($CertificateArn) {
        Write-Host "`n✅ HTTPS is enabled with your SSL certificate" -ForegroundColor Green
    } else {
        Write-Host "`n⚠️  WARNING: Consider adding SSL certificate for production use" -ForegroundColor Yellow
    }
    
    Write-Host "`n🔐 Secrets saved in: secrets-$StackName.txt" -ForegroundColor Yellow
    Write-Host "   Keep this file secure!" -ForegroundColor Yellow
    
    Write-Host "`n📊 Monitor deployment status:" -ForegroundColor Cyan
    Write-Host "   Backend service: aws ecs describe-services --cluster $StackName-cluster --services $StackName-backend --query 'services[0].{Status:status,Running:runningCount,Desired:desiredCount}'" -ForegroundColor Gray
    Write-Host "   Frontend service: aws ecs describe-services --cluster $StackName-cluster --services $StackName-frontend --query 'services[0].{Status:status,Running:runningCount,Desired:desiredCount}'" -ForegroundColor Gray
    Write-Host "   Backend logs: aws logs tail /ecs/$StackName/backend --follow" -ForegroundColor Gray
}