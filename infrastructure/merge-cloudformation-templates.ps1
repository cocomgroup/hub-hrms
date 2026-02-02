# Merge CloudFormation Templates
# Combines base Hub-HRMS template with Bastion/SES additions

param(
    [Parameter(Mandatory=$true)]
    [string]$BaseTemplate,
    
    [Parameter(Mandatory=$false)]
    [string]$OutputFile = "cloudformation-stack-complete.yaml"
)

$ErrorActionPreference = "Stop"

Write-Host "`n=== CloudFormation Template Merger ===" -ForegroundColor Cyan
Write-Host ""

# Read the base template
Write-Host "[*] Reading base template: $BaseTemplate" -ForegroundColor Yellow
$baseContent = Get-Content $BaseTemplate -Raw

if (-not $baseContent) {
    Write-Host "[ERROR] Could not read base template" -ForegroundColor Red
    exit 1
}

# Check if it's already merged
if ($baseContent -match "BastionKeyName" -or $baseContent -match "SESFromEmail") {
    Write-Host "[WARNING] Template appears to already have Bastion/SES parameters" -ForegroundColor Yellow
    $continue = Read-Host "Continue anyway? (y/n)"
    if ($continue -ne 'y') {
        exit 0
    }
}

Write-Host "[OK] Base template loaded" -ForegroundColor Green

# Define new parameters to add
$newParameters = @"

  # Bastion Host Configuration
  BastionKeyName:
    Type: AWS::EC2::KeyPair::KeyName
    Description: EC2 Key Pair for SSH access to bastion host
    
  BastionInstanceType:
    Type: String
    Default: t3.micro
    AllowedValues:
      - t3.micro
      - t3.small
      - t3.medium
    Description: Bastion host instance type
    
  BastionAllowedCIDR:
    Type: String
    Default: 0.0.0.0/0
    Description: CIDR block allowed to SSH into bastion (restrict to your IP for security)
    AllowedPattern: (\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})/(\d{1,2})
    ConstraintDescription: Must be a valid CIDR block (e.g., 1.2.3.4/32)
    
  EnableBastion:
    Type: String
    Default: 'true'
    AllowedValues: ['true', 'false']
    Description: Enable bastion host for database access
    
  # SES Configuration
  SESFromEmail:
    Type: String
    Description: Email address to send emails from (must be verified in SES)
    Default: noreply@yourdomain.com
    
  SESReplyToEmail:
    Type: String
    Description: Reply-to email address
    Default: support@yourdomain.com
    
  SESConfigurationSetName:
    Type: String
    Default: hub-hrms-emails
    Description: SES configuration set name
    
  EnableSES:
    Type: String
    Default: 'true'
    AllowedValues: ['true', 'false']
    Description: Enable SES email service
"@

# Add new conditions
$newConditions = @"

  UseBastion: !Equals [!Ref EnableBastion, 'true']
  UseSES: !Equals [!Ref EnableSES, 'true']
"@

# Add parameter groups to metadata
$newParameterGroups = @"
      - Label:
          default: Bastion Host Configuration
        Parameters:
          - EnableBastion
          - BastionKeyName
          - BastionInstanceType
          - BastionAllowedCIDR
      - Label:
          default: Email Configuration
        Parameters:
          - EnableSES
          - SESFromEmail
          - SESReplyToEmail
          - SESConfigurationSetName
"@

Write-Host "[*] Adding parameters..." -ForegroundColor Yellow

# Find where to insert parameters (before Conditions:)
if ($baseContent -match "(?s)(Parameters:.*?)(Conditions:)") {
    $baseContent = $baseContent -replace "(Parameters:.*?)(Conditions:)", "`$1$newParameters`n`n`$2"
    Write-Host "[OK] Parameters added" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Could not find Parameters section" -ForegroundColor Red
    exit 1
}

Write-Host "[*] Adding conditions..." -ForegroundColor Yellow

# Add conditions (before Metadata:)
if ($baseContent -match "(?s)(Conditions:.*?)(Metadata:)") {
    $baseContent = $baseContent -replace "(Conditions:.*?)(Metadata:)", "`$1$newConditions`n`n`$2"
    Write-Host "[OK] Conditions added" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Could not find Conditions section" -ForegroundColor Red
    exit 1
}

