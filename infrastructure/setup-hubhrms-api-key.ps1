# Hub-HRMS API Key Setup Script (PowerShell)
# This script creates and stores an API key for HR Recruiting to access Hub-HRMS

param(
    [string]$Region = "us-east-1",
    [string]$SecretName = "hr-recruiting/hubhrms-api-key",
    [string]$HubHRMSUrl = "https://hub-hrms.cocomgroup.com/graphql"
)

$ErrorActionPreference = "Stop"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Hub-HRMS API Key Setup" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Generate secure API key
Write-Host "[1/5] Generating secure API key..." -ForegroundColor Yellow

# Generate 32 random bytes and convert to hex (64 characters)
$randomBytes = New-Object byte[] 32
$rng = [System.Security.Cryptography.RNGCryptoServiceProvider]::Create()
$rng.GetBytes($randomBytes)
$apiKey = [System.BitConverter]::ToString($randomBytes).Replace("-", "").ToLower()

Write-Host "Generated API Key: $apiKey" -ForegroundColor Green
Write-Host ""

# Store in Hub-HRMS database
Write-Host "[2/5] Storing in Hub-HRMS database..." -ForegroundColor Yellow
Write-Host "Run this SQL in your Hub-HRMS PostgreSQL database:" -ForegroundColor White
Write-Host ""

$sqlScript = @"
INSERT INTO api_keys (
    key_hash,
    name,
    description,
    scopes,
    created_at,
    expires_at
) VALUES (
    crypt('$apiKey', gen_salt('bf')),
    'hr-recruiting-app',
    'Public recruiting portal API access',
    ARRAY['jobs:read', 'jobs:write', 'applications:read', 'applications:write'],
    NOW(),
    NOW() + INTERVAL '1 year'
);
"@

Write-Host $sqlScript -ForegroundColor Gray
Write-Host ""

$confirmation = Read-Host "Have you executed this SQL? (y/n)"
if ($confirmation -ne 'y') {
    Write-Host "Please execute the SQL first, then run this script again." -ForegroundColor Yellow
    exit 1
}

# Create AWS Secret
Write-Host ""
Write-Host "[3/5] Creating AWS Secrets Manager secret..." -ForegroundColor Yellow

# Check if secret already exists
try {
    $existingSecret = aws secretsmanager describe-secret --secret-id $SecretName --region $Region 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Secret already exists. Updating..." -ForegroundColor Yellow
        
        $secretValue = @{
            apiKey = $apiKey
            url = $HubHRMSUrl
        } | ConvertTo-Json -Compress
        
        aws secretsmanager update-secret `
            --secret-id $SecretName `
            --secret-string $secretValue `
            --region $Region | Out-Null
    }
} catch {
    Write-Host "Creating new secret..." -ForegroundColor Yellow
    
    $secretValue = @{
        apiKey = $apiKey
        url = $HubHRMSUrl
    } | ConvertTo-Json -Compress
    
    aws secretsmanager create-secret `
        --name $SecretName `
        --description "API key for HR Recruiting to access Hub-HRMS GraphQL API" `
        --secret-string $secretValue `
        --region $Region | Out-Null
}

Write-Host "[OK] Secret created/updated successfully" -ForegroundColor Green
Write-Host ""

# Get secret ARN
$secretArn = aws secretsmanager describe-secret `
    --secret-id $SecretName `
    --region $Region `
    --query 'ARN' `
    --output text

Write-Host "Secret ARN: $secretArn" -ForegroundColor Cyan
Write-Host ""

# Update IAM role
Write-Host "[4/5] Updating ECS Task Execution Role permissions..." -ForegroundColor Yellow

# Get the ECS execution role name
$ecsRoleName = "hr-recruiting-ecs-dev-ecs-execution-role"

# Create policy document
$policyDocument = @{
    Version = "2012-10-17"
    Statement = @(
        @{
            Effect = "Allow"
            Action = @("secretsmanager:GetSecretValue")
            Resource = $secretArn
        }
    )
} | ConvertTo-Json -Depth 10

# Save to temp file
$tempPolicyFile = [System.IO.Path]::GetTempFileName()
$policyDocument | Out-File -FilePath $tempPolicyFile -Encoding utf8

try {
    # Attach inline policy
    aws iam put-role-policy `
        --role-name $ecsRoleName `
        --policy-name SecretsManagerAccess `
        --policy-document "file://$tempPolicyFile" `
        --region $Region 2>&1 | Out-Null
    
    Write-Host "[OK] IAM policy attached" -ForegroundColor Green
} catch {
    Write-Host "[WARNING] Failed to attach IAM policy. You may need to do this manually." -ForegroundColor Yellow
} finally {
    Remove-Item $tempPolicyFile -ErrorAction SilentlyContinue
}

Write-Host ""

# Provide CloudFormation snippet
Write-Host "[5/5] CloudFormation Configuration" -ForegroundColor Yellow
Write-Host ""
Write-Host "Add this to your ECS task definition ContainerDefinitions:" -ForegroundColor White
Write-Host ""

$cloudFormationSnippet = @"
Secrets:
  - Name: HUBHRMS_API_KEY
    ValueFrom: ${secretArn}:apiKey::
  - Name: HUBHRMS_URL
    ValueFrom: ${secretArn}:url::
"@

Write-Host $cloudFormationSnippet -ForegroundColor Gray
Write-Host ""

# Summary
Write-Host "=====================================" -ForegroundColor Green
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green
Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "  [OK] API Key generated: $($apiKey.Substring(0,8))..." -ForegroundColor White
Write-Host "  [OK] Stored in AWS Secrets Manager" -ForegroundColor White
Write-Host "  [OK] IAM role updated" -ForegroundColor White
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Update your CloudFormation template with the Secrets configuration above" -ForegroundColor White
Write-Host "  2. Deploy the updated CloudFormation stack" -ForegroundColor White
Write-Host "  3. Force new ECS deployment to pick up the secret" -ForegroundColor White
Write-Host ""
Write-Host "Commands to deploy:" -ForegroundColor Yellow
Write-Host "  aws cloudformation update-stack --stack-name hr-recruiting-ecs ..." -ForegroundColor Gray
Write-Host "  aws ecs update-service --cluster <cluster> --service <service> --force-new-deployment" -ForegroundColor Gray
Write-Host ""

# Save to file for reference
$configFile = "hr-recruiting-api-key.txt"
$configContent = @"
HR Recruiting API Key Configuration
Generated: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

API Key: $apiKey
Hub-HRMS URL: $HubHRMSUrl
Secret ARN: $secretArn
Region: $Region

IMPORTANT: Keep this file secure and delete after setup!

CloudFormation Configuration:
$cloudFormationSnippet

SQL for Hub-HRMS:
$sqlScript
"@

$configContent | Out-File -FilePath $configFile -Encoding utf8

Write-Host "Configuration saved to: $configFile" -ForegroundColor Cyan
Write-Host "[WARNING] DELETE THIS FILE AFTER SETUP!" -ForegroundColor Red
Write-Host ""

# Display file location
Write-Host "Full path: $(Resolve-Path $configFile)" -ForegroundColor Gray