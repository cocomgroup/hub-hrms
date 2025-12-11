# Create hrmsdb Database via SSM
# This version doesn't rely on UserData signals

param(
    [Parameter(Mandatory=$false)]
    [string]$Region = 'us-east-1',
    
    [Parameter(Mandatory=$false)]
    [string]$DBHost = 'hub-doc-search-postgres.cqxc6o2kin1t.us-east-1.rds.amazonaws.com',
    
    [Parameter(Mandatory=$false)]
    [string]$DBPassword = 'postgresql123!',
    
    [Parameter(Mandatory=$false)]
    [string]$VpcId = 'vpc-0d52d043c7a7be124',
    
    [Parameter(Mandatory=$false)]
    [string]$SubnetId = 'subnet-0ec7d481181fa4ac6',
    
    [Parameter(Mandatory=$false)]
    [string]$RDSSecurityGroup = 'sg-0d88a4ede08f5f08c'
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Creating hrmsdb Database" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Create security group for bastion
Write-Host "Step 1: Creating security group..." -ForegroundColor Cyan

# Check if security group already exists
$existingSg = aws ec2 describe-security-groups `
  --filters "Name=group-name,Values=hrms-db-creator-sg" "Name=vpc-id,Values=$VpcId" `
  --region $Region `
  --query 'SecurityGroups[0].GroupId' `
  --output text 2>$null

if ($existingSg -and $existingSg -ne "None") {
    Write-Host "Using existing security group: $existingSg" -ForegroundColor Yellow
    $sgId = $existingSg
} else {
    $sgId = aws ec2 create-security-group `
      --group-name hrms-db-creator-sg `
      --description "Temporary SG for database creation" `
      --vpc-id $VpcId `
      --region $Region `
      --query 'GroupId' `
      --output text
    
    Write-Host "OK: Security group created: $sgId" -ForegroundColor Green
}

# Allow RDS to accept from bastion
$null = aws ec2 authorize-security-group-ingress `
  --group-id $RDSSecurityGroup `
  --protocol tcp `
  --port 5432 `
  --source-group $sgId `
  --region $Region 2>&1

Write-Host "OK: Security group rules configured" -ForegroundColor Green
Write-Host ""

# Step 2: Get latest Amazon Linux 2023 AMI
Write-Host "Step 2: Getting latest AMI..." -ForegroundColor Cyan

$amiId = aws ssm get-parameter `
  --name /aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64 `
  --region $Region `
  --query 'Parameter.Value' `
  --output text

Write-Host "OK: AMI ID: $amiId" -ForegroundColor Green
Write-Host ""

# Step 3: Create IAM role for SSM
Write-Host "Step 3: Creating IAM role..." -ForegroundColor Cyan

$roleName = "hrms-db-creator-role-$(Get-Random -Maximum 99999)"

# Create trust policy file
$trustPolicy = @"
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
"@

$policyFile = "$env:TEMP\trust-policy-$roleName.json"
$trustPolicy | Out-File -FilePath $policyFile -Encoding ASCII

# Create role
$roleOutput = aws iam create-role `
  --role-name $roleName `
  --assume-role-policy-document "file://$policyFile" `
  2>&1

Remove-Item $policyFile -Force

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to create IAM role" -ForegroundColor Red
    Write-Host $roleOutput -ForegroundColor Red
    
    # Cleanup security group
    $null = aws ec2 revoke-security-group-ingress --group-id $RDSSecurityGroup --protocol tcp --port 5432 --source-group $sgId --region $Region 2>&1
    $null = aws ec2 delete-security-group --group-id $sgId --region $Region 2>&1
    exit 1
}

# Attach policy
$null = aws iam attach-role-policy `
  --role-name $roleName `
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore `
  2>&1

# Create instance profile
$null = aws iam create-instance-profile `
  --instance-profile-name $roleName `
  2>&1

$null = aws iam add-role-to-instance-profile `
  --instance-profile-name $roleName `
  --role-name $roleName `
  2>&1

Write-Host "OK: IAM role created" -ForegroundColor Green
Write-Host "Waiting for IAM propagation..." -ForegroundColor Yellow
Start-Sleep -Seconds 10
Write-Host ""

# Step 4: Launch instance
Write-Host "Step 4: Launching bastion instance..." -ForegroundColor Cyan