Write-Host "[*] Adding parameter groups to metadata..." -ForegroundColor Yellow

# Add to parameter groups (before Parameters section in Metadata)
if ($baseContent -match "(?s)(ParameterGroups:.*?)(ParameterLabels:|Parameters:)") {
    $baseContent = $baseContent -replace "(ParameterGroups:)", "`$1`n$newParameterGroups"
    Write-Host "[OK] Parameter groups added" -ForegroundColor Green
}

# Read the bastion/SES resources
Write-Host "[*] Reading Bastion/SES resources..." -ForegroundColor Yellow
$resourcesContent = Get-Content "cloudformation-additions-bastion-ses.yaml" -Raw

# Extract just the resources section (everything between "# Bastion Host" and "# OUTPUTS")
if ($resourcesContent -match "(?s)(# =+ Bastion Host.*?)(# =+\s*OUTPUTS)") {
    $bastionSESResources = $matches[1]
    
    # Remove the "# RESOURCES TO ADD" header
    $bastionSESResources = $bastionSESResources -replace "(?m)^  # ==========================================", "  # =========================================="
    
    Write-Host "[OK] Resources extracted" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Could not extract resources from additions file" -ForegroundColor Red
    exit 1
}

Write-Host "[*] Adding resources..." -ForegroundColor Yellow

# Add resources (before Outputs:)
if ($baseContent -match "(?s)(Resources:.*?)(Outputs:)") {
    $baseContent = $baseContent -replace "(Resources:.*?)(Outputs:)", "`$1`n$bastionSESResources`n`n`$2"
    Write-Host "[OK] Resources added" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Could not find Resources section" -ForegroundColor Red
    exit 1
}

# Extract outputs section
if ($resourcesContent -match "(?s)(# OUTPUTS TO ADD.*?)(BastionHostPublicIP:.*?$)") {
    $newOutputs = $matches[2]
    
    # Get everything up to the last output
    if ($resourcesContent -match "(?s)(BastionHostPublicIP:.*SESDashboardURL:.*?https://console\.aws\.amazon\.com.*?')") {
        $newOutputs = $matches[1]
        
        Write-Host "[*] Adding outputs..." -ForegroundColor Yellow
        
        # Add outputs at the end
        $baseContent = $baseContent + "`n" + $newOutputs
        Write-Host "[OK] Outputs added" -ForegroundColor Green
    }
}

# Save the merged template
Write-Host "[*] Saving merged template..." -ForegroundColor Yellow
$baseContent | Out-File -FilePath $OutputFile -Encoding utf8

Write-Host "[OK] Template saved: $OutputFile" -ForegroundColor Green

# Validate the template
Write-Host "`n[*] Validating template..." -ForegroundColor Yellow

try {
    aws cloudformation validate-template `
        --template-body file://$OutputFile `
        --region us-east-1 | Out-Null
    
    Write-Host "[OK] Template is valid!" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Template validation failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}

# Summary
Write-Host "`n=== Merge Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Output file: $OutputFile" -ForegroundColor Cyan
Write-Host ""
Write-Host "What was added:" -ForegroundColor Yellow
Write-Host "  - 8 new parameters (Bastion + SES configuration)" -ForegroundColor White
Write-Host "  - 2 new conditions (UseBastion, UseSES)" -ForegroundColor White
Write-Host "  - 15 new resources (Bastion host + SES email service)" -ForegroundColor White
Write-Host "  - 7 new outputs (Bastion IP, SSH commands, SES details)" -ForegroundColor White
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Review the generated file: $OutputFile"  -ForegroundColor White
Write-Host "  2. Update parameters.json with new parameters" -ForegroundColor White
Write-Host "  3. Deploy with:" -ForegroundColor White
Write-Host "     aws cloudformation update-stack --stack-name hub-hrms-dev --template-body file://$OutputFile --parameters file://parameters.json --capabilities CAPABILITY_NAMED_IAM" -ForegroundColor Gray
Write-Host ""