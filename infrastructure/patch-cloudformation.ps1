# CloudFormation Template Auto-Patcher
# This script automatically fixes the CloudFormation template

param(
    [Parameter(Mandatory=$false)]
    [string]$TemplateFile = "cloudformation-stack.yaml",
    
    [Parameter(Mandatory=$false)]
    [string]$OutputFile = "cloudformation-stack-fixed.yaml"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "CloudFormation Template Auto-Patcher" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $TemplateFile)) {
    Write-Host "ERROR: Template file not found: $TemplateFile" -ForegroundColor Red
    exit 1
}

Write-Host "Reading template: $TemplateFile" -ForegroundColor Yellow
$content = Get-Content $TemplateFile -Raw

# Fix 1: Change ragdb to hrmsdb
Write-Host "Fix 1: Changing database name ragdb -> hrmsdb..." -ForegroundColor Cyan
$content = $content -replace 'ragdb', 'hrmsdb'
Write-Host "  Done: Database name fixed" -ForegroundColor Green

# Fix 2: Add JWT Secret after DatabaseSecret
Write-Host "Fix 2: Adding JWT Secret resource..." -ForegroundColor Cyan
$jwtSecretBlock = @"

  JWTSecret:
    Type: AWS::SecretsManager::Secret
    Properties:
      Name: !Sub `${EnvironmentName}/jwt-secret
      Description: JWT signing secret for Hub HRMS
      GenerateSecretString:
        SecretStringTemplate: '{}'
        GenerateStringKey: "secret"
        PasswordLength: 64
        ExcludeCharacters: '"@/\\'
"@

# Find DatabaseSecret and add JWT Secret after it
if ($content -match '(?s)(  DatabaseSecret:.*?\n\n)(  #===)') {
    $content = $content -replace '(?s)(  DatabaseSecret:.*?\n\n)(  #===)', "`$1$jwtSecretBlock`n`n`$2"
    Write-Host "  Done: JWT Secret resource added" -ForegroundColor Green
} else {
    Write-Host "  Warning: Could not find DatabaseSecret section" -ForegroundColor Yellow
}

# Fix 3: Update IAM roles to include JWT Secret
Write-Host "Fix 3: Updating IAM roles..." -ForegroundColor Cyan

# Update ECSTaskExecutionRole
$content = $content -replace '(?s)(ECSTaskExecutionRole:.*?Resource:\s*- !Ref DatabaseSecret)', "`$1`n                  - !Ref JWTSecret"

# Update ECSTaskRole  
$content = $content -replace '(?s)(ECSTaskRole:.*?Resource:\s*- !Ref DatabaseSecret)(?!\s*- !Ref JWTSecret)', "`$1`n                  - !Ref JWTSecret"

Write-Host "  Done: IAM roles updated" -ForegroundColor Green

# Fix 4: Replace Environment and Secrets in BackendTaskDefinition
Write-Host "Fix 4: Updating Backend Task Definition..." -ForegroundColor Cyan

$oldEnvBlock = @'
          Environment:
            - Name: SERVER_ADDR
              Value: ':8080'
            - Name: GIN_MODE
              Value: release
          Secrets:
            - Name: DATABASE_URL
              ValueFrom: !Sub '\$\{DatabaseSecret\}:connectionString::'
'@

$newEnvBlock = @"
          Environment:
            - Name: SERVER_ADDR
              Value: ':8080'
            - Name: GIN_MODE
              Value: release
            - Name: PORT
              Value: '8080'
            - Name: ENVIRONMENT
              Value: 'production'
          Secrets:
            - Name: DB_HOST
              ValueFrom: !Sub '`${DatabaseSecret}:host::'
            - Name: DB_PORT
              ValueFrom: !Sub '`${DatabaseSecret}:port::'
            - Name: DB_NAME
              ValueFrom: !Sub '`${DatabaseSecret}:dbname::'
            - Name: DB_USER
              ValueFrom: !Sub '`${DatabaseSecret}:username::'
            - Name: DB_PASSWORD
              ValueFrom: !Sub '`${DatabaseSecret}:password::'
            - Name: JWT_SECRET
              ValueFrom: !Sub '`${JWTSecret}:secret::'