$instanceId = aws ec2 run-instances `
  --image-id $amiId `
  --instance-type t3.micro `
  --subnet-id $SubnetId `
  --security-group-ids $sgId `
  --iam-instance-profile "Name=$roleName" `
  --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=hrms-db-creator}]" `
  --region $Region `
  --query 'Instances[0].InstanceId' `
  --output text

Write-Host "OK: Instance launched: $instanceId" -ForegroundColor Green
Write-Host ""

# Step 5: Wait for SSM to be ready
Write-Host "Step 5: Waiting for instance to be ready for SSM..." -ForegroundColor Yellow
Start-Sleep -Seconds 60

# Step 6: Run commands via SSM
Write-Host "Step 6: Creating database..." -ForegroundColor Cyan

$commands = @"
#!/bin/bash
set -e
dnf install -y postgresql16
PGPASSWORD='$DBPassword' psql -h $DBHost -U postgres -d postgres -c "CREATE DATABASE hrmsdb;" 2>&1 || echo "Database might exist"
PGPASSWORD='$DBPassword' psql -h $DBHost -U postgres -d hrmsdb -c "CREATE EXTENSION IF NOT EXISTS \\"uuid-ossp\\";"
PGPASSWORD='$DBPassword' psql -h $DBHost -U postgres -d hrmsdb -c "CREATE EXTENSION IF NOT EXISTS \\"pgcrypto\\";"
echo "SUCCESS: Database created"
"@

$commandFile = [System.IO.Path]::GetTempFileName()
$commands | Out-File -FilePath $commandFile -Encoding utf8

$commandId = aws ssm send-command `
  --instance-ids $instanceId `
  --document-name "AWS-RunShellScript" `
  --parameters "commands=$(Get-Content $commandFile -Raw)" `
  --region $Region `
  --query 'Command.CommandId' `
  --output text

Remove-Item $commandFile

Write-Host "Command sent: $commandId" -ForegroundColor White
Write-Host "Waiting for completion..." -ForegroundColor Yellow

# Wait for command to complete
Start-Sleep -Seconds 30

$output = aws ssm get-command-invocation `
  --command-id $commandId `
  --instance-id $instanceId `
  --region $Region `
  --query '[Status,StandardOutputContent,StandardErrorContent]' `
  --output text

Write-Host $output -ForegroundColor White

if ($output -match "SUCCESS") {
    Write-Host ""
    Write-Host "OK: Database created successfully!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "WARNING: Check output above for any errors" -ForegroundColor Yellow
}

Write-Host ""

# Step 7: Cleanup
Write-Host "Step 7: Cleaning up..." -ForegroundColor Cyan

# Terminate instance
aws ec2 terminate-instances --instance-ids $instanceId --region $Region | Out-Null
Write-Host "OK: Instance terminated" -ForegroundColor Green

# Wait a bit for cleanup
Start-Sleep -Seconds 10

# Remove security group rule from RDS
aws ec2 revoke-security-group-ingress `
  --group-id $RDSSecurityGroup `
  --protocol tcp `
  --port 5432 `
  --source-group $sgId `
  --region $Region 2>&1 | Out-Null

# Delete security group
aws ec2 delete-security-group --group-id $sgId --region $Region 2>&1 | Out-Null
Write-Host "OK: Security group deleted" -ForegroundColor Green

# Delete IAM resources
aws iam remove-role-from-instance-profile --instance-profile-name $roleName --role-name $roleName --region $Region 2>&1 | Out-Null
aws iam delete-instance-profile --instance-profile-name $roleName --region $Region 2>&1 | Out-Null
aws iam detach-role-policy --role-name $roleName --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore --region $Region 2>&1 | Out-Null
aws iam delete-role --role-name $roleName --region $Region 2>&1 | Out-Null
Write-Host "OK: IAM resources deleted" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "COMPLETE!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Database 'hrmsdb' is ready!" -ForegroundColor White
Write-Host ""
Write-Host "Next step - Deploy your application:" -ForegroundColor Yellow
Write-Host "  .\deploy-aws.ps1 -Environment dev -Region us-east-1 -DBHost $DBHost -DBPassword `"$DBPassword`" -DBName hrmsdb" -ForegroundColor White
Write-Host ""