"@

$content = $content -replace [regex]::Escape($oldEnvBlock), $newEnvBlock
Write-Host "  Done: Backend environment variables updated" -ForegroundColor Green

# Fix 5: Add or update Outputs section
Write-Host "Fix 5: Adding Outputs section..." -ForegroundColor Cyan

$outputsSection = @"

Outputs:
  RDSEndpoint:
    Description: RDS Database Endpoint
    Value: !GetAtt Database.Endpoint.Address
    Export:
      Name: !Sub `${EnvironmentName}-RDS-Endpoint

  DatabaseSecretArn:
    Description: Database Secret ARN
    Value: !Ref DatabaseSecret
    Export:
      Name: !Sub `${EnvironmentName}-DB-Secret-ARN

  JWTSecretArn:
    Description: JWT Secret ARN
    Value: !Ref JWTSecret
    Export:
      Name: !Sub `${EnvironmentName}-JWT-Secret-ARN

  ALBDNSName:
    Description: Application Load Balancer DNS Name
    Value: !GetAtt ApplicationLoadBalancer.DNSName
    Export:
      Name: !Sub `${EnvironmentName}-ALB-DNS

  ApplicationURL:
    Description: Application URL
    Value: !Sub 'http://`${ApplicationLoadBalancer.DNSName}'

  BackendTargetGroupArn:
    Description: Backend Target Group ARN
    Value: !Ref BackendTargetGroup
    Export:
      Name: !Sub `${EnvironmentName}-Backend-TG

  ECSClusterName:
    Description: ECS Cluster Name
    Value: !Ref ECSCluster
    Export:
      Name: !Sub `${EnvironmentName}-ECS-Cluster

  BackendServiceName:
    Description: Backend ECS Service Name
    Value: !GetAtt BackendService.Name
    Export:
      Name: !Sub `${EnvironmentName}-Backend-Service
"@

# Check if Outputs section exists
if ($content -match 'Outputs:') {
    Write-Host "  Info: Outputs section already exists, skipping" -ForegroundColor Yellow
} else {
    $content += $outputsSection
    Write-Host "  Done: Outputs section added" -ForegroundColor Green
}

# Write the fixed template
Write-Host ""
Write-Host "Writing fixed template to: $OutputFile" -ForegroundColor Yellow
$content | Set-Content $OutputFile -NoNewline

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Template Fixed Successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Changes applied:" -ForegroundColor Yellow
Write-Host "  * Fixed database name (ragdb -> hrmsdb)" -ForegroundColor Green
Write-Host "  * Added JWT Secret resource" -ForegroundColor Green
Write-Host "  * Updated IAM roles for JWT Secret access" -ForegroundColor Green
Write-Host "  * Changed from DATABASE_URL to individual DB_* vars" -ForegroundColor Green
Write-Host "  * Added JWT_SECRET environment variable" -ForegroundColor Green
Write-Host "  * Added comprehensive Outputs section" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Review the fixed template:" -ForegroundColor White
Write-Host "   code $OutputFile" -ForegroundColor Cyan
Write-Host ""
Write-Host "2. Validate the template:" -ForegroundColor White
$validateCmd = "aws cloudformation validate-template --template-body file://$OutputFile"
Write-Host "   $validateCmd" -ForegroundColor Cyan
Write-Host ""
Write-Host "3. Update your stack:" -ForegroundColor White
Write-Host "   aws cloudformation update-stack --stack-name hub-hrms \`" -ForegroundColor Cyan
Write-Host "     --template-body file://$OutputFile \`" -ForegroundColor Cyan
Write-Host "     --capabilities CAPABILITY_NAMED_IAM \`" -ForegroundColor Cyan
Write-Host "     --parameters ParameterKey=DBPassword,UsePreviousValue=true \`" -ForegroundColor Cyan
Write-Host "     ParameterKey=BackendImageUri,UsePreviousValue=true \`" -ForegroundColor Cyan
Write-Host "     ParameterKey=FrontendImageUri,UsePreviousValue=true\`" -ForegroundColor Cyan
Write-Host ""